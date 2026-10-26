package service

import (
	"context"
	"encoding/json"
	"fmt"
	m "imed/pkg/models"
	st "imed/storage/postgres"
	"time"

	"github.com/redis/go-redis/v9"
)

type TeacherService struct {
	storage st.Storage
	rdb     *redis.Client
}

func NewTeacherService(storage *st.Storage, rdb *redis.Client) *TeacherService {
	return &TeacherService{storage: *storage, rdb: rdb}
}

func (s *TeacherService) getVersion() string {
	version := "1"
	if s.rdb != nil {
		if v, err := s.rdb.Get(context.Background(), "teacher:version").Result(); err == nil && v != "" {
			version = v
		}
	}
	return version
}

func (s *TeacherService) Create(ctx context.Context, req *m.TeacherCreateReq) (*m.ById, error) {
	res, err := s.storage.TeacherS.Create(req)
	if err != nil {
		return nil, err
	}

	if s.rdb != nil {
		_ = s.rdb.Incr(context.Background(), "teacher:version").Err()
	}

	return res, nil
}

func (s *TeacherService) GetById(ctx context.Context, req *m.ById) (*m.TeacherRes, error) {
	return s.storage.TeacherS.GetById(req)
}

func (s *TeacherService) GetAll(ctx context.Context, req *m.TeacherGetListReq) (*m.PaginationResponse, *m.TeacherList, error) {

	version := s.getVersion()

	key := fmt.Sprintf(
		"teacher:list:v%s_name:%s_phonenumber:%s_page:%d_limit:%d",
		version,
		req.Name,
		req.PhoneNumber,
		req.Filter.Page,
		req.Filter.Limit,
	)

	if s.rdb != nil {
		if b, err := s.rdb.Get(context.Background(), key).Bytes(); err == nil && len(b) > 0 {
			var cached m.PaginationResponse
			if err := json.Unmarshal(b, &cached); err == nil {
				return &cached, nil, nil
			}
		}
	}

	res, list, err := s.storage.TeacherS.GetAll(req)
	if err != nil {
		return nil, nil, err
	}

	if s.rdb != nil {
		if b, err := json.Marshal(res); err == nil {
			_ = s.rdb.Set(context.Background(), key, b, 5*time.Minute).Err()
		}
	}

	return res, list, nil
}

func (s *TeacherService) Update(ctx context.Context, req *m.TeacherUpdateReq) (*m.Void, error) {
	res, err := s.storage.TeacherS.Update(req)
	if err != nil {
		return nil, err
	}

	if s.rdb != nil {
		_ = s.rdb.Incr(context.Background(), "teacher:version").Err()
	}

	return res, nil
}

func (s *TeacherService) Delete(ctx context.Context, req *m.ById) (*m.Void, error) {
	res, err := s.storage.TeacherS.Delete(req)
	if err != nil {
		return nil, err
	}

	if s.rdb != nil {
		_ = s.rdb.Incr(context.Background(), "teacher:version").Err()
	}

	return res, nil
}
