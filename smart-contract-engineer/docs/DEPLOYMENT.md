# SecureYieldVault 部署指南

## 前置条件

- Foundry 已安装 (`forge --version` 可正常运行)
- 测试网 ETH 或本地节点运行中
- .env 文件配置 (可选)

## 环境配置

### .env 文件 (可选)

```bash
# .env
PRIVATE_KEY=0x...
SEPOLIA_RPC_URL=https://sepolia.infura.io/v3/YOUR_API_KEY
ETHERSCAN_API_KEY=YOUR_API_KEY
```

### 网络端点配置 (foundry.toml)

```toml
[rpc_endpoints]
sepolia = "${SEPOLIA_RPC_URL}"
mainnet = "${MAINNET_RPC_URL}"
```

## 部署步骤

### 1. 本地节点部署

```bash
# 终端 1: 启动 anvil
anvil

# 终端 2: 部署到本地
forge script script/Deploy.s.sol --rpc-url http://localhost:8545 --broadcast

# 验证部署
forge verify-contract <CONTRACT_ADDRESS> src/SecureYieldVault.sol:SecureYieldVault
```

### 2. Sepolia 测试网部署

```bash
# 部署
forge script script/Deploy.s.sol \
    --rpc-url $SEPOLIA_RPC_URL \
    --private-key $PRIVATE_KEY \
    --broadcast \
    --verify \
    --etherscan-api-key $ETHERSCAN_API_KEY
```

### 3. 部署后配置

```bash
# 获取合约地址 (部署输出中的 Log)
# 假设部署到 VAR_ADDRESS

# 1. 铸造测试代币给用户
cast send <ASSET_TOKEN> "mint(address,uint256)" \
    <USER_ADDRESS> \
    1000000000000000000000 \
    --rpc-url <RPC_URL> \
    --private-key <PRIVATE_KEY>

# 2. 用户授权 Vault 使用代币
cast send <ASSET_TOKEN> "approve(address,uint256)" \
    <VAULT_ADDRESS> \
    <UINT256_MAX> \
    --rpc-url <RPC_URL> \
    --private-key <USER_PRIVATE_KEY>

# 3. 授权 operator (可选)
cast send <VAULT_ADDRESS> "grantRole(bytes32,address)" \
    $(cast keccak "OPERATOR_ROLE") \
    <OPERATOR_ADDRESS> \
    --rpc-url <RPC_URL> \
    --private-key <PRIVATE_KEY>

# 4. 注入奖励 (可选)
cast send <ASSET_TOKEN> "mint(address,uint256)" \
    <OPERATOR_ADDRESS> \
    100000000000000000000 \
    --rpc-url <RPC_URL> \
    --private-key <PRIVATE_KEY>

cast send <ASSET_TOKEN> "approve(address,uint256)" \
    <VAULT_ADDRESS> \
    100000000000000000000 \
    --rpc-url <RPC_URL> \
    --private-key <OPERATOR_PRIVATE_KEY>

cast send <VAULT_ADDRESS> "injectRewards(uint256)" \
    100000000000000000000 \
    --rpc-url <RPC_URL> \
    --private-key <OPERATOR_PRIVATE_KEY>
```

## 常用操作

### 存款

```bash
# 存款 (存入资产，获得 shares)
cast send <VAULT_ADDRESS> "deposit(uint256,address)" \
    1000000000000000000 \
    <USER_ADDRESS> \
    --rpc-url <RPC_URL> \
    --private-key <USER_PRIVATE_KEY>
```

### 取款

```bash
# 取款 (提取资产，销毁 shares)
cast send <VAULT_ADDRESS> "withdraw(uint256,address,address)" \
    1000000000000000000 \
    <USER_ADDRESS> \
    <USER_ADDRESS> \
    --rpc-url <RPC_URL> \
    --private-key <USER_PRIVATE_KEY>
```

### 铸造 shares

```bash
# 指定 shares 数量，自动计算所需资产
cast send <VAULT_ADDRESS> "mint(uint256,address)" \
    1000000000000000000 \
    <USER_ADDRESS> \
    --rpc-url <RPC_URL> \
    --private-key <USER_PRIVATE_KEY>
```

### 赎回 shares

