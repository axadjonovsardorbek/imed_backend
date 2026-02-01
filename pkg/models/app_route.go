package models

type AppVersionRes struct {
	Android string `json:"android" db:"android"`
	IOS     string `json:"ios" db:"ios"`
}

type AppRouteRes struct {
	Id         string        `json:"id" db:"id"`
	CallCenter string        `json:"call_center" db:"call_center"`
	SupportUrl string        `json:"support_url" db:"support_url"`
	AppVersion AppVersionRes `json:"app_version" db:"app_version" swaggertype:"object"`
	CreatedAt  string        `json:"created_at" db:"created_at"`
	UpdatedAt  string        `json:"updated_at" db:"updated_at"`
}

type AppRouteCreateReq struct {
	CallCenter string        `json:"call_center" binding:"required"`
	SupportUrl string        `json:"support_url" binding:"required"`
	AppVersion AppVersionRes `json:"app_version" binding:"required" swaggertype:"object"`
}

type AppRouteUpdateReq struct {
	Id         string        `json:"id" binding:"required"`
	CallCenter string        `json:"call_center" binding:"required"`
	SupportUrl string        `json:"support_url" binding:"required"`
	AppVersion AppVersionRes `json:"app_version" binding:"required" swaggertype:"object"`
}

type AppRouteUpdateBody struct {
	CallCenter string        `json:"call_center" binding:"required"`
	SupportUrl string        `json:"support_url" binding:"required"`
	AppVersion AppVersionRes `json:"app_version" binding:"required" swaggertype:"object"`
}

type AppRouteGetListReq struct {
}

type AppRouteList struct {
	AppRoutes []*AppRouteRes `json:"app_routes"`
	Count     int            `json:"count"`
}

type JSONRaw []byte

func (j JSONRaw) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte(`{}`), nil
	}
	return j, nil
}

func (j *JSONRaw) UnmarshalJSON(b []byte) error {
	*j = append((*j)[:0], b...)
	return nil
}
