package storage

import (
	m "imed/pkg/models"
)

type AdminI interface {
	Login(req *m.LoginReq) (*m.TokenRes, error)
	GetById(req *m.ById) (*m.AdminRes, error)
}

type BannerI interface {
	Create(req *m.BannerCreateReq) (*m.ById, error)
	GetById(req *m.ById) (*m.BannerRes, error)
	GetAll(req *m.BannerGetListReq) (*m.PaginationResponse, *m.BannerList, error)
	Update(req *m.BannerUpdateReq) (*m.Void, error)
	Delete(req *m.ById) (*m.Void, error)
	GetByIdMobile(req *m.ByIdMobile) (*m.BannerMobileRes, error)
	GetAllMobile(req *m.BannerGetListMobileReq) (*m.PaginationResponse, *m.BannerMobileList, error)
}

type TeacherI interface {
	Create(req *m.TeacherCreateReq) (*m.ById, error)
	GetById(req *m.ById) (*m.TeacherRes, error)
	GetAll(req *m.TeacherGetListReq) (*m.PaginationResponse, *m.TeacherList, error)
	Update(req *m.TeacherUpdateReq) (*m.Void, error)
	Delete(req *m.ById) (*m.Void, error)
}

type SubjectI interface {
	Create(req *m.SubjectCreateReq) (*m.ById, error)
	GetById(req *m.ById) (*m.SubjectRes, error)
	GetAll(req *m.SubjectGetListReq) (*m.PaginationResponse, *m.SubjectList, error)
	Update(req *m.SubjectUpdateReq) (*m.Void, error)
	Delete(req *m.ById) (*m.Void, error)
	GetByIdMobile(req *m.ByIdMobile) (*m.SubjectMobileRes, error)
	GetAllMobile(req *m.SubjectGetListMobileReq) (*m.PaginationResponse, *m.SubjectMobileList, error)
}

type CourseI interface {
	Create(req *m.CourseCreateReq) (*m.ById, error)
	GetById(req *m.ById) (*m.CourseRes, error)
	GetAll(req *m.CourseGetListReq) (*m.PaginationResponse, *m.CourseList, error)
	Update(req *m.CourseUpdateReq) (*m.Void, error)
	Delete(req *m.ById) (*m.Void, error)
	BuyCourse(req *m.UserCourseCreateReq) (*m.Void, error)
	UserCourse(req *m.ById) (*m.UserCourseRes, error)
	UserCourses(req *m.UserCourseGetListReq) (*m.PaginationResponse, *m.UserCourseList, error)
	UserUpdateCourse(req *m.UserCourseUpdateReq) (*m.Void, error)
	DeleteUserCourse(req *m.ById) (*m.Void, error)
	GetByIdMobile(req *m.ByIdMobile) (*m.MobileCourseRes, error)
	GetAllMobile(req *m.CourseGetListMobileReq) (*m.PaginationResponse, *m.CourseMobileList, error)
	UserCourseMobile(req *m.ByIdMobile) (*m.UserCourseMobileRes, error)
	UserCoursesMobile(req *m.UserCourseGetListMobileReq) (*m.PaginationResponse, *m.UserCourseMobileList, error)
}

type ModuleI interface {
	Create(req *m.ModuleCreateReq) (*m.ById, error)
	GetById(req *m.ById) (*m.ModuleRes, error)
	GetAll(req *m.ModuleGetListReq) (*m.PaginationResponse, *m.ModuleList, error)
	Update(req *m.ModuleUpdateReq) (*m.Void, error)
	Delete(req *m.ById) (*m.Void, error)
	// GetByIdMobile(req *m.ByIdMobile) (*m.ModuleMobileRes, error)
	// GetAllMobile(req *m.ModuleGetListMobileReq) (*m.PaginationResponse, *m.ModuleMobileList, error)
}

type LessonI interface {
	Create(req *m.LessonCreateReq) (*m.ById, error)
	GetById(req *m.ById) (*m.LessonRes, error)
	GetAll(req *m.LessonGetListReq) (*m.PaginationResponse, *m.LessonList, error)
	Update(req *m.LessonUpdateReq) (*m.Void, error)
	Delete(req *m.ById) (*m.Void, error)
	EndLesson(req *m.EndUserLessonReq) (*m.Void, error)
	GetByIdMobile(req *m.ByIdMobile) (*m.SourceLessonMobileRes, error)
	// GetAllMobile(req *m.LessonGetListMobileReq) (*m.PaginationResponse, *m.LessonMobileList, error)
}

type ContactI interface {
	Create(req *m.ContactCreateReq) (*m.ById, error)
	GetById(req *m.ById) (*m.ContactRes, error)
	GetAll(req *m.ContactGetListReq) (*m.PaginationResponse, *m.ContactList, error)
	Update(req *m.ContactUpdateReq) (*m.Void, error)
	Delete(req *m.ById) (*m.Void, error)
}

