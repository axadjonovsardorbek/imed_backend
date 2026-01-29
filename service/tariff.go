package service

import (
	"context"
	m "imed/pkg/models"
	st "imed/storage/postgres"
)

type TariffService struct {
	storage st.Storage
}

func NewTariffService(storage *st.Storage) *TariffService {
	return &TariffService{storage: *storage}
}

func (s *TariffService) Create(ctx context.Context, req *m.TariffCreateReq) (*m.ById, error) {
	return s.storage.TariffS.Create(req)
}

func (s *TariffService) GetById(ctx context.Context, req *m.ById) (*m.TariffRes, error) {
	return s.storage.TariffS.GetById(req)
}

func (s *TariffService) GetAll(ctx context.Context, req *m.TariffGetListReq) (*m.PaginationResponse, *m.TariffList, error) {
	return s.storage.TariffS.GetAll(req)
}

func (s *TariffService) Update(ctx context.Context, req *m.TariffUpdateReq) (*m.Void, error) {
	return s.storage.TariffS.Update(req)
}

func (s *TariffService) Delete(ctx context.Context, req *m.ById) (*m.Void, error) {
	return s.storage.TariffS.Delete(req)
}
