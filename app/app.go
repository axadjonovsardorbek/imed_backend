package app

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"runtime/debug"
	"time"

	"imed/api"
	"imed/api/handler"
	"imed/internal/bot"
	"imed/internal/config"
	"imed/internal/push"
	"imed/pkg/clients"
	"imed/pkg/minio"
	"imed/pkg/s3"
	"imed/storage/postgres"

	"github.com/redis/go-redis/v9"
)

func Run(cf *config.Config) {
	ctx := context.Background()

	rAddr := fmt.Sprintf("%s%s", cf.REDIS_HOST, cf.REDIS_PORT)

	rdb := redis.NewClient(&redis.Options{
		Addr: rAddr,
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		slog.Error("Failed to connect redis:", "err", err)
		return
	}
	slog.Info("Redis connected successfully")

	conn, err := postgres.NewPostgresStorage(*cf, rdb)

	if err != nil {
		slog.Error("Failed to connect postgres:", "err", err)
		return
	}
	defer conn.Db.Close()
	slog.Info("Postgres storage connected successfully")

	pushClient, err := push.NewClientFromServiceAccountFile(
		ctx,
		cf.FIREBASE_CREDENTIALS,
	)
	if err != nil {
		log.Fatal("firebase init failed:", err)
	}

	//MinIO
	minioClient, err := minio.MinIOConnect(cf)
	if err != nil {
		slog.Error("Failed to connect to MinIO", "err", err)
		return
	}

	s3Client, err := s3.NewS3(cf)
	if err != nil {
		slog.Error("Failed to connect to S3", "err", err)
		return
	}

	// services
	services, err := clients.NewClients(cf, conn, rdb, pushClient, s3Client)
	if err != nil {
		log.Fatalf("error while connecting clients. err: %s", err.Error())
	}

	handlers := handler.NewHandler(*services, minioClient, cf, s3Client)

	// Bot alohida goroutine'da ishlaydi: undagi panic butun serverni o'ldirmasligi
	// uchun recover qilinadi (bot to'xtaydi, API ishlashda davom etadi).
	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("bot goroutine panicked", "panic", rec, "stack", string(debug.Stack()))
			}
		}()
		bot.BotAuth(cf)
	}()

	// 24 soat kirmaganlarga eslatma (T-05)
	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("inactive reminder goroutine panicked", "panic", rec, "stack", string(debug.Stack()))
			}
		}()
		services.Push.RunInactiveReminder(ctx, 15*time.Minute)
	}()

	router := api.NewApi(handlers, rdb, cf)
	log.Println("Server is running on port ", cf)
	if err := router.Run(cf.AUTH_PORT); err != nil {
		slog.Error("Error:", "err", err)
	}
}
