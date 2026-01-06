package service

import (
	"context"
	m "imed/pkg/models"
	st "imed/storage/postgres"
)

type LessonService struct {
	storage st.Storage
}

func NewLessonService(storage *st.Storage) *LessonService {
	return &LessonService{storage: *storage}
}

func (s *LessonService) Create(ctx context.Context, req *m.LessonCreateReq) (*m.ById, error) {
	return s.storage.LessonS.Create(req)
}

func (s *LessonService) GetById(ctx context.Context, req *m.ById) (*m.LessonRes, error) {
	return s.storage.LessonS.GetById(req)
}

func (s *LessonService) GetAll(ctx context.Context, req *m.LessonGetListReq) (*m.PaginationResponse, *m.LessonList, error) {
	return s.storage.LessonS.GetAll(req)
}

func (s *LessonService) Update(ctx context.Context, req *m.LessonUpdateReq) (*m.Void, error) {
	return s.storage.LessonS.Update(req)
}

func (s *LessonService) Delete(ctx context.Context, req *m.ById) (*m.Void, error) {
	return s.storage.LessonS.Delete(req)
}

func (s *LessonService) EndLesson(ctx context.Context, req *m.EndUserLessonReq) (*m.Void, error) {
	return s.storage.LessonS.EndLesson(req)
}