```bash
# 销毁 shares，获得资产
cast send <VAULT_ADDRESS> "redeem(uint256,address,address)" \
    1000000000000000000 \
    <USER_ADDRESS> \
    <USER_ADDRESS> \
    --rpc-url <RPC_URL> \
    --private-key <USER_PRIVATE_KEY>
```

### 紧急提取

```bash
# 1. 暂停合约 (admin)
cast send <VAULT_ADDRESS> "pause" \
    --rpc-url <RPC_URL> \
    --private-key <ADMIN_PRIVATE_KEY>

# 2. 紧急提取 (admin)
cast send <VAULT_ADDRESS> "emergencyWithdraw(uint256,address)" \
    1000000000000000000 \
    <ADMIN_ADDRESS> \
    --rpc-url <RPC_URL> \
    --private-key <ADMIN_PRIVATE_KEY>

# 3. 恢复合约 (admin)
cast send <VAULT_ADDRESS> "unpause" \
    --rpc-url <RPC_URL> \
    --private-key <ADMIN_PRIVATE_KEY>
```

## 状态查询

```bash
# 查询 totalAssets
cast call <VAULT_ADDRESS> "totalAssets()" --rpc-url <RPC_URL>

# 查询 totalAssetsWithRewards
cast call <VAULT_ADDRESS> "totalAssetsWithRewards()" --rpc-url <RPC_URL>

# 查询用户 shares 余额
cast call <VAULT_ADDRESS> "balanceOf(address)" <USER_ADDRESS> --rpc-url <RPC_URL>

# 查询用户资产余额
cast call <VAULT_ADDRESS> "balanceOfAssets(address)" <USER_ADDRESS> --rpc-url <RPC_URL>

# 转换为 shares
cast call <VAULT_ADDRESS> "convertToShares(uint256)" 1000000000000000000 --rpc-url <RPC_URL>

# 转换为 assets
cast call <VAULT_ADDRESS> "convertToAssets(uint256)" 1000000000000000000 --rpc-url <RPC_URL>

# 检查暂停状态
cast call <VAULT_ADDRESS> "paused()" --rpc-url <RPC_URL>
```

## 角色检查

```bash
# 检查是否有 ADMIN_ROLE
cast call <VAULT_ADDRESS> "hasRole(bytes32,address)" \
    $(cast keccak "ADMIN_ROLE") \
    <ADDRESS> \
    --rpc-url <RPC_URL>

# 检查是否有 OPERATOR_ROLE
cast call <VAULT_ADDRESS> "hasRole(bytes32,address)" \
    $(cast keccak "OPERATOR_ROLE") \
    <ADDRESS> \
    --rpc-url <RPC_URL>
```

## 事件监听

```python
# scripts/listener.py 示例
from web3 import Web3

w3 = Web3(Web3.HTTPProvider("http://localhost:8545"))

vault_address = "<VAULT_ADDRESS>"
vault_contract = w3.eth.contract(address=vault_address, abi=...)

# 监听 Deposited 事件
events = vault_contract.events.Deposited.get_logs(from_block=0)
for event in events:
    print(f"User {event.args.owner} deposited {event.args.assets} assets")
```

## 故障排除

### 交易失败

```bash
# 检查交易详情
cast receipt <TX_HASH> --rpc-url <RPC_URL>
```

### 余额不足

```bash
# 检查 ETH 余额
cast balance <ADDRESS> --rpc-url <RPC_URL>

# 检查 token 余额
cast call <TOKEN_ADDRESS> "balanceOf(address)" <ADDRESS> --rpc-url <RPC_URL>
```

### 权限问题

```bash
# 检查 allowance
cast call <TOKEN_ADDRESS> "allowance(address,address)" \
    <OWNER_ADDRESS> \
    <SPENDER_ADDRESS> \
    --rpc-url <RPC_URL>
```

## 安全检查清单

部署后确认:

- [ ] admin 地址正确且私钥安全保管
- [ ] operator 地址已授权
- [ ] asset token 地址正确
- [ ] 合约已暂停或正常状态符合预期
- [ ] 管理费在允许范围内 (0-1000 bps)
- [ ] Timelock delay 已设置 (RewardDistributor)
