package models

type AdminRes struct {
	Id        string `json:"id"`
	Login     string `json:"login"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
