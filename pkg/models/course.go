package models

type CourseCreateReq struct {
	SubjectId   string `json:"subject_id"`
	TeacherId   string `json:"teacher_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       string `json:"price"`
	ImageUrl    string `json:"image_url"`
	IsPublic    bool   `json:"is_public"`
	OrderNum    int    `json:"order_num"`
}

type CourseCreateBody struct {
	SubjectId   string            `json:"subject_id"`
	TeacherId   string            `json:"teacher_id"`
	Name        map[string]string `json:"name"`
	Description map[string]string `json:"description"`
	Price       map[string]string `json:"price"`
	ImageUrl    string            `json:"image_url"`
	IsPublic    bool              `json:"is_public"`
	OrderNum    int               `json:"order_num"`
}

type CourseRes struct {
	Id          string            `json:"id"`
	Name        map[string]string `json:"name"`
	Description map[string]string `json:"description"`
	Subject     SubjectRes        `json:"subject"`
	Teacher     TeacherRes        `json:"teacher"`
	Price       map[string]string `json:"price"`
	ImageUrl    *string           `json:"image_url"`
	IsPublic    *bool             `json:"is_public"`
	OrderNum    *int              `json:"order_num"`
	HasAccess   bool              `json:"has_access"`
	CreatedAt   *string           `json:"created_at"`
	UpdatedAt   *string           `json:"updated_at"`
}

type CourseList struct {
	Courses []*CourseRes `json:"courses"`
	Count   int          `json:"count"`
}

type CourseGetListReq struct {
	SubjectId string     `json:"subject_id"`
	TeacherId string     `json:"teacher_id"`
	Name      string     `json:"name"`
	IsPublic  string     `json:"is_public"`
	Price     int        `json:"price"`
	UserId    string     `json:"user_id"`
	Filter    Pagination `json:"filter"`
}

type CourseUpdateBody struct {
	Name        map[string]string `json:"name"`
	Description map[string]string `json:"description"`
	SubjectId   string            `json:"subject_id"`
	TeacherId   string            `json:"teacher_id"`
	Price       map[string]string `json:"price"`
	ImageUrl    string            `json:"image_url"`
	IsPublic    bool              `json:"is_public"`
	OrderNum    int               `json:"order_num"`
}

type CourseUpdateReq struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	SubjectId   string `json:"subject_id"`
	TeacherId   string `json:"teacher_id"`
	Price       string `json:"price"`
	ImageUrl    string `json:"image_url"`
	IsPublic    bool   `json:"is_public"`
	OrderNum    int    `json:"order_num"`
}

type UserCourseCreateReq struct {
	UserId    string `json:"user_id"`
	CourseId  string `json:"course_id"`
	Type      string `json:"type"`
	StartedAt string `json:"started_at"`
	EndedAt   string `json:"ended_at"`
}

type UserCourseRes struct {
	Id         *string `json:"id"`
	UserId     *string `json:"user_id"`
	UserName   *string `json:"user_name"`
	CourseId   *string `json:"course_id"`
	CourseName *string `json:"course_name"`
	IsActive   *bool   `json:"is_active"`
	Type       *string `json:"type"`
	StartedAt  *string `json:"started_at"`
	EndedAt    *string `json:"ended_at"`
	CreatedAt  *string `json:"created_at"`
	UpdatedAt  *string `json:"updated_at"`
}

type UserCourseList struct {
	UserCourses []*UserCourseRes `json:"user_courses"`
	Count       int              `json:"count"`
}

type UserCourseGetListReq struct {
	UserId   string     `json:"user_id"`
	CourseId string     `json:"course_id"`
	Type     string     `json:"type"`
	IsActive string     `json:"is_active"`
	Filter   Pagination `json:"filter"`
}

type UserCourseUpdateBody struct {
	UserId    string `json:"user_id"`
	CourseId  string `json:"course_id"`
	Type      string `json:"type"`
	StartedAt string `json:"started_at"`
	EndedAt   string `json:"ended_at"`
}

type UserCourseUpdateReq struct {
	Id        string `json:"id"`
	UserId    string `json:"user_id"`
	CourseId  string `json:"course_id"`
	Type      string `json:"type"`
	StartedAt string `json:"started_at"`
	EndedAt   string `json:"ended_at"`
}
