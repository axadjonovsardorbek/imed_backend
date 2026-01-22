package app

import (
	"fmt"
	"log"
	"log/slog"

	"imed/api"
	"imed/api/handler"
	"imed/internal/config"
	"imed/pkg/clients"
	"imed/pkg/minio"
	"imed/storage/postgres"

	"github.com/go-redis/redis"
)

func Run(cfg *config.Config) {
	cf := config.Load()

	conn, err := postgres.NewPostgresStorage(cf)

	if err != nil {
		slog.Error("Failed to connect postgres:", "err", err)
	}
	defer conn.Db.Close()
	slog.Info("Postgres storage connected successfully")

	rAddr := fmt.Sprintf("%s%s", cf.REDIS_HOST, cf.REDIS_PORT)

	rdb := redis.NewClient(&redis.Options{
		Addr: rAddr,
	})

	_, err = rdb.Ping().Result()
	if err != nil {
		slog.Error("Failed to connect redis:", "err", err)
		return
	}
	slog.Info("Redis connected successfully")

	//MinIO
	minioClient, err := minio.MinIOConnect(cfg)
	if err != nil {
		slog.Error("Failed to connect to MinIO", "err", err)
		return
	}

	// services
	services, err := clients.NewClients(cfg, conn, rdb)
	if err != nil {
		log.Fatalf("error while connecting clients. err: %s", err.Error())
	}

	handlers := handler.NewHandler(*services, minioClient)

	router := api.NewApi(handlers)
	log.Println("Server is running on port ", cf)
	if err := router.Run(cf.AUTH_PORT); err != nil {
		slog.Error("Error:", "err", err)
	}
}
