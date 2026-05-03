FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main ./cmd/main.go

FROM alpine:3.16

WORKDIR /app

COPY --from=builder /app/main .

COPY --from=builder /app/api/casbin/model.conf ./api/casbin/
COPY --from=builder /app/api/casbin/policy.csv ./api/casbin/


RUN apk add --no-cache python3 ffmpeg wget nodejs
RUN wget https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -O /usr/local/bin/yt-dlp && chmod a+rx /usr/local/bin/yt-dlp

COPY .env .
COPY internal/secrets ./internal/secrets/
COPY cookies.txt ./cookies.txt

RUN chmod +x ./main

CMD ["./main"]
