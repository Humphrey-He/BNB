package repository

import (
	"database/sql"
	"time"
)

// DepositStatus represents deposit status constants
type DepositStatus string

const (
	DepositStatusDetected            DepositStatus = "detected"
	DepositStatusPendingConfirmation DepositStatus = "pending_confirmation"
	DepositStatusConfirmed           DepositStatus = "confirmed"
	DepositStatusOrphaned            DepositStatus = "orphaned"
	DepositStatusFailed              DepositStatus = "failed"
)

// Deposit represents the deposits table
type Deposit struct {
	ID                      int64         `json:"id"`
	ChainID                 int64         `json:"chain_id"`
	TokenID                 int64         `json:"token_id"`
	TxHash                  string        `json:"tx_hash"`
	LogIndex                int           `json:"log_index"`
	FromAddress             string        `json:"from_address"`
	ToAddress               string        `json:"to_address"`
	Amount                  string        `json:"amount"`
	BlockNumber             int64         `json:"block_number"`
	Status                  DepositStatus `json:"status"`
	Confirmations           int           `json:"confirmations"`
	TargetConfirmations     int           `json:"target_confirmations"`
	IdempotencyKey          string        `json:"idempotency_key"`
	ProcessedAt             *time.Time    `json:"processed_at,omitempty"`
	ConfirmedEventPublished bool          `json:"confirmed_event_published"`
	CreatedAt               time.Time     `json:"created_at"`
	UpdatedAt               time.Time     `json:"updated_at"`
}

// DepositRepository defines the interface for deposit data access
type DepositRepository interface {
	Create(deposit *Deposit) error
	GetByID(id int64) (*Deposit, error)
	GetByIdempotencyKey(key string) (*Deposit, error)
	GetByChainIDTxHashAndLogIndex(chainID int64, txHash string, logIndex int) (*Deposit, error)
	Update(deposit *Deposit) error
	UpdateStatus(id int64, status DepositStatus) error
	IncrementConfirmations(id int64) error
	SetConfirmations(id int64, confirmations int) error
	// ConfirmWithCondition updates status to confirmed only if current status is pending_confirmation
	// and confirmations >= targetConfirmations. Returns (updated bool, error)
	ConfirmWithCondition(id int64, confirmations, targetConfirmations int) (bool, error)
	MarkConfirmedEventPublished(id int64) error
	Delete(id int64) error
	List(limit int) ([]*Deposit, error)
	ListByChainID(chainID int64, limit int) ([]*Deposit, error)
	ListByAddress(chainID int64, address string, limit int) ([]*Deposit, error)
	ListByStatus(status DepositStatus, limit int) ([]*Deposit, error)
	ListConfirmedUnpublished(limit int) ([]*Deposit, error)
	ListByBlockNumber(chainID int64, blockNumber int64) ([]*Deposit, error)
}

// depositRepository implements DepositRepository
type depositRepository struct {
	db *sql.DB
}

// NewDepositRepository creates a new DepositRepository
func NewDepositRepository(db *sql.DB) DepositRepository {
	return &depositRepository{db: db}
}

func (r *depositRepository) Create(deposit *Deposit) error {
	query := `
		INSERT INTO deposits (chain_id, token_id, tx_hash, log_index, from_address, to_address, amount, block_number, status, confirmations, target_confirmations, idempotency_key, processed_at, confirmed_event_published, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id
	`
	now := time.Now()
	return r.db.QueryRow(
		query,
		deposit.ChainID,
		deposit.TokenID,
		deposit.TxHash,
		deposit.LogIndex,
		deposit.FromAddress,
		deposit.ToAddress,
		deposit.Amount,
		deposit.BlockNumber,
		deposit.Status,
		deposit.Confirmations,
		deposit.TargetConfirmations,
		deposit.IdempotencyKey,
		deposit.ProcessedAt,
		deposit.ConfirmedEventPublished,
		now,
		now,
	).Scan(&deposit.ID)
}

