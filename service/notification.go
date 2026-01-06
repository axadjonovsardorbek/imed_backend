package service

import (
	"context"
	m "imed/pkg/models"
	st "imed/storage/postgres"
)

type NotificationService struct {
	storage st.Storage
}

func NewNotificationService(storage *st.Storage) *NotificationService {
	return &NotificationService{storage: *storage}
}

func (s *NotificationService) Create(ctx context.Context, req *m.NotificationCreateReq) (*m.ById, error) {
	return s.storage.NotificationS.Create(req)
}

func (s *NotificationService) GetById(ctx context.Context, req *m.ById) (*m.MasterNotificationRes, error) {
	return s.storage.NotificationS.GetById(req)
}

func (s *NotificationService) GetAll(ctx context.Context, req *m.NotificationGetListReq) (*m.PaginationResponse, *m.MasterNotificationList, error) {
	return s.storage.NotificationS.GetAll(req)
}

func (s *NotificationService) Update(ctx context.Context, req *m.MasterNotificationUpdateReq) (*m.Void, error) {
	return s.storage.NotificationS.Update(req)
}

func (s *NotificationService) Delete(ctx context.Context, req *m.ById) (*m.Void, error) {
	return s.storage.NotificationS.Delete(req)
}
