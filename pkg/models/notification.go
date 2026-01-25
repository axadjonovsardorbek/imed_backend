package models

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
	Id         *string           `json:"id"`
	Title      map[string]string `json:"title"`
	Message    map[string]string `json:"message"`
	Type       *string           `json:"type"`
	CourseId   *string           `json:"course_id"`
	CourseName *string           `json:"course_name"`
	CreatedAt  *string           `json:"created_at"`
	UpdatedAt  *string           `json:"updated_at"`
}

type MasterNotificationList struct {
	Notifications []*MasterNotificationRes `json:"notifications"`
	Count         int                      `json:"count"`
}

type UserNotificationRes struct {
	Id        *string `json:"id"`
	Title     *string `json:"title"`
	Message   *string `json:"message"`
	IsRead    *bool   `json:"is_read"`
	CreatedAt *string `json:"created_at"`
	UpdatedAt *string `json:"updated_at"`
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
	Language string     `json:"language"`
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
