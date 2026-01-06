package service

import (
	"context"

	m "imed/pkg/models"
	st "imed/storage/postgres"
)

type AdminService struct {
	storage st.Storage
}

func NewAdminService(storage *st.Storage) *AdminService {
	return &AdminService{storage: *storage}
}

func (s *AdminService) Login(ctx context.Context, req *m.LoginReq) (*m.TokenRes, error) {
	return s.storage.AdminS.Login(req)
}
func (s *AdminService) GetById(ctx context.Context, req *m.ById) (*m.AdminRes, error) {
	return s.storage.AdminS.GetById(req)
}
