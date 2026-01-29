package handler

import (
	"imed/internal/config"
	"imed/pkg/clients"
	"imed/pkg/minio"
)

type Handler struct {
	Clients *clients.Clients
	MinIO   *minio.MinIO
	Cfg     *config.Config
}

func NewHandler(client clients.Clients, minio *minio.MinIO, cfg *config.Config) *Handler {
	return &Handler{
		Clients: &client,
		MinIO:   minio,
		Cfg:     cfg,
	}
}
