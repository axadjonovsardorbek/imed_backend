package models

type DashboardRes struct {
	Tariffs     int `json:"tariffs"`
	Subjects    int `json:"subjects"`
	Courses     int `json:"courses"`
	Modules     int `json:"modules"`
	Lessons     int `json:"lessons"`
	Tests       int `json:"tests"`
	Documents   int `json:"documents"`
	Videos      int `json:"videos"`
	Users       int `json:"users"`
	Teachers    int `json:"teachers"`
	ActiveUsers int `json:"active_users"`

	UserCourses         int `json:"user_courses"`
	ActiveUserCourses   int `json:"active_user_courses"`
	InactiveUserCourses int `json:"inactive_user_courses"`
	DeletedUserCourses  int `json:"deleted_user_courses"`

	PaidOrders      int `json:"paid_orders"`
	AdminOrders     int `json:"admin_orders"`
	ClickOrders     int `json:"click_orders"`
	CancelledOrders int `json:"cancelled_orders"`
	OrdersWithPromo int `json:"orders_with_promo"`

	PromoCodes       int `json:"promocodes"`
	ActivePromoCodes int `json:"active_promocodes"`

	PaidAmountTotal float64 `json:"paid_amount_total"`
	PaidAmountClick float64 `json:"paid_amount_click"`
	PaidAmountAdmin float64 `json:"paid_amount_admin"`

	PaidDiscountTotal float64 `json:"paid_discount_total"`
	PaidDiscountClick float64 `json:"paid_discount_click"`
	PaidDiscountAdmin float64 `json:"paid_discount_admin"`

	CourseAmountTotal float64 `json:"course_amount_total"`
	CourseAmountClick float64 `json:"course_amount_click"`
	CourseAmountAdmin float64 `json:"course_amount_admin"`
}

type GetDashboardReq struct {
	Type string `json:"type"` //range, day, week, month, year
	Day  string `json:"day"`  //2024-06-01
	From string `json:"from"` //2024-06-01
	To   string `json:"to"`   //2024-06-30
}
