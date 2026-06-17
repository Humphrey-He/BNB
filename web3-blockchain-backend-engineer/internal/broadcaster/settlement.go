package broadcaster

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
)

func settleBroadcastedWithdrawal(ctx context.Context, db *sql.DB, withdrawal *repository.Withdrawal) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin broadcast settlement transaction: %w", err)
	}
	defer tx.Rollback()

	var existingID int64
	err = tx.QueryRowContext(ctx,
		`SELECT id
		 FROM ledger_entries
		 WHERE reference_type = $1 AND reference_id = $2 AND entry_type = $3
		 FOR UPDATE`,
		repository.ReferenceTypeWithdrawal,
		withdrawal.ID,
		repository.LedgerEntryTypeWithdrawal,
	).Scan(&existingID)
	if err == nil {
		return tx.Commit()
	}
	if err != sql.ErrNoRows {
		return fmt.Errorf("failed to check existing withdrawal settlement entry: %w", err)
	}

	var frozenBefore string
	var frozenAfter string
	debitQuery := `
		UPDATE balances
		SET frozen_balance = frozen_balance - $4::numeric,
		    updated_at = now()
		WHERE account_address = $1
		  AND chain_id = $2
		  AND token_id = $3
		  AND frozen_balance::numeric >= $4::numeric
		RETURNING (frozen_balance + $4::numeric)::text AS frozen_before,
		          frozen_balance::text AS frozen_after
	`
	if err := tx.QueryRowContext(ctx, debitQuery,
		withdrawal.FromAddress,
		withdrawal.ChainID,
		withdrawal.TokenID,
		withdrawal.Amount,
	).Scan(&frozenBefore, &frozenAfter); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("insufficient frozen balance for broadcast settlement")
		}
		return fmt.Errorf("failed to consume frozen balance: %w", err)
	}

	var ledgerID int64
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO ledger_entries (account_address, chain_id, token_id, direction, amount, balance_before, balance_after, entry_type, reference_type, reference_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`,
		withdrawal.FromAddress,
		withdrawal.ChainID,
		withdrawal.TokenID,
		repository.LedgerDirectionDebit,
		withdrawal.Amount,
		frozenBefore,
		frozenAfter,
		repository.LedgerEntryTypeWithdrawal,
		repository.ReferenceTypeWithdrawal,
		withdrawal.ID,
		time.Now(),
	).Scan(&ledgerID); err != nil {
		return fmt.Errorf("failed to create withdrawal settlement ledger entry: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit withdrawal settlement transaction: %w", err)
	}

	return nil
}

func releaseFailedWithdrawal(ctx context.Context, db *sql.DB, withdrawal *repository.Withdrawal) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin withdrawal release transaction: %w", err)
	}
	defer tx.Rollback()

	var existingID int64
	err = tx.QueryRowContext(ctx,
		`SELECT id
		 FROM ledger_entries
		 WHERE reference_type = $1 AND reference_id = $2 AND entry_type = $3
		 FOR UPDATE`,
		repository.ReferenceTypeWithdrawal,
		withdrawal.ID,
		repository.LedgerEntryTypeUnfreeze,
	).Scan(&existingID)
	if err == nil {
		return tx.Commit()
	}
	if err != sql.ErrNoRows {
		return fmt.Errorf("failed to check existing unfreeze entry: %w", err)
	}

	var balanceBefore string
	var balanceAfter string
	unfreezeQuery := `
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
	`
	if err := tx.QueryRowContext(ctx, unfreezeQuery,
		withdrawal.FromAddress,
		withdrawal.ChainID,
		withdrawal.TokenID,
		withdrawal.Amount,
	).Scan(&balanceBefore, &balanceAfter); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("insufficient frozen balance for withdrawal release")
		}
		return fmt.Errorf("failed to release frozen balance: %w", err)
	}

	var ledgerID int64
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO ledger_entries (account_address, chain_id, token_id, direction, amount, balance_before, balance_after, entry_type, reference_type, reference_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`,
		withdrawal.FromAddress,
		withdrawal.ChainID,
		withdrawal.TokenID,
		repository.LedgerDirectionCredit,
		withdrawal.Amount,
		balanceBefore,
		balanceAfter,
		repository.LedgerEntryTypeUnfreeze,
		repository.ReferenceTypeWithdrawal,
		withdrawal.ID,
		time.Now(),
	).Scan(&ledgerID); err != nil {
		return fmt.Errorf("failed to create withdrawal unfreeze ledger entry: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit withdrawal release transaction: %w", err)
	}

	return nil
}

func compensateRevertedWithdrawal(ctx context.Context, db *sql.DB, withdrawal *repository.Withdrawal) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin withdrawal compensation transaction: %w", err)
	}
	defer tx.Rollback()

	var existingID int64
	err = tx.QueryRowContext(ctx,
		`SELECT id
		 FROM ledger_entries
		 WHERE reference_type = $1 AND reference_id = $2 AND entry_type = $3
		 FOR UPDATE`,
		repository.ReferenceTypeWithdrawal,
		withdrawal.ID,
		repository.LedgerEntryTypeReversal,
	).Scan(&existingID)
	if err == nil {
		return tx.Commit()
	}
	if err != sql.ErrNoRows {
		return fmt.Errorf("failed to check existing reversal entry: %w", err)
	}

	var balanceBefore string
	var balanceAfter string
	creditQuery := `
		UPDATE balances
		SET available_balance = available_balance + $4::numeric,
		    updated_at = now()
		WHERE account_address = $1
		  AND chain_id = $2
		  AND token_id = $3
		RETURNING (available_balance - $4::numeric)::text AS balance_before,
		          available_balance::text AS balance_after
	`
	if err := tx.QueryRowContext(ctx, creditQuery,
		withdrawal.FromAddress,
		withdrawal.ChainID,
		withdrawal.TokenID,
		withdrawal.Amount,
	).Scan(&balanceBefore, &balanceAfter); err != nil {
		return fmt.Errorf("failed to compensate reverted withdrawal balance: %w", err)
	}

	var ledgerID int64
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO ledger_entries (account_address, chain_id, token_id, direction, amount, balance_before, balance_after, entry_type, reference_type, reference_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`,
		withdrawal.FromAddress,
		withdrawal.ChainID,
		withdrawal.TokenID,
		repository.LedgerDirectionCredit,
		withdrawal.Amount,
		balanceBefore,
		balanceAfter,
		repository.LedgerEntryTypeReversal,
		repository.ReferenceTypeWithdrawal,
		withdrawal.ID,
		time.Now(),
	).Scan(&ledgerID); err != nil {
		return fmt.Errorf("failed to create reverted withdrawal reversal entry: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit withdrawal compensation transaction: %w", err)
	}

	return nil
}
