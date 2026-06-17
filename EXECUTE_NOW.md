# 🚀 真实链上数据配置 - 执行文档

> 服务器: 101.43.127.178
> 用户: ubuntu
> 日期: 2026-06-16

---

## Step 1: 更新 Go API .env 配置

```bash
cat > /home/ubuntu/opt/bnb/api/.env << 'EOF'
POSTGRES_HOST=127.0.0.1
POSTGRES_PORT=5432
POSTGRES_DB=asset_platform
POSTGRES_USER=user_i2nZ7w
POSTGRES_PASSWORD=password_2Bda7K
REDIS_HOST=127.0.0.1
REDIS_PORT=6379
NATS_URL=nats://127.0.0.1:4222
APP_ENV=production
LOG_LEVEL=info
PORT=8080

# Real Chain RPC Configuration
ETHEREUM_RPC_URL=https://sepolia.infura.io/v3/aa4778679f4e4e64a48621c2b6c0c8b8
BSC_RPC_URL=https://data-seed-prebsc-1-s1.bnbchain.org:8545
EOF
```

---

## Step 2: 配置数据库 - 插入链信息

```bash
sudo docker exec -i psql-dev psql -U user_i2nZ7w -d asset_platform << 'SQL'
-- 插入 Ethereum Sepolia 配置
INSERT INTO chains (chain_id, name, native_symbol, finality_confirmations, is_active)
VALUES (11155111, 'Ethereum Sepolia', 'SEP', 12, true)
ON CONFLICT (chain_id) DO UPDATE SET name = EXCLUDED.name;

-- 插入 BSC Testnet 配置
INSERT INTO chains (chain_id, name, native_symbol, finality_confirmations, is_active)
VALUES (97, 'BSC Testnet', 'BNB', 12, true)
ON CONFLICT (chain_id) DO UPDATE SET name = EXCLUDED.name;

-- 验证插入结果
SELECT * FROM chains;
SQL
```

---

## Step 3: 配置 RPC Provider

```bash
sudo docker exec -i psql-dev psql -U user_i2nZ7w -d asset_platform << 'SQL'
-- 获取链 ID
DO $$
DECLARE
    sep_id BIGINT;
    bsc_id BIGINT;
BEGIN
    SELECT id INTO sep_id FROM chains WHERE chain_id = 11155111;
    SELECT id INTO bsc_id FROM chains WHERE chain_id = 97;
    
    -- Sepolia RPC Provider
    INSERT INTO rpc_providers (chain_id, name, url, weight, is_active)
    VALUES (sep_id, 'Infura Sepolia', 'https://sepolia.infura.io/v3/aa4778679f4e4e64a48621c2b6c0c8b8', 100, true)
    ON CONFLICT DO NOTHING;
    
    -- BSC Testnet RPC Provider
    INSERT INTO rpc_providers (chain_id, name, url, weight, is_active)
    VALUES (bsc_id, 'BNB Chain Testnet', 'https://data-seed-prebsc-1-s1.bnbchain.org:8545', 100, true)
    ON CONFLICT DO NOTHING;
END $$;

-- 验证插入结果
SELECT c.name, rp.name, rp.url FROM rpc_providers rp
JOIN chains c ON rp.chain_id = c.id;
SQL
```

---

## Step 4: 重启 Go API 服务

```bash
sudo systemctl restart asset-platform-api
sudo systemctl status asset-platform-api
```

---

## Step 5: 验证配置

```bash
# 测试 RPC 连接
curl -s -X POST "https://sepolia.infura.io/v3/aa4778679f4e4e64a48621c2b6c0c8b8" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'

# 测试 Go API
curl -s http://127.0.0.1:8080/healthz
```

---

## 完成后检查清单

| 项目 | 预期结果 |
|------|----------|
| .env 文件 | 包含 ETHEREUM_RPC_URL 和 BSC_RPC_URL |
| chains 表 | 有 11155111 和 97 两条记录 |
| rpc_providers 表 | 有 2 条 provider 记录 |
| asset-platform-api | 状态 active |
| RPC 测试 | 返回区块号 (如 0xa8f22c) |
