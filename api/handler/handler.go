package handler

import (
	"imed/internal/config"
	"imed/pkg/clients"
	"imed/pkg/minio"
	"imed/pkg/s3"
)

type Handler struct {
	Clients *clients.Clients
	MinIO   *minio.MinIO
	S3      *s3.S3Service
	Cfg     *config.Config
}

func NewHandler(client clients.Clients, minio *minio.MinIO, cfg *config.Config, s3 *s3.S3Service) *Handler {
	return &Handler{
		Clients: &client,
		MinIO:   minio,
		Cfg:     cfg,
		S3:      s3,
	}
}
