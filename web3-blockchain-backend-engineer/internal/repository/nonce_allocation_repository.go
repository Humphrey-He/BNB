package repository

import (
	"database/sql"
	"time"
)

// NonceAllocationStatus represents nonce allocation status
type NonceAllocationStatus string

const (
	NonceStatusAllocated NonceAllocationStatus = "allocated"
	NonceStatusUsed      NonceAllocationStatus = "used"
	NonceStatusExpired   NonceAllocationStatus = "expired"
)

// NonceAllocation represents the nonce_allocations table
type NonceAllocation struct {
	ID           int64                 `json:"id"`
	ChainID     int64                 `json:"chain_id"`
	FromAddress string                `json:"from_address"`
	Nonce       int64                 `json:"nonce"`
	WithdrawalID int64                `json:"withdrawal_id,omitempty"`
	Status      NonceAllocationStatus `json:"status"`
	ExpiresAt   time.Time             `json:"expires_at"`
	CreatedAt   time.Time             `json:"created_at"`
}

// NonceAllocationRepository defines the interface for nonce allocation data access
type NonceAllocationRepository interface {
	Create(allocation *NonceAllocation) error
	GetByID(id int64) (*NonceAllocation, error)
	GetByChainIDAndAddressAndNonce(chainID int64, fromAddress string, nonce int64) (*NonceAllocation, error)
	GetNextAvailableNonce(chainID int64, fromAddress string) (int64, error)
	MarkUsed(chainID int64, fromAddress string, nonce int64) error
	MarkExpired(chainID int64, fromAddress string, nonce int64) error
	CleanupExpired() (int64, error)
}

// nonceAllocationRepository implements NonceAllocationRepository
type nonceAllocationRepository struct {
	db *sql.DB
}

// NewNonceAllocationRepository creates a new NonceAllocationRepository
func NewNonceAllocationRepository(db *sql.DB) NonceAllocationRepository {
	return &nonceAllocationRepository{db: db}
}

func (r *nonceAllocationRepository) Create(a *NonceAllocation) error {
	query := `
		INSERT INTO nonce_allocations (chain_id, from_address, nonce, withdrawal_id, status, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	a.CreatedAt = time.Now()
	var withdrawalID sql.NullInt64
	if a.WithdrawalID > 0 {
		withdrawalID = sql.NullInt64{Int64: a.WithdrawalID, Valid: true}
	}
	return r.db.QueryRow(
		query,
		a.ChainID,
		a.FromAddress,
		a.Nonce,
		withdrawalID,
		a.Status,
		a.ExpiresAt,
		a.CreatedAt,
	).Scan(&a.ID)
}

func (r *nonceAllocationRepository) GetByID(id int64) (*NonceAllocation, error) {
	query := `
		SELECT id, chain_id, from_address, nonce, withdrawal_id, status, expires_at, created_at
		FROM nonce_allocations WHERE id = $1
	`
	a := &NonceAllocation{}
	var withdrawalID sql.NullInt64
	err := r.db.QueryRow(query, id).Scan(
		&a.ID,
		&a.ChainID,
		&a.FromAddress,
		&a.Nonce,
		&withdrawalID,
		&a.Status,
		&a.ExpiresAt,
		&a.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	if withdrawalID.Valid {
		a.WithdrawalID = withdrawalID.Int64
	}
	return a, nil
}

func (r *nonceAllocationRepository) GetByChainIDAndAddressAndNonce(chainID int64, fromAddress string, nonce int64) (*NonceAllocation, error) {
	query := `
		SELECT id, chain_id, from_address, nonce, withdrawal_id, status, expires_at, created_at
		FROM nonce_allocations WHERE chain_id = $1 AND from_address = $2 AND nonce = $3
	`
	a := &NonceAllocation{}
	var withdrawalID sql.NullInt64
	err := r.db.QueryRow(query, chainID, fromAddress, nonce).Scan(
		&a.ID,
		&a.ChainID,
		&a.FromAddress,
		&a.Nonce,
		&withdrawalID,
		&a.Status,
		&a.ExpiresAt,
		&a.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	if withdrawalID.Valid {
		a.WithdrawalID = withdrawalID.Int64
	}
	return a, nil
}

func (r *nonceAllocationRepository) GetNextAvailableNonce(chainID int64, fromAddress string) (int64, error) {
	query := `
		SELECT COALESCE(MAX(nonce), -1) + 1
		FROM nonce_allocations
		WHERE chain_id = $1 AND from_address = $2
	`
	var nonce int64
	err := r.db.QueryRow(query, chainID, fromAddress).Scan(&nonce)
	return nonce, err
}

func (r *nonceAllocationRepository) MarkUsed(chainID int64, fromAddress string, nonce int64) error {
	query := `
		UPDATE nonce_allocations SET status = $4
		WHERE chain_id = $1 AND from_address = $2 AND nonce = $3
	`
	_, err := r.db.Exec(query, chainID, fromAddress, nonce, NonceStatusUsed)
	return err
}

func (r *nonceAllocationRepository) MarkExpired(chainID int64, fromAddress string, nonce int64) error {
	query := `
		UPDATE nonce_allocations SET status = $4
		WHERE chain_id = $1 AND from_address = $2 AND nonce = $3
	`
	_, err := r.db.Exec(query, chainID, fromAddress, nonce, NonceStatusExpired)
	return err
}

func (r *nonceAllocationRepository) CleanupExpired() (int64, error) {
	query := `
		UPDATE nonce_allocations SET status = $1
		WHERE expires_at < $2 AND status = $3
	`
	result, err := r.db.Exec(query, NonceStatusExpired, time.Now(), NonceStatusAllocated)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
