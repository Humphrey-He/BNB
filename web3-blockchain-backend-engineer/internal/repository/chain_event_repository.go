package repository

import (
	"database/sql"
	"time"
)

// ChainEvent represents the chain_events table
type ChainEvent struct {
	ID             int64     `json:"id"`
	ChainID        int64     `json:"chain_id"`
	TxHash         string    `json:"tx_hash"`
	LogIndex       int       `json:"log_index"`
	BlockNumber    int64     `json:"block_number"`
	BlockHash      string    `json:"block_hash"`
	ContractAddress string   `json:"contract_address"`
	EventName      string    `json:"event_name"`
	FromAddress    string    `json:"from_address,omitempty"`
	ToAddress      string    `json:"to_address,omitempty"`
	Amount         string    `json:"amount"`
	IsOrphaned     bool      `json:"is_orphaned"`
	CreatedAt      time.Time `json:"created_at"`
}

// ChainEventRepository defines the interface for chain event data access
type ChainEventRepository interface {
	Create(event *ChainEvent) error
	GetByID(id int64) (*ChainEvent, error)
	GetByChainIDTxHashAndLogIndex(chainID int64, txHash string, logIndex int) (*ChainEvent, error)
	Update(event *ChainEvent) error
	Delete(id int64) error
	List(limit int) ([]*ChainEvent, error)
	ListByChainID(chainID int64, limit int) ([]*ChainEvent, error)
	ListByTxHash(chainID int64, txHash string) ([]*ChainEvent, error)
	MarkOrphaned(id int64) error
}

// chainEventRepository implements ChainEventRepository
type chainEventRepository struct {
	db *sql.DB
}

// NewChainEventRepository creates a new ChainEventRepository
func NewChainEventRepository(db *sql.DB) ChainEventRepository {
	return &chainEventRepository{db: db}
}

func (r *chainEventRepository) Create(event *ChainEvent) error {
	query := `
		INSERT INTO chain_events (chain_id, tx_hash, log_index, block_number, block_hash, contract_address, event_name, from_address, to_address, amount, is_orphaned, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
	`
	return r.db.QueryRow(
		query,
		event.ChainID,
		event.TxHash,
		event.LogIndex,
		event.BlockNumber,
		event.BlockHash,
		event.ContractAddress,
		event.EventName,
		event.FromAddress,
		event.ToAddress,
		event.Amount,
		event.IsOrphaned,
		event.CreatedAt,
	).Scan(&event.ID)
}

func (r *chainEventRepository) GetByID(id int64) (*ChainEvent, error) {
	query := `
		SELECT id, chain_id, tx_hash, log_index, block_number, block_hash, contract_address, event_name, from_address, to_address, amount, is_orphaned, created_at
		FROM chain_events WHERE id = $1
	`
	event := &ChainEvent{}
	var fromAddress, toAddress sql.NullString
	err := r.db.QueryRow(query, id).Scan(
		&event.ID,
		&event.ChainID,
		&event.TxHash,
		&event.LogIndex,
		&event.BlockNumber,
		&event.BlockHash,
		&event.ContractAddress,
		&event.EventName,
		&fromAddress,
		&toAddress,
		&event.Amount,
		&event.IsOrphaned,
		&event.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	event.FromAddress = fromAddress.String
	event.ToAddress = toAddress.String
	return event, nil
}

func (r *chainEventRepository) GetByChainIDTxHashAndLogIndex(chainID int64, txHash string, logIndex int) (*ChainEvent, error) {
	query := `
		SELECT id, chain_id, tx_hash, log_index, block_number, block_hash, contract_address, event_name, from_address, to_address, amount, is_orphaned, created_at
		FROM chain_events WHERE chain_id = $1 AND tx_hash = $2 AND log_index = $3
	`
	event := &ChainEvent{}
	var fromAddress, toAddress sql.NullString
	err := r.db.QueryRow(query, chainID, txHash, logIndex).Scan(
		&event.ID,
		&event.ChainID,
		&event.TxHash,
		&event.LogIndex,
		&event.BlockNumber,
		&event.BlockHash,
		&event.ContractAddress,
		&event.EventName,
		&fromAddress,
		&toAddress,
		&event.Amount,
		&event.IsOrphaned,
		&event.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	event.FromAddress = fromAddress.String
	event.ToAddress = toAddress.String
	return event, nil
}

func (r *chainEventRepository) Update(event *ChainEvent) error {
	query := `
		UPDATE chain_events
		SET block_number = $2, block_hash = $3, from_address = $4, to_address = $5, amount = $6, is_orphaned = $7
		WHERE id = $1
	`
	_, err := r.db.Exec(
		query,
		event.ID,
		event.BlockNumber,
		event.BlockHash,
		event.FromAddress,
		event.ToAddress,
		event.Amount,
		event.IsOrphaned,
	)
	return err
}

func (r *chainEventRepository) Delete(id int64) error {
	query := `DELETE FROM chain_events WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *chainEventRepository) List(limit int) ([]*ChainEvent, error) {
	query := `
		SELECT id, chain_id, tx_hash, log_index, block_number, block_hash, contract_address, event_name, from_address, to_address, amount, is_orphaned, created_at
		FROM chain_events ORDER BY id DESC LIMIT $1
	`
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanChainEvents(rows)
}

func (r *chainEventRepository) ListByChainID(chainID int64, limit int) ([]*ChainEvent, error) {
	query := `
		SELECT id, chain_id, tx_hash, log_index, block_number, block_hash, contract_address, event_name, from_address, to_address, amount, is_orphaned, created_at
		FROM chain_events WHERE chain_id = $1 ORDER BY id DESC LIMIT $2
	`
	rows, err := r.db.Query(query, chainID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanChainEvents(rows)
}

func (r *chainEventRepository) ListByTxHash(chainID int64, txHash string) ([]*ChainEvent, error) {
	query := `
		SELECT id, chain_id, tx_hash, log_index, block_number, block_hash, contract_address, event_name, from_address, to_address, amount, is_orphaned, created_at
		FROM chain_events WHERE chain_id = $1 AND tx_hash = $2 ORDER BY log_index
	`
	rows, err := r.db.Query(query, chainID, txHash)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanChainEvents(rows)
}

func (r *chainEventRepository) MarkOrphaned(id int64) error {
	query := `UPDATE chain_events SET is_orphaned = true WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *chainEventRepository) scanChainEvents(rows *sql.Rows) ([]*ChainEvent, error) {
	var events []*ChainEvent
	for rows.Next() {
		event := &ChainEvent{}
		var fromAddress, toAddress sql.NullString
		err := rows.Scan(
			&event.ID,
			&event.ChainID,
			&event.TxHash,
			&event.LogIndex,
			&event.BlockNumber,
			&event.BlockHash,
			&event.ContractAddress,
			&event.EventName,
			&fromAddress,
			&toAddress,
			&event.Amount,
			&event.IsOrphaned,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		event.FromAddress = fromAddress.String
		event.ToAddress = toAddress.String
		events = append(events, event)
	}
	return events, rows.Err()
}
