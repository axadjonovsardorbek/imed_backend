package models

import "fmt"

type TokenFilter struct {
	CourseID string
}

type PushTargetType string

const (
	TargetAllUsers    PushTargetType = "all_users"    // kurs qo‘shilganda
	TargetCourseUsers PushTargetType = "course_users" // kursga dars qo‘shilganda
	TargetUser        PushTargetType = "user"         // kurs biriktirilsa / payment status
	TargetTopic       PushTargetType = "topic"        // topic uchun (news_uz, news_ru, news_en)
)

type PushTarget struct {
	Type     PushTargetType
	UserID   string // TargetUser uchun
	CourseID string // TargetCourseUsers uchun
	Topic    string // TargetTopic uchun
}

type PushEventType string

const (
	SourceTest     = "test"
	SourceDocument = "document"
	SourceVideo    = "video"
	SourceAudio    = "audio"
)

const (
	EventNewCourse         PushEventType = "new_course"
	EventLessonAdded       PushEventType = "lesson_added"
	EventCourseAssigned    PushEventType = "course_assigned"
	EventPaymentSuccess    PushEventType = "payment_success"
	EventPaymentFailed     PushEventType = "payment_failed"
	EventAdminNotification PushEventType = "admin_notification"
	EventModuleAdded       PushEventType = "module_added"
	EventSourceAdded       PushEventType = "source_added"

	EventCourseUpdated = "course_updated"
	EventModuleUpdated = "module_updated"
	EventLessonUpdated = "lesson_updated"
	EventSourceUpdated = "source_updated"
	EventBannerAdded   = "banner_added"
	EventSubjectAdded  = "subject_added"
	EventSubjectUpdated = "subject_updated"
)




type PushEvent struct {
	Type PushEventType

	// umumiy
	CourseID   string
	CourseName string // UZ nom (name->>'uz')

	// admin notification uchun (JSON string)
	TitleJSON   string
	MessageJSON string

	// Direct overrides (used for localized admin notifications)
	Title string
	Body  string

	// lesson uchun
	LessonID   string
	LessonName string

	// module uchun
	ModuleID   string
	ModuleName string




	// payment uchun
	OrderID    string
	Amount     string // "1000.00" formatda yuborsang yaxshi
	FailReason string // optional

	SourceType string // source_added uchun: test, document, video
}

type PushMessage struct {
	Title string
	Body  string
	Data  map[string]string
}

