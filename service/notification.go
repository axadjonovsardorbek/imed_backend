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

type NotificationService struct {
	storage st.Storage
	rdb     *redis.Client
	push    *PushService
}

func NewNotificationService(storage *st.Storage, rdb *redis.Client, push *PushService) *NotificationService {
	return &NotificationService{storage: *storage, rdb: rdb, push: push}
}

func (s *NotificationService) Create(ctx context.Context, req *m.NotificationCreateReq) (*m.ById, error) {

	res, err := s.storage.NotificationS.Create(req)
	if err != nil {
		return nil, err
	}

	titleUz, _ := helper.PickFromJSONStr(req.Title, "uz")
	messageUz, _ := helper.PickFromJSONStr(req.Message, "uz")
	push_target := m.PushTarget{Type: m.TargetAllUsers}
	push_event := m.PushEvent{
		Type:       m.EventNewCourse,
		CourseID:   titleUz,
		CourseName: messageUz,
	}

	if req.Type == "selected" && req.CourseId != "" {
		course, _ := s.storage.CourseS.GetById(&m.ById{Id: req.CourseId})

		push_target = m.PushTarget{Type: m.TargetCourseUsers, CourseID: req.CourseId}
		push_event = m.PushEvent{
			Type:       m.EventNewCourse,
			CourseID:   req.CourseId,
			CourseName: course.Name["uz"],
		}
	}

	err = s.push.SendPush(ctx,
		push_target,
		push_event,
	)
	if err != nil {
		fmt.Println("\n\n\n\\Failed to send push:", err)
	}

	fmt.Println("Push event sent:", push_event)

	fmt.Printf("Notification created: %s\n", res.Id)
	fmt.Println("Push event sent:", push_event)
	fmt.Println()

	return res, nil
}

func (s *NotificationService) GetById(ctx context.Context, req *m.ById) (*m.MasterNotificationRes, error) {
	return s.storage.NotificationS.GetById(req)
}

func (s *NotificationService) GetAll(ctx context.Context, req *m.NotificationGetListReq) (*m.PaginationResponse, *m.MasterNotificationList, error) {
	res := &m.PaginationResponse{}

	key := fmt.Sprintf("course_id:%s_title:%s_page:%d_limit:%d",
		req.CourseId, req.Title, req.Filter.Page, req.Filter.Limit)

	if s.rdb != nil {
		if b, err := s.rdb.Get(key).Bytes(); err == nil && len(b) > 0 {
			_ = json.Unmarshal(b, &res)
		}
	}

	if res.Total == 0 {
		res, _, err := s.storage.NotificationS.GetAll(req)
		if err != nil {
			return nil, nil, err
		}

		if s.rdb != nil {
			if b, err := json.Marshal(res); err == nil {
				_ = s.rdb.Set(key, b, 30*time.Second).Err()
			}
		}

		return res, nil, nil
	}

	return res, nil, nil
}

func (s *NotificationService) Update(ctx context.Context, req *m.MasterNotificationUpdateReq) (*m.Void, error) {
	return s.storage.NotificationS.Update(req)
}

func (s *NotificationService) Delete(ctx context.Context, req *m.ById) (*m.Void, error) {
	return s.storage.NotificationS.Delete(req)
}

func (s *NotificationService) UserNotifications(ctx context.Context, req *m.NotificationGetListReq) (*m.PaginationResponse, *m.UserNotificationList, error) {
	res := &m.PaginationResponse{}

	key := fmt.Sprintf("lang:%s_user_id:%s_is_read:%s",
		req.Language, req.UserId, req.IsRead)

	if s.rdb != nil {
		if b, err := s.rdb.Get(key).Bytes(); err == nil && len(b) > 0 {
			_ = json.Unmarshal(b, &res)
		}
	}

	if res.Total == 0 {
		res, _, err := s.storage.NotificationS.UserNotifications(req)
		if err != nil {
			return nil, nil, err
		}

		if s.rdb != nil {
			if b, err := json.Marshal(res); err == nil {
				_ = s.rdb.Set(key, b, 30*time.Second).Err()
			}
		}

		return res, nil, nil
	}

	return res, nil, nil
}

func (s *NotificationService) MarkAsRead(ctx context.Context, req *m.NotificationMarkReadReq) (*m.Void, error) {
	return s.storage.NotificationS.MarkAsRead(req)
}
