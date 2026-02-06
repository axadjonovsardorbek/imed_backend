package models

// CREATE TABLE IF NOT EXISTS promocodes (
//     id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
//     code              TEXT NOT NULL,

//     is_active         BOOLEAN NOT NULL DEFAULT TRUE,

//     discount_type     promo_discount_type NOT NULL,
//     discount_value    NUMERIC(12,2) NOT NULL CHECK (discount_value > 0),
//     CHECK (
//       (discount_type = 'percent' AND discount_value <= 100)
//       OR (discount_type = 'fixed')
//     ),

//     starts_at         TIMESTAMPTZ,
//     ends_at           TIMESTAMPTZ,

//     min_order_amount  NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (min_order_amount >= 0),
//     max_discount      NUMERIC(12,2),

//     max_uses_total    INT,
//     max_uses_per_user INT NOT NULL DEFAULT 1,

//     created_at        TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
//     updated_at        TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
//     deleted_at        BIGINT NOT NULL DEFAULT 0
// );

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
	PromoCodes []*PromoCodeRes `json:"promo_codes"`
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
	PromoCodeId    *string  `json:"promo_code_id"`
	CourseAmount   *float64 `json:"course_amount"`
	DiscountAmount *float64 `json:"discount_amount"`
	TotalAmount    *float64 `json:"total_amount"`
}

type PromoCodeCheckReq struct {
	Code     string `json:"code"`
	CourseId string `json:"course_id"`
	UserId   string `json:"user_id"`
	TariffId string `json:"tariff_id"`
}

type PromoCodeCheckBody struct {
	Code     string `json:"code"`
	CourseId string `json:"course_id"`
	TariffId string `json:"tariff_id"`
}

type PromoRedemptionRes struct {
	Id             *string  `json:"id"`
	PromoId        *string  `json:"promo_id"`
	UserId         *string  `json:"user_id"`
	OrderId        *string  `json:"order_id"`
	OrderAmount    *float64 `json:"order_amount"`
	DiscountAmount *float64 `json:"discount_amount"`
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
