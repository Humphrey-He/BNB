# SecureYieldVault

一个安全的 ERC-4626 风格收益金库，展示了生产级 DeFi 合约的关键安全实践。

## 项目概述

| 属性 | 值 |
|------|-----|
| 合约类型 | ERC-4626 风格 Vault |
| Solidity 版本 | ^0.8.20 |
| 测试框架 | Foundry (Forge) |
| 安全库 | OpenZeppelin v5.0 |

## 核心功能

- **Shares/Assets 模型**: 1:1 初始比例，支持奖励累积
- **存款/取款**: deposit, mint, withdraw, redeem
- **奖励注入**: Operator 可注入奖励，提高 share 价值
- **权限控制**: AccessControl 多角色 + Timelock
- **暂停机制**: granular pause 支持
- **紧急提取**: 暂停状态下可提取本金

## 合约

### SecureYieldVault

主金库合约，处理存款、取款、shares 铸造/销毁。

**角色:**
- `ADMIN_ROLE`: 暂停/恢复、设置费率、紧急提取
- `OPERATOR_ROLE`: 注入奖励

**安全特性:**
- CEI (Checks-Effects-Interactions) 模式
- ReentrancyGuard 防护
- fee-on-transfer token 处理（测量实际到账）
- 乘除法使用 Math.mulDiv 防止溢出
- principal/reward 分离会计

### RewardDistributor

独立的奖励分发器，支持 timelock 保护。

**角色:**
- `OPERATOR_ROLE`: 添加奖励池
- `TIMELOCK_ADMIN_ROLE`: 提议和执行 timelock 操作
- `DEFAULT_ADMIN_ROLE`: 取消 timelock 操作

## 快速开始

### 前置条件

- Git
- Foundry (nightly 或 stable)

### 安装

```bash
# 克隆项目
git clone https://github.com/your-org/secure-yield-vault.git
cd secure-yield-vault

# 安装依赖
forge install

# 编译
forge build

# 运行测试
forge test
```

### 基本使用

#### 部署到本地节点

```bash
# 启动 anvil (本地测试网络)
anvil

# 在另一个终端部署
forge script script/Deploy.s.sol --rpc-url http://localhost:8545 --private-key 0x... --broadcast
```

#### 与合约交互

```bash
# 铸造测试代币
cast send <TOKEN_ADDRESS> "mint(address,uint256)" <USER_ADDRESS> 1000000000000000000 --rpc-url http://localhost:8545

# 批准 Vault 使用代币
cast send <TOKEN_ADDRESS> "approve(address,uint256)" <VAULT_ADDRESS> 1000000000000000000 --rpc-url http://localhost:8545

# 存款
cast send <VAULT_ADDRESS> "deposit(uint256,address)" 1000000000000000000 <USER_ADDRESS> --rpc-url http://localhost:8545

# 查询 shares
cast call <VAULT_ADDRESS> "balanceOf(address)" <USER_ADDRESS> --rpc-url http://localhost:8545

# 提取
cast send <VAULT_ADDRESS> "withdraw(uint256,address,address)" 1000000000000000000 <USER_ADDRESS> <USER_ADDRESS> --rpc-url http://localhost:8545
```

## 合约接口

### SecureYieldVault

```solidity
// 存款：存入 assets，获得 shares
function deposit(uint256 assets, address receiver) external returns (uint256 shares)

// 铸造：指定 shares 数量，存入所需 assets
function mint(uint256 shares, address receiver) external returns (uint256 assets)

// 取款：提取 assets，销毁 shares
function withdraw(uint256 assets, address receiver, address owner) external returns (uint256 shares)

// 赎回：销毁 shares，获得 assets
function redeem(uint256 shares, address receiver, address owner) external returns (uint256 assets)

// 注入奖励 (需要 OPERATOR_ROLE)
function injectRewards(uint256 amount) external

// 紧急提取 (需要 ADMIN_ROLE，合约必须已暂停)
function emergencyWithdraw(uint256 assets, address receiver) external

// 暂停/恢复 (需要 ADMIN_ROLE)
function pause() external
function unpause() external

// 查询
function totalAssets() external view returns (uint256)
function totalAssetsWithRewards() external view returns (uint256)
function convertToShares(uint256 assets) external view returns (uint256)
function convertToAssets(uint256 shares) external view returns (uint256)
function balanceOfAssets(address owner) external view returns (uint256)
```

### RewardDistributor

```solidity
// 添加奖励池 (需要 OPERATOR_ROLE)
function addRewardPool(uint256 amount) external

// 设置奖励率 (需要 DEFAULT_ADMIN_ROLE)
function setRewardRate(uint256 newRate) external

// Timelock 操作 (需要 TIMELOCK_ADMIN_ROLE)
function scheduleTimelock(address target, bytes calldata data, uint256 value) external returns (bytes32)
function executeTimelock(address target, bytes calldata data, uint256 value) external

// 暂停/恢复
function pause() external
function unpause() external
```

## 测试

```bash
# 运行所有测试
forge test

# 详细输出
forge test -vv

# 模糊测试
forge test --match-test "testFuzz" -vv

# 不变量测试
forge test --match-test "testInvariant" -vv

# 覆盖率
forge coverage
```

## 安全特性

| 特性 | 实现 |
|------|------|
| 重入防护 | ReentrancyGuard + CEI 模式 |
| 假 ERC20 防护 | balanceBefore/balanceAfter 测量实际到账 |
| 乘法溢出防护 | OpenZeppelin Math.mulDiv |
| 权限控制 | AccessControl 多角色 |
| 暂停保护 | whenNotPaused 修饰符 |
| Timelock | 可配置的延迟执行 |
| 紧急提取 | burn shares + 暂停要求 |

## 事件

```solidity
// SecureYieldVault
event Deposited(address indexed sender, address indexed owner, uint256 assets, uint256 shares)
event Withdrawn(address indexed sender, address indexed receiver, address indexed owner, uint256 assets, uint256 shares)
event RewardAdded(uint256 amount)
event EmergencyWithdrawn(address indexed sender, uint256 assets, uint256 shares)
event FeeUpdated(uint256 oldFee, uint256 newFee)

// RewardDistributor
event RewardRateUpdated(uint256 oldRate, uint256 newRate)
event RewardsDistributed(address indexed vault, uint256 amount)
event TimelockPending(bytes32 indexed txHash, uint256 eta)
event TimelockExecuted(bytes32 indexed txHash, uint256 eta)
```

## 项目结构

```
contracts/
├── SecureYieldVault.sol      # 主金库合约
├── RewardDistributor.sol     # 奖励分发合约
├── interfaces/
│   └── ISecureYieldVault.sol # 合约接口
└── mocks/
    ├── MockERC20.sol         # 标准测试代币
    ├── FeeOnTransferToken.sol # 带手续费的代币
    ├── MaliciousERC20.sol    # 非标准返回值代币
    └── ReentrantReceiver.sol  # 重入测试合约

script/
└── Deploy.s.sol              # 部署脚本

test/
└── *.t.sol                   # 测试文件
```

## 参考

- [Foundry Book](https://book.getfoundry.sh/)
- [OpenZeppelin Contracts](https://docs.openzeppelin.com/contracts/)
- [ERC-4626 Standard](https://eips.ethereum.org/EIPS/eip-4626)
- [Solidity Docs](https://docs.soliditylang.org/)
