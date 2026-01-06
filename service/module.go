package service

import (
	"context"
	m "imed/pkg/models"
	st "imed/storage/postgres"
)

type ModuleService struct {
	storage st.Storage
}

func NewModuleService(storage *st.Storage) *ModuleService {
	return &ModuleService{storage: *storage}
}

func (s *ModuleService) Create(ctx context.Context, req *m.ModuleCreateReq) (*m.ById, error) {
	return s.storage.ModuleS.Create(req)
}

func (s *ModuleService) GetById(ctx context.Context, req *m.ById) (*m.ModuleRes, error) {
	return s.storage.ModuleS.GetById(req)
}

func (s *ModuleService) GetAll(ctx context.Context, req *m.ModuleGetListReq) (*m.PaginationResponse, *m.ModuleList, error) {
	return s.storage.ModuleS.GetAll(req)
}

func (s *ModuleService) Update(ctx context.Context, req *m.ModuleUpdateReq) (*m.Void, error) {
	return s.storage.ModuleS.Update(req)
}

func (s *ModuleService) Delete(ctx context.Context, req *m.ById) (*m.Void, error) {
	return s.storage.ModuleS.Delete(req)
}
