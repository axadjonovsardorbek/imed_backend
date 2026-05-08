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
		if v, err := s.rdb.Get("lesson:version").Result(); err == nil && v != "" {
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
		_ = s.rdb.Incr("lesson:version").Err()
	}

	info, err := s.storage.LessonS.GetLessonCourseInfo(&m.ById{Id: res.Id})
	if err == nil {
		course, err := s.storage.CourseS.GetById(&m.ById{Id: info.CourseId})
		if err == nil {
			var push_target m.PushTarget
			if course.IsPublic != nil && *course.IsPublic {
				push_target = m.PushTarget{Type: m.TargetAllUsers}
			} else {
				push_target = m.PushTarget{Type: m.TargetCourseUsers, CourseID: course.Id}
			}

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

			ntfType := "selected"
			if course.IsPublic != nil && *course.IsPublic {
				ntfType = "all"
			}

			_, _ = s.storage.NotificationS.Create(&m.NotificationCreateReq{
				Title:    title,
				Message:  body,
				Type:     ntfType,
				CourseId: info.CourseId,
			})
		}
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
		if b, err := s.rdb.Get(key).Bytes(); err == nil && len(b) > 0 {
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
			_ = s.rdb.Set(key, b, 5*time.Minute).Err()
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
		_ = s.rdb.Incr("lesson:version").Err()
	}

	info, err := s.storage.LessonS.GetLessonCourseInfo(&m.ById{Id: req.Id})
	if err == nil {
		course, err := s.storage.CourseS.GetById(&m.ById{Id: info.CourseId})
		if err == nil {
			var push_target m.PushTarget
			if course.IsPublic != nil && *course.IsPublic {
				push_target = m.PushTarget{Type: m.TargetAllUsers}
			} else {
				push_target = m.PushTarget{Type: m.TargetCourseUsers, CourseID: course.Id}
			}

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

			ntfType := "selected"
			if course.IsPublic != nil && *course.IsPublic {
				ntfType = "all"
			}

			_, _ = s.storage.NotificationS.Create(&m.NotificationCreateReq{
				Title:    title,
				Message:  body,
				Type:     ntfType,
				CourseId: info.CourseId,
			})
		}
	}

	return res, nil
}


func (s *LessonService) Delete(ctx context.Context, req *m.ById) (*m.Void, error) {
	res, err := s.storage.LessonS.Delete(req)
	if err != nil {
		return nil, err
	}
	if s.rdb != nil {
		_ = s.rdb.Incr("lesson:version").Err()
	}
	return res, nil
}

func (s *LessonService) EndLesson(ctx context.Context, req *m.EndUserLessonReq) (*m.Void, error) {
	return s.storage.LessonS.EndLesson(req)
}
func (s *LessonService) GetByIdMobile(ctx context.Context, req *m.ByIdMobile) (*m.SourceLessonMobileRes, error) {
	return s.storage.LessonS.GetByIdMobile(req)
}
