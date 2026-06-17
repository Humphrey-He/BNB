package repository

import (
	"database/sql"
	"time"
)

// Block represents the blocks table
type Block struct {
	ID          int64     `json:"id"`
	ChainID     int64     `json:"chain_id"`
	BlockNumber int64     `json:"block_number"`
	BlockHash   string    `json:"block_hash"`
	ParentHash  string    `json:"parent_hash"`
	BlockTime   time.Time `json:"block_time,omitempty"`
	IsOrphaned  bool      `json:"is_orphaned"`
	ScannedAt   time.Time `json:"scanned_at"`
}

// BlockRepository defines the interface for block data access
type BlockRepository interface {
	Create(block *Block) error
	GetByID(id int64) (*Block, error)
	GetByChainIDAndBlockNumber(chainID int64, blockNumber int64) (*Block, error)
	GetByChainIDAndBlockHash(chainID int64, blockHash string) (*Block, error)
	Update(block *Block) error
	Delete(id int64) error
	List(limit int) ([]*Block, error)
	ListByChainID(chainID int64, limit int) ([]*Block, error)
	MarkOrphaned(chainID int64, blockNumber int64) error
}

// blockRepository implements BlockRepository
type blockRepository struct {
	db *sql.DB
}

// NewBlockRepository creates a new BlockRepository
func NewBlockRepository(db *sql.DB) BlockRepository {
	return &blockRepository{db: db}
}

func (r *blockRepository) Create(block *Block) error {
	query := `
		INSERT INTO blocks (chain_id, block_number, block_hash, parent_hash, block_time, is_orphaned, scanned_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	return r.db.QueryRow(
		query,
		block.ChainID,
		block.BlockNumber,
		block.BlockHash,
		block.ParentHash,
		block.BlockTime,
		block.IsOrphaned,
		block.ScannedAt,
	).Scan(&block.ID)
}

func (r *blockRepository) GetByID(id int64) (*Block, error) {
	query := `
		SELECT id, chain_id, block_number, block_hash, parent_hash, block_time, is_orphaned, scanned_at
		FROM blocks WHERE id = $1
	`
	block := &Block{}
	var blockTime sql.NullTime
	err := r.db.QueryRow(query, id).Scan(
		&block.ID,
		&block.ChainID,
		&block.BlockNumber,
		&block.BlockHash,
		&block.ParentHash,
		&blockTime,
		&block.IsOrphaned,
		&block.ScannedAt,
	)
	if err != nil {
		return nil, err
	}
	if blockTime.Valid {
		block.BlockTime = blockTime.Time
	}
	return block, nil
}

func (r *blockRepository) GetByChainIDAndBlockNumber(chainID int64, blockNumber int64) (*Block, error) {
	query := `
		SELECT id, chain_id, block_number, block_hash, parent_hash, block_time, is_orphaned, scanned_at
		FROM blocks WHERE chain_id = $1 AND block_number = $2
		ORDER BY id DESC LIMIT 1
	`
	block := &Block{}
	var blockTime sql.NullTime
	err := r.db.QueryRow(query, chainID, blockNumber).Scan(
		&block.ID,
		&block.ChainID,
		&block.BlockNumber,
		&block.BlockHash,
		&block.ParentHash,
		&blockTime,
		&block.IsOrphaned,
		&block.ScannedAt,
	)
	if err != nil {
		return nil, err
	}
	if blockTime.Valid {
		block.BlockTime = blockTime.Time
	}
	return block, nil
}

func (r *blockRepository) GetByChainIDAndBlockHash(chainID int64, blockHash string) (*Block, error) {
	query := `
		SELECT id, chain_id, block_number, block_hash, parent_hash, block_time, is_orphaned, scanned_at
		FROM blocks WHERE chain_id = $1 AND block_hash = $2
	`
	block := &Block{}
	var blockTime sql.NullTime
	err := r.db.QueryRow(query, chainID, blockHash).Scan(
		&block.ID,
		&block.ChainID,
		&block.BlockNumber,
		&block.BlockHash,
		&block.ParentHash,
		&blockTime,
		&block.IsOrphaned,
		&block.ScannedAt,
	)
	if err != nil {
		return nil, err
	}
	if blockTime.Valid {
		block.BlockTime = blockTime.Time
	}
	return block, nil
}

func (r *blockRepository) Update(block *Block) error {
	query := `
		UPDATE blocks
		SET block_hash = $2, parent_hash = $3, block_time = $4, is_orphaned = $5
		WHERE id = $1
	`
	_, err := r.db.Exec(
		query,
		block.ID,
		block.BlockHash,
		block.ParentHash,
		block.BlockTime,
		block.IsOrphaned,
	)
	return err
}

func (r *blockRepository) Delete(id int64) error {
	query := `DELETE FROM blocks WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *blockRepository) List(limit int) ([]*Block, error) {
	query := `
		SELECT id, chain_id, block_number, block_hash, parent_hash, block_time, is_orphaned, scanned_at
		FROM blocks ORDER BY id DESC LIMIT $1
	`
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocks []*Block
	for rows.Next() {
		block := &Block{}
		var blockTime sql.NullTime
		err := rows.Scan(
			&block.ID,
			&block.ChainID,
			&block.BlockNumber,
			&block.BlockHash,
			&block.ParentHash,
			&blockTime,
			&block.IsOrphaned,
			&block.ScannedAt,
		)
		if err != nil {
			return nil, err
		}
		if blockTime.Valid {
			block.BlockTime = blockTime.Time
		}
		blocks = append(blocks, block)
	}
	return blocks, rows.Err()
}

func (r *blockRepository) ListByChainID(chainID int64, limit int) ([]*Block, error) {
	query := `
		SELECT id, chain_id, block_number, block_hash, parent_hash, block_time, is_orphaned, scanned_at
		FROM blocks WHERE chain_id = $1 ORDER BY block_number DESC LIMIT $2
	`
	rows, err := r.db.Query(query, chainID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocks []*Block
	for rows.Next() {
		block := &Block{}
		var blockTime sql.NullTime
		err := rows.Scan(
			&block.ID,
			&block.ChainID,
			&block.BlockNumber,
			&block.BlockHash,
			&block.ParentHash,
			&blockTime,
			&block.IsOrphaned,
			&block.ScannedAt,
		)
		if err != nil {
			return nil, err
		}
		if blockTime.Valid {
			block.BlockTime = blockTime.Time
		}
		blocks = append(blocks, block)
	}
	return blocks, rows.Err()
}

func (r *blockRepository) MarkOrphaned(chainID int64, blockNumber int64) error {
	query := `
		UPDATE blocks SET is_orphaned = true
		WHERE chain_id = $1 AND block_number >= $2
	`
	_, err := r.db.Exec(query, chainID, blockNumber)
	return err
}
