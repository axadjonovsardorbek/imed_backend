package models

type OrderStatisticRes struct {
	Id             string  `json:"id"`
	UserId         *string `json:"user_id"`
	UserName       *string `json:"user_name"`
	UserPhone      *string `json:"user_phone"`
	CourseId       *string `json:"course_id"`
	CourseName     *string `json:"course_name"`
	TariffId       *string `json:"tariff_id"`
	TariffName     *string `json:"tariff_name"`
	Duration       int     `json:"duration"`
	PromoCodeId    *string `json:"promocode_id"`
	PromoCode      *string `json:"promocode"`
	Amount         float64 `json:"amount"`
	DiscountAmount float64 `json:"discount_amount"`
	CourseAmount   float64 `json:"course_amount"`
	PaymentType    string  `json:"payment_type"`
	Status         string  `json:"status"`
	CreatedAt      string  `json:"created_at"`
	PaidAt         *string `json:"paid_at,omitempty"`
}

type OrderList struct {
	Count  int                  `json:"count"`
	Orders []*OrderStatisticRes `json:"orders"`
}

type OrderGetListReq struct {
	UserId      string     `json:"user_id"`
	CourseId    string     `json:"course_id"`
	TariffId    string     `json:"tariff_id"`
	PromoCodeId string     `json:"promocode_id"`
	Status      string     `json:"status"`
	PaymentType string     `json:"payment_type"`
	Type        string     `json:"type"` //range, day, week, month, year
	Day         string     `json:"day"`  //2024-06-01
	From        string     `json:"from"` //2024-06-01
	To          string     `json:"to"`   //2024-06-30
	Filter      Pagination `json:"filter"`
}

type OrderCreateReq struct {
	UserId      string `json:"user_id"`
	CourseId    string `json:"course_id"`
	PromoCodeId string `json:"promocode_id"`
	TariffId    string `json:"tariff_id"`
}

type OrderCreateBody struct {
	CourseId    string `json:"course_id"`
	TariffId    string `json:"tariff_id"`
	PromoCodeId string `json:"promocode_id"`
}

type OrderCreateRes struct {
	Id  string `json:"id"`
	Url string `json:"url"`
}

type OrderRes struct {
	Id                string  `json:"id"`
	UserId            string  `json:"user_id"`
	CourseId          string  `json:"course_id"`
	TariffId          string  `json:"tariff_id"`
	Duration          int     `json:"duration"`
	Amount            float64 `json:"amount"`
	Status            string  `json:"status"`
	MerchantPrepareId *int64  `json:"merchant_prepare_id,omitempty"`
	PaidAt            *string `json:"paid_at,omitempty"`
	Meta              *string `json:"meta,omitempty"` // json string ko'rinishida qaytarsang ham bo'ladi
}

type PayURLRes struct {
	OrderId string `json:"order_id"`
	PayURL  string `json:"pay_url"`
}

type ApplyPrepareParams struct {
	OrderId           string
	ClickTransId      int64
	ClickPaydocId     int64
	MerchantPrepareId int64
	Amount            float64
	MetaPatchJSON     []byte // jsonb patch (raw payload)
}

type PrepareResult struct {
	Error             int
	ErrorNote         string
	MerchantPrepareId int64
}

type ApplyCompleteParams struct {
	OrderId           string
	ClickTransId      int64
	MerchantPrepareId int64
	ErrorCode         int
	MetaPatchJSON     []byte
}

type CompleteResult struct {
	Error             int
	ErrorNote         string
	MerchantConfirmId int64
}
