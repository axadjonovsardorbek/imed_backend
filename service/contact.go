package service

import (
	"context"
	m "imed/pkg/models"
	st "imed/storage/postgres"
)

type ContactService struct {
	storage st.Storage
}

func NewContactService(storage *st.Storage) *ContactService {
	return &ContactService{storage: *storage}
}

func (s *ContactService) Create(ctx context.Context, req *m.ContactCreateReq) (*m.ById, error) {
	return s.storage.ContactS.Create(req)
}

func (s *ContactService) GetById(ctx context.Context, req *m.ById) (*m.ContactRes, error) {
	return s.storage.ContactS.GetById(req)
}

func (s *ContactService) GetAll(ctx context.Context, req *m.ContactGetListReq) (*m.PaginationResponse, *m.ContactList, error) {
	return s.storage.ContactS.GetAll(req)
}

func (s *ContactService) Update(ctx context.Context, req *m.ContactUpdateReq) (*m.Void, error) {
	return s.storage.ContactS.Update(req)
}

func (s *ContactService) Delete(ctx context.Context, req *m.ById) (*m.Void, error) {
	return s.storage.ContactS.Delete(req)
}
