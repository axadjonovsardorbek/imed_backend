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

	fmt.Println("----------------------------------")
	fmt.Println("CLICK PREPARE ORDER META OLD:", string(metaOld))
	fmt.Println("----------------------------------")
	fmt.Println("CLICK PREPARE ORDER META PATCH:", string(p.MetaPatchJSON))
	fmt.Println("----------------------------------")
	fmt.Println("CLICK PREPARE PARAMS:", p)
	fmt.Println("----------------------------------")
	fmt.Println("CLICK PREPARE EXISTING PREPARE:", existingPrepare)
	fmt.Println("----------------------------------")
	fmt.Println("CLICK PREPARE STATUS:", status)
	fmt.Println("----------------------------------")
	fmt.Println("CLICK PREPARE AMOUNT:", amount)
	fmt.Println("----------------------------------")
	fmt.Println("CLICK PREPARE PARAM AMOUNT:", p.Amount)
	fmt.Println("----------------------------------")

	if err != nil {
		fmt.Println("CHeck 1")
		return &m.PrepareResult{Error: ErrOrderNotExist, ErrorNote: "Order not found"}, nil
	}
	if amount != p.Amount {
		fmt.Println("CHeck 2")
		return &m.PrepareResult{Error: ErrIncorrectAmount, ErrorNote: "Incorrect parameter amount"}, nil
	}
	if status == "PAID" {
		fmt.Println("CHeck 2")
		return &m.PrepareResult{Error: ErrAlreadyPaid, ErrorNote: "Already paid", MerchantPrepareId: existingPrepare.Int64}, nil
	}
	if status == "CANCELLED" {
		fmt.Println("CHeck 4")
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
			fmt.Println("CHeck 5")
			return nil, fmt.Errorf("commit: %w", err)
		}
		fmt.Println("CHeck 6")
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
		fmt.Println("CHeck 7")
		return &m.PrepareResult{Error: ErrFailedUpdate, ErrorNote: "Failed to update"}, nil
	}

	if err := tx.Commit(); err != nil {
		fmt.Println("CHeck 8")
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


	fmt.Println("----------------------------------")
	fmt.Println("CLICK COMPLETE PARAMS:", p)
	fmt.Println("----------------------------------")
	fmt.Println("CLICK COMPLETE PREPARE IN DB:", prepare)
	fmt.Println("----------------------------------")
	fmt.Println("CLICK COMPLETE STATUS IN DB:", status)

	if err != nil {
		fmt.Println("Check complete 1")
		return &m.CompleteResult{Error: ErrOrderNotExist, ErrorNote: "Order not found"}, nil
	}
	if !prepare.Valid || prepare.Int64 != p.MerchantPrepareId {
		fmt.Println("Check complete 2")
		return &m.CompleteResult{Error: ErrTransNotExist, ErrorNote: "Transaction does not exist"}, nil
	}

	// meta merge
	_, _ = tx.ExecContext(ctx, `
		UPDATE orders
		SET meta = COALESCE(meta,'{}'::jsonb) || COALESCE($2::jsonb,'{}'::jsonb),
		    updated_at=now()
		WHERE id=$1::uuid
	`, p.OrderId, string(p.MetaPatchJSON))

	fmt.Println("----------------------------------")
	fmt.Println("CLICK COMPLETE ORDER META PATCH:", string(p.MetaPatchJSON))
	fmt.Println("----------------------------------")

	// idempotent
	if status == "PAID" {
		conf := int64(0)
		if paidAt.Valid {
			conf = paidAt.Time.Unix()
		}
		_ = tx.Commit()
		fmt.Println("Check complete 3")
		return &m.CompleteResult{Error: ErrOK, ErrorNote: "Success", MerchantConfirmId: conf}, nil
	}

	if p.ErrorCode != 0 {
		_, err = tx.ExecContext(ctx, `
			UPDATE orders SET status='CANCELLED', updated_at=now()
			WHERE id=$1::uuid
		`, p.OrderId)
		if err != nil {
			fmt.Println("Check complete 4")
			return &m.CompleteResult{Error: ErrFailedUpdate, ErrorNote: "Failed to update"}, nil
		}
		_ = tx.Commit()
		fmt.Println("Check complete 5")
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
		fmt.Println("Check complete 6")
		return &m.CompleteResult{Error: ErrFailedUpdate, ErrorNote: "Failed to update"}, nil
	}

	if err := tx.Commit(); err != nil {
		fmt.Println("Check complete 7")
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &m.CompleteResult{Error: ErrOK, ErrorNote: "Success", MerchantConfirmId: time.Now().Unix()}, nil
}
