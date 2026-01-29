package models

type OrderCreateReq struct {
	UserId   string `json:"user_id"`
	CourseId string `json:"course_id"`
	TariffId string `json:"tariff_id"`
}

type OrderCreateBody struct {
	CourseId string `json:"course_id"`
	TariffId string `json:"tariff_id"`
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
