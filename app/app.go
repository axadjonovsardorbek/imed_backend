package app

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"imed/api"
	"imed/api/handler"
	"imed/internal/bot"
	"imed/internal/config"
	"imed/internal/push"
	"imed/pkg/clients"
	"imed/pkg/minio"
	"imed/pkg/s3"
	"imed/storage/postgres"

	"github.com/go-redis/redis"
)

func Run(cf *config.Config) {
	ctx := context.Background()

	rAddr := fmt.Sprintf("%s%s", cf.REDIS_HOST, cf.REDIS_PORT)

	rdb := redis.NewClient(&redis.Options{
		Addr: rAddr,
	})

	_, err := rdb.Ping().Result()
	if err != nil {
		slog.Error("Failed to connect redis:", "err", err)
		return
	}
	slog.Info("Redis connected successfully")

	conn, err := postgres.NewPostgresStorage(*cf, rdb)

	if err != nil {
		slog.Error("Failed to connect postgres:", "err", err)
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

	go bot.BotAuth(cf)

	router := api.NewApi(handlers, rdb, cf)
	log.Println("Server is running on port ", cf)
	if err := router.Run(cf.AUTH_PORT); err != nil {
		slog.Error("Error:", "err", err)
	}
}