type SourceI interface {
	Create(req *m.SourceCreateReq) (*m.ById, error)
	GetById(req *m.ById) (*m.SourceRes, error)
	GetAll(req *m.SourceGetListReq) (*m.PaginationResponse, *m.SourceList, error)
	Update(req *m.SourceUpdateReq) (*m.Void, error)
	Delete(req *m.ById) (*m.Void, error)
	// GetByIdMobile(req *m.ByIdMobile) (*m.SourceMobileRes, error)
	// GetAllMobile(req *m.SourceGetListMobileReq) (*m.PaginationResponse, *m.SourceMobileList, error)
}

type UserI interface {
	CreateUser(req *m.UserLoginReq) (*m.ById, error)
	Check(req *m.UserCheckReq) (*m.UserCheckRes, error)
	Login(req *m.UserLoginReq) (*m.TokenRes, error)
	GetById(req *m.ById) (*m.UserRes, error)
	GetAll(req *m.UserGetListReq) (*m.PaginationResponse, *m.UserList, error)
	Update(req *m.UserUpdateReq) (*m.Void, error)
	Delete(req *m.ById) (*m.Void, error)
	ChangePassword(req *m.ChangePasswordReq) (*m.Void, error)
	RefreshPassword(req *m.RefreshPasswordReq) (*m.RefreshPassword, error)
	CreateActivity(req *m.UserActivityCreateReq) (*m.Void, error)
	GetActivities(req *m.UserActivityGetListReq) (*m.ActivityStatsResponse, error)
	GetRating(req *m.RatingRequest) (*m.RatingResponse, error)
	GetTop(req *m.RatingRequest) ([]m.RatingUserItem, error)
	GetMe(req *m.RatingRequest) (*m.RatingUserItem, error)
}

type NotificationI interface {
	Create(req *m.NotificationCreateReq) (*m.ById, error)
	GetById(req *m.ById) (*m.MasterNotificationRes, error)
	GetAll(req *m.NotificationGetListReq) (*m.PaginationResponse, *m.MasterNotificationList, error)
	Update(req *m.MasterNotificationUpdateReq) (*m.Void, error)
	Delete(req *m.ById) (*m.Void, error)
	UserNotifications(req *m.NotificationGetListReq) (*m.PaginationResponse, *m.UserNotificationList, error)
	MarkAsRead(req *m.NotificationMarkReadReq) (*m.Void, error)
}

type FaqI interface {
	Create(req *m.FaqCreateReq) (*m.ById, error)
	GetById(req *m.ById) (*m.FaqRes, error)
	GetAll(req *m.FaqGetListReq) (*m.PaginationResponse, *m.FaqList, error)
	Update(req *m.FaqUpdateReq) (*m.Void, error)
	Delete(req *m.ById) (*m.Void, error)
	GetByIdMobile(req *m.ByIdMobile) (*m.FaqMobileRes, error)
	GetAllMobile(req *m.FaqGetListMobileReq) (*m.PaginationResponse, *m.FaqMobileList, error)
}

type AboutI interface {
	Create(req *m.AboutCreateReq) (*m.ById, error)
	GetById(req *m.ById) (*m.AboutRes, error)
	GetAll(req *m.AboutGetListReq) (*m.PaginationResponse, *m.AboutList, error)
	Update(req *m.AboutUpdateReq) (*m.Void, error)
	Delete(req *m.ById) (*m.Void, error)
	GetByIdMobile(req *m.ByIdMobile) (*m.AboutMobileRes, error)
	GetAllMobile(req *m.AboutGetListMobileReq) (*m.PaginationResponse, *m.AboutMobileList, error)
}

type AppRouteI interface {
	Create(req *m.AppRouteCreateReq) (*m.ById, error)
	GetById(req *m.ById) (*m.AppRouteRes, error)
	GetAll(req *m.AppRouteGetListReq) (*m.PaginationResponse, *m.AppRouteList, error)
	Update(req *m.AppRouteUpdateReq) (*m.Void, error)
	Delete(req *m.ById) (*m.Void, error)
}

type TariffI interface {
	Create(req *m.TariffCreateReq) (*m.ById, error)
	GetById(req *m.ById) (*m.TariffRes, error)
	GetAll(req *m.TariffGetListReq) (*m.PaginationResponse, *m.TariffList, error)
	Update(req *m.TariffUpdateReq) (*m.Void, error)
	Delete(req *m.ById) (*m.Void, error)
}

type OrderI interface {
	Create(req *m.OrderCreateReq) (*m.OrderCreateRes, error)
	GetById(req *m.ById) (*m.OrderRes, error)
	GetPayURL(req *m.ById) (*m.PayURLRes, error)

	// Click callbacks uchun
	ApplyPrepare(req *m.ApplyPrepareParams) (*m.PrepareResult, error)
	ApplyComplete(req *m.ApplyCompleteParams) (*m.CompleteResult, error)
}