func (r *depositRepository) GetByID(id int64) (*Deposit, error) {
	query := `
		SELECT id, chain_id, token_id, tx_hash, log_index, from_address, to_address, amount, block_number, status, confirmations, target_confirmations, idempotency_key, processed_at, confirmed_event_published, created_at, updated_at
		FROM deposits WHERE id = $1
	`
	deposit := &Deposit{}
	err := r.db.QueryRow(query, id).Scan(
		&deposit.ID,
		&deposit.ChainID,
		&deposit.TokenID,
		&deposit.TxHash,
		&deposit.LogIndex,
		&deposit.FromAddress,
		&deposit.ToAddress,
		&deposit.Amount,
		&deposit.BlockNumber,
		&deposit.Status,
		&deposit.Confirmations,
		&deposit.TargetConfirmations,
		&deposit.IdempotencyKey,
		&deposit.ProcessedAt,
		&deposit.ConfirmedEventPublished,
		&deposit.CreatedAt,
		&deposit.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return deposit, nil
}

func (r *depositRepository) GetByIdempotencyKey(key string) (*Deposit, error) {
	query := `
		SELECT id, chain_id, token_id, tx_hash, log_index, from_address, to_address, amount, block_number, status, confirmations, target_confirmations, idempotency_key, processed_at, confirmed_event_published, created_at, updated_at
		FROM deposits WHERE idempotency_key = $1
	`
	deposit := &Deposit{}
	err := r.db.QueryRow(query, key).Scan(
		&deposit.ID,
		&deposit.ChainID,
		&deposit.TokenID,
		&deposit.TxHash,
		&deposit.LogIndex,
		&deposit.FromAddress,
		&deposit.ToAddress,
		&deposit.Amount,
		&deposit.BlockNumber,
		&deposit.Status,
		&deposit.Confirmations,
		&deposit.TargetConfirmations,
		&deposit.IdempotencyKey,
		&deposit.ProcessedAt,
		&deposit.ConfirmedEventPublished,
		&deposit.CreatedAt,
		&deposit.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return deposit, nil
}

func (r *depositRepository) GetByChainIDTxHashAndLogIndex(chainID int64, txHash string, logIndex int) (*Deposit, error) {
	query := `
		SELECT id, chain_id, token_id, tx_hash, log_index, from_address, to_address, amount, block_number, status, confirmations, target_confirmations, idempotency_key, processed_at, confirmed_event_published, created_at, updated_at
		FROM deposits WHERE chain_id = $1 AND tx_hash = $2 AND log_index = $3
	`
	deposit := &Deposit{}
	err := r.db.QueryRow(query, chainID, txHash, logIndex).Scan(
		&deposit.ID,
		&deposit.ChainID,
		&deposit.TokenID,
		&deposit.TxHash,
		&deposit.LogIndex,
		&deposit.FromAddress,
		&deposit.ToAddress,
		&deposit.Amount,
		&deposit.BlockNumber,
		&deposit.Status,
		&deposit.Confirmations,
		&deposit.TargetConfirmations,
		&deposit.IdempotencyKey,
		&deposit.ProcessedAt,
		&deposit.ConfirmedEventPublished,
		&deposit.CreatedAt,
		&deposit.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return deposit, nil
}

func (r *depositRepository) Update(deposit *Deposit) error {
	query := `
		UPDATE deposits
		SET token_id = $2, from_address = $3, to_address = $4, amount = $5, block_number = $6, status = $7, confirmations = $8, confirmed_event_published = $9, updated_at = $10
		WHERE id = $1
	`
	deposit.UpdatedAt = time.Now()
	_, err := r.db.Exec(
		query,
		deposit.ID,
		deposit.TokenID,
		deposit.FromAddress,
		deposit.ToAddress,
		deposit.Amount,
		deposit.BlockNumber,
		deposit.Status,
		deposit.Confirmations,
		deposit.ConfirmedEventPublished,
		deposit.UpdatedAt,
	)
	return err
}

func (r *depositRepository) UpdateStatus(id int64, status DepositStatus) error {
	query := `UPDATE deposits SET status = $2, updated_at = $3 WHERE id = $1`
	_, err := r.db.Exec(query, id, status, time.Now())
	return err
}

func (r *depositRepository) IncrementConfirmations(id int64) error {
	query := `UPDATE deposits SET confirmations = confirmations + 1, updated_at = $2 WHERE id = $1`
	_, err := r.db.Exec(query, id, time.Now())
	return err
}

func (r *depositRepository) SetConfirmations(id int64, confirmations int) error {
	query := `UPDATE deposits SET confirmations = $2, updated_at = $3 WHERE id = $1`
	_, err := r.db.Exec(query, id, confirmations, time.Now())
	return err
}

// ConfirmWithCondition updates deposit to confirmed status only if:
// 1. Current status is 'pending_confirmation'
// 2. confirmations >= targetConfirmations
// Returns (true, nil) if updated, (false, nil) if conditions not met, (_, error) on DB error
func (r *depositRepository) ConfirmWithCondition(id int64, confirmations, targetConfirmations int) (bool, error) {
	query := `
		UPDATE deposits
		SET status = $2,
		    confirmations = $3,
		    processed_at = $4,
		    confirmed_event_published = false,
		    updated_at = $4
		WHERE id = $1
		  AND status = $5
		  AND $3 >= $6
	`
	now := time.Now()
	result, err := r.db.Exec(query,
		id,
		DepositStatusConfirmed,
		confirmations,
		now,
		DepositStatusPendingConfirmation,
		targetConfirmations,
	)
	if err != nil {
		return false, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rowsAffected > 0, nil
}

func (r *depositRepository) MarkConfirmedEventPublished(id int64) error {
	query := `
		UPDATE deposits
		SET confirmed_event_published = true,
		    updated_at = $2
		WHERE id = $1
		  AND status = $3
	`
	_, err := r.db.Exec(query, id, time.Now(), DepositStatusConfirmed)
	return err
}

func (r *depositRepository) Delete(id int64) error {
	query := `DELETE FROM deposits WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *depositRepository) List(limit int) ([]*Deposit, error) {
	query := `
		SELECT id, chain_id, token_id, tx_hash, log_index, from_address, to_address, amount, block_number, status, confirmations, target_confirmations, idempotency_key, processed_at, confirmed_event_published, created_at, updated_at
		FROM deposits ORDER BY id DESC LIMIT $1
	`
	return r.scanDeposits(r.db.Query(query, limit))
}

func (r *depositRepository) ListByChainID(chainID int64, limit int) ([]*Deposit, error) {
	query := `
		SELECT id, chain_id, token_id, tx_hash, log_index, from_address, to_address, amount, block_number, status, confirmations, target_confirmations, idempotency_key, processed_at, confirmed_event_published, created_at, updated_at
		FROM deposits WHERE chain_id = $1 ORDER BY id DESC LIMIT $2
	`
	return r.scanDeposits(r.db.Query(query, chainID, limit))
}

func (r *depositRepository) ListByAddress(chainID int64, address string, limit int) ([]*Deposit, error) {
	query := `
		SELECT id, chain_id, token_id, tx_hash, log_index, from_address, to_address, amount, block_number, status, confirmations, target_confirmations, idempotency_key, processed_at, confirmed_event_published, created_at, updated_at
		FROM deposits WHERE chain_id = $1 AND to_address = $2 ORDER BY id DESC LIMIT $3
	`
	return r.scanDeposits(r.db.Query(query, chainID, address, limit))
}

func (r *depositRepository) ListByStatus(status DepositStatus, limit int) ([]*Deposit, error) {
	query := `
		SELECT id, chain_id, token_id, tx_hash, log_index, from_address, to_address, amount, block_number, status, confirmations, target_confirmations, idempotency_key, processed_at, confirmed_event_published, created_at, updated_at
		FROM deposits WHERE status = $1 ORDER BY id DESC LIMIT $2
	`
	return r.scanDeposits(r.db.Query(query, status, limit))
}

func (r *depositRepository) ListConfirmedUnpublished(limit int) ([]*Deposit, error) {
	query := `
		SELECT id, chain_id, token_id, tx_hash, log_index, from_address, to_address, amount, block_number, status, confirmations, target_confirmations, idempotency_key, processed_at, confirmed_event_published, created_at, updated_at
		FROM deposits
		WHERE status = $1 AND confirmed_event_published = false
		ORDER BY id ASC
		LIMIT $2
	`
	return r.scanDeposits(r.db.Query(query, DepositStatusConfirmed, limit))
}

func (r *depositRepository) scanDeposits(rows *sql.Rows, err error) ([]*Deposit, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deposits []*Deposit
	for rows.Next() {
		d := &Deposit{}
		err := rows.Scan(
			&d.ID,
			&d.ChainID,
			&d.TokenID,
			&d.TxHash,
			&d.LogIndex,
			&d.FromAddress,
			&d.ToAddress,
			&d.Amount,
			&d.BlockNumber,
			&d.Status,
			&d.Confirmations,
			&d.TargetConfirmations,
			&d.IdempotencyKey,
			&d.ProcessedAt,
			&d.ConfirmedEventPublished,
			&d.CreatedAt,
			&d.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		deposits = append(deposits, d)
	}
	return deposits, rows.Err()
}

func (r *depositRepository) ListByBlockNumber(chainID int64, blockNumber int64) ([]*Deposit, error) {
	query := `
		SELECT id, chain_id, token_id, tx_hash, log_index, from_address, to_address, amount, block_number, status, confirmations, target_confirmations, idempotency_key, processed_at, confirmed_event_published, created_at, updated_at
		FROM deposits WHERE chain_id = $1 AND block_number = $2 ORDER BY id
	`
	return r.scanDeposits(r.db.Query(query, chainID, blockNumber))
}
