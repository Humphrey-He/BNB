-- Migration: 000001_init_schema
-- Description: Create initial database schema for multi-chain-asset-platform
-- Created: 2026-05-06

-- chains
CREATE TABLE IF NOT EXISTS chains (
    id BIGSERIAL PRIMARY KEY,
    chain_id BIGINT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    native_symbol TEXT NOT NULL,
    finality_confirmations INT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- rpc_providers
CREATE TABLE IF NOT EXISTS rpc_providers (
    id BIGSERIAL PRIMARY KEY,
    chain_id BIGINT NOT NULL REFERENCES chains(id),
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    weight INT NOT NULL DEFAULT 100,
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_error TEXT,
    last_checked_at TIMESTAMPTZ
);

-- tokens
CREATE TABLE IF NOT EXISTS tokens (
    id BIGSERIAL PRIMARY KEY,
    chain_id BIGINT NOT NULL REFERENCES chains(id),
    contract_address TEXT NOT NULL,
    symbol TEXT NOT NULL,
    decimals INT NOT NULL,
    is_native BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    UNIQUE(chain_id, contract_address)
);

-- watched_addresses
CREATE TABLE IF NOT EXISTS watched_addresses (
    id BIGSERIAL PRIMARY KEY,
    chain_id BIGINT NOT NULL REFERENCES chains(id),
    address TEXT NOT NULL,
    owner_ref TEXT,
    label TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    UNIQUE(chain_id, address)
);

-- blocks
CREATE TABLE IF NOT EXISTS blocks (
    id BIGSERIAL PRIMARY KEY,
    chain_id BIGINT NOT NULL REFERENCES chains(id),
    block_number BIGINT NOT NULL,
    block_hash TEXT NOT NULL,
    parent_hash TEXT NOT NULL,
    block_time TIMESTAMPTZ,
    is_orphaned BOOLEAN NOT NULL DEFAULT false,
    scanned_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(chain_id, block_number, block_hash)
);

-- chain_events
CREATE TABLE IF NOT EXISTS chain_events (
    id BIGSERIAL PRIMARY KEY,
    chain_id BIGINT NOT NULL REFERENCES chains(id),
    tx_hash TEXT NOT NULL,
    log_index INT NOT NULL,
    block_number BIGINT NOT NULL,
    block_hash TEXT NOT NULL,
    contract_address TEXT NOT NULL,
    event_name TEXT NOT NULL,
    from_address TEXT,
    to_address TEXT,
    amount NUMERIC(78, 0),
    is_orphaned BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(chain_id, tx_hash, log_index)
);

-- deposits
CREATE TABLE IF NOT EXISTS deposits (
    id BIGSERIAL PRIMARY KEY,
    chain_id BIGINT NOT NULL REFERENCES chains(id),
    token_id BIGINT NOT NULL REFERENCES tokens(id),
    tx_hash TEXT NOT NULL,
    log_index INT NOT NULL,
    from_address TEXT NOT NULL,
    to_address TEXT NOT NULL,
    amount NUMERIC(78, 0) NOT NULL,
    block_number BIGINT NOT NULL,
    status TEXT NOT NULL DEFAULT 'detected',
    confirmations INT NOT NULL DEFAULT 0,
    target_confirmations INT NOT NULL DEFAULT 12,
    idempotency_key TEXT NOT NULL UNIQUE,
    processed_at TIMESTAMPTZ,
    confirmed_event_published BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- withdrawals
CREATE TABLE IF NOT EXISTS withdrawals (
    id BIGSERIAL PRIMARY KEY,
    chain_id BIGINT NOT NULL REFERENCES chains(id),
    token_id BIGINT NOT NULL REFERENCES tokens(id),
    from_address TEXT NOT NULL,
    to_address TEXT NOT NULL,
    amount NUMERIC(78, 0) NOT NULL,
    fee NUMERIC(78, 0) NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'created',
    tx_hash TEXT,
    nonce BIGINT,
    idempotency_key TEXT NOT NULL UNIQUE,
    failure_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ledger_entries
CREATE TABLE IF NOT EXISTS ledger_entries (
    id BIGSERIAL PRIMARY KEY,
    account_address TEXT NOT NULL,
    chain_id BIGINT NOT NULL REFERENCES chains(id),
    token_id BIGINT NOT NULL REFERENCES tokens(id),
    direction TEXT NOT NULL,
    amount NUMERIC(78, 0) NOT NULL,
    balance_before NUMERIC(78, 0) NOT NULL DEFAULT 0,
    balance_after NUMERIC(78, 0) NOT NULL DEFAULT 0,
    entry_type TEXT NOT NULL,
    reference_type TEXT NOT NULL,
    reference_id BIGINT NOT NULL,
    reversal_of BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(reference_type, reference_id, entry_type)
);

-- balances
CREATE TABLE IF NOT EXISTS balances (
    id BIGSERIAL PRIMARY KEY,
    account_address TEXT NOT NULL,
    chain_id BIGINT NOT NULL REFERENCES chains(id),
    token_id BIGINT NOT NULL REFERENCES tokens(id),
    available_balance NUMERIC(78, 0) NOT NULL DEFAULT 0,
    frozen_balance NUMERIC(78, 0) NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(account_address, chain_id, token_id)
);

-- nonce_allocations
CREATE TABLE IF NOT EXISTS nonce_allocations (
    id BIGSERIAL PRIMARY KEY,
    chain_id BIGINT NOT NULL REFERENCES chains(id),
    from_address TEXT NOT NULL,
    nonce BIGINT NOT NULL,
    withdrawal_id BIGINT REFERENCES withdrawals(id),
    status TEXT NOT NULL DEFAULT 'allocated',
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(chain_id, from_address, nonce)
);

-- scan_checkpoints
CREATE TABLE IF NOT EXISTS scan_checkpoints (
    id BIGSERIAL PRIMARY KEY,
    chain_id BIGINT NOT NULL REFERENCES chains(id),
    last_scanned_block BIGINT NOT NULL,
    last_scanned_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(chain_id)
);

-- schema_migrations
CREATE TABLE IF NOT EXISTS schema_migrations (
    id BIGSERIAL PRIMARY KEY,
    version BIGINT NOT NULL UNIQUE,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Indexes for performance
CREATE INDEX idx_blocks_chain_block ON blocks(chain_id, block_number);
CREATE INDEX idx_blocks_orphaned ON blocks(is_orphaned) WHERE is_orphaned = false;
CREATE INDEX idx_chain_events_chain_tx ON chain_events(chain_id, tx_hash);
CREATE INDEX idx_deposits_status ON deposits(status);
CREATE INDEX idx_deposits_address ON deposits(to_address);
CREATE INDEX idx_deposits_chain_time ON deposits(chain_id, created_at);
CREATE INDEX idx_deposits_confirmed_unpublished ON deposits(status, confirmed_event_published)
    WHERE status = 'confirmed' AND confirmed_event_published = false;
CREATE INDEX idx_withdrawals_status ON withdrawals(status);
CREATE INDEX idx_withdrawals_address ON withdrawals(from_address);
CREATE INDEX idx_ledger_account ON ledger_entries(account_address, chain_id, token_id);
CREATE INDEX idx_ledger_reference ON ledger_entries(reference_type, reference_id);
