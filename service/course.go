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
	"log/slog"
)

type CourseService struct {
	storage st.Storage
	push    *PushService
	rdb     *redis.Client
}

func NewCourseService(storage *st.Storage, push *PushService, rdb *redis.Client) *CourseService {
	return &CourseService{storage: *storage, push: push, rdb: rdb}
}

func (s *CourseService) getVersion() string {
	version := "1"
	if s.rdb != nil {
		if v, err := s.rdb.Get("course:version").Result(); err == nil && v != "" {
			version = v
		}
	}
	return version
}

func (s *CourseService) Create(ctx context.Context, req *m.CourseCreateReq) (*m.ById, error) {
	res, err := s.storage.CourseS.Create(req)
	if err != nil {
		return nil, err
	}

	if s.rdb != nil {
		_ = s.rdb.Incr("course:version").Err()
	}

	courseNameUz, _ := helper.PickFromJSONStr(req.Name, "uz")

	push_target := m.PushTarget{Type: m.TargetAllUsers}
	push_event := m.PushEvent{
		Type:       m.EventNewCourse,
		CourseID:   res.Id,
		CourseName: courseNameUz,
	}

	err = s.push.SendPush(ctx,
		push_target,
		push_event,
	)

	if err != nil {
		fmt.Println("Failed to send push:", err)
	}

	msg := m.BuildUzPushMessage(push_event)

	title, body := helper.StringToJSON(msg.Title, msg.Body)

	ntf, err := s.storage.NotificationS.Create(&m.NotificationCreateReq{
		Title:   title,
		Message: body,
		Type:    "all",
	})

	if err != nil {
		fmt.Println("Failed to create notification:", err)
	}

	fmt.Println("Push event sent:", push_event)
	fmt.Println("Push message:", msg)
	fmt.Println("Notification created with ID:", ntf.Id)

	return res, nil
}

func (s *CourseService) GetById(ctx context.Context, req *m.ById) (*m.CourseRes, error) {
	return s.storage.CourseS.GetById(req)
}
func (s *CourseService) GetAll(ctx context.Context, req *m.CourseGetListReq) (*m.PaginationResponse, error) {
	version := s.getVersion()
	key := fmt.Sprintf(
		"course:list:v%s_s:%s_t:%s_name:%s_pub:%s_page:%d_limit:%d",
		version,
		req.SubjectId,
		req.TeacherId,
		req.Name,
		req.IsPublic,
		req.Filter.Page,
		req.Filter.Limit,
	)

	if s.rdb != nil {
		if b, err := s.rdb.Get(key).Bytes(); err == nil && len(b) > 0 {
			var cached m.PaginationResponse
			if err := json.Unmarshal(b, &cached); err == nil {
				slog.Info("Cache Hit", "service", "course", "key", key)
				return &cached, nil
			}
		}
	}

	slog.Info("Cache Miss", "service", "course", "key", key)

	res, _, err := s.storage.CourseS.GetAll(req)
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

func (s *CourseService) Update(ctx context.Context, req *m.CourseUpdateReq) (*m.Void, error) {
	res, err := s.storage.CourseS.Update(req)
	if err != nil {
		return nil, err
	}
	if s.rdb != nil {
		_ = s.rdb.Incr("course:version").Err()
	}

	info, err := s.storage.CourseS.GetById(&m.ById{Id: req.Id})
	if err == nil {
		push_target := m.PushTarget{Type: m.TargetAllUsers}

		courseNameUz, _ := helper.PickFromJSONStr(req.Name, "uz")

		push_event := m.PushEvent{
			Type:       m.EventCourseUpdated,
			CourseID:   info.Id,
			CourseName: courseNameUz,
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


func (s *CourseService) Delete(ctx context.Context, req *m.ById) (*m.Void, error) {
	res, err := s.storage.CourseS.Delete(req)
	if err != nil {
		return nil, err
	}
	if s.rdb != nil {
		_ = s.rdb.Incr("course:version").Err()
	}
	return res, nil
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
	err = s.push.SendPush(ctx,
		push_target,
		push_event,
	)

	if err != nil {
		fmt.Println("Failed to send push:", err)
	}

	msg := m.BuildUzPushMessage(push_event)

	title, body := helper.StringToJSON(msg.Title, msg.Body)

	_, err = s.storage.NotificationS.Create(&m.NotificationCreateReq{
		Title:    title,
		Message:  body,
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
	// fmt.Println("Notification created with ID:", ntf.Id)

	return res, nil
}

func (s *CourseService) UserCourse(ctx context.Context, req *m.ById) (*m.UserCourseRes, error) {
	return s.storage.CourseS.UserCourse(req)
}
func (s *CourseService) UserCourses(ctx context.Context, req *m.UserCourseGetListReq) (*m.PaginationResponse, *m.UserCourseList, error) {
	return s.storage.CourseS.UserCourses(req)
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

func (s *CourseService) GetAllMobile(ctx context.Context, req *m.CourseGetListMobileReq) (*m.PaginationResponse, error) {
	version := s.getVersion()

	key := fmt.Sprintf(
		"course:mobile:list:v%s_s:%s_u:%s_name:%s_pub:%s_l:%s",
		version,
		req.SubjectId,
		req.UserId,
		req.Name,
		req.IsPublic,
		req.Language,
	)

	if s.rdb != nil {
		if b, err := s.rdb.Get(key).Bytes(); err == nil && len(b) > 0 {
			var cached m.PaginationResponse
			if err := json.Unmarshal(b, &cached); err == nil {
				slog.Info("Cache Hit (Mobile)", "service", "course", "key", key)
				return &cached, nil
			}
		}
	}

	slog.Info("Cache Miss (Mobile)", "service", "course", "key", key)

	res, _, err := s.storage.CourseS.GetAllMobile(req)
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
func (s *CourseService) UserCourseMobile(ctx context.Context, req *m.ByIdMobile) (*m.UserCourseMobileRes, error) {
	return s.storage.CourseS.UserCourseMobile(req)
}
func (s *CourseService) UserCoursesMobile(ctx context.Context, req *m.UserCourseGetListMobileReq) (*m.PaginationResponse, *m.UserCourseMobileList, error) {

	return s.storage.CourseS.UserCoursesMobile(req)
}
