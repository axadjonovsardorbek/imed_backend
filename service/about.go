package service

import (
	"context"
	m "imed/pkg/models"
	st "imed/storage/postgres"
)

type AboutService struct {
	storage st.Storage
}

func NewAboutService(storage *st.Storage) *AboutService {
	return &AboutService{storage: *storage}
}

func (s *AboutService) Create(ctx context.Context, req *m.AboutCreateReq) (*m.ById, error) {
	return s.storage.AboutS.Create(req)
}
func (s *AboutService) GetById(ctx context.Context, req *m.ById) (*m.AboutRes, error) {
	return s.storage.AboutS.GetById(req)
}
func (s *AboutService) GetAll(ctx context.Context, req *m.AboutGetListReq) (*m.PaginationResponse, *m.AboutList, error) {
	return s.storage.AboutS.GetAll(req)
}
func (s *AboutService) Update(ctx context.Context, req *m.AboutUpdateReq) (*m.Void, error) {
	return s.storage.AboutS.Update(req)
}
func (s *AboutService) Delete(ctx context.Context, req *m.ById) (*m.Void, error) {
	return s.storage.AboutS.Delete(req)
}
