package models

type LessonCreateReq struct {
	ModuleId string `json:"module_id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Duration int    `json:"duration"`
	IsPublic bool   `json:"is_public"`
	OrderNum int    `json:"order_num"`
}

type LessonCreateBody struct {
	ModuleId string            `json:"module_id"`
	Name     map[string]string `json:"name"`
	Type     string            `json:"type"`
	Duration int               `json:"duration"`
	IsPublic bool              `json:"is_public"`
	OrderNum int               `json:"order_num"`
}

type LessonCourseInfo struct {
	CourseId   string `json:"course_id"`
	CourseName string `json:"course_name"`
	LessonId   string `json:"lesson_id"`
	LessonName string `json:"lesson_name"`
}

type LessonRes struct {
	Id         string            `json:"id"`
	Name       map[string]string `json:"name"`
	ModuleId   *string           `json:"module_id"`
	ModuleName map[string]string `json:"module_name"`
	Type       *string           `json:"type"`
	Duration   *int              `json:"duration"`
	IsPublic   bool              `json:"is_public"`
	OrderNum   *int              `json:"order_num"`
	CreatedAt  *string           `json:"created_at"`
	UpdatedAt  *string           `json:"updated_at"`
}

type LessonList struct {
	Lessons []*LessonRes `json:"lessons"`
	Count   int          `json:"count"`
}

type LessonGetListReq struct {
	ModuleId string     `json:"module_id"`
	Name     string     `json:"name"`
	Type     string     `json:"type"`
	IsPublic string     `json:"is_public"`
	Filter   Pagination `json:"filter"`
}

type LessonUpdateBody struct {
	Name     map[string]string `json:"name"`
	ModuleId string            `json:"module_id"`
	Type     string            `json:"type"`
	Duration int               `json:"duration"`
	IsPublic bool              `json:"is_public"`
	OrderNum int               `json:"order_num"`
}

type LessonUpdateReq struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	ModuleId string `json:"module_id"`
	Type     string `json:"type"`
	Duration int    `json:"duration"`
	IsPublic bool   `json:"is_public"`
	OrderNum int    `json:"order_num"`
}

type EndUserLessonReq struct {
	UserId   string `json:"user_id"`
	LessonId string `json:"lesson_id"`
}

type LessonMobileRes struct {
	Id          string  `json:"id"`
	Name        *string `json:"name"`
	ModuleId    *string `json:"module_id"`
	ModuleName  *string `json:"module_name"`
	Type        *string `json:"type"`
	Duration    *int    `json:"duration"`
	OrderNum    *int    `json:"order_num"`
	IsCompleted bool    `json:"is_completed"`
	CreatedAt   *string `json:"created_at"`
	UpdatedAt   *string `json:"updated_at"`
}

type LessonMobileList struct {
	Lessons []*LessonRes `json:"lessons"`
	Count   int          `json:"count"`
}

type LessonGetListMobileReq struct {
	ModuleId string `json:"module_id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	UserId   string `json:"user_id"`
	Language string `json:"language"`
}

type SourceLessonMobileRes struct {
	Id          string         `json:"id"`
	Name        *string        `json:"name"`
	Type        *string        `json:"type"`
	Duration    *int           `json:"duration"`
	OrderNum    *int           `json:"order_num"`
	IsCompleted bool           `json:"is_completed"`
	Tests       []SourceMobile `json:"tests"`
	Documents   []SourceMobile `json:"documents"`
	Videos      []SourceMobile `json:"videos"`
}

type SourceMobile struct {
	Id       string  `json:"id"`
	Name     *string `json:"name"`
	Url      *string `json:"url"`
	Type     *string `json:"type"`
	OrderNum *int    `json:"order_num"`
}
