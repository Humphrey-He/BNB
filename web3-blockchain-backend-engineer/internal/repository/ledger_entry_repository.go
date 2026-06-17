package repository

import (
	"database/sql"
	"time"
)

// LedgerDirection represents ledger entry direction
type LedgerDirection string

const (
	LedgerDirectionCredit LedgerDirection = "credit"
	LedgerDirectionDebit  LedgerDirection = "debit"
)

// LedgerEntryType represents the type of ledger entry
type LedgerEntryType string

const (
	LedgerEntryTypeDeposit        LedgerEntryType = "deposit"
	LedgerEntryTypeWithdrawal     LedgerEntryType = "withdrawal"
	LedgerEntryTypeFreeze         LedgerEntryType = "freeze"
	LedgerEntryTypeUnfreeze       LedgerEntryType = "unfreeze"
	LedgerEntryTypeReversal       LedgerEntryType = "reversal"
)

// ReferenceType represents the type of reference
type ReferenceType string

const (
	ReferenceTypeDeposit     ReferenceType = "deposit"
	ReferenceTypeWithdrawal  ReferenceType = "withdrawal"
	ReferenceTypeAdjustment  ReferenceType = "adjustment"
	ReferenceTypeReorg       ReferenceType = "reorg"
)

// LedgerEntry represents the ledger_entries table
type LedgerEntry struct {
	ID             int64           `json:"id"`
	AccountAddress string          `json:"account_address"`
	ChainID        int64           `json:"chain_id"`
	TokenID        int64           `json:"token_id"`
	Direction      LedgerDirection `json:"direction"`
	Amount         string          `json:"amount"`
	BalanceBefore  string          `json:"balance_before"`
	BalanceAfter   string          `json:"balance_after"`
	EntryType      LedgerEntryType `json:"entry_type"`
	ReferenceType  ReferenceType   `json:"reference_type"`
	ReferenceID    int64           `json:"reference_id"`
	ReversalOf     int64           `json:"reversal_of,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

// LedgerEntryRepository defines the interface for ledger entry data access
type LedgerEntryRepository interface {
	Create(entry *LedgerEntry) error
	GetByID(id int64) (*LedgerEntry, error)
	Update(entry *LedgerEntry) error
	Delete(id int64) error
	List(limit int) ([]*LedgerEntry, error)
	ListByAccountAddress(accountAddress string, limit int) ([]*LedgerEntry, error)
	ListByReference(referenceType ReferenceType, referenceID int64) ([]*LedgerEntry, error)
	ListByChainIDAndTokenID(chainID int64, tokenID int64, limit int) ([]*LedgerEntry, error)
	ListByBlockNumber(chainID int64, blockNumber int64) ([]*LedgerEntry, error)
}

// ledgerEntryRepository implements LedgerEntryRepository
type ledgerEntryRepository struct {
	db *sql.DB
}

// NewLedgerEntryRepository creates a new LedgerEntryRepository
func NewLedgerEntryRepository(db *sql.DB) LedgerEntryRepository {
	return &ledgerEntryRepository{db: db}
}

func (r *ledgerEntryRepository) Create(entry *LedgerEntry) error {
	query := `
		INSERT INTO ledger_entries (account_address, chain_id, token_id, direction, amount, balance_before, balance_after, entry_type, reference_type, reference_id, reversal_of, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
	`
	var reversalOf sql.NullInt64
	if entry.ReversalOf > 0 {
		reversalOf = sql.NullInt64{Int64: entry.ReversalOf, Valid: true}
	}
	return r.db.QueryRow(
		query,
		entry.AccountAddress,
		entry.ChainID,
		entry.TokenID,
		entry.Direction,
		entry.Amount,
		entry.BalanceBefore,
		entry.BalanceAfter,
		entry.EntryType,
		entry.ReferenceType,
		entry.ReferenceID,
		reversalOf,
		entry.CreatedAt,
	).Scan(&entry.ID)
}

func (r *ledgerEntryRepository) GetByID(id int64) (*LedgerEntry, error) {
	query := `
		SELECT id, account_address, chain_id, token_id, direction, amount, balance_before, balance_after, entry_type, reference_type, reference_id, reversal_of, created_at
		FROM ledger_entries WHERE id = $1
	`
	entry := &LedgerEntry{}
	var reversalOf sql.NullInt64
	err := r.db.QueryRow(query, id).Scan(
		&entry.ID,
		&entry.AccountAddress,
		&entry.ChainID,
		&entry.TokenID,
		&entry.Direction,
		&entry.Amount,
		&entry.BalanceBefore,
		&entry.BalanceAfter,
		&entry.EntryType,
		&entry.ReferenceType,
		&entry.ReferenceID,
		&reversalOf,
		&entry.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	if reversalOf.Valid {
		entry.ReversalOf = reversalOf.Int64
	}
	return entry, nil
}

func (r *ledgerEntryRepository) Update(entry *LedgerEntry) error {
	query := `
		UPDATE ledger_entries
		SET account_address = $2, chain_id = $3, token_id = $4, direction = $5, amount = $6, entry_type = $7, reference_type = $8, reference_id = $9, reversal_of = $10
		WHERE id = $1
	`
	var reversalOf sql.NullInt64
	if entry.ReversalOf > 0 {
		reversalOf = sql.NullInt64{Int64: entry.ReversalOf, Valid: true}
	}
	_, err := r.db.Exec(
		query,
		entry.ID,
		entry.AccountAddress,
		entry.ChainID,
		entry.TokenID,
		entry.Direction,
		entry.Amount,
		entry.EntryType,
		entry.ReferenceType,
		entry.ReferenceID,
		reversalOf,
	)
	return err
}

func (r *ledgerEntryRepository) Delete(id int64) error {
	query := `DELETE FROM ledger_entries WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *ledgerEntryRepository) List(limit int) ([]*LedgerEntry, error) {
	query := `
		SELECT id, account_address, chain_id, token_id, direction, amount, balance_before, balance_after, entry_type, reference_type, reference_id, reversal_of, created_at
		FROM ledger_entries ORDER BY id DESC LIMIT $1
	`
	return r.scanLedgerEntries(r.db.Query(query, limit))
}

func (r *ledgerEntryRepository) ListByAccountAddress(accountAddress string, limit int) ([]*LedgerEntry, error) {
	query := `
		SELECT id, account_address, chain_id, token_id, direction, amount, balance_before, balance_after, entry_type, reference_type, reference_id, reversal_of, created_at
		FROM ledger_entries WHERE account_address = $1 ORDER BY id DESC LIMIT $2
	`
	return r.scanLedgerEntries(r.db.Query(query, accountAddress, limit))
}

func (r *ledgerEntryRepository) ListByReference(referenceType ReferenceType, referenceID int64) ([]*LedgerEntry, error) {
	query := `
		SELECT id, account_address, chain_id, token_id, direction, amount, balance_before, balance_after, entry_type, reference_type, reference_id, reversal_of, created_at
		FROM ledger_entries WHERE reference_type = $1 AND reference_id = $2 ORDER BY id
	`
	return r.scanLedgerEntries(r.db.Query(query, referenceType, referenceID))
}

func (r *ledgerEntryRepository) ListByChainIDAndTokenID(chainID int64, tokenID int64, limit int) ([]*LedgerEntry, error) {
	query := `
		SELECT id, account_address, chain_id, token_id, direction, amount, balance_before, balance_after, entry_type, reference_type, reference_id, reversal_of, created_at
		FROM ledger_entries WHERE chain_id = $1 AND token_id = $2 ORDER BY id DESC LIMIT $3
	`
	return r.scanLedgerEntries(r.db.Query(query, chainID, tokenID, limit))
}

// ListByBlockNumber returns ledger entries related to a specific block number.
// Since ledger_entries don't have a direct block_number column, this implementation
// queries entries and filters them in memory based on their reference_id.
// For production, consider adding a block_number column to ledger_entries
// or creating a separate index table mapping entries to blocks.
func (r *ledgerEntryRepository) ListByBlockNumber(chainID int64, blockNumber int64) ([]*LedgerEntry, error) {
	// Query all entries for this chain and filter by reference_id matching block number
	// This is a simplified implementation - in production you'd want proper indexing
	query := `
		SELECT id, account_address, chain_id, token_id, direction, amount, balance_before, balance_after, entry_type, reference_type, reference_id, reversal_of, created_at
		FROM ledger_entries WHERE chain_id = $1 ORDER BY id
	`
	entries, err := r.scanLedgerEntries(r.db.Query(query, chainID))
	if err != nil {
		return nil, err
	}

	// Filter entries where reference_id matches the block number
	// This works for entries that directly reference block-scoped entities
	var filtered []*LedgerEntry
	for _, entry := range entries {
		if entry.ReferenceID == blockNumber {
			filtered = append(filtered, entry)
		}
	}

	return filtered, nil
}

func (r *ledgerEntryRepository) scanLedgerEntries(rows *sql.Rows, err error) ([]*LedgerEntry, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*LedgerEntry
	for rows.Next() {
		entry := &LedgerEntry{}
		var reversalOf sql.NullInt64
		err := rows.Scan(
			&entry.ID,
			&entry.AccountAddress,
			&entry.ChainID,
			&entry.TokenID,
			&entry.Direction,
			&entry.Amount,
			&entry.BalanceBefore,
			&entry.BalanceAfter,
			&entry.EntryType,
			&entry.ReferenceType,
			&entry.ReferenceID,
			&reversalOf,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		if reversalOf.Valid {
			entry.ReversalOf = reversalOf.Int64
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}
