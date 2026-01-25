package service

import (
	"context"
	m "imed/pkg/models"
	st "imed/storage/postgres"
)

type FaqService struct {
	storage st.Storage
}

func NewFaqService(storage *st.Storage) *FaqService {
	return &FaqService{storage: *storage}
}

func (s *FaqService) Create(ctx context.Context, req *m.FaqCreateReq) (*m.ById, error) {
	return s.storage.FaqS.Create(req)
}
func (s *FaqService) GetById(ctx context.Context, req *m.ById) (*m.FaqRes, error) {
	return s.storage.FaqS.GetById(req)
}
func (s *FaqService) GetAll(ctx context.Context, req *m.FaqGetListReq) (*m.PaginationResponse, *m.FaqList, error) {
	return s.storage.FaqS.GetAll(req)
}
func (s *FaqService) Update(ctx context.Context, req *m.FaqUpdateReq) (*m.Void, error) {
	return s.storage.FaqS.Update(req)
}
func (s *FaqService) Delete(ctx context.Context, req *m.ById) (*m.Void, error) {
	return s.storage.FaqS.Delete(req)
}
func (s *FaqService) GetByIdMobile(ctx context.Context, req *m.ByIdMobile) (*m.FaqMobileRes, error) {
	return s.storage.FaqS.GetByIdMobile(req)
}
func (s *FaqService) GetAllMobile(ctx context.Context, req *m.FaqGetListMobileReq) (*m.PaginationResponse, *m.FaqMobileList, error) {
	return s.storage.FaqS.GetAllMobile(req)
}
