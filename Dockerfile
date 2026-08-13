FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0: statik binary, runtime bazasidan (musl/glibc) mustaqil.
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

# Runtime bazasi debian: deno faqat glibc uchun build chiqaradi
# (deno-*-unknown-linux-gnu), musl/alpine'da dynamic linker topilmay ishga tushmaydi.
# yt-dlp esa `n challenge` yechish uchun JS runtime'siz format tanlay olmaydi.
FROM debian:bookworm-slim

WORKDIR /app

ENV DEBIAN_FRONTEND=noninteractive

COPY --from=builder /app/main .

COPY --from=builder /app/api/casbin/model.conf ./api/casbin/
COPY --from=builder /app/api/casbin/policy.csv ./api/casbin/

# YouTube uchun: yt-dlp (Python), EJS challenge solver uchun deno, merge uchun ffmpeg.
RUN apt-get update \
	&& apt-get install -y --no-install-recommends \
	ffmpeg ca-certificates curl unzip wget bash nodejs python3 python3-venv \
	&& rm -rf /var/lib/apt/lists/*

# deno — yt-dlp ning standart JS runtime'i (sandboxli, shuning uchun default yoqilgan).
# `deno --version` build vaqtida tekshiriladi: jim o'rnatilib, runtime'da ishlamay
# qolishining oldi olinadi (avvalgi alpine build'da aynan shu bo'lgan).
RUN ARCH="$(uname -m)" \
	&& case "$ARCH" in \
	x86_64) DENO_ARCH=x86_64-unknown-linux-gnu ;; \
	aarch64 | arm64) DENO_ARCH=aarch64-unknown-linux-gnu ;; \
	*) echo "unsupported arch=$ARCH" >&2; exit 1 ;; \
	esac \
	&& curl -fsSL "https://github.com/denoland/deno/releases/latest/download/deno-${DENO_ARCH}.zip" -o /tmp/deno.zip \
	&& unzip -q -o /tmp/deno.zip deno -d /usr/local/bin \
	&& chmod +x /usr/local/bin/deno \
	&& rm /tmp/deno.zip \
	&& deno --version

# yt-dlp standalone binary o'rniga venv: PO token plugin faqat shu holda ko'rinadi
# (standalone build o'z Python'ini olib yuradi, pip paketlarini ko'rmaydi).
# Venv yana `pip install -U` orqali runtime'da yangilanish imkonini beradi.
RUN python3 -m venv /opt/ytdlp \
	&& /opt/ytdlp/bin/pip install --no-cache-dir --disable-pip-version-check -U \
	pip yt-dlp bgutil-ytdlp-pot-provider \
	&& ln -sf /opt/ytdlp/bin/yt-dlp /usr/local/bin/yt-dlp \
	&& yt-dlp --version

COPY .env .
COPY internal/secrets ./internal/secrets/
COPY docker/entrypoint.sh /usr/local/bin/entrypoint.sh

# Cookies rotatsiyasi shu katalogda saqlanadi; compose'da volume mount qilinadi.
RUN mkdir -p /data/ytdlp \
	&& chmod +x ./main /usr/local/bin/entrypoint.sh

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["./main"]
