package postgres

import (
	"context"
	"database/sql"
	"fmt"
	m "imed/pkg/models"
	"log/slog"
	"time"
)

const (
	ErrOK              = 0
	ErrSignFailed      = -1
	ErrIncorrectAmount = -2
	ErrActionNotFound  = -3
	ErrAlreadyPaid     = -4
	ErrOrderNotExist   = -5
	ErrTransNotExist   = -6
	ErrFailedUpdate    = -7
	ErrBadRequest      = -8
	ErrCancelled       = -9
)

func (r *OrderRepo) ApplyPrepare(p *m.ApplyPrepareParams) (*m.PrepareResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var amount float64
	var status string
	var existingPrepare sql.NullInt64
	var metaOld []byte

	err = tx.QueryRowContext(ctx, `
		SELECT amount::float8, status::text, merchant_prepare_id, meta
		FROM orders
		WHERE id=$1::uuid
		FOR UPDATE
	`, p.OrderId).Scan(&amount, &status, &existingPrepare, &metaOld)

	// meta, MetaPatchJSON va p loglanmaydi: ular ichida click_prepare_raw,
	// ya'ni Click imzosi (sign_string) bor.
	slog.Info("click prepare order state",
		"order_id", p.OrderId,
		"click_trans_id", p.ClickTransId,
		"existing_prepare", existingPrepare.Int64,
		"status", status,
		"db_amount", amount,
		"req_amount", p.Amount,
	)

	if err != nil {
		return &m.PrepareResult{Error: ErrOrderNotExist, ErrorNote: "Order not found"}, nil
	}
	if amount != p.Amount {
		return &m.PrepareResult{Error: ErrIncorrectAmount, ErrorNote: "Incorrect parameter amount"}, nil
	}
	if status == "PAID" {
		return &m.PrepareResult{Error: ErrAlreadyPaid, ErrorNote: "Already paid", MerchantPrepareId: existingPrepare.Int64}, nil
	}
	if status == "CANCELLED" {
		return &m.PrepareResult{Error: ErrCancelled, ErrorNote: "Transaction cancelled", MerchantPrepareId: existingPrepare.Int64}, nil
	}

	// meta merge: meta = meta || patch
	_, _ = tx.ExecContext(ctx, `
		UPDATE orders
		SET meta = COALESCE(meta,'{}'::jsonb) || COALESCE($2::jsonb,'{}'::jsonb),
		    updated_at=now()
		WHERE id=$1::uuid
	`, p.OrderId, string(p.MetaPatchJSON))

	// idempotent prepare
	if existingPrepare.Valid {
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit: %w", err)
		}
		return &m.PrepareResult{Error: ErrOK, ErrorNote: "Success", MerchantPrepareId: existingPrepare.Int64}, nil
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE orders
		SET status='RESERVED',
		    click_trans_id=$2,
		    click_paydoc_id=$3,
		    merchant_prepare_id=$4,
		    updated_at=now()
		WHERE id=$1::uuid
	`, p.OrderId, p.ClickTransId, p.ClickPaydocId, p.MerchantPrepareId)
	if err != nil {
		return &m.PrepareResult{Error: ErrFailedUpdate, ErrorNote: "Failed to update"}, nil
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &m.PrepareResult{Error: ErrOK, ErrorNote: "Success", MerchantPrepareId: p.MerchantPrepareId}, nil
}

func (r *OrderRepo) ApplyComplete(p *m.ApplyCompleteParams) (*m.CompleteResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var status string
	var prepare sql.NullInt64
	var paidAt sql.NullTime

	err = tx.QueryRowContext(ctx, `
		SELECT status::text, merchant_prepare_id, paid_at
		FROM orders
		WHERE id=$1::uuid
		FOR UPDATE
	`, p.OrderId).Scan(&status, &prepare, &paidAt)

	// p loglanmaydi: MetaPatchJSON ichida Click imzosi bor.
	slog.Info("click complete order state",
		"order_id", p.OrderId,
		"click_trans_id", p.ClickTransId,
		"db_prepare", prepare.Int64,
		"status", status,
	)

	if err != nil {
		return &m.CompleteResult{Error: ErrOrderNotExist, ErrorNote: "Order not found"}, nil
	}
	if !prepare.Valid || prepare.Int64 != p.MerchantPrepareId {
		return &m.CompleteResult{Error: ErrTransNotExist, ErrorNote: "Transaction does not exist"}, nil
	}

	// meta merge
	_, _ = tx.ExecContext(ctx, `
		UPDATE orders
		SET meta = COALESCE(meta,'{}'::jsonb) || COALESCE($2::jsonb,'{}'::jsonb),
		    updated_at=now()
		WHERE id=$1::uuid
	`, p.OrderId, string(p.MetaPatchJSON))

	// idempotent
	if status == "PAID" {
		conf := int64(0)
		if paidAt.Valid {
			conf = paidAt.Time.Unix()
		}
		_ = tx.Commit()
		return &m.CompleteResult{Error: ErrOK, ErrorNote: "Success", MerchantConfirmId: conf}, nil
	}

	if p.ErrorCode != 0 {
		_, err = tx.ExecContext(ctx, `
			UPDATE orders SET status='CANCELLED', updated_at=now()
			WHERE id=$1::uuid
		`, p.OrderId)
		if err != nil {
			return &m.CompleteResult{Error: ErrFailedUpdate, ErrorNote: "Failed to update"}, nil
		}
		_ = tx.Commit()
		return &m.CompleteResult{Error: ErrCancelled, ErrorNote: "Transaction cancelled", MerchantConfirmId: 0}, nil
	}

	// mark paid
	_, err = tx.ExecContext(ctx, `
		UPDATE orders
		SET status='PAID',
		    paid_at=now(),
		    click_trans_id=$2,
		    updated_at=now()
		WHERE id=$1::uuid
	`, p.OrderId, p.ClickTransId)
	if err != nil {
		return &m.CompleteResult{Error: ErrFailedUpdate, ErrorNote: "Failed to update"}, nil
	}

	var ucID string

	err = tx.QueryRowContext(ctx, `
	INSERT INTO user_courses (
		user_id, course_id, tariff_id,
		started_at, ended_at,
		created_at, updated_at, deleted_at
	  )
	  SELECT
		o.user_id,
		o.course_id,
		o.tariff_id,
		now(),
		now() + (o.duration || ' months')::interval,
		now(), now(), 0
	  FROM orders o
	  WHERE o.id = $1::uuid
		AND o.status = 'PAID'
		AND NOT EXISTS (
		  SELECT 1
		  FROM user_courses uc
		  WHERE uc.user_id = o.user_id
			AND uc.course_id = o.course_id
			AND uc.deleted_at = 0
			AND (uc.ended_at IS NULL OR uc.ended_at > now())
		)
	  RETURNING id
	`, p.OrderId).Scan(&ucID)

	if err == sql.ErrNoRows {
		slog.Info("no user course created, it may already exist", "order_id", p.OrderId)
	} else if err != nil {
		return &m.CompleteResult{Error: ErrFailedUpdate, ErrorNote: "Failed to create user course"}, nil
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &m.CompleteResult{Error: ErrOK, ErrorNote: "Success", MerchantConfirmId: time.Now().Unix()}, nil
}
