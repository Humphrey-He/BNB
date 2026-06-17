package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
)

type withdrawalCreationService struct {
	db *sql.DB
}

func newWithdrawalCreationService(db *sql.DB) *withdrawalCreationService {
	return &withdrawalCreationService{db: db}
}

type withdrawalReviewService struct {
	db *sql.DB
}

func newWithdrawalReviewService(db *sql.DB) *withdrawalReviewService {
	return &withdrawalReviewService{db: db}
}

func (s *withdrawalCreationService) Create(ctx context.Context, withdrawal *repository.Withdrawal) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin withdrawal transaction: %w", err)
	}
	defer tx.Rollback()

	var balanceBefore string
	var balanceAfter string
	freezeQuery := `
		UPDATE balances
		SET available_balance = available_balance - $4::numeric,
		    frozen_balance = frozen_balance + $4::numeric,
		    updated_at = now()
		WHERE account_address = $1
		  AND chain_id = $2
		  AND token_id = $3
		  AND available_balance::numeric >= $4::numeric
		RETURNING (available_balance + $4::numeric)::text AS balance_before,
		          available_balance::text AS balance_after
	`
	err = tx.QueryRowContext(ctx, freezeQuery,
		withdrawal.FromAddress,
		withdrawal.ChainID,
		withdrawal.TokenID,
		withdrawal.Amount,
	).Scan(&balanceBefore, &balanceAfter)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("insufficient available balance")
		}
		return fmt.Errorf("failed to freeze balance: %w", err)
	}

	now := time.Now()
	withdrawal.CreatedAt = now
	withdrawal.UpdatedAt = now

	var txHash sql.NullString
	if withdrawal.TxHash != "" {
		txHash = sql.NullString{String: withdrawal.TxHash, Valid: true}
	}
	var nonce sql.NullInt64
	if withdrawal.Nonce > 0 {
		nonce = sql.NullInt64{Int64: withdrawal.Nonce, Valid: true}
	}
	var failureReason sql.NullString
	if withdrawal.FailureReason != "" {
		failureReason = sql.NullString{String: withdrawal.FailureReason, Valid: true}
	}

	createWithdrawalQuery := `
		INSERT INTO withdrawals (chain_id, token_id, from_address, to_address, amount, status, tx_hash, nonce, idempotency_key, failure_reason, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
	`
	if err := tx.QueryRowContext(ctx, createWithdrawalQuery,
		withdrawal.ChainID,
		withdrawal.TokenID,
		withdrawal.FromAddress,
		withdrawal.ToAddress,
		withdrawal.Amount,
		withdrawal.Status,
		txHash,
		nonce,
		withdrawal.IdempotencyKey,
		failureReason,
		withdrawal.CreatedAt,
		withdrawal.UpdatedAt,
	).Scan(&withdrawal.ID); err != nil {
		return fmt.Errorf("failed to create withdrawal: %w", err)
	}

	createLedgerQuery := `
		INSERT INTO ledger_entries (account_address, chain_id, token_id, direction, amount, balance_before, balance_after, entry_type, reference_type, reference_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`
	var ledgerID int64
	if err := tx.QueryRowContext(ctx, createLedgerQuery,
		withdrawal.FromAddress,
		withdrawal.ChainID,
		withdrawal.TokenID,
		repository.LedgerDirectionDebit,
		withdrawal.Amount,
		balanceBefore,
		balanceAfter,
		repository.LedgerEntryTypeFreeze,
		repository.ReferenceTypeWithdrawal,
		withdrawal.ID,
		now,
	).Scan(&ledgerID); err != nil {
		return fmt.Errorf("failed to create freeze ledger entry: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit withdrawal transaction: %w", err)
	}

	return nil
}

func (s *withdrawalReviewService) Approve(ctx context.Context, withdrawalID int64) (*repository.Withdrawal, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin approval transaction: %w", err)
	}
	defer tx.Rollback()

	withdrawal, err := loadWithdrawalForUpdate(ctx, tx, withdrawalID)
	if err != nil {
		return nil, err
	}
	if withdrawal.Status != repository.WithdrawalStatusManualReview {
		return nil, fmt.Errorf("withdrawal is not in manual_review status")
	}

	withdrawal.Status = repository.WithdrawalStatusApproved
	withdrawal.FailureReason = ""
	withdrawal.UpdatedAt = time.Now()

	if _, err := tx.ExecContext(ctx, `
		UPDATE withdrawals
		SET status = $2, failure_reason = NULL, updated_at = $3
		WHERE id = $1
	`, withdrawal.ID, withdrawal.Status, withdrawal.UpdatedAt); err != nil {
		return nil, fmt.Errorf("failed to approve withdrawal: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit approval transaction: %w", err)
	}

	return withdrawal, nil
}

