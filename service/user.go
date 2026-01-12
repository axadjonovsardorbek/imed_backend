package service

import (
	"context"
	m "imed/pkg/models"
	st "imed/storage/postgres"
)

type UserService struct {
	storage st.Storage
}

func NewUserService(storage *st.Storage) *UserService {
	return &UserService{storage: *storage}
}

func (s *UserService) Check(ctx context.Context, req *m.UserCheckReq) (bool, error) {
	return s.storage.UserS.Check(req)
}
func (s *UserService) Login(ctx context.Context, req *m.UserLoginReq) (*m.TokenRes, error) {
	return s.storage.UserS.Login(req)
}
func (s *UserService) GetById(ctx context.Context, req *m.ById) (*m.UserRes, error) {
	return s.storage.UserS.GetById(req)
}
func (s *UserService) GetAll(ctx context.Context, req *m.UserGetListReq) (*m.PaginationResponse, *m.UserList, error) {
	return s.storage.UserS.GetAll(req)
}
func (s *UserService) Update(ctx context.Context, req *m.UserUpdateReq) (*m.Void, error) {
	return s.storage.UserS.Update(req)
}
func (s *UserService) Delete(ctx context.Context, req *m.ById) (*m.Void, error) {
	return s.storage.UserS.Delete(req)
}
func (s *UserService) ChangePassword(ctx context.Context, req *m.ChangePasswordReq) (*m.Void, error) {
	return s.storage.UserS.ChangePassword(req)
}
func (s *UserService) RefreshPassword(ctx context.Context, req *m.RefreshPasswordReq) (*m.RefreshPassword, error) {
	return s.storage.UserS.RefreshPassword(req)
}
func (s *UserService) CreateActivity(ctx context.Context, req *m.UserActivityCreateReq) (*m.Void, error) {
	return s.storage.UserS.CreateActivity(req)
}
func (s *UserService) GetActivities(ctx context.Context, req *m.UserActivityGetListReq) (*m.PaginationResponse, *m.UserActivityList, error) {
	return s.storage.UserS.GetActivities(req)
}
