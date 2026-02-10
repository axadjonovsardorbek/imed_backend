package models

type DashboardRes struct {
	Tariffs   int `json:"tariffs"`
	Subjects  int `json:"subjects"`
	Courses   int `json:"courses"`
	Modules   int `json:"modules"`
	Lessons   int `json:"lessons"`
	Tests     int `json:"tests"`
	Documents int `json:"documents"`
	Videos    int `json:"videos"`
	Users     int `json:"users"`
	Teachers  int `json:"teachers"`
}

type GetDashboardReq struct {
	Type string `json:"type"` //range, day, week, month, year
	Day  string `json:"day"`  //2024-06-01
	From string `json:"from"` //2024-06-01
	To   string `json:"to"`   //2024-06-30
}
