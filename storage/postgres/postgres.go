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
	OrderS        storage.OrderI
	TariffS       storage.TariffI
	PromoCodeS    storage.PromoCodeI
	DashboardS    storage.DashBoardI
	UserS         storage.UserI
}

func NewPostgresStorage(cf config.Config) (*Storage, error) {
	conn := fmt.Sprintf("host=%s user=%s dbname=%s password=%s port=%d sslmode=disable",
		cf.POSTGRES_HOST, cf.POSTGRES_USER, cf.POSTGRES_NAME, cf.POSTGRES_PASSWORD, cf.POSTGRES_PORT)
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
	order := NewOrderRepo(db)
	tariff := NewTariffRepo(db)
	promocode := NewPromoCodeRepo(db)
	dashboard := NewDashboardRepo(db)
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
		OrderS:        order,
		TariffS:       tariff,
		PromoCodeS:    promocode,
		DashboardS:    dashboard,
		UserS:         user,
	}, nil
}