func BuildUzPushMessage(e PushEvent) PushMessage {
	switch e.Type {

	case EventNewCourse:
		return PushMessage{
			Title: "Yangi kurs! 🔥",
			Body:  fmt.Sprintf("Yangi kurs qo‘shildi: %s 🔥", e.CourseName),
			Data: map[string]string{
				"type":      "new_course",
				"course_id": e.CourseID,
			},
		}


	case EventLessonAdded:
		return PushMessage{
			Title: "Yangi dars qo‘shildi 📚",
			Body:  fmt.Sprintf("%s kursiga yangi dars: %s 📚", e.CourseName, e.LessonName),
			Data: map[string]string{
				"type":      "lesson_added",
				"course_id": e.CourseID,
				"lesson_id": e.LessonID,
			},
		}


	case EventCourseUpdated:
		return PushMessage{
			Title: "Kurs yangilandi! 🔄",
			Body:  fmt.Sprintf("%s kursi yangilandi 🔄", e.CourseName),
			Data: map[string]string{
				"type":      "course_updated",
				"course_id": e.CourseID,
			},
		}


	case EventModuleAdded:
		return PushMessage{
			Title: "Yangi modul qo‘shildi 📚",
			Body:  fmt.Sprintf("%s kursiga yangi %s nomli modul qo‘shildi 📚", e.CourseName, e.ModuleName),
			Data: map[string]string{
				"type":      "module_added",
				"course_id": e.CourseID,
				"module_id": e.ModuleID,
			},
		}


	case EventModuleUpdated:
		return PushMessage{
			Title: "Modul yangilandi 🔄",
			Body:  fmt.Sprintf("%s kursidagi %s moduli yangilandi 🔄", e.CourseName, e.ModuleName),
			Data: map[string]string{
				"type":      "module_updated",
				"course_id": e.CourseID,
				"module_id": e.ModuleID,
			},
		}


	case EventLessonUpdated:
		return PushMessage{
			Title: "Dars yangilandi 🔄",
			Body:  fmt.Sprintf("%s kursidagi %s darsi yangilandi 🔄", e.CourseName, e.LessonName),
			Data: map[string]string{
				"type":      "lesson_updated",
				"course_id": e.CourseID,
				"lesson_id": e.LessonID,
			},
		}


	case EventSourceAdded:
		return PushMessage{
			Title: "Yangi material qo‘shildi 📚",
			Body:  fmt.Sprintf("%s kursining %s darsiga yangi material qo‘shildi 📚", e.CourseName, e.LessonName),
			Data: map[string]string{
				"type":      "source_added",
				"course_id": e.CourseID,
				"lesson_id": e.LessonID,
			},
		}


	case EventSourceUpdated:
		return PushMessage{
			Title: "Material yangilandi 🔄",
			Body:  fmt.Sprintf("%s kursining %s darsidagi material yangilandi 🔄", e.CourseName, e.LessonName),
			Data: map[string]string{
				"type":      "source_updated",
				"course_id": e.CourseID,
				"lesson_id": e.LessonID,
			},
		}


	case EventCourseAssigned:
		return PushMessage{
			Title: "Kurs biriktirildi ✅",
			Body:  fmt.Sprintf("Sizga kurs biriktirildi: %s ✅", e.CourseName),
			Data: map[string]string{
				"type":      "course_assigned",
				"course_id": e.CourseID,
			},
		}


	case EventPaymentSuccess:
		return PushMessage{
			Title: "To‘lov qabul qilindi ✅",
			Body:  fmt.Sprintf("To‘lov muvaffaqiyatli. Kurs: %s", e.CourseName),
			Data: map[string]string{
				"type":      "payment_success",
				"order_id":  e.OrderID,
				"course_id": e.CourseID,
				"amount":    e.Amount,
			},
		}

	case EventPaymentFailed:
		body := fmt.Sprintf("To‘lov qabul qilinmadi. Kurs: %s", e.CourseName)
		if e.FailReason != "" {
			body = body + ". Sabab: " + e.FailReason
		}
		return PushMessage{
			Title: "To‘lov bekor qilindi ❌",
			Body:  body,
			Data: map[string]string{
				"type":      "payment_failed",
				"order_id":  e.OrderID,
				"course_id": e.CourseID,
				"amount":    e.Amount,
			},
		}

	case EventAdminNotification:
		return PushMessage{
			Title: e.Title,
			Body:  e.Body,
			Data: map[string]string{
				"type": "admin_notification",
				"info": e.Body,
			},
		}

	case EventBannerAdded:
		return PushMessage{
			Title: e.Title,
			Body:  e.Body,
			Data: map[string]string{
				"type": "banner_added",
			},
		}

	case EventSubjectAdded:
		return PushMessage{
			Title: e.Title,
			Body:  fmt.Sprintf("Yangi fan qo‘shildi: %s 📖", e.Body),
			Data: map[string]string{
				"type": "subject_added",
			},
		}


	case EventSubjectUpdated:
		return PushMessage{
			Title: e.Title,
			Body:  fmt.Sprintf("Fan yangilandi: %s 🔄", e.Body),
			Data: map[string]string{
				"type": "subject_updated",
			},
		}





	default:
		return PushMessage{
			Title: "Bildirishnoma",
			Body:  "Yangi xabar bor",
			Data:  map[string]string{"type": "unknown"},
		}
	}
}

