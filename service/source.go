package service

import (
	"context"
	"encoding/json"
	"fmt"
	"imed/internal/helper"
	m "imed/pkg/models"
	st "imed/storage/postgres"
	"time"

	"github.com/go-redis/redis"
)

type SourceService struct {
	storage st.Storage
	rdb     *redis.Client
	push    *PushService
}

func NewSourceService(storage *st.Storage, rdb *redis.Client, push *PushService) *SourceService {
	return &SourceService{storage: *storage, rdb: rdb, push: push}
}

func (s *SourceService) getVersion() string {
	version := "1"
	if s.rdb != nil {
		if v, err := s.rdb.Get("source:version").Result(); err == nil && v != "" {
			version = v
		}
	}
	return version
}

func (s *SourceService) Create(ctx context.Context, req *m.SourceCreateReq) (*m.ById, error) {
	res, err := s.storage.SourceS.Create(req)
	if err != nil {
		return nil, err
	}

	// 🔥 cache invalidation
	if s.rdb != nil {
		_ = s.rdb.Incr("source:version").Err()
	}

	info, err := s.storage.LessonS.GetLessonCourseInfo(&m.ById{Id: res.Id})
	if err == nil {

		push_target := m.PushTarget{
			Type:     m.TargetCourseUsers,
			CourseID: info.CourseId,
		}

		push_event := m.PushEvent{
			Type:       m.EventSourceAdded,
			CourseID:   info.CourseId,
			CourseName: info.CourseName,
			LessonID:   info.LessonId,
			LessonName: info.LessonName,
			SourceType: req.Type,
		}

		_ = s.push.SendPush(ctx, push_target, push_event)

		msg := m.BuildUzPushMessage(push_event)
		title, body := helper.StringToJSON(msg.Title, msg.Body)

		_, _ = s.storage.NotificationS.Create(&m.NotificationCreateReq{
			Title:    title,
			Message:  body,
			Type:     "selected",
			CourseId: info.CourseId,
		})
	}

	return res, nil
}

func (s *SourceService) GetById(ctx context.Context, req *m.ById) (*m.SourceRes, error) {
	return s.storage.SourceS.GetById(req)
}

func (s *SourceService) GetAll(ctx context.Context, req *m.SourceGetListReq) (*m.PaginationResponse, *m.SourceList, error) {

	version := s.getVersion()

	key := fmt.Sprintf(
		"source:list:v%s_lesson:%s_name:%s_type:%s_page:%d_limit:%d",
		version,
		req.LessonId,
		req.Name,
		req.Type,
		req.Filter.Page,
		req.Filter.Limit,
	)

	if s.rdb != nil {
		if b, err := s.rdb.Get(key).Bytes(); err == nil && len(b) > 0 {
			var cached m.PaginationResponse
			if err := json.Unmarshal(b, &cached); err == nil {
				return &cached, nil, nil
			}
		}
	}

	res, list, err := s.storage.SourceS.GetAll(req)
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

func (s *SourceService) Update(ctx context.Context, req *m.SourceUpdateReq) (*m.Void, error) {
	res, err := s.storage.SourceS.Update(req)
	if err != nil {
		return nil, err
	}

	if s.rdb != nil {
		_ = s.rdb.Incr("source:version").Err()
	}

	return res, nil
}

func (s *SourceService) Delete(ctx context.Context, req *m.ById) (*m.Void, error) {
	res, err := s.storage.SourceS.Delete(req)
	if err != nil {
		return nil, err
	}

	if s.rdb != nil {
		_ = s.rdb.Incr("source:version").Err()
	}

	return res, nil
}
