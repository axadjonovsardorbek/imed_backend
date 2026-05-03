package models

type SourceCreateReq struct {
	LessonId string `json:"lesson_id"`
	Name     string `json:"name"`
	Url      string `json:"url"`
	VideoUrl string `json:"video_url"`
	Type     string `json:"type"`
	OrderNum int    `json:"order_num"`
}

type SourceCreateBody struct {
	LessonId string            `json:"lesson_id"`
	Name     map[string]string `json:"name"`
	Url      map[string]string `json:"url"`
	Type     string            `json:"type"`
	OrderNum int               `json:"order_num"`
}

type SourceRes struct {
	Id         string            `json:"id"`
	Name       map[string]string `json:"name"`
	LessonId   string            `json:"lesson_id"`
	LessonName map[string]string `json:"lesson_name"`
	Url        map[string]string `json:"url"`
	VideoUrl   map[string]string `json:"video_url"`
	Type       *string           `json:"type"`
	OrderNum   *int              `json:"order_num"`
	CreatedAt  *string           `json:"created_at"`
	UpdatedAt  *string           `json:"updated_at"`
}

type SourceList struct {
	Sources []*SourceRes `json:"source"`
	Count   int          `json:"count"`
}

type SourceGetListReq struct {
	LessonId string     `json:"lesson_id"`
	Name     string     `json:"name"`
	Type     string     `json:"type"`
	Filter   Pagination `json:"filter"`
}

type SourceUpdateBody struct {
	Name     map[string]string `json:"name"`
	LessonId string            `json:"lesson_id"`
	Url      map[string]string `json:"url"`
	VideoUrl map[string]string `json:"video_url"`
	Type     string            `json:"type"`
	OrderNum int               `json:"order_num"`
}

type SourceUpdateReq struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	LessonId string `json:"lesson_id"`
	Url      string `json:"url"`
	VideoUrl string `json:"video_url"`
	Type     string `json:"type"`
	OrderNum int    `json:"order_num"`
}

// type SourceMobileRes struct {
// 	Id         string  `json:"id"`
// 	Name       *string `json:"name"`
// 	LessonId   string  `json:"lesson_id"`
// 	LessonName *string `json:"lesson_name"`
// 	Url        *string `json:"url"`
// 	Type       *string `json:"type"`
// 	OrderNum   *int    `json:"order_num"`
// 	CreatedAt  *string `json:"created_at"`
// 	UpdatedAt  *string `json:"updated_at"`
// }

// type SourceMobileList struct {
// 	Sources []*SourceRes `json:"source"`
// 	Count   int          `json:"count"`
// }

// type SourceGetListMobileReq struct {
// 	LessonId string `json:"lesson_id"`
// 	Name     string `json:"name"`
// 	Type     string `json:"type"`
// 	Language string `json:"language"`
// }
