package postgres

import (
	"context"
	"database/sql"
	"fmt"
	m "imed/pkg/models"
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

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &m.CompleteResult{Error: ErrOK, ErrorNote: "Success", MerchantConfirmId: time.Now().Unix()}, nil
}
