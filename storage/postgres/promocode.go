package postgres

import (
	"database/sql"
	"fmt"
	m "imed/pkg/models"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type PromoCodeRepo struct {
	db *sql.DB
}

func NewPromoCodeRepo(db *sql.DB) *PromoCodeRepo {
	return &PromoCodeRepo{
		db: db,
	}
}

func (r *PromoCodeRepo) Create(req *m.PromoCodeCreateReq) (*m.ById, error) {
	id := uuid.NewString()
	query := `
	INSERT INTO promocodes (
		id,
		code,
		discount_type,
		discount_value,
		starts_at,
		ends_at,
		min_order_amount,
		max_discount,
		max_uses_total,
		max_uses_per_user
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.Exec(query,
		id,
		req.Code,
		req.DiscountType,
		req.DiscountValue,
		req.StartsAt,
		req.EndsAt,
		req.MinOrderAmount,
		req.MaxDiscount,
		req.MaxUsesTotal,
		req.MaxUsesPerUser,
	)

	if err != nil {
		return nil, fmt.Errorf("error while creating promocode: %w", err)
	}

	return &m.ById{Id: id}, nil
}
func (r *PromoCodeRepo) GetById(req *m.ById) (*m.PromoCodeRes, error) {
	query := `
	SELECT 
		id,
		code,
		is_active,
		discount_type,
		discount_value,
		starts_at,
		ends_at,
		min_order_amount,
		max_discount,
		max_uses_total,
		max_uses_per_user,
		created_at,
		updated_at
	FROM
		promocodes
	WHERE
		deleted_at = 0
	AND
		id = $1
	`

	row := r.db.QueryRow(query, req.Id)

	promocode := m.PromoCodeRes{}
	err := row.Scan(
		&promocode.Id,
		&promocode.Code,
		&promocode.IsActive,
		&promocode.DiscountType,
		&promocode.DiscountValue,
		&promocode.StartsAt,
		&promocode.EndsAt,
		&promocode.MinOrderAmount,
		&promocode.MaxDiscount,
		&promocode.MaxUsesTotal,
		&promocode.MaxUsesPerUser,
		&promocode.CreatedAt,
		&promocode.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("promocode not found with id: %s", req.Id)
	}

	if err != nil {
		return nil, fmt.Errorf("error while getting promocode by id: %w", err)
	}

	return &promocode, nil
}
func (r *PromoCodeRepo) GetAll(req *m.PromoCodeGetListReq) (*m.PaginationResponse, *m.PromoCodeList, error) {
	query := `
	SELECT 
		COUNT(id) OVER() AS total_count,
		id,
		code,
		is_active,
		discount_type,
		discount_value,
		starts_at,
		ends_at,
		min_order_amount,
		max_discount,
		max_uses_total,
		max_uses_per_user,
		created_at,
		updated_at
	FROM
		promocodes
	WHERE
		deleted_at = 0
	`

	var args []interface{}
	if req.Code != "" && req.Code != "string" {
		query += fmt.Sprintf(" AND code ILIKE $%d", len(args)+1)
		args = append(args, "%"+req.Code+"%")
	}
	if req.DiscountType != "" && req.DiscountType != "string" {
		query += fmt.Sprintf(" AND discount_type = $%d", len(args)+1)
		args = append(args, req.DiscountType)
	}
	if req.IsActive != "" && req.IsActive != "string" {
		if req.IsActive == "true" {
			query += fmt.Sprintf(" AND is_active = TRUE")
		} else if req.IsActive == "false" {
			query += fmt.Sprintf(" AND is_active = FALSE")
		}
	}

	query += " ORDER BY created_at DESC"

	args = append(args, req.Filter.Limit, req.Filter.Page)
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)-1, len(args))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("error while getting promocode: %w", err)
	}
	defer rows.Close()

	promocodes := m.PromoCodeList{}

	for rows.Next() {

		promocode := m.PromoCodeRes{}

		var totalCount int

		err := rows.Scan(
			&totalCount,
			&promocode.Id,
			&promocode.Code,
			&promocode.IsActive,
			&promocode.DiscountType,
			&promocode.DiscountValue,
			&promocode.StartsAt,
			&promocode.EndsAt,
			&promocode.MinOrderAmount,
			&promocode.MaxDiscount,
			&promocode.MaxUsesTotal,
			&promocode.MaxUsesPerUser,
			&promocode.CreatedAt,
			&promocode.UpdatedAt,
		)

		if err != nil {
			return nil, nil, fmt.Errorf("error while scanning promocode row: %w", err)
		}

		promocodes.PromoCodes = append(promocodes.PromoCodes, &promocode)
		promocodes.Count = totalCount

	}

	response := m.NewPaginationResponse(req.Filter, promocodes.Count, promocodes.PromoCodes)

	return &response, &promocodes, nil
}
func (r *PromoCodeRepo) Update(req *m.PromoCodeUpdateReq) (*m.Void, error) {
	query := `
	UPDATE promocodes
	SET
	`

	var conditions []string
	var args []interface{}

	conditions = append(conditions, " is_active = $"+strconv.Itoa(len(args)+1))
	args = append(args, req.IsActive)

	if req.DiscountType != "" && req.DiscountType != "string" {
		conditions = append(conditions, " discount_type = $"+strconv.Itoa(len(args)+1))
		args = append(args, req.DiscountType)
	}

	if req.DiscountValue > 0 {
		conditions = append(conditions, " discount_value = $"+strconv.Itoa(len(args)+1))
		args = append(args, req.DiscountValue)
	}

	if req.EndsAt != "" && req.EndsAt != "string" {
		conditions = append(conditions, " ends_at = $"+strconv.Itoa(len(args)+1))
		args = append(args, req.EndsAt)
	}

	if req.StartsAt != "" && req.StartsAt != "string" {
		conditions = append(conditions, " starts_at = $"+strconv.Itoa(len(args)+1))
		args = append(args, req.StartsAt)
	}

	if req.MinOrderAmount > 0 {
		conditions = append(conditions, " min_order_amount = $"+strconv.Itoa(len(args)+1))
		args = append(args, req.MinOrderAmount)
	}

	if req.MaxDiscount > 0 {
		conditions = append(conditions, " max_discount = $"+strconv.Itoa(len(args)+1))
		args = append(args, req.MaxDiscount)
	}

	if req.MaxUsesTotal > 0 {
		conditions = append(conditions, " max_uses_total = $"+strconv.Itoa(len(args)+1))
		args = append(args, req.MaxUsesTotal)
	}

	if req.MaxUsesPerUser > 0 {
		conditions = append(conditions, " max_uses_per_user = $"+strconv.Itoa(len(args)+1))
		args = append(args, req.MaxUsesPerUser)
	}

	if len(conditions) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	conditions = append(conditions, " updated_at = now()")
	query += strings.Join(conditions, ", ")
	query += " WHERE id = $" + strconv.Itoa(len(args)+1) + " AND deleted_at = 0"

	args = append(args, req.Id)

	res, err := r.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error while updating promocode: %w", err)
	}

	if r, err := res.RowsAffected(); r == 0 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("promocode with id %s couldnt update", req.Id)
	}
	return &m.Void{}, nil
}
func (r *PromoCodeRepo) Delete(req *m.ById) (*m.Void, error) {
	query := `
	UPDATE promocodes
	SET
		deleted_at = EXTRACT(EPOCH FROM NOW())
	WHERE
		id = $1
	AND
		deleted_at = 0
	`

	res, err := r.db.Exec(query, req.Id)
	if err != nil {
		return nil, fmt.Errorf("error while deleting promocode: %w", err)
	}

	if r, err := res.RowsAffected(); r == 0 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("promocode with id %s couldnt delete", req.Id)
	}

	return &m.Void{}, nil
}

func (r *PromoCodeRepo) Check(req *m.PromoCodeCheckReq) (*m.PromoCodeCheckRes, error) {
	res := &m.PromoCodeCheckRes{
		IsValid: false,
	}

	var amount float64
	err := r.db.QueryRow(`
		SELECT (opt->>'price')::numeric(12,2)
		FROM courses c
		CROSS JOIN LATERAL jsonb_array_elements(c.price) AS opt
		WHERE c.deleted_at = 0
		  AND c.id = $1::uuid
		  AND (opt->>'tariff_id')::uuid = $2::uuid
		LIMIT 1
	`, req.CourseId, req.TariffId).Scan(&amount)
	if err == sql.ErrNoRows {
		res.Message = fmt.Sprintf("course or tariff not found")
		return res, nil
	}
	if err != nil {
		return nil, err
	}

	// 2) promo
	var (
		promoID        string
		isActive       bool
		discountType   string
		discountValue  float64
		minOrderAmount float64
		maxDiscount    sql.NullFloat64
		startsAt       sql.NullTime
		endsAt         sql.NullTime
		maxUsesTotal   sql.NullInt64
		maxUsesPerUser int
	)

	err = r.db.QueryRow(`
		SELECT id, is_active, discount_type, discount_value,
			   min_order_amount, max_discount,
			   starts_at, ends_at,
			   max_uses_total, max_uses_per_user
		FROM promocodes
		WHERE lower(code) = lower($1)
		  AND deleted_at = 0
		LIMIT 1
	`, req.PromoCode).Scan(
		&promoID, &isActive, &discountType, &discountValue,
		&minOrderAmount, &maxDiscount,
		&startsAt, &endsAt,
		&maxUsesTotal, &maxUsesPerUser,
	)
	if err == sql.ErrNoRows {
		res.Message = fmt.Sprintf("promocode not found")
		return res, nil
	}
	if err != nil {
		return nil, err
	}

	now := time.Now()

	// ✅ 3) auto-deactivate: expired bo'lsa DBda is_active=false qilib qo'yamiz
	if isActive && endsAt.Valid && now.After(endsAt.Time) {
		// best-effort update (agar parallel update bo'lsa ham mayli)
		_, _ = r.db.Exec(`
			UPDATE promocodes
			SET is_active = FALSE,
				updated_at = now()
			WHERE id = $1::uuid
			  AND is_active = TRUE
			  AND deleted_at = 0
		`, promoID)

		res.Message = fmt.Sprintf("promocode expired")
		return res, nil
	}

	// 3.1) validate promo (endidan oldin)
	if !isActive {
		res.Message = fmt.Sprintf("promocode inactive")
		return res, nil
	}
	if startsAt.Valid && now.Before(startsAt.Time) {
		res.Message = fmt.Sprintf("promocode not started")
		return res, nil
	}
	if amount < minOrderAmount {
		res.Message = fmt.Sprintf("order amount is too small for this promocode")
		return res, nil
	}

	// 4) per-user limit
	var usedByUser int
	err = r.db.QueryRow(`
		SELECT COUNT(*)
		FROM promo_redemptions
		WHERE promo_id = $1::uuid
		  AND user_id  = $2::uuid
		  AND is_active = TRUE
	`, promoID, req.UserId).Scan(&usedByUser)
	if err != nil {
		return nil, err
	}

	if usedByUser >= maxUsesPerUser {
		res.Message = fmt.Sprintf("promocode usage limit reached for user")
		return res, nil
	}

	// 5) total limit (optional) + ✅ auto-deactivate if full
	if maxUsesTotal.Valid {
		var usedTotal int
		err = r.db.QueryRow(`
			SELECT COUNT(*)
			FROM promo_redemptions
			WHERE promo_id = $1::uuid
			  AND is_active = TRUE
		`, promoID).Scan(&usedTotal)
		if err != nil {
			return nil, err
		}

		if usedTotal >= int(maxUsesTotal.Int64) {
			// ✅ limit to'lgan bo'lsa promo'ni inactive qilib qo'yamiz
			_, _ = r.db.Exec(`
				UPDATE promocodes
				SET is_active = FALSE,
					updated_at = now()
				WHERE id = $1::uuid
				  AND is_active = TRUE
				  AND deleted_at = 0
				  AND max_uses_total IS NOT NULL
			`, promoID)

			res.Message = fmt.Sprintf("promocode usage limit reached")
			return res, nil
		}
	}

	// 6) compute discount
	var md *float64
	if maxDiscount.Valid {
		v := maxDiscount.Float64
		md = &v
	}
	discount := ComputeDiscount(amount, discountType, discountValue, md)
	total := amount - discount

	// 7) response

	msg := "promocode applied successfully"
	res.IsValid = true
	res.Message = msg
	res.DiscountType = &discountType
	res.DiscountValue = &discountValue
	res.TarriffId = &req.TariffId
	res.CourseId = &req.CourseId
	res.PromoCodeId = &promoID
	res.CourseAmount = &amount
	res.DiscountAmount = &discount
	res.TotalAmount = &total

	return res, nil

}

func (r *PromoCodeRepo) GetByIdUserPromo(req *m.ById) (*m.PromoRedemptionRes, error) {
	query := `
	SELECT 
		r.id,
		r.promo_id,
		p.code,
		r.user_id,
		u.name,
		u.phone_number,
		r.order_id,
		r.order_amount,
		r.discount_amount,
		r.course_amount,
		r.is_active,
		r.created_at,
		r.deactivated_at
	FROM
		promo_redemptions r
	JOIN
		users u ON u.id = r.user_id
	JOIN
		promocodes p ON p.id = r.promo_id
	JOIN
		orders o ON o.id = r.order_id
	WHERE
		r.id = $1
	`

	row := r.db.QueryRow(query, req.Id)

	redemption := m.PromoRedemptionRes{}
	err := row.Scan(
		&redemption.Id,
		&redemption.PromoId,
		&redemption.PromoCode,
		&redemption.UserId,
		&redemption.UserName,
		&redemption.UserPhone,
		&redemption.OrderId,
		&redemption.OrderAmount,
		&redemption.DiscountAmount,
		&redemption.CourseAmount,
		&redemption.IsActive,
		&redemption.CreatedAt,
		&redemption.DeactivatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("promo redemption not found with id: %s", req.Id)
	}

	if err != nil {
		return nil, fmt.Errorf("error while getting promo redemption by id: %w", err)
	}

	return &redemption, nil
}
func (r *PromoCodeRepo) GetAllUserPromo(req *m.PromoRedemptionGetListReq) (*m.PaginationResponse, *m.PromoRedemptionList, error) {
	query := `
	SELECT 
		COUNT(r.id) OVER() AS total_count,
		r.id,
		r.promo_id,
		p.code,
		r.user_id,
		u.name,
		u.phone_number,
		r.order_id,
		r.order_amount,
		r.discount_amount,
		r.course_amount,
		r.is_active,
		r.created_at,
		r.deactivated_at
	FROM
		promo_redemptions r
	JOIN
		users u ON u.id = r.user_id
	JOIN
		promocodes p ON p.id = r.promo_id
	JOIN
		orders o ON o.id = r.order_id
	WHERE
		1=1
	`

	var args []interface{}
	if req.UserId != "" && req.UserId != "string" {
		query += fmt.Sprintf(" AND r.user_id = $%d", len(args)+1)
		args = append(args, req.UserId)
	}
	if req.PromoId != "" && req.PromoId != "string" {
		query += fmt.Sprintf(" AND r.promo_id = $%d", len(args)+1)
		args = append(args, req.PromoId)
	}
	if req.IsActive != "" && req.IsActive != "string" {
		if req.IsActive == "true" {
			query += fmt.Sprintf(" AND r.is_active = TRUE")
		} else if req.IsActive == "false" {
			query += fmt.Sprintf(" AND r.is_active = FALSE")
		}
	}

	query += " ORDER BY r.created_at DESC"

	args = append(args, req.Filter.Limit, req.Filter.Page)
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)-1, len(args))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("error while getting promocode: %w", err)
	}
	defer rows.Close()

	redemptions := m.PromoRedemptionList{}

	for rows.Next() {

		redemption := m.PromoRedemptionRes{}

		var totalCount int

		err := rows.Scan(
			&totalCount,
			&redemption.Id,
			&redemption.PromoId,
			&redemption.PromoCode,
			&redemption.UserId,
			&redemption.UserName,
			&redemption.UserPhone,
			&redemption.OrderId,
			&redemption.OrderAmount,
			&redemption.DiscountAmount,
			&redemption.CourseAmount,
			&redemption.IsActive,
			&redemption.CreatedAt,
			&redemption.DeactivatedAt,
		)

		if err != nil {
			return nil, nil, fmt.Errorf("error while scanning promocode row: %w", err)
		}

		redemptions.Redemptions = append(redemptions.Redemptions, &redemption)
		redemptions.Count = totalCount

	}

	response := m.NewPaginationResponse(req.Filter, redemptions.Count, redemptions.Redemptions)

	return &response, &redemptions, nil
}

func ComputeDiscount(amount float64, discountType string, discountValue float64, maxDiscount *float64) float64 {
	var d float64

	switch discountType {
	case "percent":
		d = amount * discountValue / 100.0
		if maxDiscount != nil && d > *maxDiscount {
			d = *maxDiscount
		}
	case "fixed":
		d = discountValue
	default:
		return 0
	}

	if d > amount {
		d = amount
	}
	if d < 0 {
		d = 0
	}
	return d
}
