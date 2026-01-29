// storage/postgres/order_repo.go
package postgres

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"imed/internal/config"
	m "imed/pkg/models"

	"github.com/google/uuid"
)

type OrderRepo struct {
	db *sql.DB
}

func NewOrderRepo(db *sql.DB) *OrderRepo {
	return &OrderRepo{db: db}
}

// func round2(v float64) float64 { return math.Round(v*100) / 100 }

func (r *OrderRepo) Create(req *m.OrderCreateReq) (*m.OrderCreateRes, error) {
	tz, err := r.db.Begin()
	if err != nil {
		return nil, err
	}

	defer func() {
		if p := recover(); p != nil {
			tz.Rollback()
			err = fmt.Errorf("panic occurred: %v", p)
		} else if err != nil {
			tz.Rollback()
		} else {
			err = tz.Commit()
		}
	}()
	id := uuid.NewString()

	meta := map[string]any{
		"created_from": "api",
	}
	metaBytes, _ := json.Marshal(meta)

	var duration, price int
	var amount float64

	getTariffQuery := `
	SELECT
	  t.duration,
	  (opt->>'price')::bigint AS price
	FROM tariffs t
	JOIN courses c ON c.id = $1
	CROSS JOIN LATERAL jsonb_array_elements(c.price) AS opt
	WHERE t.id = $2
	  AND t.deleted_at = 0
	  AND c.deleted_at = 0
	  AND (opt->>'duration')::bigint = t.duration
	LIMIT 1
	`

	err = tz.QueryRow(getTariffQuery, req.CourseId, req.TariffId).Scan(&duration, &price)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("price option not found for course=%s tariff=%s", req.CourseId, req.TariffId)
		}
		return nil, err
	}

	query := `
	INSERT INTO orders (
		id,
		user_id,
		course_id,
		tariff_id,
		duration,
		amount,
		meta
	) VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb)
	RETURNING amount
	`
	err = tz.QueryRow(query,
		id,
		req.UserId,
		req.CourseId,
		req.TariffId,
		duration,
		price,
		string(metaBytes),
	).Scan(&amount)

	if err != nil {
		return nil, fmt.Errorf("error while creating order: %w", err)
	}

	cfg := config.Load()

	payURL := fmt.Sprintf(
		"https://my.click.uz/services/pay?merchant_id=%d&service_id=%d&amount=%.2f&transaction_param=%s&return_url=%s",
		cfg.CLICK_MERCHANT_ID,
		cfg.CLICK_SERVICE_ID,
		amount,
		id,
		url.QueryEscape("https://youtube.com"),
	)

	return &m.OrderCreateRes{Id: id, Url: payURL}, nil
}

func (r *OrderRepo) GetById(req *m.ById) (*m.OrderRes, error) {
	query := `
	SELECT
	  id::text,
	  user_id::text,
	  course_id::text,
	  tariff_id::text,
	  duration,
	  amount::float8,
	  status::text,
	  merchant_prepare_id,
	  paid_at,
	  meta::text
	FROM orders
	WHERE id=$1::uuid
	`
	var (
		id, userId, courseId, tariff_id, status, metaText string
		duration                                          int
		amount                                            float64
		merchantPrepareID                                 sql.NullInt64
		paidAt                                            sql.NullTime
	)

	err := r.db.QueryRow(query, req.Id).Scan(
		&id, &userId, &courseId, tariff_id, &duration, &amount, &status,
		&merchantPrepareID, &paidAt, &metaText,
	)
	if err != nil {
		return nil, fmt.Errorf("error while getting order: %w", err)
	}

	res := &m.OrderRes{
		Id:       id,
		UserId:   userId,
		CourseId: courseId,
		TariffId: tariff_id,
		Duration: duration,
		Amount:   amount,
		Status:   status,
	}
	if merchantPrepareID.Valid {
		v := merchantPrepareID.Int64
		res.MerchantPrepareId = &v
	}
	if paidAt.Valid {
		s := paidAt.Time.Format(time.RFC3339)
		res.PaidAt = &s
	}
	res.Meta = &metaText
	return res, nil
}

func (r *OrderRepo) GetPayURL(req *m.ById) (*m.PayURLRes, error) {
	// Pay URL build qilishni repo emas, service/handlerda qilish ham mumkin,
	// lekin sen "repo qsmini" deding — amount/status ni olib qaytaramiz.
	// Pay URL ni handlerda yasaymiz. Shu yerda minimal qilib beraman:
	return nil, fmt.Errorf("implement in handler/service (need config: merchant_id, service_id, base_url)")
}
