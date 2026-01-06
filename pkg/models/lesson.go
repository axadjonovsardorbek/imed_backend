package models

type LessonCreateReq struct {
	ModuleId string `json:"module_id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Duration int    `json:"duration"`
	OrderNum int    `json:"order_num"`
}

type LessonCreateBody struct {
	ModuleId string            `json:"module_id"`
	Name     map[string]string `json:"name"`
	Type     string            `json:"type"`
	Duration int               `json:"duration"`
	OrderNum int               `json:"order_num"`
}

type LessonRes struct {
	Id        string            `json:"id"`
	Name      map[string]string `json:"name"`
	Module    ModuleRes         `json:"module"`
	Type      *string           `json:"type"`
	Duration  *int              `json:"duration"`
	OrderNum  *int              `json:"order_num"`
	CreatedAt *string           `json:"created_at"`
	UpdatedAt *string           `json:"updated_at"`
}

type LessonList struct {
	Lessons []*LessonRes `json:"lessons"`
	Count   int          `json:"count"`
}

type LessonGetListReq struct {
	ModuleId string     `json:"module_id"`
	Name     string     `json:"name"`
	Type     string     `json:"type"`
	Filter   Pagination `json:"filter"`
}

type LessonUpdateBody struct {
	Name     map[string]string `json:"name"`
	ModuleId string            `json:"module_id"`
	Type     string            `json:"type"`
	Duration int               `json:"duration"`
	OrderNum int               `json:"order_num"`
}

type LessonUpdateReq struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	ModuleId string `json:"module_id"`
	Type     string `json:"type"`
	Duration int    `json:"duration"`
	OrderNum int    `json:"order_num"`
}

type EndUserLessonReq struct {
	UserId   string `json:"user_id"`
	LessonId string `json:"lesson_id"`
}
