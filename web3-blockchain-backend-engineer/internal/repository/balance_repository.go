package repository

import (
	"database/sql"
	"time"
)

// Balance represents the balances table
type Balance struct {
	ID                int64     `json:"id"`
	AccountAddress    string    `json:"account_address"`
	ChainID           int64     `json:"chain_id"`
	TokenID           int64     `json:"token_id"`
	AvailableBalance  string    `json:"available_balance"`
	FrozenBalance     string    `json:"frozen_balance"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// BalanceRepository defines the interface for balance data access
type BalanceRepository interface {
	Create(balance *Balance) error
	GetByID(id int64) (*Balance, error)
	GetByAccountChainAndToken(accountAddress string, chainID int64, tokenID int64) (*Balance, error)
	Update(balance *Balance) error
	// Credit atomically adds amount to available_balance. Creates balance record if not exists.
	// Returns (newBalance, error)
	Credit(accountAddress string, chainID int64, tokenID int64, amount string) (*Balance, error)
	Freeze(accountAddress string, chainID int64, tokenID int64, amount string) error
	Unfreeze(accountAddress string, chainID int64, tokenID int64, amount string) error
	Delete(id int64) error
	List(limit int) ([]*Balance, error)
	ListByAccountAddress(accountAddress string) ([]*Balance, error)
	ListByChainID(chainID int64) ([]*Balance, error)
}

// balanceRepository implements BalanceRepository
type balanceRepository struct {
	db *sql.DB
}

// NewBalanceRepository creates a new BalanceRepository
func NewBalanceRepository(db *sql.DB) BalanceRepository {
	return &balanceRepository{db: db}
}

func (r *balanceRepository) Create(balance *Balance) error {
	query := `
		INSERT INTO balances (account_address, chain_id, token_id, available_balance, frozen_balance, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	balance.UpdatedAt = time.Now()
	return r.db.QueryRow(
		query,
		balance.AccountAddress,
		balance.ChainID,
		balance.TokenID,
		balance.AvailableBalance,
		balance.FrozenBalance,
		balance.UpdatedAt,
	).Scan(&balance.ID)
}

func (r *balanceRepository) GetByID(id int64) (*Balance, error) {
	query := `
		SELECT id, account_address, chain_id, token_id, available_balance, frozen_balance, updated_at
		FROM balances WHERE id = $1
	`
	balance := &Balance{}
	err := r.db.QueryRow(query, id).Scan(
		&balance.ID,
		&balance.AccountAddress,
		&balance.ChainID,
		&balance.TokenID,
		&balance.AvailableBalance,
		&balance.FrozenBalance,
		&balance.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return balance, nil
}

func (r *balanceRepository) GetByAccountChainAndToken(accountAddress string, chainID int64, tokenID int64) (*Balance, error) {
	query := `
		SELECT id, account_address, chain_id, token_id, available_balance, frozen_balance, updated_at
		FROM balances WHERE account_address = $1 AND chain_id = $2 AND token_id = $3
	`
	balance := &Balance{}
	err := r.db.QueryRow(query, accountAddress, chainID, tokenID).Scan(
		&balance.ID,
		&balance.AccountAddress,
		&balance.ChainID,
		&balance.TokenID,
		&balance.AvailableBalance,
		&balance.FrozenBalance,
		&balance.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return balance, nil
}

func (r *balanceRepository) Update(balance *Balance) error {
	query := `
		UPDATE balances
		SET available_balance = $2, frozen_balance = $3, updated_at = $4
		WHERE id = $1
	`
	balance.UpdatedAt = time.Now()
	_, err := r.db.Exec(
		query,
		balance.ID,
		balance.AvailableBalance,
		balance.FrozenBalance,
		balance.UpdatedAt,
	)
	return err
}

// Credit atomically adds amount to available_balance.
// Uses INSERT ... ON CONFLICT DO UPDATE to handle concurrent first-time deposits correctly.
// This ensures no deposit is lost when two concurrent requests try to create the same balance.
func (r *balanceRepository) Credit(accountAddress string, chainID int64, tokenID int64, amount string) (*Balance, error) {
	query := `
		INSERT INTO balances (account_address, chain_id, token_id, available_balance, frozen_balance, updated_at)
		VALUES ($1, $2, $3, $4::numeric, 0, $5)
		ON CONFLICT (account_address, chain_id, token_id)
		DO UPDATE SET available_balance = balances.available_balance::numeric + $4::numeric, updated_at = $5
		RETURNING id, available_balance, frozen_balance
	`
	balance := &Balance{}
	err := r.db.QueryRow(query, accountAddress, chainID, tokenID, amount, time.Now()).Scan(
		&balance.ID,
		&balance.AvailableBalance,
		&balance.FrozenBalance,
	)
	if err != nil {
		return nil, err
	}

	balance.AccountAddress = accountAddress
	balance.ChainID = chainID
	balance.TokenID = tokenID
	balance.UpdatedAt = time.Now()

	return balance, nil
}

func (r *balanceRepository) Freeze(accountAddress string, chainID int64, tokenID int64, amount string) error {
	query := `
		UPDATE balances
		SET available_balance = available_balance - $4::numeric,
		    frozen_balance = frozen_balance + $4::numeric,
		    updated_at = $5
		WHERE account_address = $1 AND chain_id = $2 AND token_id = $3
		  AND available_balance::numeric >= $4::numeric
	`
	result, err := r.db.Exec(query, accountAddress, chainID, tokenID, amount, time.Now())
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *balanceRepository) Unfreeze(accountAddress string, chainID int64, tokenID int64, amount string) error {
	query := `
		UPDATE balances
		SET available_balance = available_balance + $4::numeric,
		    frozen_balance = frozen_balance - $4::numeric,
		    updated_at = $5
		WHERE account_address = $1 AND chain_id = $2 AND token_id = $3
		  AND frozen_balance::numeric >= $4::numeric
	`
	result, err := r.db.Exec(query, accountAddress, chainID, tokenID, amount, time.Now())
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *balanceRepository) Delete(id int64) error {
	query := `DELETE FROM balances WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *balanceRepository) List(limit int) ([]*Balance, error) {
	query := `
		SELECT id, account_address, chain_id, token_id, available_balance, frozen_balance, updated_at
		FROM balances ORDER BY id LIMIT $1
	`
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var balances []*Balance
	for rows.Next() {
		b := &Balance{}
		err := rows.Scan(
			&b.ID,
			&b.AccountAddress,
			&b.ChainID,
			&b.TokenID,
			&b.AvailableBalance,
			&b.FrozenBalance,
			&b.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		balances = append(balances, b)
	}
	return balances, rows.Err()
}

func (r *balanceRepository) ListByAccountAddress(accountAddress string) ([]*Balance, error) {
	query := `
		SELECT id, account_address, chain_id, token_id, available_balance, frozen_balance, updated_at
		FROM balances WHERE account_address = $1 ORDER BY id
	`
	rows, err := r.db.Query(query, accountAddress)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var balances []*Balance
	for rows.Next() {
		b := &Balance{}
		err := rows.Scan(
			&b.ID,
			&b.AccountAddress,
			&b.ChainID,
			&b.TokenID,
			&b.AvailableBalance,
			&b.FrozenBalance,
			&b.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		balances = append(balances, b)
	}
	return balances, rows.Err()
}

func (r *balanceRepository) ListByChainID(chainID int64) ([]*Balance, error) {
	query := `
		SELECT id, account_address, chain_id, token_id, available_balance, frozen_balance, updated_at
		FROM balances WHERE chain_id = $1 ORDER BY id
	`
	rows, err := r.db.Query(query, chainID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var balances []*Balance
	for rows.Next() {
		b := &Balance{}
		err := rows.Scan(
			&b.ID,
			&b.AccountAddress,
			&b.ChainID,
			&b.TokenID,
			&b.AvailableBalance,
			&b.FrozenBalance,
			&b.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		balances = append(balances, b)
	}
	return balances, rows.Err()
}
