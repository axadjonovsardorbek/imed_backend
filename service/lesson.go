package service

import (
	"context"
	"encoding/json"
	"fmt"
	"imed/internal/helper"
	m "imed/pkg/models"
	st "imed/storage/postgres"
	"time"

	"github.com/redis/go-redis/v9"
)

type LessonService struct {
	storage st.Storage
	rdb     *redis.Client
	push    *PushService
}

func NewLessonService(storage *st.Storage, rdb *redis.Client, push *PushService) *LessonService {
	return &LessonService{storage: *storage, rdb: rdb, push: push}
}

func (s *LessonService) getVersion() string {
	version := "1"
	if s.rdb != nil {
		if v, err := s.rdb.Get(context.Background(), "lesson:version").Result(); err == nil && v != "" {
			version = v
		}
	}
	return version
}

func (s *LessonService) Create(ctx context.Context, req *m.LessonCreateReq) (*m.ById, error) {
	res, err := s.storage.LessonS.Create(req)
	if err != nil {
		return nil, err
	}

	if s.rdb != nil {
		_ = s.rdb.Incr(context.Background(), "lesson:version").Err()
	}

	info, err := s.storage.LessonS.GetLessonCourseInfo(&m.ById{Id: res.Id})
	if err == nil {
		push_target := m.PushTarget{Type: m.TargetAllUsers}

		push_event := m.PushEvent{
			Type:       m.EventLessonAdded,
			CourseID:   info.CourseId,
			CourseName: info.CourseName,
			LessonID:   info.LessonId,
			LessonName: info.LessonName,
		}

		_ = s.push.SendPush(ctx, push_target, push_event)

		msg := m.BuildUzPushMessage(push_event)
		title, body := helper.StringToJSON(msg.Title, msg.Body)

		_, _ = s.storage.NotificationS.Create(&m.NotificationCreateReq{
			Title:   title,
			Message: body,
			Type:    "all",
		})
	}

	return res, nil
}

func (s *LessonService) GetById(ctx context.Context, req *m.ById) (*m.LessonRes, error) {
	return s.storage.LessonS.GetById(req)
}

func (s *LessonService) GetAll(ctx context.Context, req *m.LessonGetListReq) (*m.PaginationResponse, error) {
	version := s.getVersion()
	key := fmt.Sprintf(
		"lesson:list:v%s_module:%s_name:%s_type:%s_pub:%s_page:%d_limit:%d",
		version,
		req.ModuleId,
		req.Name,
		req.Type,
		req.IsPublic,
		req.Filter.Page,
		req.Filter.Limit,
	)

	if s.rdb != nil {
		if b, err := s.rdb.Get(context.Background(), key).Bytes(); err == nil && len(b) > 0 {
			var cached m.PaginationResponse
			if err := json.Unmarshal(b, &cached); err == nil {
				return &cached, nil
			}
		}
	}

	res, _, err := s.storage.LessonS.GetAll(req)
	if err != nil {
		return nil, err
	}

	if s.rdb != nil {
		if b, err := json.Marshal(res); err == nil {
			_ = s.rdb.Set(context.Background(), key, b, 5*time.Minute).Err()
		}
	}

	return res, nil
}

func (s *LessonService) Update(ctx context.Context, req *m.LessonUpdateReq) (*m.Void, error) {
	res, err := s.storage.LessonS.Update(req)
	if err != nil {
		return nil, err
	}
	if s.rdb != nil {
		_ = s.rdb.Incr(context.Background(), "lesson:version").Err()
	}

	info, err := s.storage.LessonS.GetLessonCourseInfo(&m.ById{Id: req.Id})
	if err == nil {
		push_target := m.PushTarget{Type: m.TargetAllUsers}

		push_event := m.PushEvent{
			Type:       m.EventLessonUpdated,
			CourseID:   info.CourseId,
			CourseName: info.CourseName,
			LessonID:   info.LessonId,
			LessonName: info.LessonName,
		}

		_ = s.push.SendPush(ctx, push_target, push_event)

		msg := m.BuildUzPushMessage(push_event)
		title, body := helper.StringToJSON(msg.Title, msg.Body)

		_, _ = s.storage.NotificationS.Create(&m.NotificationCreateReq{
			Title:   title,
			Message: body,
			Type:    "all",
		})
	}

	return res, nil
}

func (s *LessonService) Delete(ctx context.Context, req *m.ById) (*m.Void, error) {
	res, err := s.storage.LessonS.Delete(req)
	if err != nil {
		return nil, err
	}
	if s.rdb != nil {
		_ = s.rdb.Incr(context.Background(), "lesson:version").Err()
	}
	return res, nil
}

func (s *LessonService) EndLesson(ctx context.Context, req *m.EndUserLessonReq) (*m.Void, error) {
	return s.storage.LessonS.EndLesson(req)
}
func (s *LessonService) GetByIdMobile(ctx context.Context, req *m.ByIdMobile) (*m.SourceLessonMobileRes, error) {
	res, err := s.storage.LessonS.GetByIdMobile(req)
	if err != nil {
		return nil, err
	}

	res.Quizzes, err = s.storage.TestS.GetByLessonMobile(ctx, req.Id, req.UserId, req.Language)
	if err != nil {
		return nil, err
	}

	// T-06: mobil faqat shu ro'yxatdagi bo'limlarni ko'rsatadi, bo'sh bo'lim chiqmaydi.
	res.ContentTypes = make([]string, 0, 4)
	if len(res.Videos) > 0 {
		res.ContentTypes = append(res.ContentTypes, m.SourceVideo)
	}
	if len(res.Audios) > 0 {
		res.ContentTypes = append(res.ContentTypes, m.SourceAudio)
	}
	if len(res.Documents) > 0 {
		res.ContentTypes = append(res.ContentTypes, m.SourceDocument)
	}
	if len(res.Quizzes) > 0 {
		res.ContentTypes = append(res.ContentTypes, m.SourceTest)
	}
	return res, nil
}
