package service

import (
	"context"
	"imed/internal/helper"
	m "imed/pkg/models"
	st "imed/storage/postgres"

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

func (s *SourceService) Create(ctx context.Context, req *m.SourceCreateReq) (*m.ById, error) {
	res, err := s.storage.SourceS.Create(req)
	if err != nil {
		return nil, err
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
	// res := &m.PaginationResponse{}

	// key := fmt.Sprintf("lesson_id:%s_name:%s_type:%s_page:%d_limit:%d",
	// 	req.LessonId, req.Name, req.Type, req.Filter.Page, req.Filter.Limit)

	// if s.rdb != nil {
	// 	if b, err := s.rdb.Get(key).Bytes(); err == nil && len(b) > 0 {
	// 		_ = json.Unmarshal(b, &res)
	// 	}
	// }

	// if res.Total == 0 {
	// 	res, _, err := s.storage.SourceS.GetAll(req)
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

	return s.storage.SourceS.GetAll(req)
}

func (s *SourceService) Update(ctx context.Context, req *m.SourceUpdateReq) (*m.Void, error) {
	return s.storage.SourceS.Update(req)
}

func (s *SourceService) Delete(ctx context.Context, req *m.ById) (*m.Void, error) {
	return s.storage.SourceS.Delete(req)
}
