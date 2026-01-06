package service

import (
	"context"
	m "imed/pkg/models"
	st "imed/storage/postgres"
)

type SourceService struct {
	storage st.Storage
}

func NewSourceService(storage *st.Storage) *SourceService {
	return &SourceService{storage: *storage}
}

func (s *SourceService) Create(ctx context.Context, req *m.SourceCreateReq) (*m.ById, error) {
	return s.storage.SourceS.Create(req)
}

func (s *SourceService) GetById(ctx context.Context, req *m.ById) (*m.SourceRes, error) {
	return s.storage.SourceS.GetById(req)
}

func (s *SourceService) GetAll(ctx context.Context, req *m.SourceGetListReq) (*m.PaginationResponse, *m.SourceList, error) {
	return s.storage.SourceS.GetAll(req)
}

func (s *SourceService) Update(ctx context.Context, req *m.SourceUpdateReq) (*m.Void, error) {
	return s.storage.SourceS.Update(req)
}

func (s *SourceService) Delete(ctx context.Context, req *m.ById) (*m.Void, error) {
	return s.storage.SourceS.Delete(req)
}