func BuildRuPushMessage(e PushEvent) PushMessage {
	switch e.Type {

	case EventNewCourse:
		return PushMessage{
			Title: "Новый курс! 🔥",
			Body:  fmt.Sprintf("Добавлен новый курс: %s 🔥", e.CourseName),
			Data: map[string]string{
				"type":      "new_course",
				"course_id": e.CourseID,
			},
		}


	case EventLessonAdded:
		return PushMessage{
			Title: "Добавлен новый урок 📚",
			Body:  fmt.Sprintf("В курсе %s новый урок: %s 📚", e.CourseName, e.LessonName),
			Data: map[string]string{
				"type":      "lesson_added",
				"course_id": e.CourseID,
				"lesson_id": e.LessonID,
			},
		}


	case EventCourseUpdated:
		return PushMessage{
			Title: "Курс обновлен! 🔄",
			Body:  fmt.Sprintf("Курс %s обновлен 🔄", e.CourseName),
			Data: map[string]string{
				"type":      "course_updated",
				"course_id": e.CourseID,
			},
		}


	case EventModuleAdded:
		return PushMessage{
			Title: "Добавлен новый модуль 📚",
			Body:  fmt.Sprintf("В курсе %s добавлен новый модуль: %s", e.CourseName, e.ModuleName),
			Data: map[string]string{
				"type":      "module_added",
				"course_id": e.CourseID,
				"module_id": e.ModuleID,
			},
		}

	case EventModuleUpdated:
		return PushMessage{
			Title: "Модуль обновлен 🔄",
			Body:  fmt.Sprintf("Модуль %s в курсе %s обновлен", e.ModuleName, e.CourseName),
			Data: map[string]string{
				"type":      "module_updated",
				"course_id": e.CourseID,
				"module_id": e.ModuleID,
			},
		}

	case EventLessonUpdated:
		return PushMessage{
			Title: "Урок обновлен 🔄",
			Body:  fmt.Sprintf("Урок %s в курсе %s обновлен", e.LessonName, e.CourseName),
			Data: map[string]string{
				"type":      "lesson_updated",
				"course_id": e.CourseID,
				"lesson_id": e.LessonID,
			},
		}

	case EventSourceAdded:
		return PushMessage{
			Title: "Добавлен новый материал 📚",
			Body:  fmt.Sprintf("В уроке %s курса %s добавлен новый материал", e.LessonName, e.CourseName),
			Data: map[string]string{
				"type":      "source_added",
				"course_id": e.CourseID,
				"lesson_id": e.LessonID,
			},
		}

	case EventSourceUpdated:
		return PushMessage{
			Title: "Материал обновлен 🔄",
			Body:  fmt.Sprintf("Материал в уроке %s курса %s обновлен", e.LessonName, e.CourseName),
			Data: map[string]string{
				"type":      "source_updated",
				"course_id": e.CourseID,
				"lesson_id": e.LessonID,
			},
		}

	case EventCourseAssigned:
		return PushMessage{
			Title: "Курс назначен ✅",
			Body:  fmt.Sprintf("Вам назначен курс: %s ✅", e.CourseName),
			Data: map[string]string{
				"type":      "course_assigned",
				"course_id": e.CourseID,
			},
		}


	case EventPaymentSuccess:
		return PushMessage{
			Title: "Оплата принята ✅",
			Body:  fmt.Sprintf("Оплата прошла успешно. Курс: %s", e.CourseName),
			Data: map[string]string{
				"type":      "payment_success",
				"order_id":  e.OrderID,
				"course_id": e.CourseID,
				"amount":    e.Amount,
			},
		}

	case EventPaymentFailed:
		body := fmt.Sprintf("Оплата не принята. Курс: %s", e.CourseName)
		if e.FailReason != "" {
			body = body + ". Причина: " + e.FailReason
		}
		return PushMessage{
			Title: "Оплата отклонена ❌",
			Body:  body,
			Data: map[string]string{
				"type":      "payment_failed",
				"order_id":  e.OrderID,
				"course_id": e.CourseID,
				"amount":    e.Amount,
			},
		}

	case EventAdminNotification:
		return PushMessage{
			Title: e.Title,
			Body:  e.Body,
			Data: map[string]string{
				"type": "admin_notification",
				"info": e.Body,
			},
		}

	case EventBannerAdded:
		return PushMessage{
			Title: e.Title,
			Body:  e.Body,
			Data: map[string]string{
				"type": "banner_added",
			},
		}

	case EventSubjectAdded:
		return PushMessage{
			Title: e.Title,
			Body:  fmt.Sprintf("Добавлен новый предмет: %s 📖", e.Body),
			Data: map[string]string{
				"type": "subject_added",
			},
		}


	case EventSubjectUpdated:
		return PushMessage{
			Title: e.Title,
			Body:  fmt.Sprintf("Предмет обновлен: %s 🔄", e.Body),
			Data: map[string]string{
				"type": "subject_updated",
			},
		}





	default:
		return PushMessage{
			Title: "Уведомление",
			Body:  "Есть новое сообщение",
			Data:  map[string]string{"type": "unknown"},
		}
	}
}

