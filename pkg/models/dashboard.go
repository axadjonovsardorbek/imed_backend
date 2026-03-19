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

	PromoCodes       int `json:"promocodes"`
	ActivePromoCodes int `json:"active_promocodes"`
}

type GetDashboardReq struct {
	Type string `json:"type"` //range, day, week, month, year
	Day  string `json:"day"`  //2024-06-01
	From string `json:"from"` //2024-06-01
	To   string `json:"to"`   //2024-06-30
}
