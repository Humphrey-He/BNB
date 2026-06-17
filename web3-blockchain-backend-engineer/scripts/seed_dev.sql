-- Seed data for development environment
-- Created: 2026-05-06

-- ============================================
-- Chains Configuration
-- ============================================

INSERT INTO chains (chain_id, name, native_symbol, finality_confirmations, is_active)
VALUES
    (1, 'Ethereum', 'ETH', 12, true),
    (56, 'BNB Chain', 'BNB', 15, true)
ON CONFLICT (chain_id) DO NOTHING;

-- ============================================
-- RPC Providers
-- ============================================

INSERT INTO rpc_providers (chain_id, name, url, weight, is_active)
VALUES
    ((SELECT id FROM chains WHERE chain_id = 1), 'Infura', 'https://mainnet.infura.io/v3/YOUR_API_KEY', 100, true),
    ((SELECT id FROM chains WHERE chain_id = 1), 'LlamaRPC', 'https://eth.llamarpc.com', 80, true),
    ((SELECT id FROM chains WHERE chain_id = 56), 'Binance', 'https://bsc-dataseed.binance.org', 100, true),
    ((SELECT id FROM chains WHERE chain_id = 56), 'LlamaRPC BSC', 'https://bsc.llamarpc.com', 80, true)
ON CONFLICT DO NOTHING;

-- ============================================
-- Native Tokens (ETH and BNB)
-- ============================================

INSERT INTO tokens (chain_id, contract_address, symbol, decimals, is_native, is_active)
VALUES
    ((SELECT id FROM chains WHERE chain_id = 1), '0x0000000000000000000000000000000000000000', 'ETH', 18, true, true),
    ((SELECT id FROM chains WHERE chain_id = 56), '0x0000000000000000000000000000000000000000', 'BNB', 18, true, true)
ON CONFLICT (chain_id, contract_address) DO NOTHING;

-- ============================================
-- ERC-20 Tokens
-- ============================================

INSERT INTO tokens (chain_id, contract_address, symbol, decimals, is_native, is_active)
VALUES
    ((SELECT id FROM chains WHERE chain_id = 1), '0xdAC17F958D2ee523a2206206994597C13D831ec7', 'USDT', 6, false, true),
    ((SELECT id FROM chains WHERE chain_id = 56), '0x55d398326f99059fF775485246999027B3197955', 'USDT', 6, false, true),
    ((SELECT id FROM chains WHERE chain_id = 56), '0xe9e7CEA3DedcA5984780Bafc599bD69ADd087D56', 'BUSD', 18, false, true)
ON CONFLICT (chain_id, contract_address) DO NOTHING;

-- ============================================
-- Watched Addresses
-- ============================================

INSERT INTO watched_addresses (chain_id, address, owner_ref, label, is_active)
VALUES
    ((SELECT id FROM chains WHERE chain_id = 1), '0x742d35Cc6634C0532925a3b844Bc9e7595f2bD61', 'user_001', 'Ethereum Main Deposit', true),
    ((SELECT id FROM chains WHERE chain_id = 1), '0x8Ba1f109551bD432803012645Ac136ddd64DBA72', 'user_002', 'Ethereum Trading Account', true),
    ((SELECT id FROM chains WHERE chain_id = 56), '0x30C5452B9bC2F4d6b3D1B2a3F4E5d6C7b8A9a0B1', 'user_001', 'BSC Main Deposit', true),
    ((SELECT id FROM chains WHERE chain_id = 56), '0x4B20993Bc481177ec7E8f571ceCaE8A9e22C02db', 'user_002', 'BSC Trading Account', true),
    ((SELECT id FROM chains WHERE chain_id = 56), '0x787B64B0a5D6c7d3a5F6e7E8F9a0B1C2D3E4F5A6', 'hot_wallet_01', 'BSC Hot Wallet', true)
ON CONFLICT (chain_id, address) DO NOTHING;

-- ============================================
-- Initial Scan Checkpoints
-- ============================================

INSERT INTO scan_checkpoints (chain_id, last_scanned_block, last_scanned_at)
VALUES
    ((SELECT id FROM chains WHERE chain_id = 1), 19500000, now()),
    ((SELECT id FROM chains WHERE chain_id = 56), 35000000, now())
ON CONFLICT (chain_id) DO NOTHING;
