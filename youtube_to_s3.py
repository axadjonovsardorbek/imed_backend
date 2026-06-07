#!/usr/bin/env python3
import os, re, json, tempfile, subprocess, logging
from datetime import datetime
from concurrent.futures import ThreadPoolExecutor, as_completed

import psycopg2, psycopg2.extras
import boto3
from boto3.s3.transfer import TransferConfig
from botocore.client import Config

DB_CONFIG = {
    "host": "localhost",
    "port": 15432,
    "user": "postgres",
    "password": "1111",
    "dbname": "imed",
}

S3_ACCESS_KEY = "c90d2e513a2b9c00c64b0ebb75f66b72"
S3_SECRET_KEY = "da365d7dfcdba126b8ecd95058a4f2cb"
S3_TENANT     = "89dfc571989b45c8a67be7db3b6d8bf7"
S3_BUCKET     = "imed"
S3_ENDPOINT   = "https://eu2.contabostorage.com"
S3_DOMAIN     = "https://cdn.imedteam.uz"
S3_PREFIX     = "youtube"
WORKERS       = 2
COOKIES       = os.path.expanduser("~/Downloads/cookies.txt")
NODE_PATH     = "/opt/homebrew/bin/node"
REPORT_FILE   = "migration_report.json"

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
    handlers=[
        logging.FileHandler("migration.log"),
        logging.StreamHandler(),
    ],
)
log = logging.getLogger(__name__)


def get_youtube_id(url):
    for p in [
        r"youtu\.be/([A-Za-z0-9_-]{11})",
        r"youtube\.com/watch\?.*v=([A-Za-z0-9_-]{11})",
        r"youtube\.com/embed/([A-Za-z0-9_-]{11})",
    ]:
        m = re.search(p, url)
        if m:
            return m.group(1)
    return None


def make_s3():
    return boto3.client(
        "s3",
        endpoint_url=S3_ENDPOINT,
        aws_access_key_id=S3_ACCESS_KEY,
        aws_secret_access_key=S3_SECRET_KEY,
        config=Config(
            signature_version="s3v4",
            s3={"addressing_style": "path"}
        ),
    )


def upload_to_s3(s3, local_path, video_id):
    key = f"{S3_PREFIX}/{video_id}.mp4"
    # multipart o'chirish — UnseekableStreamError ni hal qiladi
    transfer_config = TransferConfig(
        multipart_threshold=10 * 1024 * 1024 * 1024,  # 10GB — amalda multipart ishlatmaydi
        use_threads=False,
    )
    s3.upload_file(
        local_path,
        S3_BUCKET,
        key,
        ExtraArgs={"ContentType": "video/mp4", "ACL": "public-read"},
        Config=transfer_config,
    )
    return f"{S3_DOMAIN}/{S3_TENANT}:{S3_BUCKET}/{key}"


def download_youtube(url, out_path):
    cmd = [
        "yt-dlp",
        "--no-playlist",
        "--cookies", COOKIES,
        "--js-runtimes", f"node:{NODE_PATH}",
        "--remote-components", "ejs:github",
        "-f", "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best",
        "--merge-output-format", "mp4",
        "-o", out_path,
        "--quiet",
        "--no-warnings",
        url,
    ]
    r = subprocess.run(cmd, capture_output=True, text=True, timeout=7200)
    if r.returncode != 0:
        log.error(f"  yt-dlp: {r.stderr.strip()}")
        return False
    return True


def fetch_sources(conn):
    with conn.cursor(cursor_factory=psycopg2.extras.DictCursor) as cur:
        cur.execute(
            "SELECT id, url FROM sources WHERE type = 'video' ORDER BY id"
        )
        return cur.fetchall()


def update_source(conn, source_id, cdn_url):
    v = json.dumps({"en": cdn_url, "ru": cdn_url, "uz": cdn_url})
    with conn.cursor() as cur:
        cur.execute(
            "UPDATE sources SET video_url = %s WHERE id = %s",
            (v, source_id)
        )
    conn.commit()


def already_done(conn, source_id):
    with conn.cursor() as cur:
        cur.execute(
            "SELECT video_url FROM sources WHERE id = %s", (source_id,)
        )
        row = cur.fetchone()
        return row and row[0] is not None


