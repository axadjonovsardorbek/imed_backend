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

type ModuleService struct {
	storage st.Storage
	rdb     *redis.Client
}

func NewModuleService(storage *st.Storage, rdb *redis.Client) *ModuleService {
	return &ModuleService{storage: *storage, rdb: rdb}
}

func (s *ModuleService) Create(ctx context.Context, req *m.ModuleCreateReq) (*m.ById, error) {
	return s.storage.ModuleS.Create(req)
}

func (s *ModuleService) GetById(ctx context.Context, req *m.ById) (*m.ModuleRes, error) {
	return s.storage.ModuleS.GetById(req)
}

func (s *ModuleService) GetAll(ctx context.Context, req *m.ModuleGetListReq) (*m.PaginationResponse, *m.ModuleList, error) {
	res := &m.PaginationResponse{}

	key := fmt.Sprintf("course_id:%s_name:%s_page:%d_limit:%d",
		req.CourseId, req.Name, req.Filter.Page, req.Filter.Limit)

	if s.rdb != nil {
		if b, err := s.rdb.Get(key).Bytes(); err == nil && len(b) > 0 {
			_ = json.Unmarshal(b, &res)
		}
	}

	if res.Total == 0 {
		res, _, err := s.storage.ModuleS.GetAll(req)
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

func (s *ModuleService) Update(ctx context.Context, req *m.ModuleUpdateReq) (*m.Void, error) {
	return s.storage.ModuleS.Update(req)
}

func (s *ModuleService) Delete(ctx context.Context, req *m.ById) (*m.Void, error) {
	return s.storage.ModuleS.Delete(req)
}
