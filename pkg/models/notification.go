package models

// CREATE TABLE IF NOT EXISTS master_notifications (
//     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
//     title JSONB NOT NULL,
//     message JSONB NOT NULL,
//     type VARCHAR(50) NOT NULL CHECK (type IN ('selected', 'all')),
//     course_id UUID REFERENCES courses(id) DEFAULT NULL,
//     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
//     updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
// );

// CREATE TABLE IF NOT EXISTS user_notifications (
//     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
//     master_notification_id UUID NOT NULL REFERENCES master_notifications(id) ON DELETE CASCADE,
//     user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
//     is_read BOOLEAN DEFAULT false,
//     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
//     UNIQUE(master_notification_id, user_id)
// );

type NotificationCreateBody struct {
	Title    map[string]string `json:"title"`
	Message  map[string]string `json:"message"`
	Type     string            `json:"type"`
	CourseId string            `json:"course_id"`
}

type NotificationCreateReq struct {
	Title    string `json:"title"`
	Message  string `json:"message"`
	Type     string `json:"type"`
	CourseId string `json:"course_id"`
}

type MasterNotificationRes struct {
	Id        *string           `json:"id"`
	Title     map[string]string `json:"title"`
	Message   map[string]string `json:"message"`
	Type      *string           `json:"type"`
	Course    *CourseRes        `json:"course"`
	CreatedAt *string           `json:"created_at"`
	UpdatedAt *string           `json:"updated_at"`
}

type MasterNotificationList struct {
	Notifications []*MasterNotificationRes `json:"notifications"`
	Count         int                      `json:"count"`
}

type UserNotificationRes struct {
	Id                 *string                `json:"id"`
	MasterNotification *MasterNotificationRes `json:"master_notification"`
	IsRead             *bool                  `json:"is_read"`
	CreatedAt          *string                `json:"created_at"`
	UpdatedAt          *string                `json:"updated_at"`
}

type UserNotificationList struct {
	Notifications []*UserNotificationRes `json:"notifications"`
	Count         int                    `json:"count"`
}

type NotificationGetListReq struct {
	CourseId string     `json:"course_id"`
	UserId   string     `json:"user_id"`
	Title    string     `json:"title"`
	IsRead   string     `json:"is_read"`
	Filter   Pagination `json:"filter"`
}

type NotificationMarkReadReq struct {
	Id     string `json:"id"`
	UserId string `json:"user_id"`
}

type MasterNotificationUpdateBody struct {
	Title   map[string]string `json:"title"`
	Message map[string]string `json:"message"`
}

type MasterNotificationUpdateReq struct {
	Id      string `json:"id"`
	Title   string `json:"title"`
	Message string `json:"message"`
}
