package repository

import (
	"database/sql"
	"time"
)

// WithdrawalStatus represents withdrawal status constants
type WithdrawalStatus string

const (
	WithdrawalStatusCreated      WithdrawalStatus = "created"
	WithdrawalStatusRiskChecking WithdrawalStatus = "risk_checking"
	WithdrawalStatusManualReview WithdrawalStatus = "manual_review"
	WithdrawalStatusApproved     WithdrawalStatus = "approved"
	WithdrawalStatusSigning      WithdrawalStatus = "signing"
	WithdrawalStatusBroadcasting WithdrawalStatus = "broadcasting"
	WithdrawalStatusBroadcasted  WithdrawalStatus = "broadcasted"
	WithdrawalStatusConfirmed    WithdrawalStatus = "confirmed"
	WithdrawalStatusFailed       WithdrawalStatus = "failed"
	WithdrawalStatusCanceled     WithdrawalStatus = "canceled"
)

// Withdrawal represents the withdrawals table
type Withdrawal struct {
	ID             int64            `json:"id"`
	ChainID        int64            `json:"chain_id"`
	TokenID        int64            `json:"token_id"`
	FromAddress    string           `json:"from_address"`
	ToAddress      string           `json:"to_address"`
	Amount         string           `json:"amount"`
	Status         WithdrawalStatus `json:"status"`
	TxHash         string           `json:"tx_hash,omitempty"`
	Nonce          int64            `json:"nonce,omitempty"`
	IdempotencyKey string           `json:"idempotency_key"`
	FailureReason  string           `json:"failure_reason,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

// WithdrawalRepository defines the interface for withdrawal data access
type WithdrawalRepository interface {
	Create(withdrawal *Withdrawal) error
	GetByID(id int64) (*Withdrawal, error)
	GetByIdempotencyKey(key string) (*Withdrawal, error)
	GetByTxHash(chainID int64, txHash string) (*Withdrawal, error)
	Update(withdrawal *Withdrawal) error
	UpdateStatus(id int64, status WithdrawalStatus) error
	Delete(id int64) error
	List(limit int) ([]*Withdrawal, error)
	ListByChainID(chainID int64, limit int) ([]*Withdrawal, error)
	ListByFromAddress(chainID int64, fromAddress string, limit int) ([]*Withdrawal, error)
	ListByStatus(status WithdrawalStatus, limit int) ([]*Withdrawal, error)
}

// withdrawalRepository implements WithdrawalRepository
type withdrawalRepository struct {
	db *sql.DB
}

// NewWithdrawalRepository creates a new WithdrawalRepository
func NewWithdrawalRepository(db *sql.DB) WithdrawalRepository {
	return &withdrawalRepository{db: db}
}

func (r *withdrawalRepository) Create(withdrawal *Withdrawal) error {
	query := `
		INSERT INTO withdrawals (chain_id, token_id, from_address, to_address, amount, status, tx_hash, nonce, idempotency_key, failure_reason, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
	`
	now := time.Now()
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
	return r.db.QueryRow(
		query,
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
		now,
		now,
	).Scan(&withdrawal.ID)
}

func (r *withdrawalRepository) GetByID(id int64) (*Withdrawal, error) {
	query := `
		SELECT id, chain_id, token_id, from_address, to_address, amount, status, tx_hash, nonce, idempotency_key, failure_reason, created_at, updated_at
		FROM withdrawals WHERE id = $1
	`
	w := &Withdrawal{}
	var txHash, failureReason sql.NullString
	var nonce sql.NullInt64
	err := r.db.QueryRow(query, id).Scan(
		&w.ID,
		&w.ChainID,
		&w.TokenID,
		&w.FromAddress,
		&w.ToAddress,
		&w.Amount,
		&w.Status,
		&txHash,
		&nonce,
		&w.IdempotencyKey,
		&failureReason,
		&w.CreatedAt,
		&w.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if txHash.Valid {
		w.TxHash = txHash.String
	}
	if nonce.Valid {
		w.Nonce = nonce.Int64
	}
	if failureReason.Valid {
		w.FailureReason = failureReason.String
	}
	return w, nil
}

func (r *withdrawalRepository) GetByIdempotencyKey(key string) (*Withdrawal, error) {
	query := `
		SELECT id, chain_id, token_id, from_address, to_address, amount, status, tx_hash, nonce, idempotency_key, failure_reason, created_at, updated_at
		FROM withdrawals WHERE idempotency_key = $1
	`
	w := &Withdrawal{}
	var txHash, failureReason sql.NullString
	var nonce sql.NullInt64
	err := r.db.QueryRow(query, key).Scan(
		&w.ID,
		&w.ChainID,
		&w.TokenID,
		&w.FromAddress,
		&w.ToAddress,
		&w.Amount,
		&w.Status,
		&txHash,
		&nonce,
		&w.IdempotencyKey,
		&failureReason,
		&w.CreatedAt,
		&w.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if txHash.Valid {
		w.TxHash = txHash.String
	}
	if nonce.Valid {
		w.Nonce = nonce.Int64
	}
	if failureReason.Valid {
		w.FailureReason = failureReason.String
	}
	return w, nil
}

func (r *withdrawalRepository) GetByTxHash(chainID int64, txHash string) (*Withdrawal, error) {
	query := `
		SELECT id, chain_id, token_id, from_address, to_address, amount, status, tx_hash, nonce, idempotency_key, failure_reason, created_at, updated_at
		FROM withdrawals WHERE chain_id = $1 AND tx_hash = $2
	`
	w := &Withdrawal{}
	var dbTxHash, failureReason sql.NullString
	var nonce sql.NullInt64
	err := r.db.QueryRow(query, chainID, txHash).Scan(
		&w.ID,
		&w.ChainID,
		&w.TokenID,
		&w.FromAddress,
		&w.ToAddress,
		&w.Amount,
		&w.Status,
		&dbTxHash,
		&nonce,
		&w.IdempotencyKey,
		&failureReason,
		&w.CreatedAt,
		&w.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if dbTxHash.Valid {
		w.TxHash = dbTxHash.String
	}
	if nonce.Valid {
		w.Nonce = nonce.Int64
	}
	if failureReason.Valid {
		w.FailureReason = failureReason.String
	}
	return w, nil
}

func (r *withdrawalRepository) Update(w *Withdrawal) error {
	query := `
		UPDATE withdrawals
		SET token_id = $2, from_address = $3, to_address = $4, amount = $5, status = $6, tx_hash = $7, nonce = $8, failure_reason = $9, updated_at = $10
		WHERE id = $1
	`
	w.UpdatedAt = time.Now()
	var txHash sql.NullString
	if w.TxHash != "" {
		txHash = sql.NullString{String: w.TxHash, Valid: true}
	}
	var nonce sql.NullInt64
	if w.Nonce > 0 {
		nonce = sql.NullInt64{Int64: w.Nonce, Valid: true}
	}
	var failureReason sql.NullString
	if w.FailureReason != "" {
		failureReason = sql.NullString{String: w.FailureReason, Valid: true}
	}
	_, err := r.db.Exec(
		query,
		w.ID,
		w.TokenID,
		w.FromAddress,
		w.ToAddress,
		w.Amount,
		w.Status,
		txHash,
		nonce,
		failureReason,
		w.UpdatedAt,
	)
	return err
}

func (r *withdrawalRepository) UpdateStatus(id int64, status WithdrawalStatus) error {
	query := `UPDATE withdrawals SET status = $2, updated_at = $3 WHERE id = $1`
	_, err := r.db.Exec(query, id, status, time.Now())
	return err
}

func (r *withdrawalRepository) Delete(id int64) error {
	query := `DELETE FROM withdrawals WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *withdrawalRepository) List(limit int) ([]*Withdrawal, error) {
	query := `
		SELECT id, chain_id, token_id, from_address, to_address, amount, status, tx_hash, nonce, idempotency_key, failure_reason, created_at, updated_at
		FROM withdrawals ORDER BY id DESC LIMIT $1
	`
	return r.scanWithdrawals(r.db.Query(query, limit))
}

func (r *withdrawalRepository) ListByChainID(chainID int64, limit int) ([]*Withdrawal, error) {
	query := `
		SELECT id, chain_id, token_id, from_address, to_address, amount, status, tx_hash, nonce, idempotency_key, failure_reason, created_at, updated_at
		FROM withdrawals WHERE chain_id = $1 ORDER BY id DESC LIMIT $2
	`
	return r.scanWithdrawals(r.db.Query(query, chainID, limit))
}

func (r *withdrawalRepository) ListByFromAddress(chainID int64, fromAddress string, limit int) ([]*Withdrawal, error) {
	query := `
		SELECT id, chain_id, token_id, from_address, to_address, amount, status, tx_hash, nonce, idempotency_key, failure_reason, created_at, updated_at
		FROM withdrawals WHERE chain_id = $1 AND from_address = $2 ORDER BY id DESC LIMIT $3
	`
	return r.scanWithdrawals(r.db.Query(query, chainID, fromAddress, limit))
}

func (r *withdrawalRepository) ListByStatus(status WithdrawalStatus, limit int) ([]*Withdrawal, error) {
	query := `
		SELECT id, chain_id, token_id, from_address, to_address, amount, status, tx_hash, nonce, idempotency_key, failure_reason, created_at, updated_at
		FROM withdrawals WHERE status = $1 ORDER BY id DESC LIMIT $2
	`
	return r.scanWithdrawals(r.db.Query(query, status, limit))
}

func (r *withdrawalRepository) scanWithdrawals(rows *sql.Rows, err error) ([]*Withdrawal, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var withdrawals []*Withdrawal
	for rows.Next() {
		w := &Withdrawal{}
		var txHash, failureReason sql.NullString
		var nonce sql.NullInt64
		err := rows.Scan(
			&w.ID,
			&w.ChainID,
			&w.TokenID,
			&w.FromAddress,
			&w.ToAddress,
			&w.Amount,
			&w.Status,
			&txHash,
			&nonce,
			&w.IdempotencyKey,
			&failureReason,
			&w.CreatedAt,
			&w.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if txHash.Valid {
			w.TxHash = txHash.String
		}
		if nonce.Valid {
			w.Nonce = nonce.Int64
		}
		if failureReason.Valid {
			w.FailureReason = failureReason.String
		}
		withdrawals = append(withdrawals, w)
	}
	return withdrawals, rows.Err()
}
