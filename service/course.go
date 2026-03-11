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

type CourseService struct {
	storage st.Storage
	push    *PushService
	rdb     *redis.Client
}

func NewCourseService(storage *st.Storage, push *PushService, rdb *redis.Client) *CourseService {
	return &CourseService{storage: *storage, push: push, rdb: rdb}
}

func (s *CourseService) Create(ctx context.Context, req *m.CourseCreateReq) (*m.ById, error) {
	res, err := s.storage.CourseS.Create(req)
	if err != nil {
		return nil, err
	}

	courseNameUz, _ := helper.PickFromJSONStr(req.Name, "uz")

	push_target := m.PushTarget{Type: m.TargetAllUsers}
	push_event := m.PushEvent{
		Type:       m.EventNewCourse,
		CourseID:   res.Id,
		CourseName: courseNameUz,
	}

	_ = s.push.SendPush(ctx,
		push_target,
		push_event,
	)

	msg := m.BuildUzPushMessage(push_event)

	s.storage.NotificationS.Create(&m.NotificationCreateReq{
		Title:   msg.Title,
		Message: msg.Body,
		Type:    "all",
	})

	return res, nil
}

func (s *CourseService) GetById(ctx context.Context, req *m.ById) (*m.CourseRes, error) {
	return s.storage.CourseS.GetById(req)
}

func (s *CourseService) GetAll(ctx context.Context, req *m.CourseGetListReq) (*m.PaginationResponse, *m.CourseList, error) {

	res := &m.PaginationResponse{}

	key := fmt.Sprintf("s_id:%s_t_id:%s_name:%s_is_pub:%s_price:%d_user_id:%s_role:%s_page:%d_limit:%d",
		req.SubjectId, req.TeacherId, req.Name, req.IsPublic, req.Price, req.UserId, req.Role, req.Filter.Page, req.Filter.Limit)

	if s.rdb != nil {
		if b, err := s.rdb.Get(key).Bytes(); err == nil && len(b) > 0 {
			_ = json.Unmarshal(b, &res)
		}
	}

	if res.Total == 0 {
		res, _, err := s.storage.CourseS.GetAll(req)
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

func (s *CourseService) Update(ctx context.Context, req *m.CourseUpdateReq) (*m.Void, error) {
	return s.storage.CourseS.Update(req)
}

func (s *CourseService) Delete(ctx context.Context, req *m.ById) (*m.Void, error) {
	return s.storage.CourseS.Delete(req)
}

func (s *CourseService) BuyCourse(ctx context.Context, req *m.UserCourseCreateReq) (*m.Void, error) {

	if req.StartedAt == "" || req.EndedAt == "" {
		tariff, err := s.storage.TariffS.GetById(&m.ById{Id: req.TariffId})
		if err != nil {
			fmt.Println("Error getting tariff by ID:", err)
			return nil, err
		}
		start, end := helper.GetStartEndDate(*tariff.Duration)
		req.StartedAt = start
		req.EndedAt = end
	}

	fmt.Println("Buying course with details:", req)
	res, err := s.storage.CourseS.BuyCourse(req)
	if err != nil {
		fmt.Println("Error buying course:", err)
		return nil, err
	}

	course, _ := s.storage.CourseS.GetById(&m.ById{Id: req.CourseId})

	push_target := m.PushTarget{Type: m.TargetUser, UserID: req.UserId}
	push_event := m.PushEvent{
		Type:       m.EventCourseAssigned,
		CourseID:   course.Id,
		CourseName: course.Name["uz"],
	}
	_ = s.push.SendPush(ctx,
		push_target,
		push_event,
	)

	msg := m.BuildUzPushMessage(push_event)

	ntf, err := s.storage.NotificationS.Create(&m.NotificationCreateReq{
		Title:    msg.Title,
		Message:  msg.Body,
		Type:     "selected",
		UserId:   req.UserId,
		CourseId: req.CourseId,
	})

	if err != nil {
		fmt.Println("Failed to create notification:", err)
	}

	fmt.Println("Push event sent:", push_event)
	fmt.Println("Push message:", msg)
	fmt.Println("Course bought successfully for user:", req.UserId)
	fmt.Println("Notification created with ID:", ntf.Id)

	return res, nil
}
func (s *CourseService) UserCourse(ctx context.Context, req *m.ById) (*m.UserCourseRes, error) {
	return s.storage.CourseS.UserCourse(req)
}
func (s *CourseService) UserCourses(ctx context.Context, req *m.UserCourseGetListReq) (*m.PaginationResponse, *m.UserCourseList, error) {

	res := &m.PaginationResponse{}

	key := fmt.Sprintf("user_id:%s_course_id:%s_tariff_id:%s_is_active:%s_page:%d_limit:%d",
		req.UserId, req.CourseId, req.TariffId, req.IsActive, req.Filter.Page, req.Filter.Limit)

	if s.rdb != nil {
		if b, err := s.rdb.Get(key).Bytes(); err == nil && len(b) > 0 {
			_ = json.Unmarshal(b, &res)
		}
	}

	if res.Total == 0 {
		res, _, err := s.storage.CourseS.UserCourses(req)
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
func (s *CourseService) UserUpdateCourse(ctx context.Context, req *m.UserCourseUpdateReq) (*m.Void, error) {
	return s.storage.CourseS.UserUpdateCourse(req)
}
func (s *CourseService) DeleteUserCourse(ctx context.Context, req *m.ById) (*m.Void, error) {
	return s.storage.CourseS.DeleteUserCourse(req)
}
func (s *CourseService) GetByIdMobile(ctx context.Context, req *m.ByIdMobile) (*m.MobileCourseRes, error) {
	return s.storage.CourseS.GetByIdMobile(req)
}
func (s *CourseService) GetAllMobile(ctx context.Context, req *m.CourseGetListMobileReq) (*m.PaginationResponse, *m.CourseMobileList, error) {

	res := &m.PaginationResponse{}

	key := fmt.Sprintf("s_id:%s_name:%s_is_pub:%s_user_id:%s_lang:%s",
		req.SubjectId, req.Name, req.IsPublic, req.UserId, req.Language)

	if s.rdb != nil {
		if b, err := s.rdb.Get(key).Bytes(); err == nil && len(b) > 0 {
			_ = json.Unmarshal(b, &res)
		}
	}

	if res.Total == 0 {
		res, _, err := s.storage.CourseS.GetAllMobile(req)
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
func (s *CourseService) UserCourseMobile(ctx context.Context, req *m.ByIdMobile) (*m.UserCourseMobileRes, error) {
	return s.storage.CourseS.UserCourseMobile(req)
}
func (s *CourseService) UserCoursesMobile(ctx context.Context, req *m.UserCourseGetListMobileReq) (*m.PaginationResponse, *m.UserCourseMobileList, error) {

	res := &m.PaginationResponse{}

	key := fmt.Sprintf("user_id:%s_tariff_id:%s_is_active:%s_lang:%s",
		req.UserId, req.TariffId, req.IsActive, req.Language)

	if s.rdb != nil {
		if b, err := s.rdb.Get(key).Bytes(); err == nil && len(b) > 0 {
			_ = json.Unmarshal(b, &res)
		}
	}

	if res.Total == 0 {
		res, _, err := s.storage.CourseS.UserCoursesMobile(req)
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
