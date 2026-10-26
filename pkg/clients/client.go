package clients

import (
	"imed/internal/config"
	"imed/internal/push"
	"imed/pkg/s3"
	"imed/service"
	"imed/storage/postgres"

	"github.com/redis/go-redis/v9"
)

type Clients struct {
	Admin        *service.AdminService
	About        *service.AboutService
	Banner       *service.BannerService
	Contact      *service.ContactService
	Course       *service.CourseService
	Faq          *service.FaqService
	Lesson       *service.LessonService
	Module       *service.ModuleService
	Source       *service.SourceService
	Subject      *service.SubjectService
	Teacher      *service.TeacherService
	Notification *service.NotificationService
	AppRoute     *service.AppRouteService
	Order        *service.OrderService
	Tariff       *service.TariffService
	PromoCode    *service.PromoCodeService
	Dashboard    *service.DashboardService
	User         *service.UserService
	Push         *service.PushService
	Test         *service.TestService
	Premium      *service.PremiumService
}

func NewClients(cfg *config.Config, conn *postgres.Storage, rdb *redis.Client, pushClient *push.Client, s3Svc *s3.S3Service) (*Clients, error) {
	pushSvc := service.NewPushService(conn, pushClient, cfg.AUTO_CONTENT_PUSH)
	return &Clients{
		Push:         pushSvc,
		Test:         service.NewTestService(conn),
		Premium:      service.NewPremiumService(conn, rdb),
		Admin:        service.NewAdminService(conn),
		About:        service.NewAboutService(conn, rdb),
		Banner:       service.NewBannerService(conn, rdb, pushSvc),
		Contact:      service.NewContactService(conn, rdb),
		Course:       service.NewCourseService(conn, pushSvc, rdb),
		Faq:          service.NewFaqService(conn, rdb),
		Lesson:       service.NewLessonService(conn, rdb, pushSvc),
		Module:       service.NewModuleService(conn, rdb, pushSvc),
		Source:       service.NewSourceService(conn, rdb, pushSvc, s3Svc),
		Subject:      service.NewSubjectService(conn, rdb, pushSvc),
		Teacher:      service.NewTeacherService(conn, rdb),
		Notification: service.NewNotificationService(conn, rdb, pushSvc),
		AppRoute:     service.NewAppRouteService(conn, rdb),
		Order:        service.NewOrderService(conn, rdb, pushSvc),
		Tariff:       service.NewTariffService(conn, rdb),
		PromoCode:    service.NewPromoCodeService(conn, rdb),
		Dashboard:    service.NewDashboardService(conn, rdb),
		User:         service.NewUserService(conn, rdb),
	}, nil
}
