package handler

import (
	"imed/pkg/clients"
	"imed/pkg/minio"
)

type Handler struct {
	Clients *clients.Clients
	MinIO   *minio.MinIO
}

func NewHandler(client clients.Clients, minio *minio.MinIO) *Handler {
	return &Handler{
		Clients: &client,
		MinIO:   minio,
	}
}
