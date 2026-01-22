package clients

import (
	"imed/internal/config"
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
	User         *service.UserService
}

func NewClients(cfg *config.Config, conn *postgres.Storage, rdb *redis.Client) (*Clients, error) {
	return &Clients{
		Admin:        service.NewAdminService(conn),
		About:        service.NewAboutService(conn),
		Banner:       service.NewBannerService(conn),
		Contact:      service.NewContactService(conn),
		Course:       service.NewCourseService(conn),
		Faq:          service.NewFaqService(conn),
		Lesson:       service.NewLessonService(conn),
		Module:       service.NewModuleService(conn),
		Source:       service.NewSourceService(conn),
		Subject:      service.NewSubjectService(conn),
		Teacher:      service.NewTeacherService(conn),
		Notification: service.NewNotificationService(conn),
		User:         service.NewUserService(conn, rdb),
	}, nil
}
