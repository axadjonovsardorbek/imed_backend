package clients

import (
	"imed/internal/config"
	"imed/internal/push"
	"imed/service"
	"imed/storage/postgres"

	"github.com/go-redis/redis"
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
	User         *service.UserService
}

func NewClients(cfg *config.Config, conn *postgres.Storage, rdb *redis.Client, push *push.Client) (*Clients, error) {
	return &Clients{
		Admin:        service.NewAdminService(conn),
		About:        service.NewAboutService(conn, rdb),
		Banner:       service.NewBannerService(conn, rdb),
		Contact:      service.NewContactService(conn, rdb),
		Course:       service.NewCourseService(conn, push, rdb),
		Faq:          service.NewFaqService(conn, rdb),
		Lesson:       service.NewLessonService(conn, rdb),
		Module:       service.NewModuleService(conn, rdb),
		Source:       service.NewSourceService(conn, rdb),
		Subject:      service.NewSubjectService(conn, rdb),
		Teacher:      service.NewTeacherService(conn, rdb),
		Notification: service.NewNotificationService(conn, rdb),
		AppRoute:     service.NewAppRouteService(conn, rdb),
		Order:        service.NewOrderService(conn, rdb),
		Tariff:       service.NewTariffService(conn, rdb),
		PromoCode:    service.NewPromoCodeService(conn, rdb),
		User:         service.NewUserService(conn, rdb),
	}, nil
}
