package postgres

import (
	"database/sql"
	"fmt"

	"imed/internal/config"
	"imed/storage"

	_ "github.com/lib/pq"
)

type Storage struct {
	Db            *sql.DB
	AdminS        storage.AdminI
	AboutS        storage.AboutI
	BannerS       storage.BannerI
	ContactS      storage.ContactI
	CourseS       storage.CourseI
	FaqS          storage.FaqI
	LessonS       storage.LessonI
	ModuleS       storage.ModuleI
	SourceS       storage.SourceI
	SubjectS      storage.SubjectI
	TeacherS      storage.TeacherI
	NotificationS storage.NotificationI
	AppRouteS     storage.AppRouteI
	UserS         storage.UserI
}

func NewPostgresStorage(config config.Config) (*Storage, error) {
	conn := fmt.Sprintf("host=%s user=%s dbname=%s password=%s port=%d sslmode=disable",
		config.DB_HOST, config.DB_USER, config.DB_NAME, config.DB_PASSWORD, config.DB_PORT)
	db, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	admin := NewAdminRepo(db)
	about := NewAboutRepo(db)
	banner := NewBannerRepo(db)
	contact := NewContactRepo(db)
	course := NewCourseRepo(db)
	faq := NewFaqRepo(db)
	lesson := NewLessonRepo(db)
	module := NewModuleRepo(db)
	source := NewSourceRepo(db)
	subject := NewSubjectRepo(db)
	teacher := NewTeacherRepo(db)
	notification := NewNotificationRepo(db)
	app_route := NewAppRouteRepo(db)
	user := NewUserRepo(db)

	return &Storage{
		Db:            db,
		AdminS:        admin,
		AboutS:        about,
		BannerS:       banner,
		ContactS:      contact,
		CourseS:       course,
		FaqS:          faq,
		LessonS:       lesson,
		ModuleS:       module,
		SourceS:       source,
		TeacherS:      teacher,
		SubjectS:      subject,
		NotificationS: notification,
		AppRouteS:     app_route,
		UserS:         user,
	}, nil
}
