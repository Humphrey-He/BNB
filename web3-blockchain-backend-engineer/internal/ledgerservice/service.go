package ledgerservice

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
	"github.com/nats-io/nats.go"
)

// NATS subjects
const (
	SubjectDepositConfirmed = "deposit_confirmed"
)

// LedgerService processes ledger entries and balance updates
type LedgerService struct {
	db                *sql.DB
	natsClient        *nats.Conn
	ledgerRepo        repository.LedgerEntryRepository
	balanceRepo       repository.BalanceRepository
	depositRepo       repository.DepositRepository
	logger            *slog.Logger
	reconcileInterval time.Duration
}

// NewLedgerService creates a new LedgerService
func NewLedgerService(
	db *sql.DB,
	natsClient *nats.Conn,
	ledgerRepo repository.LedgerEntryRepository,
	balanceRepo repository.BalanceRepository,
	depositRepo repository.DepositRepository,
	logger *slog.Logger,
) *LedgerService {
	return &LedgerService{
		db:                db,
		natsClient:        natsClient,
		ledgerRepo:        ledgerRepo,
		balanceRepo:       balanceRepo,
		depositRepo:       depositRepo,
		logger:            logger,
		reconcileInterval: 30 * time.Second,
	}
}

// Start begins the ledger service loop
func (s *LedgerService) Start(ctx context.Context) error {
	// Subscribe to confirmed deposits
	sub, err := s.natsClient.Subscribe(SubjectDepositConfirmed, s.handleDepositConfirmed)
	if err != nil {
		s.logger.Error("Failed to subscribe to deposit_confirmed", "error", err)
		return err
	}

	s.logger.Info("Ledger service started, listening for deposit_confirmed events")

	if err := s.reconcileConfirmedDeposits(ctx); err != nil {
		s.logger.Warn("Initial confirmed deposit reconciliation failed", "error", err)
	}

	ticker := time.NewTicker(s.reconcileInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			sub.Unsubscribe()
			s.logger.Info("Ledger service stopped")
			return ctx.Err()
		case <-ticker.C:
			if err := s.reconcileConfirmedDeposits(ctx); err != nil {
				s.logger.Warn("Confirmed deposit reconciliation failed", "error", err)
			}
		}
	}
}

// handleDepositConfirmed processes a confirmed deposit event
func (s *LedgerService) handleDepositConfirmed(msg *nats.Msg) {
	var event DepositConfirmedEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		s.logger.Error("Failed to unmarshal deposit confirmed event", "error", err)
		return
	}

	s.logger.Info("Processing deposit confirmed event",
		"deposit_id", event.DepositID,
		"account", event.Account,
		"amount", event.Amount,
		"chain_id", event.ChainID,
		"token_id", event.TokenID,
	)

	// Create ledger entry and update balance atomically
	if err := s.processDepositCredit(context.Background(), &event); err != nil {
		s.logger.Error("Failed to process deposit credit",
			"deposit_id", event.DepositID,
			"error", err,
		)
		return
	}

	s.logger.Info("Deposit credit processed successfully",
		"deposit_id", event.DepositID,
		"account", event.Account,
		"amount", event.Amount,
	)

	msg.Ack()
}

