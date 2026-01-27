package models

type AppRouteRes struct {
	Id          string  `json:"id" db:"id"`
	Key         string  `json:"key" db:"key"`
	Value       JSONRaw `json:"value" db:"value" swaggertype:"object"`
	Description string  `json:"description,omitempty" db:"description"`
	OrderNum    int     `json:"order_num" db:"order_num"`
	CreatedAt   string  `json:"created_at" db:"created_at"`
	UpdatedAt   string  `json:"updated_at" db:"updated_at"`
}

type AppRouteCreateReq struct {
	Key         string  `json:"key" binding:"required"`
	Value       JSONRaw `json:"value" binding:"required" swaggertype:"object"`
	Description string  `json:"description" binding:"omitempty"`
	OrderNum    int     `json:"order_num" binding:"omitempty"`
}

type AppRouteUpdateReq struct {
	Id          string  `json:"id" binding:"required"`
	Key         string  `json:"key" binding:"required"`
	Value       JSONRaw `json:"value" binding:"required" swaggertype:"object"`
	Description string  `json:"description" binding:"omitempty"`
	OrderNum    int     `json:"order_num" binding:"omitempty"`
}

type AppRouteUpdateBody struct {
	Key         string  `json:"key"`
	Value       JSONRaw `json:"value" swaggertype:"object"`
	Description string  `json:"description"`
	OrderNum    int     `json:"order_num"`
}

type AppRouteGetListReq struct {
	Key string `json:"key"`
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
