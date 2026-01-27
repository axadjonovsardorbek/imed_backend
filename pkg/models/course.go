package models

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type CoursePriceOption struct {
	Price    int `json:"price"`
	Duration int `json:"duration"`
}

type CourseCreateReq struct {
	SubjectId   string     `json:"subject_id"`
	TeacherId   string     `json:"teacher_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Price       PriceField `json:"price"`
	ImageUrl    string     `json:"image_url"`
	IsPublic    bool       `json:"is_public"`
	OrderNum    int        `json:"order_num"`
}

type CourseCreateBody struct {
	SubjectId   string            `json:"subject_id"`
	TeacherId   string            `json:"teacher_id"`
	Name        map[string]string `json:"name"`
	Description map[string]string `json:"description"`
	Price       PriceField        `json:"price"`
	ImageUrl    string            `json:"image_url"`
	IsPublic    bool              `json:"is_public"`
	OrderNum    int               `json:"order_num"`
}

type CourseRes struct {
	Id                 string            `json:"id"`
	Name               map[string]string `json:"name"`
	Description        map[string]string `json:"description"`
	SubjectId          string            `json:"subject_id"`
	SubjectName        map[string]string `json:"subject_name"`
	TeacherId          string            `json:"teacher_id"`
	TeacherName        string            `json:"teacher_name"`
	TeacherPhoneNumber string            `json:"teacher_phone_number"`
	Price              PriceField        `db:"price" json:"price"`
	ImageUrl           *string           `json:"image_url"`
	IsPublic           *bool             `json:"is_public"`
	OrderNum           *int              `json:"order_num"`
	CreatedAt          *string           `json:"created_at"`
	UpdatedAt          *string           `json:"updated_at"`
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
	Role      string     `json:"role"`
	Filter    Pagination `json:"filter"`
}

type CourseUpdateBody struct {
	Name        map[string]string `json:"name"`
	Description map[string]string `json:"description"`
	SubjectId   string            `json:"subject_id"`
	TeacherId   string            `json:"teacher_id"`
	Price       PriceField        `json:"price"`
	ImageUrl    string            `json:"image_url"`
	IsPublic    bool              `json:"is_public"`
	OrderNum    int               `json:"order_num"`
}

type CourseUpdateReq struct {
	Id          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	SubjectId   string     `json:"subject_id"`
	TeacherId   string     `json:"teacher_id"`
	Price       PriceField `json:"price"`
	ImageUrl    string     `json:"image_url"`
	IsPublic    bool       `json:"is_public"`
	OrderNum    int        `json:"order_num"`
}

type UserCourseCreateReq struct {
	UserId    string `json:"user_id"`
	CourseId  string `json:"course_id"`
	Type      string `json:"type"`
	StartedAt string `json:"started_at"`
	EndedAt   string `json:"ended_at"`
}

type UserCourseRes struct {
	Id         *string           `json:"id"`
	UserId     *string           `json:"user_id"`
	UserName   *string           `json:"user_name"`
	CourseId   *string           `json:"course_id"`
	CourseName map[string]string `json:"course_name"`
	IsActive   *bool             `json:"is_active"`
	Type       *string           `json:"type"`
	StartedAt  *string           `json:"started_at"`
	EndedAt    *string           `json:"ended_at"`
	CreatedAt  *string           `json:"created_at"`
	UpdatedAt  *string           `json:"updated_at"`
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

type CourseMobileRes struct {
	Id          string  `json:"id"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	TeacherId   string  `json:"teacher_id"`
	TeacherName string  `json:"teacher_name"`
	ImageUrl    *string `json:"image_url"`
	IsPublic    *bool   `json:"is_public"`
	OrderNum    *int    `json:"order_num"`
}

type CourseMobileList struct {
	Courses []*CourseMobileRes `json:"courses"`
	Count   int                `json:"count"`
}

type CourseGetListMobileReq struct {
	SubjectId string `json:"subject_id"`
	Name      string `json:"name"`
	IsPublic  string `json:"is_public"`
	UserId    string `json:"user_id"`
	Language  string `json:"language"`
}

type UserCourseMobileRes struct {
	Id               *string `json:"id"`
	UserId           *string `json:"user_id"`
	UserName         *string `json:"user_name"`
	CourseId         *string `json:"course_id"`
	CourseName       *string `json:"course_name"`
	CourseImageUrl   *string `json:"course_image_url"`
	IsActive         *bool   `json:"is_active"`
	Type             *string `json:"type"`
	StartedAt        *string `json:"started_at"`
	EndedAt          *string `json:"ended_at"`
	CreatedAt        *string `json:"created_at"`
	UpdatedAt        *string `json:"updated_at"`
	TotalLessons     *int    `json:"total_lessons"`
	CompletedLessons *int    `json:"completed_lessons"`
}

type UserCourseMobileList struct {
	UserCourses []*UserCourseMobileRes `json:"user_courses"`
	Count       int                    `json:"count"`
}

type UserCourseGetListMobileReq struct {
	UserId   string `json:"user_id"`
	Type     string `json:"type"`
	IsActive string `json:"is_active"`
	Language string `json:"language"`
}

type MobileCourseRes struct {
	Id          string                  `json:"id"`
	Name        *string                 `json:"name"`
	Description *string                 `json:"description"`
	ImageUrl    *string                 `json:"image_url"`
	TeacherId   string                  `json:"teacher_id"`
	TeacherName *string                 `json:"teacher_name"`
	HasAccess   bool                    `json:"has_access"`
	Price       PriceField              `json:"price"`
	Tests       int                     `json:"tests"`
	Documents   int                     `json:"documents"`
	Duration    int                     `json:"duration"`
	Lessons     int                     `json:"lessons"`
	Modules     []MobileCourseModuleRes `json:"modules"`
}

type MobileCourseModuleRes struct {
	Id       string                  `json:"id"`
	Name     *string                 `json:"name"`
	OrderNum *int                    `json:"order_num"`
	CourseId *string                 `json:"course_id"`
	Lessons  []MobileCourseLessonRes `json:"lessons"`
}

type MobileCourseLessonRes struct {
	Id       string  `json:"id"`
	Name     *string `json:"name"`
	ModuleId *string `json:"module_id"`
	OrderNum *int    `json:"order_num"`
}

type PriceField []CoursePriceOption

func (p *PriceField) Scan(src any) error {
	if src == nil {
		*p = nil
		return nil
	}

	var b []byte
	switch v := src.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("PriceField.Scan: unsupported type %T", src)
	}

	b = bytes.TrimSpace(b)
	if len(b) == 0 || bytes.Equal(b, []byte("null")) {
		*p = nil
		return nil
	}

	switch b[0] {
	case '[':
		var arr []CoursePriceOption
		if err := json.Unmarshal(b, &arr); err != nil {
			return fmt.Errorf("PriceField.Scan array unmarshal: %w", err)
		}
		*p = arr
		return nil

	case '{':
		var one CoursePriceOption
		if err := json.Unmarshal(b, &one); err != nil {
			return fmt.Errorf("PriceField.Scan object unmarshal: %w", err)
		}
		*p = []CoursePriceOption{one}
		return nil

	default:
		return fmt.Errorf("PriceField.Scan: invalid json (must be array/object): %s", string(b))
	}
}

// Value Go -> DB (INSERT/UPDATE)
func (p PriceField) Value() (driver.Value, error) {
	if p == nil {
		return []byte("[]"), nil
	}
	j, err := json.Marshal([]CoursePriceOption(p))
	if err != nil {
		return nil, fmt.Errorf("PriceField.Value marshal: %w", err)
	}
	return j, nil
}