// processDepositCredit creates a credit ledger entry and updates the balance in a single transaction
// This ensures idempotency: if the ledger entry already exists, no changes are made.
// If credit succeeds but ledger entry fails, the transaction rolls back - no double credit.
func (s *LedgerService) processDepositCredit(ctx context.Context, event *DepositConfirmedEvent) error {
	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Lock an existing ledger row when present. If another worker inserts the same
	// reference later, the unique constraint rolls this transaction back safely.
	var existingID int64
	err = tx.QueryRowContext(ctx,
		`SELECT id
		 FROM ledger_entries
		 WHERE reference_type = $1 AND reference_id = $2 AND entry_type = $3
		 FOR UPDATE`,
		repository.ReferenceTypeDeposit, event.DepositID, repository.LedgerEntryTypeDeposit,
	).Scan(&existingID)
	if err == nil {
		s.logger.Info("Ledger entry already exists for deposit, skipping",
			"deposit_id", event.DepositID,
			"entry_id", existingID,
		)
		if commitErr := tx.Commit(); commitErr != nil {
			return fmt.Errorf("failed to commit idempotent skip: %w", commitErr)
		}
		return nil
	}
	if err != sql.ErrNoRows {
		return fmt.Errorf("failed to check existing ledger entry: %w", err)
	}

	// Credit the balance within the same transaction and return the audited before/after values.
	creditQuery := `
		INSERT INTO balances (account_address, chain_id, token_id, available_balance, frozen_balance, updated_at)
		VALUES ($1, $2, $3, $4::numeric, 0, now())
		ON CONFLICT (account_address, chain_id, token_id)
		DO UPDATE SET available_balance = balances.available_balance::numeric + $4::numeric, updated_at = now()
		RETURNING (available_balance - $4::numeric)::text AS balance_before,
		          available_balance::text AS balance_after
	`
	var balanceBefore string
	var balanceAfter string
	err = tx.QueryRowContext(ctx, creditQuery,
		event.Account, event.ChainID, event.TokenID, event.Amount,
	).Scan(&balanceBefore, &balanceAfter)
	if err != nil {
		return fmt.Errorf("failed to credit balance: %w", err)
	}

	s.logger.Info("Credited balance in transaction",
		"account", event.Account,
		"chain_id", event.ChainID,
		"token_id", event.TokenID,
		"amount", event.Amount,
		"balance_before", balanceBefore,
		"balance_after", balanceAfter,
	)

	// Create ledger entry within the same transaction
	ledgerQuery := `
		INSERT INTO ledger_entries (account_address, chain_id, token_id, direction, amount, balance_before, balance_after, entry_type, reference_type, reference_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, now())
		RETURNING id
	`
	entry := &repository.LedgerEntry{}
	err = tx.QueryRowContext(ctx, ledgerQuery,
		event.Account,
		event.ChainID,
		event.TokenID,
		repository.LedgerDirectionCredit,
		event.Amount,
		balanceBefore,
		balanceAfter,
		repository.LedgerEntryTypeDeposit,
		repository.ReferenceTypeDeposit,
		event.DepositID,
	).Scan(&entry.ID)
	if err != nil {
		// Transaction will rollback - no partial state
		return fmt.Errorf("failed to create ledger entry: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Info("Created credit ledger entry",
		"entry_id", entry.ID,
		"account", event.Account,
		"amount", event.Amount,
		"deposit_id", event.DepositID,
		"balance_before", balanceBefore,
		"balance_after", balanceAfter,
	)

	return nil
}

// subtractStrings computes a - b for string-encoded big integers
func subtractStrings(a, b string) string {
	aInt, ok := new(big.Int).SetString(a, 10)
	if !ok {
		return ""
	}
	bInt, ok := new(big.Int).SetString(b, 10)
	if !ok {
		return ""
	}
	result := new(big.Int).Sub(aInt, bInt)
	if result.Sign() < 0 {
		return "0"
	}
	return result.String()
}

func (s *LedgerService) reconcileConfirmedDeposits(ctx context.Context) error {
	rows, err := s.db.QueryContext(ctx, `
		SELECT d.id, d.chain_id, d.token_id, d.to_address, d.amount, d.tx_hash, d.block_number
		FROM deposits d
		LEFT JOIN ledger_entries le
			ON le.reference_type = $1
			AND le.reference_id = d.id
			AND le.entry_type = $2
		WHERE d.status = $3
			AND le.id IS NULL
		ORDER BY d.id ASC
		LIMIT 100
	`, repository.ReferenceTypeDeposit, repository.LedgerEntryTypeDeposit, repository.DepositStatusConfirmed)
	if err != nil {
		return fmt.Errorf("failed to query uncredited confirmed deposits: %w", err)
	}
	defer rows.Close()

	var events []*DepositConfirmedEvent
	for rows.Next() {
		event := &DepositConfirmedEvent{}
		if err := rows.Scan(
			&event.DepositID,
			&event.ChainID,
			&event.TokenID,
			&event.Account,
			&event.Amount,
			&event.TxHash,
			&event.BlockNumber,
		); err != nil {
			return fmt.Errorf("failed to scan uncredited deposit: %w", err)
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, event := range events {
		if err := s.processDepositCredit(ctx, event); err != nil {
			return fmt.Errorf("failed to reconcile deposit %d: %w", event.DepositID, err)
		}
	}
	if len(events) > 0 {
		s.logger.Info("Reconciled confirmed deposits", "count", len(events))
	}
	return nil
}

// updateBalance is deprecated - use Credit() directly for atomic operations
// Kept for backwards compatibility with other use cases
func (s *LedgerService) updateBalance(account string, chainID, tokenID int64, amount string) error {
	newBalance, err := s.balanceRepo.Credit(account, chainID, tokenID, amount)
	if err != nil {
		return err
	}
	s.logger.Info("Updated balance",
		"account", account,
		"chain_id", chainID,
		"token_id", tokenID,
		"new_available", newBalance.AvailableBalance,
	)
	return nil
}

// DepositConfirmedEvent represents a confirmed deposit event from NATS
type DepositConfirmedEvent struct {
	DepositID   int64  `json:"deposit_id"`
	ChainID     int64  `json:"chain_id"`
	TokenID     int64  `json:"token_id"`
	Account     string `json:"account"`
	Amount      string `json:"amount"`
	TxHash      string `json:"tx_hash"`
	BlockNumber int64  `json:"block_number"`
}
