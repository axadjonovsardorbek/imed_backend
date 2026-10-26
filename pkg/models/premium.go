package models

const (
	OrderTypeCourse  = "course"
	OrderTypePremium = "premium"
)

type PremiumSettings struct {
	Price          int `json:"price"`
	DurationMonths int `json:"duration_months"`
}

type PremiumOrderBody struct {
	PromoCodeId string `json:"promocode_id"`
}

type PremiumOrderReq struct {
	UserId      string
	PromoCodeId string
}

// AccessStatus foydalanuvchining premium va ustoz holati (T-12, T-15, T-16).
type AccessStatus struct {
	IsPremium    bool    `json:"is_premium"`
	PremiumUntil *string `json:"premium_until"`
	IsUstoz      bool    `json:"is_ustoz"`
}

type PremiumInfoRes struct {
	PremiumSettings
	AccessStatus
}

type UstozBody struct {
	IsUstoz bool `json:"is_ustoz"`
}

type UstozReq struct {
	UserId  string
	IsUstoz bool
}
