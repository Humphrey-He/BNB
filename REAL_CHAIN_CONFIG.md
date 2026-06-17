# BNB Web3 项目 - 真实链上数据配置指南

> 更新日期: 2026-06-16
> 配置目标: 让项目连接真实区块链 (Sepolia + BSC Testnet)

---

## 一、概述

本文档说明如何将项目从 Demo 模式升级到真实链上数据模式。

### 当前 RPC 配置

| Provider | Network | URL | 状态 |
|----------|---------|-----|------|
| **Infura** | Ethereum Sepolia | `https://sepolia.infura.io/v3/aa4778679f4e4e64a48621c2b6c0c8b8` | ✅ 可用 |
| **BNB Chain** | BSC Testnet | `https://data-seed-prebsc-1-s1.bnbchain.org:8545` | ✅ 可用 |

### 升级前后对比

| 功能 | 升级前 | 升级后 |
|------|--------|--------|
| 链上存款监听 | ❌ Mock 数据 | ✅ 真实 RPC 监听 |
| 钱包签名验证 | ✅ 真实实现 | ✅ 保持 |
| 合约交互 | ❌ 未部署 | ✅ Sepolia 测试网 |
| 数据来源 | 代码中硬编码 | 区块链实时获取 |

---

## 二、前置准备

### 2.1 当前 RPC Provider 配置

使用 **Infura** 作为主要 RPC Provider：

- Project ID: `aa4778679f4e4e64a48621c2b6c0c8b8`
- Dashboard: https://app.infura.io/key/all-endpoints

### 2.2 获取测试币 (Faucet)

- **Sepolia ETH**: https://sepoliafaucet.com/
- **BSC Testnet BNB**: https://testnet.bnbchain.org/faucet-smart-chain

---

## 三、Go API 配置

### 3.1 当前 .env 配置

服务器路径: `/home/ubuntu/opt/bnb/api/.env`

```
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
```

### 3.2 更新 .env 命令

如需更新，执行：

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
LOG_LEVEL=info
PORT=8080

# ============================================
# Real Chain RPC Configuration
# ============================================

# Ethereum Sepolia (测试网)
ETHEREUM_RPC_URL=https://eth-sepolia.g.alchemy.com/v2/YOUR_ALCHEMY_API_KEY

# BSC Testnet
BSC_RPC_URL=https://data-seed-prebsc-1-s1.bnbchain.org:8545

# Chain IDs
CHAIN_ID_ETHEREUM_SEPOLIA=11155111
CHAIN_ID_BSC_TESTNET=97
```

---

## 四、数据库初始化

### 4.1 插入链配置

连接数据库执行：

```sql
-- 插入 Ethereum Sepolia 配置
INSERT INTO chains (chain_id, name, native_symbol, finality_confirmations, is_active)
VALUES (11155111, 'Ethereum Sepolia', 'SEP', 12, true)
ON CONFLICT (chain_id) DO UPDATE SET 
    name = EXCLUDED.name,
    finality_confirmations = EXCLUDED.finality_confirmations;

-- 插入 BSC Testnet 配置
INSERT INTO chains (chain_id, name, native_symbol, finality_confirmations, is_active)
VALUES (97, 'BSC Testnet', 'BNB', 12, true)
ON CONFLICT (chain_id) DO UPDATE SET 
    name = EXCLUDED.name,
    finality_confirmations = EXCLUDED.finality_confirmations;
```

### 4.2 插入 RPC Provider 配置

```sql
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
    VALUES (sep_id, 'Alchemy Sepolia', 'https://eth-sepolia.g.alchemy.com/v2/YOUR_API_KEY', 100, true)
    ON CONFLICT DO NOTHING;
    
    -- BSC Testnet RPC Provider
    INSERT INTO rpc_providers (chain_id, name, url, weight, is_active)
    VALUES (bsc_id, 'BNB Chain Testnet', 'https://data-seed-prebsc-1-s1.bnbchain.org:8545', 100, true)
    ON CONFLICT DO NOTHING;
END $$;
```

### 4.3 插入监听的合约地址

```sql
-- 插入要监听的 Vault 合约地址（部署后替换）
DO $$
DECLARE
    sep_id BIGINT;
    bsc_id BIGINT;
    sep_token_id BIGINT;
    bsc_token_id BIGINT;
BEGIN
    SELECT id INTO sep_id FROM chains WHERE chain_id = 11155111;
    SELECT id INTO bsc_id FROM chains WHERE chain_id = 97;
    
    -- Sepolia: 插入 ETH 原生币
    INSERT INTO tokens (chain_id, contract_address, symbol, decimals, is_native)
    VALUES (sep_id, '0x0000000000000000000000000000000000000000', 'ETH', 18, true)
    ON CONFLICT (chain_id, contract_address) DO NOTHING;
    
    -- BSC: 插入 BNB 原生币
    INSERT INTO tokens (chain_id, contract_address, symbol, decimals, is_native)
    VALUES (bsc_id, '0x0000000000000000000000000000000000000000', 'BNB', 18, true)
    ON CONFLICT (chain_id, contract_address) DO NOTHING;
    
    -- 获取 token IDs
    SELECT id INTO sep_token_id FROM tokens WHERE chain_id = sep_id AND is_native = true;
    SELECT id INTO bsc_token_id FROM tokens WHERE chain_id = bsc_id AND is_native = true;
    
    -- 插入 Vault 合约地址（部署后替换下面的地址）
    INSERT INTO watched_addresses (chain_id, address, owner_ref, label, is_active)
    VALUES 
        (sep_id, '0x0000000000000000000000000000000000000000', 'platform', 'Sepolia Deposit Vault', true),
        (bsc_id, '0x0000000000000000000000000000000000000000', 'platform', 'BSC Deposit Vault', true)
    ON CONFLICT (chain_id, address) DO NOTHING;
