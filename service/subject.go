package service

import (
	"context"
	"encoding/json"
	"fmt"
	m "imed/pkg/models"
	st "imed/storage/postgres"
	"time"

	"log/slog"

	"github.com/go-redis/redis"
)

type SubjectService struct {
	storage st.Storage
	rdb     *redis.Client
	push    *PushService
}

func NewSubjectService(storage *st.Storage, rdb *redis.Client, push *PushService) *SubjectService {
	return &SubjectService{storage: *storage, rdb: rdb, push: push}
}

func (s *SubjectService) getVersion() string {
	version := "1"
	if s.rdb != nil {
		if v, err := s.rdb.Get("subject:version").Result(); err == nil && v != "" {
			version = v
		}
	}
	return version
}

func (s *SubjectService) Create(ctx context.Context, req *m.SubjectCreateReq) (*m.ById, error) {
	res, err := s.storage.SubjectS.Create(req)
	if err != nil {
		return nil, err
	}

	if s.rdb != nil {
		_ = s.rdb.Incr("subject:version").Err()
	}

	subjectTitleJSON := `{"uz": "Yangi fan qo'shildi", "ru": "Добавлен новый предмет", "en": "New subject added"}`
	_ = s.push.SendPush(ctx, m.PushTarget{Type: m.TargetAllUsers}, m.PushEvent{
		Type:        m.EventSubjectAdded,
		TitleJSON:   subjectTitleJSON,
		MessageJSON: req.Name,
	})

	return res, nil
}

func (s *SubjectService) GetById(ctx context.Context, req *m.ById) (*m.SubjectRes, error) {
	return s.storage.SubjectS.GetById(req)
}

func (s *SubjectService) GetAll(ctx context.Context, req *m.SubjectGetListReq) (*m.PaginationResponse, *m.SubjectList, error) {

	version := s.getVersion()

	key := fmt.Sprintf(
		"subject:list:v%s_name:%s_page:%d_limit:%d",
		version,
		req.Name,
		req.Filter.Page,
		req.Filter.Limit,
	)

	if s.rdb != nil {
		if b, err := s.rdb.Get(key).Bytes(); err == nil && len(b) > 0 {
			var cached m.PaginationResponse
			if err := json.Unmarshal(b, &cached); err == nil {
				slog.Info("Cache Hit", "service", "subject", "key", key)
				return &cached, nil, nil
			}
		}
	}

	slog.Info("Cache Miss", "service", "subject", "key", key)

	res, list, err := s.storage.SubjectS.GetAll(req)
	if err != nil {
		return nil, nil, err
	}

	if s.rdb != nil {
		if b, err := json.Marshal(res); err == nil {
			_ = s.rdb.Set(key, b, 5*time.Minute).Err()
		}
	}

	return res, list, nil
}

func (s *SubjectService) Update(ctx context.Context, req *m.SubjectUpdateReq) (*m.Void, error) {
	res, err := s.storage.SubjectS.Update(req)
	if err != nil {
		return nil, err
	}

	if s.rdb != nil {
		_ = s.rdb.Incr("subject:version").Err()
	}

	subjectTitleJSON := `{"uz": "Yangi fan qo'shildi", "ru": "Добавлен новый предмет", "en": "New subject added"}`
	_ = s.push.SendPush(ctx, m.PushTarget{Type: m.TargetAllUsers}, m.PushEvent{
		Type:        m.EventSubjectAdded,
		TitleJSON:   subjectTitleJSON,
		MessageJSON: req.Name,
	})

	return res, nil
}

func (s *SubjectService) Delete(ctx context.Context, req *m.ById) (*m.Void, error) {
	res, err := s.storage.SubjectS.Delete(req)
	if err != nil {
		return nil, err
	}

	if s.rdb != nil {
		_ = s.rdb.Incr("subject:version").Err()
	}

	return res, nil
}

func (s *SubjectService) GetByIdMobile(ctx context.Context, req *m.ByIdMobile) (*m.SubjectMobileRes, error) {
	return s.storage.SubjectS.GetByIdMobile(req)
}

func (s *SubjectService) GetAllMobile(ctx context.Context, req *m.SubjectGetListMobileReq) (*m.PaginationResponse, *m.SubjectMobileList, error) {

	version := s.getVersion()

	key := fmt.Sprintf(
		"subject:mobile:v%s_name:%s_lang:%s",
		version,
		req.Name,
		req.Language,
	)

	if s.rdb != nil {
		if b, err := s.rdb.Get(key).Bytes(); err == nil && len(b) > 0 {
			var cached m.PaginationResponse
			if err := json.Unmarshal(b, &cached); err == nil {
				slog.Info("Cache Hit (Mobile)", "service", "subject", "key", key)
				return &cached, nil, nil
			}
		}
	}

	slog.Info("Cache Miss (Mobile)", "service", "subject", "key", key)

	res, list, err := s.storage.SubjectS.GetAllMobile(req)
	if err != nil {
		return nil, nil, err
	}

	if s.rdb != nil {
		if b, err := json.Marshal(res); err == nil {
			_ = s.rdb.Set(key, b, 5*time.Minute).Err()
		}
	}

	return res, list, nil
}
