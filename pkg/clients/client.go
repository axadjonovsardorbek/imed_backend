package clients

import (
	"imed/internal/config"
	"imed/service"
	"imed/storage/postgres"
)

type Clients struct {
	Admin        *service.AdminService
	Banner       *service.BannerService
	Contact      *service.ContactService
	Course       *service.CourseService
	Lesson       *service.LessonService
	Module       *service.ModuleService
	Source       *service.SourceService
	Subject      *service.SubjectService
	Teacher      *service.TeacherService
	Notification *service.NotificationService
	User         *service.UserService
}

func NewClients(cfg *config.Config, conn *postgres.Storage) (*Clients, error) {
	return &Clients{
		Admin:        service.NewAdminService(conn),
		Banner:       service.NewBannerService(conn),
		Contact:      service.NewContactService(conn),
		Course:       service.NewCourseService(conn),
		Lesson:       service.NewLessonService(conn),
		Module:       service.NewModuleService(conn),
		Source:       service.NewSourceService(conn),
		Subject:      service.NewSubjectService(conn),
		Teacher:      service.NewTeacherService(conn),
		Notification: service.NewNotificationService(conn),
		User:         service.NewUserService(conn),
	}, nil
}
