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

type SubjectService struct {
	storage st.Storage
	rdb     *redis.Client
}

func NewSubjectService(storage *st.Storage, rdb *redis.Client) *SubjectService {
	return &SubjectService{storage: *storage, rdb: rdb}
}

func (s *SubjectService) Create(ctx context.Context, req *m.SubjectCreateReq) (*m.ById, error) {
	return s.storage.SubjectS.Create(req)
}

func (s *SubjectService) GetById(ctx context.Context, req *m.ById) (*m.SubjectRes, error) {
	return s.storage.SubjectS.GetById(req)
}

func (s *SubjectService) GetAll(ctx context.Context, req *m.SubjectGetListReq) (*m.PaginationResponse, *m.SubjectList, error) {
	// res := &m.PaginationResponse{}

	// key := fmt.Sprintf("name:%s_page:%d_limit:%d",
	// 	req.Name, req.Filter.Page, req.Filter.Limit)

	// if s.rdb != nil {
	// 	if b, err := s.rdb.Get(key).Bytes(); err == nil && len(b) > 0 {
	// 		_ = json.Unmarshal(b, &res)
	// 	}
	// }

	// if res.Total == 0 {
	// 	res, _, err := s.storage.SubjectS.GetAll(req)
	// 	if err != nil {
	// 		return nil, nil, err
	// 	}

	// 	if s.rdb != nil {
	// 		if b, err := json.Marshal(res); err == nil {
	// 			_ = s.rdb.Set(key, b, 60*time.Second).Err()
	// 		}
	// 	}

	// 	return res, nil, nil
	// }

	return s.storage.SubjectS.GetAll(req)
}

func (s *SubjectService) Update(ctx context.Context, req *m.SubjectUpdateReq) (*m.Void, error) {
	return s.storage.SubjectS.Update(req)
}

func (s *SubjectService) Delete(ctx context.Context, req *m.ById) (*m.Void, error) {
	return s.storage.SubjectS.Delete(req)
}

func (s *SubjectService) GetByIdMobile(ctx context.Context, req *m.ByIdMobile) (*m.SubjectMobileRes, error) {
	return s.storage.SubjectS.GetByIdMobile(req)
}

func (s *SubjectService) GetAllMobile(ctx context.Context, req *m.SubjectGetListMobileReq) (*m.PaginationResponse, *m.SubjectMobileList, error) {
	res := &m.PaginationResponse{}

	key := fmt.Sprintf("name:%s_lang:%s",
		req.Name, req.Language)

	if s.rdb != nil {
		if b, err := s.rdb.Get(key).Bytes(); err == nil && len(b) > 0 {
			_ = json.Unmarshal(b, &res)
		}
	}

	if res.Total == 0 {
		res, _, err := s.storage.SubjectS.GetAllMobile(req)
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
