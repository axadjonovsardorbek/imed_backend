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

type TeacherService struct {
	storage st.Storage
	rdb     *redis.Client
}

func NewTeacherService(storage *st.Storage, rdb *redis.Client) *TeacherService {
	return &TeacherService{storage: *storage, rdb: rdb}
}

func (s *TeacherService) Create(ctx context.Context, req *m.TeacherCreateReq) (*m.ById, error) {
	return s.storage.TeacherS.Create(req)
}

func (s *TeacherService) GetById(ctx context.Context, req *m.ById) (*m.TeacherRes, error) {
	return s.storage.TeacherS.GetById(req)
}

func (s *TeacherService) GetAll(ctx context.Context, req *m.TeacherGetListReq) (*m.PaginationResponse, *m.TeacherList, error) {
	res := &m.PaginationResponse{}

	key := fmt.Sprintf("teacher_name:%s_phone:%s_page:%d_limit:%d",
		req.Name, req.PhoneNumber, req.Filter.Page, req.Filter.Limit)

	if s.rdb != nil {
		if b, err := s.rdb.Get(key).Bytes(); err == nil && len(b) > 0 {
			_ = json.Unmarshal(b, &res)
		}
	}

	if res.Total == 0 {
		res, _, err := s.storage.TeacherS.GetAll(req)
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

func (s *TeacherService) Update(ctx context.Context, req *m.TeacherUpdateReq) (*m.Void, error) {
	return s.storage.TeacherS.Update(req)
}

func (s *TeacherService) Delete(ctx context.Context, req *m.ById) (*m.Void, error) {
	return s.storage.TeacherS.Delete(req)
}
