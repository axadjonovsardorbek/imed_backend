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
)

type PushTarget struct {
	Type     PushTargetType
	UserID   string // TargetUser uchun
	CourseID string // TargetCourseUsers uchun
}

type PushEventType string

const (
	SourceTest     = "test"
	SourceDocument = "document"
	SourceVideo    = "video"
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
)

type PushEvent struct {
	Type PushEventType

	// umumiy
	CourseID   string
	CourseName string // UZ nom (name->>'uz')

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
			Body:  fmt.Sprintf("Yangi kurs qo‘shildi: %s", e.CourseName),
			Data: map[string]string{
				"type":      "new_course",
				"course_id": e.CourseID,
			},
		}

	case EventLessonAdded:
		return PushMessage{
			Title: "Yangi dars qo‘shildi 📚",
			Body:  fmt.Sprintf("%s kursiga yangi dars: %s", e.CourseName, e.LessonName),
			Data: map[string]string{
				"type":      "lesson_added",
				"course_id": e.CourseID,
				"lesson_id": e.LessonID,
			},
		}

	case EventCourseAssigned:
		return PushMessage{
			Title: "Kurs biriktirildi ✅",
			Body:  fmt.Sprintf("Sizga kurs biriktirildi: %s", e.CourseName),
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
			Title: e.CourseID,
			Body:  e.CourseName,
			Data: map[string]string{
				"type": "admin_notification",
				"info": e.CourseName,
			},
		}

	case EventModuleAdded:
		return PushMessage{
			Title: "Yangi modul qo‘shildi 📚",
			Body:  fmt.Sprintf("%s kursiga yangi '%s' modul qo‘shildi", e.CourseName, e.ModuleName),
		}

	case EventSourceAdded:
		var sourceName string

		switch e.SourceType {
		case SourceTest:
			sourceName = "test"
		case SourceDocument:
			sourceName = "hujjat"
		case SourceVideo:
			sourceName = "video"
		}

		return PushMessage{
			Title: "Yangi material qo‘shildi 📚",
			Body: fmt.Sprintf(
				"%s kursining %s darsiga yangi %s qo‘shildi",
				e.CourseName,
				e.LessonName,
				sourceName,
			),
		}

	default:
		return PushMessage{
			Title: "Bildirishnoma",
			Body:  "Yangi xabar bor",
			Data:  map[string]string{"type": "unknown"},
		}
	}
}
