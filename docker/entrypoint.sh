#!/bin/sh
# Entrypoint: keeps yt-dlp and its PO-token plugin current, then runs the app.
#
# Why this exists: YouTube changes its extraction/anti-bot logic every few weeks.
# The single most common cause of "downloads suddenly stopped working" is a stale
# yt-dlp. Refreshing it on boot and daily means most breakages heal without a
# redeploy and without anyone touching the server.
#
# Env vars:
#   YT_DLP_AUTO_UPDATE       "true" (default) | "false" — disable updates entirely
#   YT_DLP_UPDATE_INTERVAL   seconds between background update checks (default 86400)

set -e

VENV_PIP=/opt/ytdlp/bin/pip
PACKAGES="yt-dlp bgutil-ytdlp-pot-provider"

log() {
	echo "[entrypoint] $*"
}

update_ytdlp() {
	if ! "$VENV_PIP" install --no-cache-dir --disable-pip-version-check -q -U $PACKAGES; then
		log "yt-dlp update failed; continuing with the installed version"
		return 1
	fi
	log "yt-dlp updated to $(yt-dlp --version 2>/dev/null || echo unknown)"
	return 0
}

if [ "${YT_DLP_AUTO_UPDATE:-true}" = "true" ]; then
	# Run in the background so a slow or unreachable PyPI never blocks app startup.
	(
		update_ytdlp || true

		interval="${YT_DLP_UPDATE_INTERVAL:-86400}"
		while [ "$interval" -gt 0 ] 2>/dev/null; do
			sleep "$interval"
			update_ytdlp || true
		done
	) &
else
	log "auto-update disabled (YT_DLP_AUTO_UPDATE=${YT_DLP_AUTO_UPDATE})"
fi

# Cookie rotation is persisted here; yt-dlp rewrites the cookie file on every run.
mkdir -p "${YT_DLP_COOKIES_STATE_DIR:-/data/ytdlp}" 2>/dev/null || true

log "starting: $*"
exec "$@"
