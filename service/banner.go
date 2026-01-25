package service

import (
	"context"
	m "imed/pkg/models"
	st "imed/storage/postgres"
)

type BannerService struct {
	storage st.Storage
}

func NewBannerService(storage *st.Storage) *BannerService {
	return &BannerService{storage: *storage}
}

func (s *BannerService) Create(ctx context.Context, req *m.BannerCreateReq) (*m.ById, error) {
	return s.storage.BannerS.Create(req)
}

func (s *BannerService) GetById(ctx context.Context, req *m.ById) (*m.BannerRes, error) {
	return s.storage.BannerS.GetById(req)
}

func (s *BannerService) GetAll(ctx context.Context, req *m.BannerGetListReq) (*m.PaginationResponse, *m.BannerList, error) {
	return s.storage.BannerS.GetAll(req)
}

func (s *BannerService) Update(ctx context.Context, req *m.BannerUpdateReq) (*m.Void, error) {
	return s.storage.BannerS.Update(req)
}

func (s *BannerService) Delete(ctx context.Context, req *m.ById) (*m.Void, error) {
	return s.storage.BannerS.Delete(req)
}

func (s *BannerService) GetByIdMobile(ctx context.Context, req *m.ByIdMobile) (*m.BannerMobileRes, error) {
	return s.storage.BannerS.GetByIdMobile(req)
}

func (s *BannerService) GetAllMobile(ctx context.Context, req *m.BannerGetListMobileReq) (*m.PaginationResponse, *m.BannerMobileList, error) {
	return s.storage.BannerS.GetAllMobile(req)
}