def process(row):
    source_id = str(row["id"])
    url_field = row["url"]

    if isinstance(url_field, str):
        try:
            url_data = json.loads(url_field)
        except Exception:
            url_data = {"en": url_field}
    else:
        url_data = url_field

    yt_url = url_data.get("en") or next(iter(url_data.values()))
    video_id = get_youtube_id(yt_url)

    if not video_id:
        log.warning(f"[{source_id}] YouTube ID topilmadi: {yt_url}")
        return {"id": source_id, "url": yt_url, "status": "error", "reason": "no_youtube_id"}

    conn = psycopg2.connect(**DB_CONFIG)
    s3 = make_s3()

    try:
        if already_done(conn, source_id):
            log.info(f"[{source_id}] Allaqachon bajarilgan, o'tkazib yuborildi")
            conn.close()
            return {"id": source_id, "video_id": video_id, "status": "skipped"}

        with tempfile.TemporaryDirectory() as tmp:
            out = os.path.join(tmp, f"{video_id}.mp4")
            log.info(f"[{source_id}] Yuklab olinmoqda: {video_id}")

            if not download_youtube(yt_url, out):
                conn.close()
                return {
                    "id": source_id, "video_id": video_id,
                    "url": yt_url, "status": "error", "reason": "download_failed"
                }

            if not os.path.exists(out):
                conn.close()
                return {
                    "id": source_id, "video_id": video_id,
                    "url": yt_url, "status": "error", "reason": "file_not_found"
                }

            size_mb = round(os.path.getsize(out) / 1024 / 1024, 1)
            log.info(f"[{source_id}] {size_mb} MB, S3 ga yuklanmoqda...")

            cdn_url = upload_to_s3(s3, out, video_id)
            log.info(f"[{source_id}] S3 ga yuklandi -> {cdn_url}")

        update_source(conn, source_id, cdn_url)
        log.info(f"[{source_id}] DB yangilandi ✓")
        conn.close()

        return {
            "id": source_id,
            "video_id": video_id,
            "youtube_url": yt_url,
            "cdn_url": cdn_url,
            "size_mb": size_mb,
            "status": "done",
        }

    except Exception as e:
        log.error(f"[{source_id}] Xato: {e}", exc_info=True)
        try:
            conn.close()
        except Exception:
            pass
        return {"id": source_id, "url": yt_url, "status": "error", "reason": str(e)}


def main():
    start = datetime.now()
    log.info("=== Migration boshlandi ===")

    if not os.path.exists(COOKIES):
        log.error(f"Cookies fayl topilmadi: {COOKIES}")
        return
    log.info(f"Cookies: OK")

    if not os.path.exists(NODE_PATH):
        log.error(f"Node topilmadi: {NODE_PATH}")
        return
    log.info(f"Node: OK")

    try:
        s3 = make_s3()
        s3.head_bucket(Bucket=S3_BUCKET)
        log.info(f"S3: OK")
    except Exception as e:
        log.error(f"S3 xatosi: {e}")
        return

    try:
        conn = psycopg2.connect(**DB_CONFIG)
        sources = fetch_sources(conn)
        conn.close()
        log.info(f"DB: OK | Jami: {len(sources)} ta video")
    except Exception as e:
        log.error(f"DB xatosi: {e}")
        return

    done, skipped, failed = [], [], []

    with ThreadPoolExecutor(max_workers=WORKERS) as executor:
        futures = {executor.submit(process, row): row for row in sources}
        for i, future in enumerate(as_completed(futures), 1):
            res = future.result()
            status = res["status"]
            log.info(f"Progress: {i}/{len(sources)} | {status.upper()} | id={res['id']}")
            if status == "done":
                done.append(res)
            elif status == "skipped":
                skipped.append(res)
            else:
                failed.append(res)

    elapsed = round((datetime.now() - start).total_seconds())

    report = {
        "generated_at": datetime.now().isoformat(),
        "elapsed_seconds": elapsed,
        "total": len(sources),
        "done": len(done),
        "skipped": len(skipped),
        "failed": len(failed),
        "done_list": done,
        "skipped_list": skipped,
        "failed_list": failed,
    }

    with open(REPORT_FILE, "w", encoding="utf-8") as f:
        json.dump(report, f, ensure_ascii=False, indent=2)

    log.info("=" * 50)
    log.info(f"✓ DONE:    {len(done)}")
    log.info(f"  SKIPPED: {len(skipped)}")
    log.info(f"  FAILED:  {len(failed)}")
    log.info(f"  Vaqt:    {elapsed}s")
    log.info(f"  Hisobot: {REPORT_FILE}")

    if failed:
        log.warning("Muvaffaqiyatsizlar:")
        for f in failed:
            log.warning(f"  id={f['id']} | {f.get('reason')} | {f.get('url', '')}")


if __name__ == "__main__":
    main()
