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

func (s *LessonService) Create(ctx context.Context, req *m.LessonCreateReq) (*m.ById, error) {
	res, err := s.storage.LessonS.Create(req)
	if err != nil {
		return nil, err
	}

	info, err := s.storage.LessonS.GetLessonCourseInfo(&m.ById{Id: res.Id})
	if err == nil {

		push_target := m.PushTarget{Type: m.TargetCourseUsers, CourseID: info.CourseId}
		push_event := m.PushEvent{
			Type:       m.EventLessonAdded,
			CourseID:   info.CourseId,
			CourseName: info.CourseName,
			LessonID:   info.LessonId,
			LessonName: info.LessonName,
		}
		_ = s.push.SendPush(ctx,
			push_target,
			push_event,
		)

		msg := m.BuildUzPushMessage(push_event)

		title, body := helper.StringToJSON(msg.Title, msg.Body)

		ntf, err := s.storage.NotificationS.Create(&m.NotificationCreateReq{
			Title:    title,
			Message:  body,
			Type:     "selected",
			CourseId: info.CourseId,
		})

		if err != nil {
			fmt.Println("Failed to create notification:", err)
		}

		fmt.Println("Push event sent:", push_event)
		fmt.Println("Push message:", msg)
		fmt.Println("Lesson created with ID:", res.Id)
		fmt.Println("Notification created with ID:", ntf.Id)
	}

	return res, nil
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

		return res, nil, nil
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
	return s.storage.LessonS.GetByIdMobile(req)
}
