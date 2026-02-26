package models

type SubjectCreateReq struct {
	Name     string `json:"name"`
	ImageUrl string `json:"image_url"`
	OrderNum int    `json:"order_num"`
}

type SubjectCreateBody struct {
	Name     map[string]string `json:"name"`
	ImageUrl string            `json:"image_url"`
	OrderNum int               `json:"order_num"`
}

type SubjectRes struct {
	Id           *string           `json:"id"`
	Name         map[string]string `json:"name"`
	ImageUrl     *string           `json:"image_url"`
	OrderNum     *int              `json:"order_num"`
	TotalCourses *int              `json:"total_courses"`
	CreatedAt    *string           `json:"created_at"`
	UpdatedAt    *string           `json:"updated_at"`
}

type SubjectList struct {
	Subjects []*SubjectRes `json:"subjects"`
	Count    int           `json:"count"`
}

type SubjectGetListReq struct {
	Name   string     `json:"name"`
	Filter Pagination `json:"filter"`
}

type SubjectUpdateBody struct {
	Name     map[string]string `json:"name"`
	ImageUrl string            `json:"image_url"`
	OrderNum int               `json:"order_num"`
}

type SubjectUpdateReq struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	ImageUrl string `json:"image_url"`
	OrderNum int    `json:"order_num"`
}

type SubjectMobileRes struct {
	Id           *string `json:"id"`
	Name         *string `json:"name"`
	ImageUrl     *string `json:"image_url"`
	OrderNum     *int    `json:"order_num"`
	TotalCourses *int    `json:"total_courses"`
	CreatedAt    *string `json:"created_at"`
	UpdatedAt    *string `json:"updated_at"`
}

type SubjectMobileList struct {
	Subjects []*SubjectMobileRes `json:"subjects"`
	Count    int                 `json:"count"`
}

type SubjectGetListMobileReq struct {
	Name     string `json:"name"`
	Language string `json:"language"`
}
