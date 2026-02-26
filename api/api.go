package api

import (
	"log/slog"

	"time"

	"imed/api/handler"
	mid "imed/api/middleware"

	"github.com/casbin/casbin/v2"

	_ "imed/api/docs"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Imed API
// @version 1.0
// @description API for Imed platform
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func NewApi(h *handler.Handler) *gin.Engine {
	router := gin.Default()

	enforcer, err := casbin.NewEnforcer("./api/casbin/model.conf", "./api/casbin/policy.csv")
	if err != nil {
		slog.Error("Error while creating enforcer: ", "err", err)
	}

	if enforcer == nil {
		slog.Error("Enforcer is nil after initialization!")
	} else {
		slog.Info("Enforcer initialized successfully.")
	}

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           24 * time.Hour,
	}))

	router.GET("/api/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.POST("/web/auth/admin/login", h.Login)
	router.POST("/mobile/auth/user/check", h.CheckUser)
	router.POST("/mobile/auth/user/login", h.LoginUser)
	router.POST("/click/prepare", h.ClickPrepare)
	router.POST("/click/complete", h.ClickComplete)
	router.POST("/mobile/auth/user/otp/send", h.SendConfirmation)
	router.POST("/mobile/auth/user/otp/confirm", h.VerifyConfirmation)

	web := router.Group("/web")
	{
		auth := web.Group("/auth").Use(mid.NewAuth(enforcer))
		{
			auth.POST("/user/create", h.CreateUser)
			auth.PUT("/password/change", h.ChangePasswordWeb)
			auth.PUT("/password/refresh", h.RefreshPassword)
		}

		user := web.Group("/user").Use(mid.NewAuth(enforcer))
		{
			user.GET("", h.GetAllUsers)
			user.GET("/:id", h.GetByIdUser)
			user.DELETE("/:id/delete", h.DeleteUser)
		}

		web.GET("/admin/:id", h.GetByIdAdmin).Use(mid.NewAuth(enforcer))

		banner := web.Group("/banner").Use(mid.NewAuth(enforcer))
		{
			banner.POST("", h.CreateBanner)
			banner.GET("", h.GetAllBanners)
			banner.GET("/:id", h.GetByIdBanner)
			banner.PUT("/:id/update", h.UpdateBanner)
			banner.DELETE("/:id/delete", h.DeleteBanner)
		}

		about := web.Group("/about").Use(mid.NewAuth(enforcer))
		{
			about.POST("", h.CreateAbout)
			about.GET("", h.GetAllAbouts)
			about.GET("/:id", h.GetByIdAbout)
			about.PUT("/:id/update", h.UpdateAbout)
			about.DELETE("/:id/delete", h.DeleteAbout)
		}

		app_route := web.Group("/app-route").Use(mid.NewAuth(enforcer))
		{
			app_route.POST("", h.CreateAppRoute)
			app_route.GET("", h.GetAllAppRoutes)
			app_route.GET("/:id", h.GetByIdAppRoute)
			app_route.PUT("/:id/update", h.UpdateAppRoute)
			app_route.DELETE("/:id/delete", h.DeleteAppRoute)
		}

		faq := web.Group("/faq").Use(mid.NewAuth(enforcer))
		{
			faq.POST("", h.CreateFaq)
			faq.GET("", h.GetAllFaqs)
			faq.GET("/:id", h.GetByIdFaq)
			faq.PUT("/:id/update", h.UpdateFaq)
			faq.DELETE("/:id/delete", h.DeleteFaq)
		}

		course := web.Group("/course").Use(mid.NewAuth(enforcer))
		{
			course.POST("", h.CreateCourse)
			course.GET("", h.GetAllCourses)
			course.GET("/:id", h.GetByIdCourse)
			course.PUT("/:id/update", h.UpdateCourse)
			course.DELETE("/:id/delete", h.DeleteCourse)
			course.POST("/permission", h.BuyCourse)
			course.GET("/permission/:id", h.UserCourse)
			course.GET("/permission", h.UserCourses)
			course.PUT("/permission/:id/update", h.UpdateUserCourse)
			course.DELETE("/permission/:id/delete", h.DeleteUserCourse)
		}

		contact := web.Group("/contact").Use(mid.NewAuth(enforcer))
		{
			contact.POST("", h.CreateContact)
			contact.GET("", h.GetAllContacts)
			contact.GET("/:id", h.GetByIdContact)
			contact.PUT("/:id/update", h.UpdateContact)
			contact.DELETE("/:id/delete", h.DeleteContact)
		}

		tariff := web.Group("/tariff").Use(mid.NewAuth(enforcer))
		{
			tariff.POST("", h.CreateTariff)
			tariff.GET("", h.GetAllTariffs)
			tariff.GET("/:id", h.GetByIdTariff)
			tariff.PUT("/:id/update", h.UpdateTariff)
			tariff.DELETE("/:id/delete", h.DeleteTariff)
		}

		lesson := web.Group("/lesson").Use(mid.NewAuth(enforcer))
		{
			lesson.POST("", h.CreateLesson)
			lesson.GET("", h.GetAllLessons)
			lesson.GET("/:id", h.GetByIdLesson)
			lesson.PUT("/:id/update", h.UpdateLesson)
			lesson.DELETE("/:id/delete", h.DeleteLesson)
		}

		module := web.Group("/module").Use(mid.NewAuth(enforcer))
		{
			module.POST("", h.CreateModule)
			module.GET("", h.GetAllModules)
			module.GET("/:id", h.GetByIdModule)
			module.PUT("/:id/update", h.UpdateModule)
			module.DELETE("/:id/delete", h.DeleteModule)
		}

		teacher := web.Group("/teacher").Use(mid.NewAuth(enforcer))
		{
			teacher.POST("", h.CreateTeacher)
			teacher.GET("", h.GetAllTeachers)
			teacher.GET("/:id", h.GetByIdTeacher)
			teacher.PUT("/:id/update", h.UpdateTeacher)
			teacher.DELETE("/:id/delete", h.DeleteTeacher)
		}

		source := web.Group("/source").Use(mid.NewAuth(enforcer))
		{
			source.POST("", h.CreateSource)
			source.GET("", h.GetAllSources)
			source.GET("/:id", h.GetByIdSource)
			source.PUT("/:id/update", h.UpdateSource)
			source.DELETE("/:id/delete", h.DeleteSource)
		}

		subject := web.Group("/subject").Use(mid.NewAuth(enforcer))
		{
			subject.POST("", h.CreateSubject)
			subject.GET("", h.GetAllSubjects)
			subject.GET("/:id", h.GetByIdSubject)
			subject.PUT("/:id/update", h.UpdateSubject)
			subject.DELETE("/:id/delete", h.DeleteSubject)
		}

		notification := web.Group("/notification").Use(mid.NewAuth(enforcer))
		{
			notification.POST("", h.CreateNotification)
			notification.GET("", h.GetAllNotifications)
			notification.GET("/:id", h.GetByIdNotification)
			notification.PUT("/:id/update", h.UpdateNotification)
			notification.DELETE("/:id/delete", h.DeleteNotification)
			notification.PUT("/:id/read", h.ReadNotification)
			notification.GET("/user", h.GetAllUserNotifications)
		}

		promocode := web.Group("/promocode").Use(mid.NewAuth(enforcer))
		{
			promocode.POST("", h.CreatePromoCode)
			promocode.GET("", h.GetAllPromoCodes)
			promocode.GET("/:id", h.GetByIdPromoCode)
			promocode.PUT("/:id/update", h.UpdatePromoCode)
			promocode.DELETE("/:id/delete", h.DeletePromoCode)
			promocode.GET("/redemption", h.GetAllUserPromoCodes)
			promocode.GET("/redemption/:id", h.GetByIdUserPromoCode)
		}

		order := web.Group("/order").Use(mid.NewAuth(enforcer))
		{
			order.GET("/:id", h.GetByIdOrder)
			order.GET("", h.GetAllOrders)
		}

		dashboard := web.Group("/dashboard").Use(mid.NewAuth(enforcer))
		{
			dashboard.GET("", h.GetDashboard)
		}

		web.POST("/file-upload", h.UploadFile).Use(mid.NewAuth(enforcer))
	}

	mobile := router.Group("/mobile")
	{
		auth := mobile.Group("/auth").Use(mid.NewAuth(enforcer))
		{
			auth.PUT("/password/change", h.ChangePasswordMobile)
		}

		user := mobile.Group("/user").Use(mid.NewAuth(enforcer))
		{
			user.GET("/get/profile", h.GetProfileUser)
			user.PUT("/update/profile", h.UpdateProfileUser)
			user.DELETE("/delete/profile", h.DeleteProfileUser)
			user.POST("/activity", h.CreateActivity)
			user.GET("/activity", h.GetActivityStats)
			user.GET("/rating", h.GetRating)
		}

		banner := mobile.Group("/banner").Use(mid.NewAuth(enforcer))
		{
			banner.GET("", h.GetAllMobileBanners)
			banner.GET("/:id", h.GetByIdMobileBanner)
		}

		about := mobile.Group("/about").Use(mid.NewAuth(enforcer))
		{
			about.GET("", h.GetAllMobileAbouts)
			about.GET("/:id", h.GetByIdMobileAbout)
		}

		app_route := mobile.Group("/app-route").Use(mid.NewAuth(enforcer))
		{
			app_route.GET("", h.GetAllAppRoutes)
			app_route.GET("/:id", h.GetByIdAppRoute)
		}

		faq := mobile.Group("/faq").Use(mid.NewAuth(enforcer))
		{
			faq.GET("", h.GetAllMobileFaqs)
			faq.GET("/:id", h.GetByIdMobileFaq)
		}

		course := mobile.Group("/course").Use(mid.NewAuth(enforcer))
		{
			course.GET("", h.GetAllMobileCourses)
			course.GET("/:id", h.GetByIdMobileCourse)
			course.POST("/permission", h.BuyCourse)
			course.GET("/permission/:id", h.UserMobileCourse)
			course.GET("/permission", h.UserMobileCourses)
		}

		contact := mobile.Group("/contact").Use(mid.NewAuth(enforcer))
		{
			contact.GET("", h.GetAllContacts)
			contact.GET("/:id", h.GetByIdContact)
		}

		lesson := mobile.Group("/lesson").Use(mid.NewAuth(enforcer))
		{
			// lesson.GET("", h.GetAllLessons)
			lesson.GET("/:id", h.GetByIdMobileLesson)
			lesson.PUT("/:id/end", h.EndLesson)
		}

		// module := mobile.Group("/module").Use(mid.NewAuth(enforcer))
		// {
		// 	module.GET("", h.GetAllModules)
		// 	module.GET("/:id", h.GetByIdModule)
		// }

		// source := mobile.Group("/source").Use(mid.NewAuth(enforcer))
		// {
		// 	source.GET("", h.GetAllSources)
		// 	source.GET("/:id", h.GetByIdSource)
		// }

		subject := mobile.Group("/subject").Use(mid.NewAuth(enforcer))
		{
			subject.GET("", h.GetAllMobileSubjects)
			subject.GET("/:id", h.GetByIdMobileSubject)
		}

		notification := mobile.Group("/notification").Use(mid.NewAuth(enforcer))
		{
			notification.PUT("/:id/read", h.ReadNotification)
			notification.GET("/user", h.GetAllUserNotifications)
		}

		tariff := mobile.Group("/tariff").Use(mid.NewAuth(enforcer))
		{
			tariff.GET("", h.GetAllTariffs)
			tariff.GET("/:id", h.GetByIdTariff)
		}

		order := mobile.Group("/order").Use(mid.NewAuth(enforcer))
		{
			order.POST("", h.CreateOrder)
		}

		promocode := mobile.Group("/promocode").Use(mid.NewAuth(enforcer))
		{
			promocode.POST("/check", h.CheckPromoCode)
		}

	}
	return router
}
