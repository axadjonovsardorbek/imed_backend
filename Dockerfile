FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main ./cmd/main.go

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/main .

COPY --from=builder /app/api/casbin/model.conf ./api/casbin/
COPY --from=builder /app/api/casbin/policy.csv ./api/casbin/

# YouTube uchun: yt-dlp asosan Python uchraydi (#!/usr/bin/env python3), EJS uchun deno/nodejs.
RUN apk add --no-cache ffmpeg ca-certificates curl unzip wget bash nodejs python3 \
	&& ARCH="$(uname -m)" \
	&& case "$ARCH" in \
	x86_64) DENO_ARCH=x86_64-unknown-linux-gnu ;; \
	aarch64) DENO_ARCH=aarch64-unknown-linux-gnu ;; \
	arm64) DENO_ARCH=aarch64-unknown-linux-gnu ;; \
	*) echo "unsupported arch=$ARCH"; DENO_ARCH=x86_64-unknown-linux-gnu ;; \
	esac \
	&& curl -fsSL "https://github.com/denoland/deno/releases/latest/download/deno-${DENO_ARCH}.zip" -o /tmp/deno.zip \
	&& unzip -q -o /tmp/deno.zip deno -d /usr/local/bin \
	&& chmod +x /usr/local/bin/deno \
	&& rm /tmp/deno.zip \
	&& wget -qO /usr/local/bin/yt-dlp https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp \
	&& chmod a+rx /usr/local/bin/yt-dlp

COPY .env .
COPY internal/secrets ./internal/secrets/

RUN chmod +x ./main

CMD ["./main"]
