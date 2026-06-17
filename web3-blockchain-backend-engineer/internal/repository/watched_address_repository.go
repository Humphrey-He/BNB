package repository

import (
	"database/sql"
)

// WatchedAddress represents the watched_addresses table
type WatchedAddress struct {
	ID          int64     `json:"id"`
	ChainID     int64     `json:"chain_id"`
	Address     string    `json:"address"`
	OwnerRef    string    `json:"owner_ref,omitempty"`
	Label       string    `json:"label,omitempty"`
	IsActive    bool      `json:"is_active"`
}

// WatchedAddressRepository defines the interface for watched address data access
type WatchedAddressRepository interface {
	Create(addr *WatchedAddress) error
	GetByID(id int64) (*WatchedAddress, error)
	GetByChainIDAndAddress(chainID int64, address string) (*WatchedAddress, error)
	Update(addr *WatchedAddress) error
	Delete(id int64) error
	List() ([]*WatchedAddress, error)
	ListByChainID(chainID int64) ([]*WatchedAddress, error)
	ListActive() ([]*WatchedAddress, error)
}

// watchedAddressRepository implements WatchedAddressRepository
type watchedAddressRepository struct {
	db *sql.DB
}

// NewWatchedAddressRepository creates a new WatchedAddressRepository
func NewWatchedAddressRepository(db *sql.DB) WatchedAddressRepository {
	return &watchedAddressRepository{db: db}
}

func (r *watchedAddressRepository) Create(addr *WatchedAddress) error {
	query := `
		INSERT INTO watched_addresses (chain_id, address, owner_ref, label, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	return r.db.QueryRow(
		query,
		addr.ChainID,
		addr.Address,
		addr.OwnerRef,
		addr.Label,
		addr.IsActive,
	).Scan(&addr.ID)
}

func (r *watchedAddressRepository) GetByID(id int64) (*WatchedAddress, error) {
	query := `
		SELECT id, chain_id, address, owner_ref, label, is_active
		FROM watched_addresses WHERE id = $1
	`
	addr := &WatchedAddress{}
	var ownerRef, label sql.NullString
	err := r.db.QueryRow(query, id).Scan(
		&addr.ID,
		&addr.ChainID,
		&addr.Address,
		&ownerRef,
		&label,
		&addr.IsActive,
	)
	if err != nil {
		return nil, err
	}
	addr.OwnerRef = ownerRef.String
	addr.Label = label.String
	return addr, nil
}

func (r *watchedAddressRepository) GetByChainIDAndAddress(chainID int64, address string) (*WatchedAddress, error) {
	query := `
		SELECT id, chain_id, address, owner_ref, label, is_active
		FROM watched_addresses WHERE chain_id = $1 AND lower(address) = lower($2)
	`
	addr := &WatchedAddress{}
	var ownerRef, label sql.NullString
	err := r.db.QueryRow(query, chainID, address).Scan(
		&addr.ID,
		&addr.ChainID,
		&addr.Address,
		&ownerRef,
		&label,
		&addr.IsActive,
	)
	if err != nil {
		return nil, err
	}
	addr.OwnerRef = ownerRef.String
	addr.Label = label.String
	return addr, nil
}

func (r *watchedAddressRepository) Update(addr *WatchedAddress) error {
	query := `
		UPDATE watched_addresses
		SET owner_ref = $2, label = $3, is_active = $4
		WHERE id = $1
	`
	_, err := r.db.Exec(
		query,
		addr.ID,
		addr.OwnerRef,
		addr.Label,
		addr.IsActive,
	)
	return err
}

func (r *watchedAddressRepository) Delete(id int64) error {
	query := `DELETE FROM watched_addresses WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *watchedAddressRepository) List() ([]*WatchedAddress, error) {
	query := `
		SELECT id, chain_id, address, owner_ref, label, is_active
		FROM watched_addresses ORDER BY id
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var addrs []*WatchedAddress
	for rows.Next() {
		addr := &WatchedAddress{}
		var ownerRef, label sql.NullString
		err := rows.Scan(
			&addr.ID,
			&addr.ChainID,
			&addr.Address,
			&ownerRef,
			&label,
			&addr.IsActive,
		)
		if err != nil {
			return nil, err
		}
		addr.OwnerRef = ownerRef.String
		addr.Label = label.String
		addrs = append(addrs, addr)
	}
	return addrs, rows.Err()
}

func (r *watchedAddressRepository) ListByChainID(chainID int64) ([]*WatchedAddress, error) {
	query := `
		SELECT id, chain_id, address, owner_ref, label, is_active
		FROM watched_addresses WHERE chain_id = $1 ORDER BY id
	`
	rows, err := r.db.Query(query, chainID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var addrs []*WatchedAddress
	for rows.Next() {
		addr := &WatchedAddress{}
		var ownerRef, label sql.NullString
		err := rows.Scan(
			&addr.ID,
			&addr.ChainID,
			&addr.Address,
			&ownerRef,
			&label,
			&addr.IsActive,
		)
		if err != nil {
			return nil, err
		}
		addr.OwnerRef = ownerRef.String
		addr.Label = label.String
		addrs = append(addrs, addr)
	}
	return addrs, rows.Err()
}

func (r *watchedAddressRepository) ListActive() ([]*WatchedAddress, error) {
	query := `
		SELECT id, chain_id, address, owner_ref, label, is_active
		FROM watched_addresses WHERE is_active = true ORDER BY id
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var addrs []*WatchedAddress
	for rows.Next() {
		addr := &WatchedAddress{}
		var ownerRef, label sql.NullString
		err := rows.Scan(
			&addr.ID,
			&addr.ChainID,
			&addr.Address,
			&ownerRef,
			&label,
			&addr.IsActive,
		)
		if err != nil {
			return nil, err
		}
		addr.OwnerRef = ownerRef.String
		addr.Label = label.String
		addrs = append(addrs, addr)
	}
	return addrs, rows.Err()
}
