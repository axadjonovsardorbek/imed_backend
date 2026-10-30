package postgres

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
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

func (r *OrderRepo) CreateWithPromoCode(req *m.OrderCreateReq) (*m.OrderCreateRes, error) {
	cfg := config.Load()

	tz, err := r.db.Begin()
	if err != nil {
		return nil, err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tz.Rollback()
			err = fmt.Errorf("panic occurred: %v", p)
			return
		}
		if err != nil {
			_ = tz.Rollback()
			return
		}
		err = tz.Commit()
	}()

	if req.PromoCodeId != "" {

		var check_id string
		var check_amount float64

		check_query := `
		SELECT
			id,
			amount
		FROM
			orders
		WHERE
			user_id = $1
		AND
			course_id = $2
		AND
			tariff_id = $3 
		AND 
			promo_id = $4
		AND
			status IN ('NEW','EXPIRED')
		AND 
			created_at >= now() - interval '1 days'
		ORDER BY created_at DESC
		LIMIT 1
		`

		_ = tz.QueryRow(check_query, req.UserId, req.CourseId, req.TariffId, req.PromoCodeId).Scan(&check_id, &check_amount)

		if check_id != "" && check_amount > 0 {
			payURL := fmt.Sprintf(
				"https://my.click.uz/services/pay?merchant_id=%d&service_id=%d&amount=%.2f&transaction_param=%s",
				cfg.CLICK_MERCHANT_ID,
				cfg.CLICK_SERVICE_ID,
				check_amount,
				check_id,
			)

			return &m.OrderCreateRes{Id: check_id, Url: payURL}, nil
		}
	}

	id := uuid.NewString()

	meta := map[string]any{
		"created_from": "api",
	}
	metaBytes, _ := json.Marshal(meta)

	var duration int
	var price int

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

	var (
		promoID        *string
		discountAmount float64 = 0
		totalAmount    float64 = float64(price)
	)

	if req.PromoCodeId != "" {
		promoID, discountAmount, err = applyPromo(tz, req.PromoCodeId, req.UserId, req.CourseId, float64(price))
		if err != nil {
			return nil, err
		}
		totalAmount = float64(price) - discountAmount
		if totalAmount < 0 {
			totalAmount = 0
		}
	}

	var payAmount float64
	query := `
		INSERT INTO orders (
			id,
			user_id,
			course_id,
			tariff_id,
			duration,
			amount,
			promo_id,
			discount_amount,
			course_amount,
			meta
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::jsonb)
		RETURNING amount
		`
	err = tz.QueryRow(query,
		id,
		req.UserId,
		req.CourseId,
		req.TariffId,
		duration,
		totalAmount,
		promoID,
		discountAmount,
		float64(price), // course price
		string(metaBytes),
	).Scan(&payAmount)
	if err != nil {
		return nil, fmt.Errorf("error while creating order: %w", err)
	}

	if promoID != nil && discountAmount > 0 {
		_, err = tz.Exec(`
				INSERT INTO promo_redemptions (
					promo_id, user_id, order_id,
					order_amount, discount_amount, course_amount,
					is_active
				) VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5, $6, TRUE)
			`, *promoID, req.UserId, id, payAmount, discountAmount, float64(price))
		if err != nil {
			return nil, fmt.Errorf("error while creating promo redemption: %w", err)
		}
	}

	payURL := fmt.Sprintf(
		"https://my.click.uz/services/pay?merchant_id=%d&service_id=%d&amount=%.2f&transaction_param=%s",
		cfg.CLICK_MERCHANT_ID,
		cfg.CLICK_SERVICE_ID,
		payAmount, // total_amount
		id,
	)

	return &m.OrderCreateRes{Id: id, Url: payURL}, nil
}

