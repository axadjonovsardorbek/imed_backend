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

type NotificationService struct {
	storage st.Storage
	rdb     *redis.Client
	push    *PushService
}

func NewNotificationService(storage *st.Storage, rdb *redis.Client, push *PushService) *NotificationService {
	return &NotificationService{storage: *storage, rdb: rdb, push: push}
}

func (s *NotificationService) getVersion() string {
	version := "1"
	if s.rdb != nil {
		if v, err := s.rdb.Get("notification:version").Result(); err == nil && v != "" {
			version = v
		}
	}
	return version
}

func (s *NotificationService) Create(ctx context.Context, req *m.NotificationCreateReq) (*m.ById, error) {

	res, err := s.storage.NotificationS.Create(req)
	if err != nil {
		return nil, err
	}

	// 🔥 cache invalidation
	if s.rdb != nil {
		_ = s.rdb.Incr("notification:version").Err()
	}


	push_target := m.PushTarget{Type: m.TargetAllUsers}
	push_event := m.PushEvent{
		Type:        m.EventAdminNotification,
		TitleJSON:   req.Title,
		MessageJSON: req.Message,
	}

	if req.Type == "selected" && req.CourseId != "" {
		course, err := s.storage.CourseS.GetById(&m.ById{Id: req.CourseId})
		if err == nil {
			push_target = m.PushTarget{Type: m.TargetCourseUsers, CourseID: req.CourseId}
			push_event = m.PushEvent{
				Type:        m.EventAdminNotification,
				CourseID:    req.CourseId,
				CourseName:  course.Name["uz"],
				TitleJSON:   req.Title,
				MessageJSON: req.Message,
			}
		}
	}


	_ = s.push.SendPush(ctx, push_target, push_event)

	return res, nil
}

func (s *NotificationService) GetById(ctx context.Context, req *m.ById) (*m.MasterNotificationRes, error) {
	return s.storage.NotificationS.GetById(req)
}

func (s *NotificationService) GetAll(ctx context.Context, req *m.NotificationGetListReq) (*m.PaginationResponse, *m.MasterNotificationList, error) {

	version := s.getVersion()

	key := fmt.Sprintf(
		"notification:list:v%s_course:%s_title:%s_page:%d_limit:%d",
		version,
		req.CourseId,
		req.Title,
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

	res, list, err := s.storage.NotificationS.GetAll(req)
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

func (s *NotificationService) Update(ctx context.Context, req *m.MasterNotificationUpdateReq) (*m.Void, error) {
	res, err := s.storage.NotificationS.Update(req)
	if err != nil {
		return nil, err
	}

	if s.rdb != nil {
		_ = s.rdb.Incr("notification:version").Err()
	}

	return res, nil
}

func (s *NotificationService) Delete(ctx context.Context, req *m.ById) (*m.Void, error) {
	res, err := s.storage.NotificationS.Delete(req)
	if err != nil {
		return nil, err
	}

	if s.rdb != nil {
		_ = s.rdb.Incr("notification:version").Err()
	}

	return res, nil
}

func (s *NotificationService) UserNotifications(ctx context.Context, req *m.NotificationGetListReq) (*m.PaginationResponse, *m.UserNotificationList, error) {

	version := s.getVersion()

	key := fmt.Sprintf(
		"notification:user:v%s_lang:%s_user:%s_read:%s_page:%d_limit:%d",
		version,
		req.Language,
		req.UserId,
		req.IsRead,
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

	res, list, err := s.storage.NotificationS.UserNotifications(req)
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

func (s *NotificationService) MarkAsRead(ctx context.Context, req *m.NotificationMarkReadReq) (*m.Void, error) {
	return s.storage.NotificationS.MarkAsRead(req)
}
