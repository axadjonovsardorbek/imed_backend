package storage

import (
	"context"
	m "imed/pkg/models"
	"time"
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
	GetLessonCourseInfo(req *m.ById) (*m.LessonCourseInfo, error)
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
	UpdateVideoUrl(id string, videoUrl string) error
	// GetByIdMobile(req *m.ByIdMobile) (*m.SourceMobileRes, error)
	// GetAllMobile(req *m.SourceGetListMobileReq) (*m.PaginationResponse, *m.SourceMobileList, error)
}

type UserI interface {
	EmailLogin(req *m.ConfirmationReq) (*m.TokenRes, error)
	TelegramLogin(req *m.TelegramLoginReq) (*m.TokenRes, error)
	GetToken(user_id string) (*string, error)
	GetTokens(ctx context.Context, f m.TokenFilter) ([]string, error)
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
	DeleteToken(req []string) (*m.Void, error)
	TouchLastSeen(ctx context.Context, userID string) error
	GetInactiveUsers(ctx context.Context, before time.Time, limit int) ([]m.InactiveUser, error)
	MarkInactivePushed(ctx context.Context, ids []string) error
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
	CreateWithPromoCode(req *m.OrderCreateReq) (*m.OrderCreateRes, error)
	GetById(req *m.ById) (*m.OrderStatisticRes, error)
	GetAll(req *m.OrderGetListReq) (*m.PaginationResponse, *m.OrderList, error)
	GetPayURL(req *m.ById) (*m.PayURLRes, error)

	// Click callbacks uchun
	ApplyPrepare(req *m.ApplyPrepareParams) (*m.PrepareResult, error)
	ApplyComplete(req *m.ApplyCompleteParams) (*m.CompleteResult, error)
}

type PromoCodeI interface {
	Create(req *m.PromoCodeCreateReq) (*m.ById, error)
	GetById(req *m.ById) (*m.PromoCodeRes, error)
	GetAll(req *m.PromoCodeGetListReq) (*m.PaginationResponse, *m.PromoCodeList, error)
	Update(req *m.PromoCodeUpdateReq) (*m.Void, error)
	Delete(req *m.ById) (*m.Void, error)
	Check(req *m.PromoCodeCheckReq) (*m.PromoCodeCheckRes, error)
	GetByIdUserPromo(req *m.ById) (*m.PromoRedemptionRes, error)
	GetAllUserPromo(req *m.PromoRedemptionGetListReq) (*m.PaginationResponse, *m.PromoRedemptionList, error)
}

type DashBoardI interface {
	GetDashboard(req *m.GetDashboardReq) (*m.DashboardRes, error)
}

type TestI interface {
	Create(ctx context.Context, req *m.TestCreateReq) (*m.ById, error)
	GetById(ctx context.Context, req *m.ById) (*m.TestRes, error)
	GetAll(ctx context.Context, req *m.TestGetListReq) (*m.PaginationResponse, error)
	Update(ctx context.Context, req *m.TestUpdateReq) (*m.Void, error)
	Delete(ctx context.Context, req *m.ById) (*m.Void, error)
	AddQuestions(ctx context.Context, testId string, questions []m.TestQuestionBody) ([]string, error)
	UpdateQuestion(ctx context.Context, req *m.TestQuestionReq) (*m.Void, error)
	DeleteQuestion(ctx context.Context, req *m.ById) (*m.Void, error)
	GetByIdMobile(ctx context.Context, req *m.ByIdMobile) (*m.TestMobileRes, error)
	GetByLessonMobile(ctx context.Context, lessonId, userId, lang string) ([]m.TestShortMobile, error)
	Start(ctx context.Context, req *m.ByIdMobile) (*m.TestStartRes, error)
	Answer(ctx context.Context, req *m.TestAnswerReq) (*m.TestAnswerRes, error)
	Finish(ctx context.Context, req *m.TestAttemptReq) (*m.TestFinishRes, error)
	History(ctx context.Context, req *m.TestHistoryReq) (*m.PaginationResponse, error)
}

type PremiumI interface {
	GetSettings(ctx context.Context) (*m.PremiumSettings, error)
	UpdateSettings(ctx context.Context, req *m.PremiumSettings) (*m.Void, error)
	GetStatus(ctx context.Context, userId string) (*m.AccessStatus, error)
	CreateOrder(ctx context.Context, req *m.PremiumOrderReq) (*m.OrderCreateRes, error)
	SetUstoz(ctx context.Context, req *m.UstozReq) (*m.Void, error)
}
