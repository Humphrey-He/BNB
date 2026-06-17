# SecureYieldVault 功能架构

## 1. 系统定位

`secure-yield-vault` 是一个安全导向的 ERC-4626 风格收益金库项目。它的目标不是做复杂收益策略，而是用适中的业务复杂度展示智能合约工程岗位最看重的能力：

- shares/assets 模型与 proper rounding
- 存款、赎回、奖励注入、收益结算
- 角色权限和 timelock
- 暂停和 emergency withdraw
- 完整测试：unit、fuzz、invariant、attack PoC
- 静态分析、审计报告和 known limitations
- 链下事件监听和数据校验

## 2. 系统架构

```
┌──────────────────────────────────────────────────────────────────┐
│                         User / DApp                               │
└───────────────────────────────┬──────────────────────────────────┘
                                │ tx
                                ▼
┌──────────────────────────────────────────────────────────────────┐
│                      SecureYieldVault                             │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │  deposit / mint / withdraw / redeem                        │  │
│  │  convertToShares / convertToAssets                         │  │
│  │  injectRewards / emergencyWithdraw                          │  │
│  │  pause / unpause                                          │  │
│  └────────────────────────────────────────────────────────────┘  │
│                              │                                   │
│  ┌───────────────────────────┼───────────────────────────────┐  │
│  │ ERC20 (shares)            │ SafeERC20 (asset)            │  │
│  │ AccessControl             │ Pausable                      │  │
│  │ ReentrancyGuard           │ Math.mulDiv                   │  │
│  └───────────────────────────┴───────────────────────────────┘  │
└───────────────────────────────┬──────────────────────────────────┘
                                │
                                ▼
┌──────────────────────────────────────────────────────────────────┐
│                    RewardDistributor                              │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │  addRewardPool / setRewardRate                             │  │
│  │  scheduleTimelock / executeTimelock / cancelTimelock       │  │
│  └────────────────────────────────────────────────────────────┘  │
│                              │                                   │
│  ┌───────────────────────────┼───────────────────────────────┐  │
│  │ AccessControl            │ Timelock                       │  │
│  │ Pausable                 │ ReentrancyGuard               │  │
│  └───────────────────────────┴───────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌──────────────────────────────────────────────────────────────────┐
│                         Off-chain                                 │
│  ┌─────────────────┐  ┌──────────────────┐  ┌───────────────┐   │
│  │ Event Listener  │  │   SQLite DB      │  │  REST API     │   │
│  │ (Python/Go)    │──│   (Event Store)  │──│  (Optional)   │   │
│  └─────────────────┘  └──────────────────┘  └───────────────┘   │
└──────────────────────────────────────────────────────────────────┘
```

## 3. 合约模块

### 3.1 SecureYieldVault

**职责:**
- 接收 asset 存款，发行 vault shares
- 根据 shares/assets 汇率处理 withdraw/redeem
- 接收奖励注入，提升 share 价值
- 支持暂停、紧急提现、事件输出

**继承:**
- ERC20, ERC20Permit - shares 代币化
- AccessControl - 多角色权限
- Pausable - 暂停机制
- ReentrancyGuard - 重入防护

**角色:**
| 角色 | 权限 |
|------|------|
| `ADMIN_ROLE` | pause, unpause, setManagementFee, emergencyWithdraw |
| `OPERATOR_ROLE` | injectRewards |
| `DEFAULT_ADMIN_ROLE` | 包含 ADMIN_ROLE + 角色管理 |

**关键实现:**

```solidity
// fee-on-transfer 防护
uint256 balanceBefore = asset.balanceOf(address(this));
asset.safeTransferFrom(msg.sender, address(this), assets);
uint256 received = asset.balanceOf(address(this)) - balanceBefore;
require(received >= assets, "Fee deducted");

// 奖励分离会计
uint256 principalToWithdraw = (assets * totalAssets) / (totalAssets + accumulatedRewards);
uint256 rewardToWithdraw = assets - principalToWithdraw;
totalAssets -= principalToWithdraw;
accumulatedRewards -= rewardToWithdraw;
```

### 3.2 RewardDistributor

**职责:**
- 允许 OPERATOR_ROLE 注入奖励
- 支持 timelock 保护的敏感操作
- 记录奖励来源和分发历史

**Timelock 设计:**
- 统一的 txHash 计算方式 (不含 eta)
- schedule → pending → execute 流程
- 7 天宽限期

### 3.3 AccessControl 角色层级

```
DEFAULT_ADMIN_ROLE
    │
    ├── ADMIN_ROLE ──────────────────┐
    │   ├── pause/unpause            │
    │   ├── setManagementFee         │
    │   └── emergencyWithdraw        │
    │                                │
    └── OPERATOR_ROLE                │
        └── injectRewards           │
                                     │
TIMELOCK_ADMIN_ROLE ─────────────────┘
    ├── scheduleTimelock
    └── executeTimelock
```

## 4. 数据模型

### 4.1 SecureYieldVault Storage

```solidity
IERC20  public immutable asset;        // 底层资产代币
uint256 public totalAssets;             // 本金 (不含奖励)
uint256 public accumulatedRewards;      // 累计奖励
uint256 public managementFee;           // 管理费 (bps)
```

### 4.2 Shares/Assets 转换

