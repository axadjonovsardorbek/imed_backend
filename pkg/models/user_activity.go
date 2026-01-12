package models

type UserActivityCreateReq struct {
	UserId   string `json:"user_id"`
	Activity int64  `json:"activity"`
}

type UserActivityCreateBody struct {
	Activity int64 `json:"activity"`
}

type UserActivity struct {
	UserId    *string `json:"user_id"`
	Activity  *int64  `json:"activity"`
	Day       *string `json:"day"`
	UpdatedAt *string `json:"created_at"`
}

type UserActivityList struct {
	Activities []*UserActivity `json:"activities"`
	Count      int             `json:"count"`
}

type UserActivityGetListReq struct {
	UserId   string `json:"user_id"`
	Type     string `json:"type"`
	Date     string `json:"date"`
	FromDate string `json:"from_date"`
	ToDate   string `json:"to_date"`
}