func (r *OrderRepo) Create(req *m.OrderCreateReq) (*m.OrderCreateRes, error) {
	cfg := config.Load()
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

	var check_id string
	var check_amount float64

	check_query := `
	SELECT
		id,
		amount
	FROM
		orders
	WHERE
		user_id = $1
	AND
		course_id = $2
	AND
		tariff_id = $3 
	AND 
		status IN ('NEW','EXPIRED')
	AND 
		created_at >= now() - interval '1 days'
	ORDER BY created_at DESC
	LIMIT 1
	`

	_ = tz.QueryRow(check_query, req.UserId, req.CourseId, req.TariffId).Scan(&check_id, &check_amount)

	if check_id != "" && check_amount > 0 {
		payURL := fmt.Sprintf(
			"https://my.click.uz/services/pay?merchant_id=%d&service_id=%d&amount=%.2f&transaction_param=%s",
			cfg.CLICK_MERCHANT_ID,
			cfg.CLICK_SERVICE_ID,
			check_amount,
			check_id,
		)

		return &m.OrderCreateRes{Id: check_id, Url: payURL}, nil
	}

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
		discount_amount,
		course_amount,
		meta
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb)
	RETURNING amount
	`
	err = tz.QueryRow(query,
		id,
		req.UserId,
		req.CourseId,
		req.TariffId,
		duration,
		price,
		0,
		price,
		string(metaBytes),
	).Scan(&amount)

	if err != nil {
		return nil, fmt.Errorf("error while creating order: %w", err)
	}

	payURL := fmt.Sprintf(
		"https://my.click.uz/services/pay?merchant_id=%d&service_id=%d&amount=%.2f&transaction_param=%s",
		cfg.CLICK_MERCHANT_ID,
		cfg.CLICK_SERVICE_ID,
		amount,
		id,
	)

	return &m.OrderCreateRes{Id: id, Url: payURL}, nil
}

func (r *OrderRepo) GetById(req *m.ById) (*m.OrderStatisticRes, error) {
	query := `
	SELECT
		o.id, 
		o.user_id,
		u.name,
		u.phone_number,
		o.course_id,
		c.name ->> 'uz' AS course_name,
		o.tariff_id,
		t.name AS tariff_name,
		o.duration,
		o.promo_id,
		p.code AS promocode,
		o.amount,
		o.discount_amount,
		o.course_amount,
		o.payment_type,
		o.type,
		o.package,
		o.status,
		o.comment,
		o.check_url,
		o.created_at,
		o.paid_at
	FROM orders o
	JOIN users u ON u.id = o.user_id
	LEFT JOIN courses c ON c.id = o.course_id
	LEFT JOIN tariffs t ON t.id = o.tariff_id
	LEFT JOIN promocodes p ON p.id = o.promo_id
	WHERE o.id = $1
	`

	row := r.db.QueryRow(query, req.Id)

	order := m.OrderStatisticRes{}
	err := row.Scan(
		&order.Id,
		&order.UserId,
		&order.UserName,
		&order.UserPhone,
		&order.CourseId,
		&order.CourseName,
		&order.TariffId,
		&order.TariffName,
		&order.Duration,
		&order.PromoCodeId,
		&order.PromoCode,
		&order.Amount,
		&order.DiscountAmount,
		&order.CourseAmount,
		&order.PaymentType,
		&order.OrderType,
		&order.Package,
		&order.Status,
		&order.Comment,
		&order.CheckUrl,
		&order.CreatedAt,
		&order.PaidAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("order not found with id: %s", req.Id)
	}

	if err != nil {
		return nil, fmt.Errorf("error while getting order by id: %w", err)
	}

	return &order, nil
}
func (r *OrderRepo) GetAll(req *m.OrderGetListReq) (*m.PaginationResponse, *m.OrderList, error) {

	query := `
	SELECT
		COUNT(o.id) OVER() AS total_count,
		o.id, 
		o.user_id,
		u.name,
		u.phone_number,
		o.course_id,
		c.name ->> 'uz' AS course_name,
		o.tariff_id,
		t.name AS tariff_name,
		o.duration,
		o.promo_id,
		p.code AS promocode,
		o.amount,
		o.discount_amount,
		o.course_amount,
		o.payment_type,
		o.type,
		o.package,
		o.status,
		o.comment,
		o.check_url,
		o.created_at,
		o.paid_at
	FROM orders o
	JOIN users u ON u.id = o.user_id
	LEFT JOIN courses c ON c.id = o.course_id
	LEFT JOIN tariffs t ON t.id = o.tariff_id
	LEFT JOIN promocodes p ON p.id = o.promo_id
	WHERE 1=1
	`

	var args []interface{}

	if req.UserId != "" && req.UserId != "string" {
		args = append(args, req.UserId)
		query += fmt.Sprintf(" AND o.user_id = $%d", len(args))
	}

	if req.CourseId != "" && req.CourseId != "string" {
		args = append(args, req.CourseId)
		query += fmt.Sprintf(" AND o.course_id = $%d", len(args))
	}

	if req.TariffId != "" && req.TariffId != "string" {
		args = append(args, req.TariffId)
		query += fmt.Sprintf(" AND o.tariff_id = $%d", len(args))
	}

	if req.PromoCodeId != "" && req.PromoCodeId != "string" {
		args = append(args, req.PromoCodeId)
		query += fmt.Sprintf(" AND o.promo_id = $%d", len(args))
	}

	if req.Status != "" && req.Status != "string" {
		args = append(args, req.Status)
		query += fmt.Sprintf(" AND o.status = $%d", len(args))
	}

	if req.PaymentType != "" && req.PaymentType != "string" {
		args = append(args, req.PaymentType)
		query += fmt.Sprintf(" AND o.payment_type = $%d", len(args))
	}

	if req.Type != "" && req.Type != "string" {

		loc, err := time.LoadLocation("Asia/Tashkent")
		if err != nil {
			slog.Warn("tzdata missing? fallback to fixed zone", "err", err)
			loc = time.FixedZone("Asia/Tashkent", 5*60*60)
		}

		parseDate := func(s string) (time.Time, error) {
			return time.ParseInLocation("2006-01-02", s, loc)
		}

		startOfDay := func(t time.Time) time.Time {
			return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
		}

		startOfWeekMonday := func(t time.Time) time.Time {
			t = startOfDay(t)
			wd := int(t.Weekday())
			if wd == 0 {
				wd = 7
			}
			return t.AddDate(0, 0, -(wd - 1))
		}
		now := time.Now().In(loc)

		switch req.Type {

		case "day":
			if req.Day == "" || req.Day == "string" {
				return nil, nil, fmt.Errorf("day is required for type=day")
			}
			d, err := parseDate(req.Day)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid day format, expected YYYY-MM-DD")
			}
			from := startOfDay(d)
			to := from.AddDate(0, 0, 1)

			args = append(args, from)
			query += fmt.Sprintf(" AND o.created_at >= $%d", len(args))

			args = append(args, to)
			query += fmt.Sprintf(" AND o.created_at < $%d", len(args))

		case "range":
			if req.From == "" || req.From == "string" || req.To == "" || req.To == "string" {
				return nil, nil, fmt.Errorf("from and to are required for type=range")
			}

			f, err := parseDate(req.From)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid from format, expected YYYY-MM-DD")
			}
			t, err := parseDate(req.To)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid to format, expected YYYY-MM-DD")
			}

			from := startOfDay(f)
			to := startOfDay(t).AddDate(0, 0, 1)

			args = append(args, from)
			query += fmt.Sprintf(" AND o.created_at >= $%d", len(args))

			args = append(args, to)
			query += fmt.Sprintf(" AND o.created_at < $%d", len(args))

		case "week":
			from := startOfWeekMonday(now)
			to := from.AddDate(0, 0, 7)

			args = append(args, from)
			query += fmt.Sprintf(" AND o.created_at >= $%d", len(args))

			args = append(args, to)
			query += fmt.Sprintf(" AND o.created_at < $%d", len(args))

		case "month":
			from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
			to := from.AddDate(0, 1, 0)

			args = append(args, from)
			query += fmt.Sprintf(" AND o.created_at >= $%d", len(args))

			args = append(args, to)
			query += fmt.Sprintf(" AND o.created_at < $%d", len(args))

		case "year":
			from := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, loc)
			to := from.AddDate(1, 0, 0)

			args = append(args, from)
			query += fmt.Sprintf(" AND o.created_at >= $%d", len(args))

			args = append(args, to)
			query += fmt.Sprintf(" AND o.created_at < $%d", len(args))

		default:
			return nil, nil, fmt.Errorf("invalid type: %s", req.Type)
		}
	}

	query += " ORDER BY o.created_at DESC"

	args = append(args, req.Filter.Limit, req.Filter.Page)
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)-1, len(args))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("error while getting courses: %w", err)
	}
	defer rows.Close()

	orders := m.OrderList{}

	for rows.Next() {
		order := &m.OrderStatisticRes{}
		var totalCount int

		err := rows.Scan(
			&totalCount,
			&order.Id,
			&order.UserId,
			&order.UserName,
			&order.UserPhone,
			&order.CourseId,
			&order.CourseName,
			&order.TariffId,
			&order.TariffName,
			&order.Duration,
			&order.PromoCodeId,
			&order.PromoCode,
			&order.Amount,
			&order.DiscountAmount,
			&order.CourseAmount,
			&order.PaymentType,
			&order.OrderType,
			&order.Package,
			&order.Status,
			&order.Comment,
			&order.CheckUrl,
			&order.CreatedAt,
			&order.PaidAt,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("error while scanning order: %w", err)
		}

		orders.Orders = append(orders.Orders, order)
		orders.Count = totalCount
	}

	response := m.NewPaginationResponse(req.Filter, orders.Count, orders.Orders)
	return &response, &orders, nil
}

func (r *OrderRepo) GetPayURL(req *m.ById) (*m.PayURLRes, error) {
	// Pay URL build qilishni repo emas, service/handlerda qilish ham mumkin,
	// lekin sen "repo qsmini" deding — amount/status ni olib qaytaramiz.
	// Pay URL ni handlerda yasaymiz. Shu yerda minimal qilib beraman:
	return nil, fmt.Errorf("implement in handler/service (need config: merchant_id, service_id, base_url)")
}

// applyPromo promokodni tekshiradi va chegirma miqdorini qaytaradi.
// courseId bo'sh bo'lsa (premium buyurtma), faqat "selected" bo'lmagan promokodlar o'tadi.
func applyPromo(tz *sql.Tx, promoCodeId, userId, courseId string, price float64) (*string, float64, error) {
	var err error
	type promoRow struct {
		ID            string
		IsActive      bool
		DiscountType  string
		DiscountValue float64
		MinAmount     float64
		MaxDiscount   sql.NullFloat64
		StartsAt      sql.NullTime
		EndsAt        sql.NullTime
		MaxUsesTotal  sql.NullInt64
		MaxPerUser    int
		pType         string
	}

	var p promoRow

	err = tz.QueryRow(`
			SELECT id, is_active, discount_type, discount_value,
				   min_order_amount, max_discount,
				   starts_at, ends_at,
				   max_uses_total, max_uses_per_user, type
			FROM promocodes
			WHERE id = $1
			  AND deleted_at = 0
			FOR UPDATE
		`, promoCodeId).Scan(
		&p.ID, &p.IsActive, &p.DiscountType, &p.DiscountValue,
		&p.MinAmount, &p.MaxDiscount,
		&p.StartsAt, &p.EndsAt,
		&p.MaxUsesTotal, &p.MaxPerUser, &p.pType,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, fmt.Errorf("promocode not found")
		}
		return nil, 0, err
	}

	// 🔥 COURSE MATCH CHECK
	if p.pType == "selected" {
		// Premium buyurtmada kurs yo'q: faqat barcha kurslarga amal qiladigan promokod ishlaydi.
		if courseId == "" {
			return nil, 0, fmt.Errorf(ErrPromoNotApplicable)
		}
		var exists bool

		err = tz.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM promocode_courses
			WHERE promocode_id = $1::uuid
			  AND course_id = $2::uuid
		)
		`, p.ID, courseId).Scan(&exists)

		if err != nil {
			return nil, 0, fmt.Errorf(ErrUnExpected)
		}

		if !exists {
			return nil, 0, fmt.Errorf(ErrPromoNotApplicable)
		}
	}

	now := time.Now()

	if p.IsActive && p.EndsAt.Valid && now.After(p.EndsAt.Time) {
		_, _ = tz.Exec(`
				UPDATE promocodes
				SET is_active = FALSE, updated_at = now()
				WHERE id = $1::uuid AND is_active = TRUE AND deleted_at = 0
			`, p.ID)
		return nil, 0, fmt.Errorf("promocode expired")
	}

	if !p.IsActive {
		return nil, 0, fmt.Errorf("promocode inactive")
	}
	if p.StartsAt.Valid && now.Before(p.StartsAt.Time) {
		return nil, 0, fmt.Errorf("promocode not started")
	}

	if price < p.MinAmount && p.MinAmount > 0 {
		return nil, 0, fmt.Errorf("order amount is too small for this promocode")
	}

	var usedByUser int
	err = tz.QueryRow(`
			SELECT COUNT(*)
			FROM promo_redemptions
			WHERE promo_id = $1::uuid
			  AND user_id  = $2::uuid
			  AND is_active = TRUE
		`, p.ID, userId).Scan(&usedByUser)
	if err != nil {
		return nil, 0, err
	}
	if usedByUser >= p.MaxPerUser && p.MaxPerUser > 0 {
		return nil, 0, fmt.Errorf("promocode usage limit reached for user")
	}

	if p.MaxUsesTotal.Valid {
		var usedTotal int
		err = tz.QueryRow(`
				SELECT COUNT(*)
				FROM promo_redemptions
				WHERE promo_id = $1::uuid
				  AND is_active = TRUE
			`, p.ID).Scan(&usedTotal)
		if err != nil {
			return nil, 0, err
		}

		if usedTotal >= int(p.MaxUsesTotal.Int64) && p.MaxUsesTotal.Int64 > 0 {
			_, _ = tz.Exec(`
					UPDATE promocodes
					SET is_active = FALSE, updated_at = now()
					WHERE id = $1::uuid AND is_active = TRUE AND deleted_at = 0
				`, p.ID)
			return nil, 0, fmt.Errorf("promocode usage limit reached")
		}
	}

	var md *float64
	if p.MaxDiscount.Valid {
		v := p.MaxDiscount.Float64
		md = &v
	}
	return &p.ID, ComputeDiscount(price, p.DiscountType, p.DiscountValue, md), nil
}
