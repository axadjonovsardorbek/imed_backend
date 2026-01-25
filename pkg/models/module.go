package models

type ModuleCreateReq struct {
	CourseId string `json:"course_id"`
	Name     string `json:"name"`
	OrderNum int    `json:"order_num"`
}

type ModuleCreateBody struct {
	CourseId string            `json:"course_id"`
	Name     map[string]string `json:"name"`
	OrderNum int               `json:"order_num"`
}

type ModuleRes struct {
	Id         string            `json:"id"`
	Name       map[string]string `json:"name"`
	CourseId   *string           `json:"course_id"`
	CourseName map[string]string `json:"course_name"`
	OrderNum   *int              `json:"order_num"`
	CreatedAt  *string           `json:"created_at"`
	UpdatedAt  *string           `json:"updated_at"`
}

type ModuleList struct {
	Modules []*ModuleRes `json:"modules"`
	Count   int          `json:"count"`
}

type ModuleGetListReq struct {
	CourseId string     `json:"course_id"`
	Name     string     `json:"name"`
	Filter   Pagination `json:"filter"`
}

type ModuleUpdateBody struct {
	Name     map[string]string `json:"name"`
	CourseId string            `json:"course_id"`
	OrderNum int               `json:"order_num"`
}

type ModuleUpdateReq struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	CourseId string `json:"course_id"`
	OrderNum int    `json:"order_num"`
}

// type ModuleMobileRes struct {
// 	Id         string  `json:"id"`
// 	Name       *string `json:"name"`
// 	CourseId   *string `json:"course_id"`
// 	CourseName *string `json:"course_name"`
// 	OrderNum   *int    `json:"order_num"`
// 	CreatedAt  *string `json:"created_at"`
// 	UpdatedAt  *string `json:"updated_at"`
// }

// type ModuleMobileList struct {
// 	Modules []*ModuleRes `json:"modules"`
// 	Count   int          `json:"count"`
// }

// type ModuleGetListMobileReq struct {
// 	CourseId string     `json:"course_id"`
// 	Name     string     `json:"name"`
// 	Language string `json:"language"`
// }
