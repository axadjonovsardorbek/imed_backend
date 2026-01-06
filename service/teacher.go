package service

import (
	"context"
	m "imed/pkg/models"
	st "imed/storage/postgres"
)

type TeacherService struct {
	storage st.Storage
}

func NewTeacherService(storage *st.Storage) *TeacherService {
	return &TeacherService{storage: *storage}
}

func (s *TeacherService) Create(ctx context.Context, req *m.TeacherCreateReq) (*m.ById, error) {
	return s.storage.TeacherS.Create(req)
}

func (s *TeacherService) GetById(ctx context.Context, req *m.ById) (*m.TeacherRes, error) {
	return s.storage.TeacherS.GetById(req)
}

func (s *TeacherService) GetAll(ctx context.Context, req *m.TeacherGetListReq) (*m.PaginationResponse, *m.TeacherList, error) {
	return s.storage.TeacherS.GetAll(req)
}

func (s *TeacherService) Update(ctx context.Context, req *m.TeacherUpdateReq) (*m.Void, error) {
	return s.storage.TeacherS.Update(req)
}

func (s *TeacherService) Delete(ctx context.Context, req *m.ById) (*m.Void, error) {
	return s.storage.TeacherS.Delete(req)
}
