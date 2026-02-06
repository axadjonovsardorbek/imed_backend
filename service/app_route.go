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

type AppRouteService struct {
	storage st.Storage
	rdb     *redis.Client
}

func NewAppRouteService(storage *st.Storage, rdb *redis.Client) *AppRouteService {
	return &AppRouteService{storage: *storage, rdb: rdb}
}

func (s *AppRouteService) Create(ctx context.Context, req *m.AppRouteCreateReq) (*m.ById, error) {
	return s.storage.AppRouteS.Create(req)
}
func (s *AppRouteService) GetById(ctx context.Context, req *m.ById) (*m.AppRouteRes, error) {
	return s.storage.AppRouteS.GetById(req)
}
func (s *AppRouteService) GetAll(ctx context.Context, req *m.AppRouteGetListReq) (*m.PaginationResponse, *m.AppRouteList, error) {

	res := &m.PaginationResponse{}

	key := fmt.Sprintf("app_route")

	if s.rdb != nil {
		if b, err := s.rdb.Get(key).Bytes(); err == nil && len(b) > 0 {
			_ = json.Unmarshal(b, &res)
		}
	}

	if res.Total == 0 {
		res, _, err := s.storage.AppRouteS.GetAll(req)
		if err != nil {
			return nil, nil, err
		}

		if s.rdb != nil {
			if b, err := json.Marshal(res); err == nil {
				_ = s.rdb.Set(key, b, 60*time.Second).Err()
			}
		}

		return res, nil, nil
	}

	return res, nil, nil
}
func (s *AppRouteService) Update(ctx context.Context, req *m.AppRouteUpdateReq) (*m.Void, error) {
	return s.storage.AppRouteS.Update(req)
}
func (s *AppRouteService) Delete(ctx context.Context, req *m.ById) (*m.Void, error) {
	return s.storage.AppRouteS.Delete(req)
}
