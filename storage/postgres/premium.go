package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"imed/internal/config"
	m "imed/pkg/models"

	"github.com/google/uuid"
)

var ErrPremiumNotConfigured = errors.New("premium price is not configured")

type PremiumRepo struct {
	db *sql.DB
}

func NewPremiumRepo(db *sql.DB) *PremiumRepo {
	return &PremiumRepo{db: db}
}

func (r *PremiumRepo) GetSettings(ctx context.Context) (*m.PremiumSettings, error) {
	var raw []byte
	err := r.db.QueryRowContext(ctx, `SELECT value FROM settings WHERE key = 'premium'`).Scan(&raw)
	if err == sql.ErrNoRows {
		return &m.PremiumSettings{DurationMonths: 1}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error while getting premium settings: %w", err)
	}

	res := m.PremiumSettings{}
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, err
	}
	if res.DurationMonths <= 0 {
		res.DurationMonths = 1
	}
	return &res, nil
}

func (r *PremiumRepo) UpdateSettings(ctx context.Context, req *m.PremiumSettings) (*m.Void, error) {
	raw, _ := json.Marshal(req)
	_, err := r.db.ExecContext(ctx, `
	INSERT INTO settings (key, value, updated_at) VALUES ('premium', $1, now())
	ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = now()
	`, string(raw))
	if err != nil {
		return nil, fmt.Errorf("error while updating premium settings: %w", err)
	}
	return &m.Void{}, nil
}

func (r *PremiumRepo) GetStatus(ctx context.Context, userId string) (*m.AccessStatus, error) {
	var (
		res   m.AccessStatus
		until sql.NullTime
	)
	err := r.db.QueryRowContext(ctx, `
	SELECT
		u.is_ustoz,
		(SELECT MAX(s.ended_at) FROM user_subscriptions s WHERE s.user_id = u.id AND s.ended_at > now())
	FROM users u
	WHERE u.id = $1 AND u.deleted_at = 0
	`, userId).Scan(&res.IsUstoz, &until)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found with id: %s", userId)
	}
	if err != nil {
		return nil, fmt.Errorf("error while getting premium status: %w", err)
	}
	if until.Valid {
		res.IsPremium = true
		v := until.Time.Format(time.RFC3339)
		res.PremiumUntil = &v
	}
	return &res, nil
}

// CreateOrder bitta oylik premium uchun Click buyurtma yaratadi (T-12).
func (r *PremiumRepo) CreateOrder(ctx context.Context, req *m.PremiumOrderReq) (res *m.OrderCreateRes, err error) {
	cfg := config.Load()

	settings, err := r.GetSettings(ctx)
	if err != nil {
		return nil, err
	}
	if settings.Price <= 0 {
		return nil, ErrPremiumNotConfigured
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	payURL := func(id string, amount float64) string {
		return fmt.Sprintf(
			"https://my.click.uz/services/pay?merchant_id=%d&service_id=%d&amount=%.2f&transaction_param=%s",
			cfg.CLICK_MERCHANT_ID,
			cfg.CLICK_SERVICE_ID,
			amount,
			id,
		)
	}

	// Oxirgi 1 kun ichidagi to'lanmagan xuddi shunday buyurtma qayta ishlatiladi.
	var (
		existingId     string
		existingAmount float64
	)
	_ = tx.QueryRowContext(ctx, `
	SELECT id, amount
	FROM orders
	WHERE user_id = $1
	  AND type = 'premium'
	  AND COALESCE(promo_id::text, '') = $2
	  AND course_amount = $3
	  AND status IN ('NEW', 'EXPIRED')
	  AND created_at >= now() - interval '1 days'
	ORDER BY created_at DESC
	LIMIT 1
	`, req.UserId, req.PromoCodeId, settings.Price).Scan(&existingId, &existingAmount)
	if existingId != "" && existingAmount > 0 {
		return &m.OrderCreateRes{Id: existingId, Url: payURL(existingId, existingAmount)}, nil
	}

	var (
		promoID  *string
		discount float64
		price    = float64(settings.Price)
	)
	if req.PromoCodeId != "" {
		promoID, discount, err = applyPromo(tx, req.PromoCodeId, req.UserId, "", price)
		if err != nil {
			return nil, err
		}
	}
	total := price - discount
	if total <= 0 {
		return nil, fmt.Errorf("order amount must be greater than 0")
	}

	id := uuid.NewString()
	_, err = tx.ExecContext(ctx, `
	INSERT INTO orders (
		id, user_id, type, duration, amount, promo_id, discount_amount, course_amount, meta
	) VALUES ($1, $2, 'premium', $3, $4, $5, $6, $7, '{"created_from": "api"}'::jsonb)
	`, id, req.UserId, settings.DurationMonths, total, promoID, discount, price)
	if err != nil {
		return nil, fmt.Errorf("error while creating premium order: %w", err)
	}

	if promoID != nil && discount > 0 {
		_, err = tx.ExecContext(ctx, `
		INSERT INTO promo_redemptions (
			promo_id, user_id, order_id, order_amount, discount_amount, course_amount, is_active
		) VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5, $6, TRUE)
		`, *promoID, req.UserId, id, total, discount, price)
		if err != nil {
			return nil, fmt.Errorf("error while creating promo redemption: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &m.OrderCreateRes{Id: id, Url: payURL(id, total)}, nil
}

// SetUstoz ustoz unvonini beradi yoki oladi (T-16).
func (r *PremiumRepo) SetUstoz(ctx context.Context, req *m.UstozReq) (*m.Void, error) {
	res, err := r.db.ExecContext(ctx, `
	UPDATE users
	SET
		is_ustoz = $2,
		ustoz_granted_at = CASE WHEN $2 THEN COALESCE(ustoz_granted_at, now()) ELSE NULL END,
		updated_at = now()
	WHERE id = $1 AND deleted_at = 0
	`, req.UserId, req.IsUstoz)
	if err != nil {
		return nil, fmt.Errorf("error while updating ustoz: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, fmt.Errorf("user not found with id: %s", req.UserId)
	}
	return &m.Void{}, nil
}