END $$;
```

---

## 五、部署智能合约

### 5.1 完善部署脚本

更新 `smart-contract-engineer/script/DeployScript.sol`：

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import { Script } from "forge-std/Script.sol";
import { SecureYieldVault } from "../contracts/SecureYieldVault.sol";
import { RewardDistributor } from "../contracts/RewardDistributor.sol";

contract DeployScript is Script {
    function run() external {
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY");
        
        vm.startBroadcast(deployerPrivateKey);
        
        // Deploy SecureYieldVault
        SecureYieldVault vault = new SecureYieldVault();
        
        // Deploy RewardDistributor with vault address
        RewardDistributor distributor = new RewardDistributor(address(vault));
        
        vm.stopBroadcast();
        
        console.log("SecureYieldVault deployed to:", address(vault));
        console.log("RewardDistributor deployed to:", address(distributor));
    }
}
```

### 5.2 部署命令

在本地 Windows 执行（需要 Foundry）：

```powershell
# 设置环境变量
$env:PRIVATE_KEY = "your_deployer_private_key"
$env:ETHEREUM_RPC_URL = "https://eth-sepolia.g.alchemy.com/v2/YOUR_API_KEY"

# 部署到 Sepolia
cd E:\awesomeProject\BNB\smart-contract-engineer
forge script script/DeployScript.sol --rpc-url $env:ETHEREUM_RPC_URL --broadcast --verify
```

### 5.3 保存合约地址

部署成功后会输出合约地址，记录下来并更新数据库：

```sql
-- 替换下面的合约地址为实际部署的地址
UPDATE watched_addresses 
SET address = '0xYourVaultAddress' 
WHERE label = 'Sepolia Deposit Vault';
```

---

## 六、启动 Scanner 服务

### 6.1 创建 Scanner systemd 服务

```bash
sudo tee /etc/systemd/system/asset-scanner.service > /dev/null <<'EOF'
[Unit]
Description=BNB Web3 Chain Scanner
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu/opt/bnb/api
EnvironmentFile=/home/ubuntu/opt/bnb/api/.env
ExecStart=/home/ubuntu/opt/bnb/api/api-server-linux-amd64 scanner
Restart=on-failure
RestartSec=5
Environment=RUST_LOG=info

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable asset-scanner
sudo systemctl start asset-scanner
```

### 6.2 验证 Scanner 运行

```bash
# 查看日志
journalctl -u asset-scanner -f --no-pager -n 50

# 检查是否正在监听新区块
curl -s http://127.0.0.1:8080/api/v1/scanner-status
```

---

## 七、验证配置

### 7.1 测试 RPC 连接

```bash
# 测试 Ethereum Sepolia RPC
curl -X POST "https://eth-sepolia.g.alchemy.com/v2/YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'

# 预期返回: {"jsonrpc":"2.0","result":"0x4c2a3b",...}
```

### 7.2 测试数据库连接

```bash
sudo docker exec -i psql-dev psql -U user_i2nZ7w -d asset_platform -c "SELECT * FROM chains;"
```

### 7.3 完整验证脚本

```bash
# 保存为 verify-real-chain.sh 并执行
#!/bin/bash
echo "=== Real Chain Configuration Verification ==="

echo -e "\n[1/5] Testing Ethereum Sepolia RPC..."
ETH_BLOCK=$(curl -s -X POST "$ETHEREUM_RPC_URL" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  | jq -r '.result')
echo "  Latest block: $ETH_BLOCK"

echo -e "\n[2/5] Testing BSC Testnet RPC..."
BSC_BLOCK=$(curl -s -X POST "$BSC_RPC_URL" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  | jq -r '.result')
echo "  Latest block: $BSC_BLOCK"

echo -e "\n[3/5] Checking database chains..."
sudo docker exec -i psql-dev psql -U user_i2nZ7w -d asset_platform -c "SELECT chain_id, name, is_active FROM chains;"

echo -e "\n[4/5] Checking RPC providers..."
sudo docker exec -i psql-dev psql -U user_i2nZ7w -d asset_platform -c "SELECT name, url, is_active FROM rpc_providers;"

echo -e "\n[5/5] Testing Go API..."
curl -s http://127.0.0.1:8080/healthz

echo -e "\n=== Verification Complete ==="
```

---

## 八、常见问题排查

### Q1: RPC 返回错误

```
error: connection refused
```
**解决**: 检查 API Key 是否正确，网络是否可达

### Q2: Scanner 没收到事件

```
no new events found
```
**解决**: 
1. 确认 Vault 合约地址正确
2. 确认有真实的 Transfer 事件
3. 检查 Scanner 日志

### Q3: 余额计算错误

```
balance overflow
```
**解决**: 检查是否使用了 int64 解析 NUMERIC(78,0)

---

## 九、配置检查清单

| 项目 | 状态 | 说明 |
|------|------|------|
| Alchemy 账号注册 | ☐ | |
| Sepolia API Key 获取 | ☐ | |
| 测试币领取 | ☐ | |
| .env RPC 配置 | ☐ | |
| 数据库链配置 | ☐ | |
| 合约部署 | ☐ | |
| Scanner 服务启动 | ☐ | |
| 端到端验证 | ☐ | |

---

## 十、后续优化建议

1. **多 RPC Provider**: 配置多个 RPC 实现负载均衡和故障转移
2. **监控告警**: 设置 Prometheus AlertManager 监控 RPC 失败率
3. **数据清理**: 定期清理 old blocks 和 events 表
4. **签名安全**: 将钱包私钥移到 Vault 或硬件钱包

---

**版本**: v1.0 (2026-06-16)
