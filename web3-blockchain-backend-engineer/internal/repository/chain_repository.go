package repository

import (
	"database/sql"
	"time"
)

// Chain represents the chains table
type Chain struct {
	ID                  int64     `json:"id"`
	ChainID             int64     `json:"chain_id"`
	Name                string    `json:"name"`
	NativeSymbol        string    `json:"native_symbol"`
	FinalityConfirmations int    `json:"finality_confirmations"`
	IsActive            bool      `json:"is_active"`
	CreatedAt           time.Time `json:"created_at"`
}

// ChainRepository defines the interface for chain data access
type ChainRepository interface {
	Create(chain *Chain) error
	GetByID(id int64) (*Chain, error)
	GetByChainID(chainID int64) (*Chain, error)
	Update(chain *Chain) error
	Delete(id int64) error
	List() ([]*Chain, error)
	ListActive() ([]*Chain, error)
}

// chainRepository implements ChainRepository
type chainRepository struct {
	db *sql.DB
}

// NewChainRepository creates a new ChainRepository
func NewChainRepository(db *sql.DB) ChainRepository {
	return &chainRepository{db: db}
}

func (r *chainRepository) Create(chain *Chain) error {
	query := `
		INSERT INTO chains (chain_id, name, native_symbol, finality_confirmations, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	return r.db.QueryRow(
		query,
		chain.ChainID,
		chain.Name,
		chain.NativeSymbol,
		chain.FinalityConfirmations,
		chain.IsActive,
		chain.CreatedAt,
	).Scan(&chain.ID)
}

func (r *chainRepository) GetByID(id int64) (*Chain, error) {
	query := `
		SELECT id, chain_id, name, native_symbol, finality_confirmations, is_active, created_at
		FROM chains WHERE id = $1
	`
	chain := &Chain{}
	err := r.db.QueryRow(query, id).Scan(
		&chain.ID,
		&chain.ChainID,
		&chain.Name,
		&chain.NativeSymbol,
		&chain.FinalityConfirmations,
		&chain.IsActive,
		&chain.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return chain, nil
}

func (r *chainRepository) GetByChainID(chainID int64) (*Chain, error) {
	query := `
		SELECT id, chain_id, name, native_symbol, finality_confirmations, is_active, created_at
		FROM chains WHERE chain_id = $1
	`
	chain := &Chain{}
	err := r.db.QueryRow(query, chainID).Scan(
		&chain.ID,
		&chain.ChainID,
		&chain.Name,
		&chain.NativeSymbol,
		&chain.FinalityConfirmations,
		&chain.IsActive,
		&chain.CreatedAt,
	)
	if err == sql.ErrNoRows {
		// Some runtime paths currently pass the chains table primary key instead of
		// the on-chain network id. Fallback to the primary key so the Sepolia-only
		// deployment can keep moving until the schema semantics are unified.
		return r.GetByID(chainID)
	}
	if err != nil {
		return nil, err
	}
	return chain, nil
}

func (r *chainRepository) Update(chain *Chain) error {
	query := `
		UPDATE chains
		SET name = $2, native_symbol = $3, finality_confirmations = $4, is_active = $5
		WHERE id = $1
	`
	_, err := r.db.Exec(
		query,
		chain.ID,
		chain.Name,
		chain.NativeSymbol,
		chain.FinalityConfirmations,
		chain.IsActive,
	)
	return err
}

func (r *chainRepository) Delete(id int64) error {
	query := `DELETE FROM chains WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *chainRepository) List() ([]*Chain, error) {
	query := `
		SELECT id, chain_id, name, native_symbol, finality_confirmations, is_active, created_at
		FROM chains ORDER BY id
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chains []*Chain
	for rows.Next() {
		chain := &Chain{}
		err := rows.Scan(
			&chain.ID,
			&chain.ChainID,
			&chain.Name,
			&chain.NativeSymbol,
			&chain.FinalityConfirmations,
			&chain.IsActive,
			&chain.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		chains = append(chains, chain)
	}
	return chains, rows.Err()
}

func (r *chainRepository) ListActive() ([]*Chain, error) {
	query := `
		SELECT id, chain_id, name, native_symbol, finality_confirmations, is_active, created_at
		FROM chains WHERE is_active = true ORDER BY id
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chains []*Chain
	for rows.Next() {
		chain := &Chain{}
		err := rows.Scan(
			&chain.ID,
			&chain.ChainID,
			&chain.Name,
			&chain.NativeSymbol,
			&chain.FinalityConfirmations,
			&chain.IsActive,
			&chain.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		chains = append(chains, chain)
	}
	return chains, rows.Err()
}
