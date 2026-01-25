package service

import (
	"context"
	m "imed/pkg/models"
	st "imed/storage/postgres"
)

type SubjectService struct {
	storage st.Storage
}

func NewSubjectService(storage *st.Storage) *SubjectService {
	return &SubjectService{storage: *storage}
}

func (s *SubjectService) Create(ctx context.Context, req *m.SubjectCreateReq) (*m.ById, error) {
	return s.storage.SubjectS.Create(req)
}

func (s *SubjectService) GetById(ctx context.Context, req *m.ById) (*m.SubjectRes, error) {
	return s.storage.SubjectS.GetById(req)
}

func (s *SubjectService) GetAll(ctx context.Context, req *m.SubjectGetListReq) (*m.PaginationResponse, *m.SubjectList, error) {
	return s.storage.SubjectS.GetAll(req)
}

func (s *SubjectService) Update(ctx context.Context, req *m.SubjectUpdateReq) (*m.Void, error) {
	return s.storage.SubjectS.Update(req)
}

func (s *SubjectService) Delete(ctx context.Context, req *m.ById) (*m.Void, error) {
	return s.storage.SubjectS.Delete(req)
}

func (s *SubjectService) GetByIdMobile(ctx context.Context, req *m.ByIdMobile) (*m.SubjectMobileRes, error) {
	return s.storage.SubjectS.GetByIdMobile(req)
}

func (s *SubjectService) GetAllMobile(ctx context.Context, req *m.SubjectGetListMobileReq) (*m.PaginationResponse, *m.SubjectMobileList, error) {
	return s.storage.SubjectS.GetAllMobile(req)
}