func BuildEnPushMessage(e PushEvent) PushMessage {
	switch e.Type {

	case EventNewCourse:
		return PushMessage{
			Title: "New Course! 🔥",
			Body:  fmt.Sprintf("New course added: %s 🔥", e.CourseName),
			Data: map[string]string{
				"type":      "new_course",
				"course_id": e.CourseID,
			},
		}


	case EventLessonAdded:
		return PushMessage{
			Title: "New lesson added 📚",
			Body:  fmt.Sprintf("New lesson in %s course: %s 📚", e.CourseName, e.LessonName),
			Data: map[string]string{
				"type":      "lesson_added",
				"course_id": e.CourseID,
				"lesson_id": e.LessonID,
			},
		}


	case EventCourseUpdated:
		return PushMessage{
			Title: "Course updated! 🔄",
			Body:  fmt.Sprintf("Course %s updated", e.CourseName),
			Data: map[string]string{
				"type":      "course_updated",
				"course_id": e.CourseID,
			},
		}

	case EventModuleAdded:
		return PushMessage{
			Title: "New module added 📚",
			Body:  fmt.Sprintf("New module in %s course: %s", e.CourseName, e.ModuleName),
			Data: map[string]string{
				"type":      "module_added",
				"course_id": e.CourseID,
				"module_id": e.ModuleID,
			},
		}

	case EventModuleUpdated:
		return PushMessage{
			Title: "Module updated 🔄",
			Body:  fmt.Sprintf("Module %s in course %s updated", e.ModuleName, e.CourseName),
			Data: map[string]string{
				"type":      "module_updated",
				"course_id": e.CourseID,
				"module_id": e.ModuleID,
			},
		}

	case EventLessonUpdated:
		return PushMessage{
			Title: "Lesson updated 🔄",
			Body:  fmt.Sprintf("Lesson %s in course %s updated", e.LessonName, e.CourseName),
			Data: map[string]string{
				"type":      "lesson_updated",
				"course_id": e.CourseID,
				"lesson_id": e.LessonID,
			},
		}

	case EventSourceAdded:
		return PushMessage{
			Title: "New material added 📚",
			Body:  fmt.Sprintf("New material in lesson %s of course %s added", e.LessonName, e.CourseName),
			Data: map[string]string{
				"type":      "source_added",
				"course_id": e.CourseID,
				"lesson_id": e.LessonID,
			},
		}

	case EventSourceUpdated:
		return PushMessage{
			Title: "Material updated 🔄",
			Body:  fmt.Sprintf("Material in lesson %s of course %s updated", e.LessonName, e.CourseName),
			Data: map[string]string{
				"type":      "source_updated",
				"course_id": e.CourseID,
				"lesson_id": e.LessonID,
			},
		}

	case EventCourseAssigned:
		return PushMessage{
			Title: "Course assigned ✅",
			Body:  fmt.Sprintf("You have been assigned to course: %s ✅", e.CourseName),
			Data: map[string]string{
				"type":      "course_assigned",
				"course_id": e.CourseID,
			},
		}


	case EventPaymentSuccess:
		return PushMessage{
			Title: "Payment accepted ✅",
			Body:  fmt.Sprintf("Payment successful for course: %s", e.CourseName),
			Data: map[string]string{
				"type":      "payment_success",
				"order_id":  e.OrderID,
				"course_id": e.CourseID,
				"amount":    e.Amount,
			},
		}

	case EventPaymentFailed:
		body := fmt.Sprintf("Payment failed for course: %s", e.CourseName)
		if e.FailReason != "" {
			body = body + ". Reason: " + e.FailReason
		}
		return PushMessage{
			Title: "Payment failed ❌",
			Body:  body,
			Data: map[string]string{
				"type":      "payment_failed",
				"order_id":  e.OrderID,
				"course_id": e.CourseID,
				"amount":    e.Amount,
			},
		}

	case EventAdminNotification:
		return PushMessage{
			Title: e.Title,
			Body:  e.Body,
			Data: map[string]string{
				"type": "admin_notification",
				"info": e.Body,
			},
		}

	case EventBannerAdded:
		return PushMessage{
			Title: e.Title,
			Body:  e.Body,
			Data: map[string]string{
				"type": "banner_added",
			},
		}

	case EventSubjectAdded:
		return PushMessage{
			Title: e.Title,
			Body:  fmt.Sprintf("New subject added: %s 📖", e.Body),
			Data: map[string]string{
				"type": "subject_added",
			},
		}

	case EventSubjectUpdated:
		return PushMessage{
			Title: e.Title,
			Body:  fmt.Sprintf("Subject updated: %s 🔄", e.Body),
			Data: map[string]string{
				"type": "subject_updated",
			},
		}





	default:
		return PushMessage{
			Title: "Notification",
			Body:  "New message",
			Data:  map[string]string{"type": "unknown"},
		}
	}
}

