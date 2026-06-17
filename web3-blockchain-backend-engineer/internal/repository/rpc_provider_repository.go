package repository

import (
	"database/sql"
	"time"
)

// RPCProvider represents the rpc_providers table
type RPCProvider struct {
	ID           int64     `json:"id"`
	ChainID      int64     `json:"chain_id"`
	Name         string    `json:"name"`
	URL          string    `json:"url"`
	Weight       int       `json:"weight"`
	IsActive     bool      `json:"is_active"`
	LastError    string    `json:"last_error,omitempty"`
	LastCheckedAt time.Time `json:"last_checked_at,omitempty"`
}

// RPCProviderRepository defines the interface for RPC provider data access
type RPCProviderRepository interface {
	Create(provider *RPCProvider) error
	GetByID(id int64) (*RPCProvider, error)
	GetByChainID(chainID int64) ([]*RPCProvider, error)
	GetActiveByChainID(chainID int64) ([]*RPCProvider, error)
	Update(provider *RPCProvider) error
	UpdateLastError(id int64, err string) error
	Delete(id int64) error
}

// rpcProviderRepository implements RPCProviderRepository
type rpcProviderRepository struct {
	db *sql.DB
}

// NewRPCProviderRepository creates a new RPCProviderRepository
func NewRPCProviderRepository(db *sql.DB) RPCProviderRepository {
	return &rpcProviderRepository{db: db}
}

func (r *rpcProviderRepository) Create(provider *RPCProvider) error {
	query := `
		INSERT INTO rpc_providers (chain_id, name, url, weight, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	return r.db.QueryRow(
		query,
		provider.ChainID,
		provider.Name,
		provider.URL,
		provider.Weight,
		provider.IsActive,
	).Scan(&provider.ID)
}

func (r *rpcProviderRepository) GetByID(id int64) (*RPCProvider, error) {
	query := `
		SELECT id, chain_id, name, url, weight, is_active, last_error, last_checked_at
		FROM rpc_providers WHERE id = $1
	`
	p := &RPCProvider{}
	var lastError sql.NullString
	var lastCheckedAt sql.NullTime
	err := r.db.QueryRow(query, id).Scan(
		&p.ID,
		&p.ChainID,
		&p.Name,
		&p.URL,
		&p.Weight,
		&p.IsActive,
		&lastError,
		&lastCheckedAt,
	)
	if err != nil {
		return nil, err
	}
	if lastError.Valid {
		p.LastError = lastError.String
	}
	if lastCheckedAt.Valid {
		p.LastCheckedAt = lastCheckedAt.Time
	}
	return p, nil
}

func (r *rpcProviderRepository) GetByChainID(chainID int64) ([]*RPCProvider, error) {
	query := `
		SELECT id, chain_id, name, url, weight, is_active, last_error, last_checked_at
		FROM rpc_providers WHERE chain_id = $1 ORDER BY weight DESC
	`
	rows, err := r.db.Query(query, chainID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanProviders(rows, nil)
}

func (r *rpcProviderRepository) GetActiveByChainID(chainID int64) ([]*RPCProvider, error) {
	query := `
		SELECT id, chain_id, name, url, weight, is_active, last_error, last_checked_at
		FROM rpc_providers WHERE chain_id = $1 AND is_active = true ORDER BY weight DESC
	`
	rows, err := r.db.Query(query, chainID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanProviders(rows, nil)
}

func (r *rpcProviderRepository) Update(provider *RPCProvider) error {
	query := `
		UPDATE rpc_providers
		SET name = $2, url = $3, weight = $4, is_active = $5
		WHERE id = $1
	`
	_, err := r.db.Exec(
		query,
		provider.ID,
		provider.Name,
		provider.URL,
		provider.Weight,
		provider.IsActive,
	)
	return err
}

func (r *rpcProviderRepository) UpdateLastError(id int64, errMsg string) error {
	query := `
		UPDATE rpc_providers
		SET last_error = $2, last_checked_at = $3
		WHERE id = $1
	`
	_, err := r.db.Exec(query, id, errMsg, time.Now())
	return err
}

func (r *rpcProviderRepository) Delete(id int64) error {
	query := `DELETE FROM rpc_providers WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *rpcProviderRepository) scanProviders(rows *sql.Rows, err error) ([]*RPCProvider, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []*RPCProvider
	for rows.Next() {
		p := &RPCProvider{}
		var lastError sql.NullString
		var lastCheckedAt sql.NullTime
		err := rows.Scan(
			&p.ID,
			&p.ChainID,
			&p.Name,
			&p.URL,
			&p.Weight,
			&p.IsActive,
			&lastError,
			&lastCheckedAt,
		)
		if err != nil {
			return nil, err
		}
		if lastError.Valid {
			p.LastError = lastError.String
		}
		if lastCheckedAt.Valid {
			p.LastCheckedAt = lastCheckedAt.Time
		}
		providers = append(providers, p)
	}
	return providers, rows.Err()
}
