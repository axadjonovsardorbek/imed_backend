package service

import (
	"context"
	"fmt"
	"imed/internal/helper"
	"imed/internal/push"
	m "imed/pkg/models"
	st "imed/storage/postgres"
)

type CourseService struct {
	storage st.Storage
	push    *push.Client
}

func NewCourseService(storage *st.Storage, push *push.Client) *CourseService {
	return &CourseService{storage: *storage, push: push}
}

func (s *CourseService) Create(ctx context.Context, req *m.CourseCreateReq) (*m.ById, error) {
	res, err := s.storage.CourseS.Create(req)
	if err != nil {
		return nil, err
	}

	tokens, err := s.storage.UserS.GetTokens()
	if err != nil {
		fmt.Println("Error getting tokens:", err)
	} else {
		fmt.Println("Tokens retrieved:", tokens)
		_, err = s.push.SendToManyTokens(ctx, tokens, "Yangi kurs!", "Yangi kurs qo‘shildi 🔥", map[string]string{
			"type": "new_course",
		})
		if err != nil {
			fmt.Println("Error sending push to many tokens:", err)
		}
	}

	return res, nil
}

func (s *CourseService) GetById(ctx context.Context, req *m.ById) (*m.CourseRes, error) {
	return s.storage.CourseS.GetById(req)
}

func (s *CourseService) GetAll(ctx context.Context, req *m.CourseGetListReq) (*m.PaginationResponse, *m.CourseList, error) {
	return s.storage.CourseS.GetAll(req)
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

	token, err := s.storage.UserS.GetToken(req.UserId)
	if err != nil {
		fmt.Println("Error getting token:", err)
	} else {
		fmt.Println("Token retrieved:", token)
		err = s.push.SendToToken(ctx, *token, "Kurs sotib olindi", "Kursingiz aktivlashtirildi ✅", map[string]string{
			"type": "course_purchased",
		})

		if err != nil {
			fmt.Println("Error sending push to token:", err)
		}
	}

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
func (s *CourseService) GetAllMobile(ctx context.Context, req *m.CourseGetListMobileReq) (*m.PaginationResponse, *m.CourseMobileList, error) {
	return s.storage.CourseS.GetAllMobile(req)
}
func (s *CourseService) UserCourseMobile(ctx context.Context, req *m.ByIdMobile) (*m.UserCourseMobileRes, error) {
	return s.storage.CourseS.UserCourseMobile(req)
}
func (s *CourseService) UserCoursesMobile(ctx context.Context, req *m.UserCourseGetListMobileReq) (*m.PaginationResponse, *m.UserCourseMobileList, error) {
	return s.storage.CourseS.UserCoursesMobile(req)
}