使用 OpenZeppelin Math.mulDiv 防止溢出：

```solidity
// assets -> shares (1:1 初始，奖励注入后比例变化)
shares = Math.mulDiv(assets, totalSupply, totalAssets + accumulatedRewards, Floor)

// shares -> assets
assets = Math.mulDiv(shares, totalAssets + accumulatedRewards, totalSupply, Floor)
```

### 4.3 链下数据库 (SQLite)

```sql
CREATE TABLE events (
    id INTEGER PRIMARY KEY,
    event_type TEXT,       -- 'Deposited' | 'Withdrawn' | 'RewardAdded'
    user_address TEXT,
    assets TEXT,            -- 大数用字符串
    shares TEXT,
    block_number INTEGER,
    tx_hash TEXT,
    timestamp INTEGER
);

CREATE TABLE user_positions (
    user_address TEXT PRIMARY KEY,
    shares TEXT,
    last_update_block INTEGER
);
```

## 5. 安全设计

### 5.1 攻击面与防护

| 攻击类型 | 防护措施 | 验证方式 |
|---------|---------|---------|
| 重入攻击 | ReentrancyGuard + CEI | fuzz test |
| fee-on-transfer | balanceBefore/After | 单元测试 |
| 假 ERC20 | safeTransferFrom | MaliciousERC20 mock |
| 溢出攻击 | Math.mulDiv | fuzz test |
| 权限失控 | AccessControl 多角色 | 权限测试 |
| 重复提取 | 状态更新在 transfer 前 | 单元测试 |
| 紧急提取滥用 | 必须暂停 + burn shares | 单元测试 |

### 5.2 CEI 模式

```solidity
function withdraw(...) external nonReentrant whenNotPaused {
    // 1. Check
    require(shares > 0, "Zero shares");
    require(balanceOf(owner) >= shares, "Insufficient balance");

    // 2. Effects (状态更新)
    totalAssets -= principalToWithdraw;
    accumulatedRewards -= rewardToWithdraw;

    // 3. Interactions (external calls last)
    _burn(owner, shares);
    asset.safeTransfer(receiver, assets);
}
```

## 6. 测试策略

### 6.1 测试金字塔

```
        ┌─────────────┐
        │ Invariant   │  全局不变量验证
        │   Tests     │
        └──────┬──────┘
               │
        ┌──────┴──────┐
        │    Fuzz     │  随机输入测试
        │   Tests     │
        └──────┬──────┘
               │
        ┌──────┴──────┐
        │ Unit +      │  正常流/边界/安全
        │ Integration │
        └─────────────┘
```

### 6.2 核心测试场景

**正常流程:**
- deposit → withdraw (1:1)
- deposit → injectRewards → withdraw (含奖励)
- mint → redeem
- 多用户同时操作

**边界条件:**
- 零金额 deposit/withdraw
- 超过余额的 withdraw
- 未存款就 withdraw
- fee-on-transfer token

**权限控制:**
- 非 admin 调用 pause
- 非 operator 调用 injectRewards
- emergencyWithdraw 未暂停状态

**安全测试:**
- 重入攻击 (withdraw 回调)
- 重复提取同一批次
- 奖励计算精度

### 6.3 Invariants

```solidity
// 不变量 1: totalSupply = sum(balanceOf)
invariantTotalSupplyConsistency()

// 不变量 2: totalAssets 不会为负
invariantTotalAssetsNonNegative()

// 不变量 3: 奖励不会凭空增加
invariantRewardConservation()

// 不变量 4: emergencyWithdraw 后 shares 减少
invariantEmergencyWithdrawBurnsShares()
```

## 7. 部署架构

### 7.1 环境

| 环境 | 网络 | 用途 |
|------|------|------|
| Local | `anvil` | 开发测试 |
| Testnet | Sepolia / BSC Testnet | 演示验证 |

### 7.2 部署顺序

1. 部署 MockERC20 (asset token)
2. 部署 SecureYieldVault
3. 部署 RewardDistributor (可选)
4. 授权 operator 角色
5. 注入初始奖励 (可选)
6. 验证配置

### 7.3 配置检查清单

- [ ] admin 地址正确
- [ ] operator 地址已授权
- [ ] asset token 地址正确
- [ ] pause 状态正确
- [ ] managementFee 在允许范围内

## 8. 文件结构

```
secure-yield-vault/
├── contracts/
│   ├── SecureYieldVault.sol       # 主金库
│   ├── RewardDistributor.sol      # 奖励分发
│   ├── interfaces/
│   │   └── ISecureYieldVault.sol
│   └── mocks/
│       ├── MockERC20.sol
│       ├── FeeOnTransferToken.sol
│       ├── MaliciousERC20.sol
│       └── ReentrantReceiver.sol
├── script/
│   └── Deploy.s.sol              # 部署脚本
├── test/
│   ├── SecureYieldVault.t.sol    # 主测试
│   ├── RewardDistributor.t.sol
│   └── Invariants.t.sol          # 不变量测试
├── docs/
│   ├── README.md                  # 项目说明
│   ├── ARCHITECTURE.md            # 架构文档
│   ├── PROJECT_PLAN.md            # 项目计划
│   └── ANALYSIS.md                # 分析文档
├── foundry.toml
├── remappings.txt
└── README.md                     # 入口文档
```