func (s *withdrawalReviewService) Reject(ctx context.Context, withdrawalID int64, reason string) (*repository.Withdrawal, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin rejection transaction: %w", err)
	}
	defer tx.Rollback()

	withdrawal, err := loadWithdrawalForUpdate(ctx, tx, withdrawalID)
	if err != nil {
		return nil, err
	}
	if withdrawal.Status != repository.WithdrawalStatusManualReview {
		return nil, fmt.Errorf("withdrawal is not in manual_review status")
	}

	var existingID int64
	err = tx.QueryRowContext(ctx, `
		SELECT id
		FROM ledger_entries
		WHERE reference_type = $1 AND reference_id = $2 AND entry_type = $3
		FOR UPDATE
	`, repository.ReferenceTypeWithdrawal, withdrawal.ID, repository.LedgerEntryTypeUnfreeze).Scan(&existingID)
	if err == nil {
		withdrawal.Status = repository.WithdrawalStatusCanceled
		withdrawal.FailureReason = reason
		if _, updateErr := tx.ExecContext(ctx, `
			UPDATE withdrawals
			SET status = $2, failure_reason = $3, updated_at = $4
			WHERE id = $1
		`, withdrawal.ID, withdrawal.Status, withdrawal.FailureReason, time.Now()); updateErr != nil {
			return nil, fmt.Errorf("failed to persist canceled withdrawal: %w", updateErr)
		}
		if commitErr := tx.Commit(); commitErr != nil {
			return nil, fmt.Errorf("failed to commit rejection transaction: %w", commitErr)
		}
		return withdrawal, nil
	}
	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check existing unfreeze entry: %w", err)
	}

	var balanceBefore string
	var balanceAfter string
	if err := tx.QueryRowContext(ctx, `
		UPDATE balances
		SET available_balance = available_balance + $4::numeric,
		    frozen_balance = frozen_balance - $4::numeric,
		    updated_at = now()
		WHERE account_address = $1
		  AND chain_id = $2
		  AND token_id = $3
		  AND frozen_balance::numeric >= $4::numeric
		RETURNING (available_balance - $4::numeric)::text AS available_before,
		          available_balance::text AS available_after
	`, withdrawal.FromAddress, withdrawal.ChainID, withdrawal.TokenID, withdrawal.Amount).Scan(&balanceBefore, &balanceAfter); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("insufficient frozen balance for rejection release")
		}
		return nil, fmt.Errorf("failed to release frozen balance: %w", err)
	}

	var ledgerID int64
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO ledger_entries (account_address, chain_id, token_id, direction, amount, balance_before, balance_after, entry_type, reference_type, reference_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`, withdrawal.FromAddress, withdrawal.ChainID, withdrawal.TokenID, repository.LedgerDirectionCredit, withdrawal.Amount, balanceBefore, balanceAfter, repository.LedgerEntryTypeUnfreeze, repository.ReferenceTypeWithdrawal, withdrawal.ID, time.Now()).Scan(&ledgerID); err != nil {
		return nil, fmt.Errorf("failed to create rejection unfreeze ledger entry: %w", err)
	}

	withdrawal.Status = repository.WithdrawalStatusCanceled
	withdrawal.FailureReason = reason
	withdrawal.UpdatedAt = time.Now()
	if _, err := tx.ExecContext(ctx, `
		UPDATE withdrawals
		SET status = $2, failure_reason = $3, updated_at = $4
		WHERE id = $1
	`, withdrawal.ID, withdrawal.Status, withdrawal.FailureReason, withdrawal.UpdatedAt); err != nil {
		return nil, fmt.Errorf("failed to cancel withdrawal: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit rejection transaction: %w", err)
	}

	return withdrawal, nil
}

func loadWithdrawalForUpdate(ctx context.Context, tx *sql.Tx, withdrawalID int64) (*repository.Withdrawal, error) {
	withdrawal := &repository.Withdrawal{}
	var txHash sql.NullString
	var nonce sql.NullInt64
	var failureReason sql.NullString
	err := tx.QueryRowContext(ctx, `
		SELECT id, chain_id, token_id, from_address, to_address, amount, status, tx_hash, nonce, idempotency_key, failure_reason, created_at, updated_at
		FROM withdrawals
		WHERE id = $1
		FOR UPDATE
	`, withdrawalID).Scan(
		&withdrawal.ID,
		&withdrawal.ChainID,
		&withdrawal.TokenID,
		&withdrawal.FromAddress,
		&withdrawal.ToAddress,
		&withdrawal.Amount,
		&withdrawal.Status,
		&txHash,
		&nonce,
		&withdrawal.IdempotencyKey,
		&failureReason,
		&withdrawal.CreatedAt,
		&withdrawal.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("withdrawal not found")
		}
		return nil, fmt.Errorf("failed to load withdrawal: %w", err)
	}
	if txHash.Valid {
		withdrawal.TxHash = txHash.String
	}
	if nonce.Valid {
		withdrawal.Nonce = nonce.Int64
	}
	if failureReason.Valid {
		withdrawal.FailureReason = failureReason.String
	}
	return withdrawal, nil
}
