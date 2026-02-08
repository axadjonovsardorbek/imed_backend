package models

type PromoCodeCreateReq struct {
	Code           string  `json:"code"`
	DiscountType   string  `json:"discount_type"`
	DiscountValue  float64 `json:"discount_value"`
	StartsAt       string  `json:"starts_at"`
	EndsAt         string  `json:"ends_at"`
	MinOrderAmount float64 `json:"min_order_amount"`
	MaxDiscount    float64 `json:"max_discount"`
	MaxUsesTotal   int     `json:"max_uses_total"`
	MaxUsesPerUser int     `json:"max_uses_per_user"`
}

type PromoCodeRes struct {
	Id             *string  `json:"id"`
	IsActive       *bool    `json:"is_active"`
	Code           *string  `json:"code"`
	DiscountType   *string  `json:"discount_type"`
	DiscountValue  *float64 `json:"discount_value"`
	StartsAt       *string  `json:"starts_at"`
	EndsAt         *string  `json:"ends_at"`
	MinOrderAmount *float64 `json:"min_order_amount"`
	MaxDiscount    *float64 `json:"max_discount"`
	MaxUsesTotal   *int     `json:"max_uses_total"`
	MaxUsesPerUser *int     `json:"max_uses_per_user"`
	CreatedAt      *string  `json:"created_at"`
	UpdatedAt      *string  `json:"updated_at"`
}

type PromoCodeList struct {
	PromoCodes []*PromoCodeRes `json:"promocodes"`
	Count      int             `json:"count"`
}

type PromoCodeGetListReq struct {
	Code         string     `json:"code"`
	DiscountType string     `json:"discount_type"`
	IsActive     string     `json:"is_active"`
	Filter       Pagination `json:"filter"`
}

type PromoCodeUpdateBody struct {
	IsActive       bool    `json:"is_active"`
	DiscountType   string  `json:"discount_type"`
	DiscountValue  float64 `json:"discount_value"`
	StartsAt       string  `json:"starts_at"`
	EndsAt         string  `json:"ends_at"`
	MinOrderAmount float64 `json:"min_order_amount"`
	MaxDiscount    float64 `json:"max_discount"`
	MaxUsesTotal   int     `json:"max_uses_total"`
	MaxUsesPerUser int     `json:"max_uses_per_user"`
}

type PromoCodeUpdateReq struct {
	Id             string  `json:"id"`
	IsActive       bool    `json:"is_active"`
	DiscountType   string  `json:"discount_type"`
	DiscountValue  float64 `json:"discount_value"`
	StartsAt       string  `json:"starts_at"`
	EndsAt         string  `json:"ends_at"`
	MinOrderAmount float64 `json:"min_order_amount"`
	MaxDiscount    float64 `json:"max_discount"`
	MaxUsesTotal   int     `json:"max_uses_total"`
	MaxUsesPerUser int     `json:"max_uses_per_user"`
}

type PromoCodeCheckRes struct {
	IsValid        bool     `json:"is_valid"`
	Message        string   `json:"message"`
	DiscountType   *string  `json:"discount_type"`
	DiscountValue  *float64 `json:"discount_value"`
	TarriffId      *string  `json:"tariff_id"`
	CourseId       *string  `json:"course_id"`
	PromoCodeId    *string  `json:"promocode_id"`
	CourseAmount   *float64 `json:"course_amount"`
	DiscountAmount *float64 `json:"discount_amount"`
	TotalAmount    *float64 `json:"total_amount"`
}

type PromoCodeCheckReq struct {
	PromoCode string `json:"promocode"`
	CourseId  string `json:"course_id"`
	UserId    string `json:"user_id"`
	TariffId  string `json:"tariff_id"`
}

type PromoCodeCheckBody struct {
	PromoCode string `json:"promocode"`
	CourseId  string `json:"course_id"`
	TariffId  string `json:"tariff_id"`
}

type PromoRedemptionRes struct {
	Id             *string  `json:"id"`
	PromoId        *string  `json:"promo_id"`
	PromoCode      *string  `json:"promocode"`
	UserId         *string  `json:"user_id"`
	UserName       *string  `json:"user_name"`
	UserPhone      *string  `json:"user_phone"`
	OrderId        *string  `json:"order_id"`
	OrderAmount    *float64 `json:"order_amount"`
	DiscountAmount *float64 `json:"discount_amount"`
	CourseAmount   *float64 `json:"course_amount"`
	IsActive       *bool    `json:"is_active"`
	CreatedAt      *string  `json:"created_at"`
	DeactivatedAt  *string  `json:"deactivated_at"`
}

type PromoRedemptionList struct {
	Redemptions []*PromoRedemptionRes `json:"redemptions"`
	Count       int                   `json:"count"`
}

type PromoRedemptionGetListReq struct {
	PromoId  string     `json:"promo_id"`
	UserId   string     `json:"user_id"`
	IsActive string     `json:"is_active"`
	Filter   Pagination `json:"filter"`
}

// CREATE TABLE IF NOT EXISTS promo_redemptions (
// 	id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
// 	promo_id        UUID NOT NULL REFERENCES promocodes(id),
// 	user_id         UUID NOT NULL REFERENCES users(id),
// 	order_id        UUID NOT NULL REFERENCES orders(id),
// 	order_amount    NUMERIC(12,2) NOT NULL,
// 	discount_amount NUMERIC(12,2) NOT NULL,
// 	is_active       BOOLEAN NOT NULL DEFAULT TRUE,
// 	created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
// 	deactivated_at  TIMESTAMP WITH TIME ZONE DEFAULT NULL
//   );
