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

var (
	ErrPremiumNotConfigured = errors.New("premium price is not configured")
	ErrInvalidPackage       = errors.New("package must be one of: fundamental, preclinical, clinical")
)

type PremiumRepo struct {
	db *sql.DB
}

func NewPremiumRepo(db *sql.DB) *PremiumRepo {
	return &PremiumRepo{db: db}
}

func settingsKey(pkg string) string {
	return "premium:" + pkg
}

func (r *PremiumRepo) GetSettings(ctx context.Context, pkg string) (*m.PremiumSettings, error) {
	if !m.IsPremiumPackage(pkg) {
		return nil, ErrInvalidPackage
	}
	res := m.PremiumSettings{Package: pkg, DurationMonths: 1}

	var raw []byte
	err := r.db.QueryRowContext(ctx, `SELECT value FROM settings WHERE key = $1`, settingsKey(pkg)).Scan(&raw)
	if err == sql.ErrNoRows {
		return &res, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error while getting premium settings: %w", err)
	}

	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, err
	}
	res.Package = pkg
	if res.DurationMonths <= 0 {
		res.DurationMonths = 1
	}
	return &res, nil
}

func (r *PremiumRepo) ListSettings(ctx context.Context) ([]m.PremiumSettings, error) {
	res := make([]m.PremiumSettings, 0, len(m.PremiumPackages))
	for _, pkg := range m.PremiumPackages {
		s, err := r.GetSettings(ctx, pkg)
		if err != nil {
			return nil, err
		}
		res = append(res, *s)
	}
	return res, nil
}

func (r *PremiumRepo) UpdateSettings(ctx context.Context, req *m.PremiumSettings) (*m.Void, error) {
	if !m.IsPremiumPackage(req.Package) {
		return nil, ErrInvalidPackage
	}
	raw, _ := json.Marshal(map[string]int{"price": req.Price, "duration_months": req.DurationMonths})
	_, err := r.db.ExecContext(ctx, `
	INSERT INTO settings (key, value, updated_at) VALUES ($1, $2, now())
	ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = now()
	`, settingsKey(req.Package), string(raw))
	if err != nil {
		return nil, fmt.Errorf("error while updating premium settings: %w", err)
	}
	return &m.Void{}, nil
}

// GetStatus har paket bo'yicha faol obuna tugash vaqtini qaytaradi.
// Eski umumiy ('all') obuna uchala paketni ham qoplaydi.
func (r *PremiumRepo) GetStatus(ctx context.Context, userId string) (*m.AccessStatus, error) {
	res := m.AccessStatus{}
	err := r.db.QueryRowContext(ctx, `SELECT is_ustoz FROM users WHERE id = $1 AND deleted_at = 0`, userId).Scan(&res.IsUstoz)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found with id: %s", userId)
	}
	if err != nil {
		return nil, fmt.Errorf("error while getting premium status: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `
	SELECT package, MAX(ended_at)
	FROM user_subscriptions
	WHERE user_id = $1 AND ended_at > now()
	GROUP BY package
	`, userId)
	if err != nil {
		return nil, fmt.Errorf("error while getting premium status: %w", err)
	}
	defer rows.Close()

	until := map[string]time.Time{}
	for rows.Next() {
		var (
			pkg string
			t   time.Time
		)
		if err := rows.Scan(&pkg, &t); err != nil {
			return nil, err
		}
		until[pkg] = t
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var latest time.Time
	for _, pkg := range m.PremiumPackages {
		t := until[pkg]
		if all := until[m.PackageAll]; all.After(t) {
			t = all
		}
		ps := m.PackageStatus{Package: pkg}
		if !t.IsZero() {
			ps.IsActive = true
			v := t.Format(time.RFC3339)
			ps.Until = &v
			if t.After(latest) {
				latest = t
			}
		}
		res.Packages = append(res.Packages, ps)
	}
	if !latest.IsZero() {
		res.IsPremium = true
		v := latest.Format(time.RFC3339)
		res.PremiumUntil = &v
	}
	return &res, nil
}

// CreateOrder tanlangan premium paket uchun Click buyurtma yaratadi (T-12).
func (r *PremiumRepo) CreateOrder(ctx context.Context, req *m.PremiumOrderReq) (res *m.OrderCreateRes, err error) {
	cfg := config.Load()

	settings, err := r.GetSettings(ctx, req.Package)
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
	  AND package = $4
	  AND COALESCE(promo_id::text, '') = $2
	  AND course_amount = $3
	  AND status IN ('NEW', 'EXPIRED')
	  AND created_at >= now() - interval '1 days'
	ORDER BY created_at DESC
	LIMIT 1
	`, req.UserId, req.PromoCodeId, settings.Price, req.Package).Scan(&existingId, &existingAmount)
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
		id, user_id, type, package, duration, amount, promo_id, discount_amount, course_amount, meta
	) VALUES ($1, $2, 'premium', $8, $3, $4, $5, $6, $7, '{"created_from": "api"}'::jsonb)
	`, id, req.UserId, settings.DurationMonths, total, promoID, discount, price, req.Package)
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
