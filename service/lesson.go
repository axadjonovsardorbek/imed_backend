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

type LessonService struct {
	storage st.Storage
	rdb     *redis.Client
}

func NewLessonService(storage *st.Storage, rdb *redis.Client) *LessonService {
	return &LessonService{storage: *storage, rdb: rdb}
}

func (s *LessonService) Create(ctx context.Context, req *m.LessonCreateReq) (*m.ById, error) {
	return s.storage.LessonS.Create(req)
}

func (s *LessonService) GetById(ctx context.Context, req *m.ById) (*m.LessonRes, error) {
	return s.storage.LessonS.GetById(req)
}

func (s *LessonService) GetAll(ctx context.Context, req *m.LessonGetListReq) (*m.PaginationResponse, *m.LessonList, error) {
	res := &m.PaginationResponse{}

	key := fmt.Sprintf("module_id:%s_name:%s_type:%s_is_public:%s_page:%d_limit:%d",
		req.ModuleId, req.Name, req.Type, req.IsPublic, req.Filter.Page, req.Filter.Limit)

	if s.rdb != nil {
		if b, err := s.rdb.Get(key).Bytes(); err == nil && len(b) > 0 {
			_ = json.Unmarshal(b, &res)
		}
	}

	if res.Total == 0 {
		res, _, err := s.storage.LessonS.GetAll(req)
		if err != nil {
			return nil, nil, err
		}

		if s.rdb != nil {
			if b, err := json.Marshal(res); err == nil {
				_ = s.rdb.Set(key, b, 60*time.Second).Err()
			}
		}
	}

	return res, nil, nil
}

func (s *LessonService) Update(ctx context.Context, req *m.LessonUpdateReq) (*m.Void, error) {
	return s.storage.LessonS.Update(req)
}

func (s *LessonService) Delete(ctx context.Context, req *m.ById) (*m.Void, error) {
	return s.storage.LessonS.Delete(req)
}

func (s *LessonService) EndLesson(ctx context.Context, req *m.EndUserLessonReq) (*m.Void, error) {
	return s.storage.LessonS.EndLesson(req)
}
func (s *LessonService) GetByIdMobile(ctx context.Context, req *m.ByIdMobile) (*m.SourceLessonMobileRes, error) {
	res := &m.SourceLessonMobileRes{}

	key := fmt.Sprintf("lesson_id:%s_user_id:%s_lang:%s",
		req.Id, req.UserId, req.Language)

	if s.rdb != nil {
		if b, err := s.rdb.Get(key).Bytes(); err == nil && len(b) > 0 {
			_ = json.Unmarshal(b, &res)
		}
	}

	if res.Id == "" {
		res, err := s.storage.LessonS.GetByIdMobile(req)
		if err != nil {
			return nil, err
		}

		if s.rdb != nil {
			if b, err := json.Marshal(res); err == nil {
				_ = s.rdb.Set(key, b, 60*time.Second).Err()
			}
		}
	}

	return res, nil
}
