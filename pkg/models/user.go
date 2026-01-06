package models

type UserCheckReq struct {
	PhoneNumber string `json:"phone_number"`
}

type UserLoginReq struct {
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
	Name        string `json:"name"`
}

type UserRes struct {
	Id          string  `json:"id"`
	Name        *string `json:"name"`
	PhoneNumber *string `json:"phone_number"`
	FCMToken    *string `json:"fcm_token"`
	ImageUrl    *string `json:"image_url"`
	Language    *string `json:"language"`
	CreatedAt   *string `json:"created_at"`
	UpdatedAt   *string `json:"updated_at"`
}

type UserUpdateReq struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	ImageUrl    string `json:"image_url"`
	FCMToken    string `json:"fcm_token"`
	Language    string `json:"language"`
}

type UserUpdateBody struct {
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	ImageUrl    string `json:"image_url"`
	FCMToken    string `json:"fcm_token"`
	Language    string `json:"language"`
}

type UserGetListReq struct {
	Filter      Pagination `json:"filter"`
	Name        string     `json:"name"`
	PhoneNumber string     `json:"phone_number"`
}

type UserList struct {
	Users []UserRes `json:"users"`
	Count int       `json:"count"`
}
