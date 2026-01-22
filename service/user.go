package service

import (
	"context"
	"encoding/json"
	"fmt"
	m "imed/pkg/models"
	st "imed/storage/postgres"
	"time"

	"github.com/go-redis/redis"
)

type UserService struct {
	storage st.Storage
	rdb     *redis.Client
}

func NewUserService(storage *st.Storage, rdb *redis.Client) *UserService {
	return &UserService{storage: *storage, rdb: rdb}
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
func (s *UserService) GetActivities(ctx context.Context, req *m.UserActivityGetListReq) (*m.ActivityStatsResponse, error) {
	return s.storage.UserS.GetActivities(req)
}
func (s *UserService) GetRating(ctx context.Context, req *m.RatingRequest) (*m.RatingResponse, error) {
	req.Limit = clampLimit(req.Limit)

	res := &m.RatingResponse{
		Type:  req.Type,
		Limit: req.Limit,
		Items: make([]m.RatingUserItem, 0, req.Limit),
	}

	key := topCacheKey(string(req.Type), req.Limit)

	if s.rdb != nil {
		if b, err := s.rdb.Get(key).Bytes(); err == nil && len(b) > 0 {
			_ = json.Unmarshal(b, &res.Items)
		}
	}

	if len(res.Items) == 0 {
		items, err := s.storage.UserS.GetTop(req)
		if err != nil {
			return nil, err
		}
		res.Items = items

		if s.rdb != nil {
			if b, err := json.Marshal(items); err == nil {
				_ = s.rdb.Set(key, b, 30*time.Second).Err()
			}
		}
	}

	if req.UserId != "" {
		me, err := s.storage.UserS.GetMe(req)
		if err != nil {
			return nil, err
		}
		res.Me = me
	}

	fmt.Println("RATING SERVICE RETURNING")
	return res, nil

	// return s.storage.UserS.GetRating(req)
}

func clampLimit(limit int) int {
	if limit <= 0 {
		return 10
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func topCacheKey(typ string, limit int) string {
	return fmt.Sprintf("rating:top:%s:%d", typ, limit)
}
