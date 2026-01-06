package models

type TeacherCreateReq struct {
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	Login       string `json:"login"`
	Password    string `json:"password"`
}

type TeacherRes struct {
	Id          *string `json:"id"`
	Name        string  `json:"name"`
	PhoneNumber string  `json:"phone_number"`
	Login       string  `json:"login"`
	CreatedAt   *string `json:"created_at"`
	UpdatedAt   *string `json:"updated_at"`
}

type TeacherList struct {
	Teachers []*TeacherRes `json:"teachers"`
	Count    int           `json:"count"`
}

type TeacherGetListReq struct {
	Name        string     `json:"name"`
	PhoneNumber string     `json:"phone_number"`
	Filter      Pagination `json:"filter"`
}

type TeacherUpdateBody struct {
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
}

type TeacherUpdateReq struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
}
