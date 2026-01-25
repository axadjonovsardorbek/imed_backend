package service

import (
	"context"
	m "imed/pkg/models"
	st "imed/storage/postgres"
)

type CourseService struct {
	storage st.Storage
}

func NewCourseService(storage *st.Storage) *CourseService {
	return &CourseService{storage: *storage}
}

func (s *CourseService) Create(ctx context.Context, req *m.CourseCreateReq) (*m.ById, error) {
	return s.storage.CourseS.Create(req)
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
	return s.storage.CourseS.BuyCourse(req)
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
