package repository

import (
	"database/sql"
	"time"
)

// ScanCheckpoint represents the scan_checkpoints table
type ScanCheckpoint struct {
	ID               int64     `json:"id"`
	ChainID          int64     `json:"chain_id"`
	LastScannedBlock int64     `json:"last_scanned_block"`
	LastScannedAt    time.Time `json:"last_scanned_at"`
}

// ScanCheckpointRepository defines the interface for scan checkpoint data access
type ScanCheckpointRepository interface {
	Create(checkpoint *ScanCheckpoint) error
	GetByChainID(chainID int64) (*ScanCheckpoint, error)
	List() ([]*ScanCheckpoint, error)
	UpdateLastScannedBlock(chainID int64, blockNumber int64) error
	Delete(id int64) error
}

// scanCheckpointRepository implements ScanCheckpointRepository
type scanCheckpointRepository struct {
	db *sql.DB
}

// NewScanCheckpointRepository creates a new ScanCheckpointRepository
func NewScanCheckpointRepository(db *sql.DB) ScanCheckpointRepository {
	return &scanCheckpointRepository{db: db}
}

func (r *scanCheckpointRepository) Create(checkpoint *ScanCheckpoint) error {
	query := `
		INSERT INTO scan_checkpoints (chain_id, last_scanned_block, last_scanned_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	checkpoint.LastScannedAt = time.Now()
	return r.db.QueryRow(
		query,
		checkpoint.ChainID,
		checkpoint.LastScannedBlock,
		checkpoint.LastScannedAt,
	).Scan(&checkpoint.ID)
}

func (r *scanCheckpointRepository) GetByChainID(chainID int64) (*ScanCheckpoint, error) {
	query := `
		SELECT id, chain_id, last_scanned_block, last_scanned_at
		FROM scan_checkpoints WHERE chain_id = $1
	`
	checkpoint := &ScanCheckpoint{}
	err := r.db.QueryRow(query, chainID).Scan(
		&checkpoint.ID,
		&checkpoint.ChainID,
		&checkpoint.LastScannedBlock,
		&checkpoint.LastScannedAt,
	)
	if err != nil {
		return nil, err
	}
	return checkpoint, nil
}

func (r *scanCheckpointRepository) List() ([]*ScanCheckpoint, error) {
	query := `
		SELECT id, chain_id, last_scanned_block, last_scanned_at
		FROM scan_checkpoints
		ORDER BY chain_id
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checkpoints []*ScanCheckpoint
	for rows.Next() {
		checkpoint := &ScanCheckpoint{}
		if err := rows.Scan(
			&checkpoint.ID,
			&checkpoint.ChainID,
			&checkpoint.LastScannedBlock,
			&checkpoint.LastScannedAt,
		); err != nil {
			return nil, err
		}
		checkpoints = append(checkpoints, checkpoint)
	}
	return checkpoints, rows.Err()
}

func (r *scanCheckpointRepository) UpdateLastScannedBlock(chainID int64, blockNumber int64) error {
	query := `
		UPDATE scan_checkpoints
		SET last_scanned_block = $2, last_scanned_at = $3
		WHERE chain_id = $1
	`
	result, err := r.db.Exec(query, chainID, blockNumber, time.Now())
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *scanCheckpointRepository) Delete(id int64) error {
	query := `DELETE FROM scan_checkpoints WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
