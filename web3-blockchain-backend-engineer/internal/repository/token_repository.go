package repository

import (
	"database/sql"
)

// Token represents the tokens table
type Token struct {
	ID             int64     `json:"id"`
	ChainID        int64     `json:"chain_id"`
	ContractAddress string   `json:"contract_address"`
	Symbol         string    `json:"symbol"`
	Decimals       int       `json:"decimals"`
	IsNative       bool      `json:"is_native"`
	IsActive       bool      `json:"is_active"`
}

// TokenRepository defines the interface for token data access
type TokenRepository interface {
	Create(token *Token) error
	GetByID(id int64) (*Token, error)
	GetByChainIDAndContract(chainID int64, contractAddress string) (*Token, error)
	Update(token *Token) error
	Delete(id int64) error
	List() ([]*Token, error)
	ListByChainID(chainID int64) ([]*Token, error)
	ListActive() ([]*Token, error)
}

// tokenRepository implements TokenRepository
type tokenRepository struct {
	db *sql.DB
}

// NewTokenRepository creates a new TokenRepository
func NewTokenRepository(db *sql.DB) TokenRepository {
	return &tokenRepository{db: db}
}

func (r *tokenRepository) Create(token *Token) error {
	query := `
		INSERT INTO tokens (chain_id, contract_address, symbol, decimals, is_native, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	return r.db.QueryRow(
		query,
		token.ChainID,
		token.ContractAddress,
		token.Symbol,
		token.Decimals,
		token.IsNative,
		token.IsActive,
	).Scan(&token.ID)
}

func (r *tokenRepository) GetByID(id int64) (*Token, error) {
	query := `
		SELECT id, chain_id, contract_address, symbol, decimals, is_native, is_active
		FROM tokens WHERE id = $1
	`
	token := &Token{}
	err := r.db.QueryRow(query, id).Scan(
		&token.ID,
		&token.ChainID,
		&token.ContractAddress,
		&token.Symbol,
		&token.Decimals,
		&token.IsNative,
		&token.IsActive,
	)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (r *tokenRepository) GetByChainIDAndContract(chainID int64, contractAddress string) (*Token, error) {
	query := `
		SELECT id, chain_id, contract_address, symbol, decimals, is_native, is_active
		FROM tokens WHERE chain_id = $1 AND lower(contract_address) = lower($2)
	`
	token := &Token{}
	err := r.db.QueryRow(query, chainID, contractAddress).Scan(
		&token.ID,
		&token.ChainID,
		&token.ContractAddress,
		&token.Symbol,
		&token.Decimals,
		&token.IsNative,
		&token.IsActive,
	)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (r *tokenRepository) Update(token *Token) error {
	query := `
		UPDATE tokens
		SET symbol = $2, decimals = $3, is_native = $4, is_active = $5
		WHERE id = $1
	`
	_, err := r.db.Exec(
		query,
		token.ID,
		token.Symbol,
		token.Decimals,
		token.IsNative,
		token.IsActive,
	)
	return err
}

func (r *tokenRepository) Delete(id int64) error {
	query := `DELETE FROM tokens WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *tokenRepository) List() ([]*Token, error) {
	query := `
		SELECT id, chain_id, contract_address, symbol, decimals, is_native, is_active
		FROM tokens ORDER BY id
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []*Token
	for rows.Next() {
		token := &Token{}
		err := rows.Scan(
			&token.ID,
			&token.ChainID,
			&token.ContractAddress,
			&token.Symbol,
			&token.Decimals,
			&token.IsNative,
			&token.IsActive,
		)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, rows.Err()
}

func (r *tokenRepository) ListByChainID(chainID int64) ([]*Token, error) {
	query := `
		SELECT id, chain_id, contract_address, symbol, decimals, is_native, is_active
		FROM tokens WHERE chain_id = $1 ORDER BY id
	`
	rows, err := r.db.Query(query, chainID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []*Token
	for rows.Next() {
		token := &Token{}
		err := rows.Scan(
			&token.ID,
			&token.ChainID,
			&token.ContractAddress,
			&token.Symbol,
			&token.Decimals,
			&token.IsNative,
			&token.IsActive,
		)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, rows.Err()
}

func (r *tokenRepository) ListActive() ([]*Token, error) {
	query := `
		SELECT id, chain_id, contract_address, symbol, decimals, is_native, is_active
		FROM tokens WHERE is_active = true ORDER BY id
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []*Token
	for rows.Next() {
		token := &Token{}
		err := rows.Scan(
			&token.ID,
			&token.ChainID,
			&token.ContractAddress,
			&token.Symbol,
			&token.Decimals,
			&token.IsNative,
			&token.IsActive,
		)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, rows.Err()
}
