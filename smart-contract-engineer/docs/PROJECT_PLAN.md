# Secure Yield Vault - 项目规划

## 1. 项目概述

**项目名称**: secure-yield-vault
**项目类型**: 安全导向智能合约开发 + 审计流程 + 链下事件校验
**目标**: 建立一个能证明 Solidity、DeFi vault 设计、安全测试、审计意识和链上/链下协作能力的作品集项目

## 2. 核心功能范围

| 功能模块 | 描述 | 优先级 | 状态 |
|---------|------|--------|------|
| ERC-4626 风格 Vault | deposit/mint/withdraw/redeem | P0 | ✅ |
| Shares/Assets 模型 | 份额和资产转换，处理 rounding | P0 | ✅ |
| 奖励注入 | operator 注入收益，提高 share 价值 | P0 | ✅ |
| 权限控制 | AccessControl + timelock 保护关键参数 | P0 | ✅ |
| 暂停机制 | deposit/withdraw granular pause | P0 | ✅ |
| 紧急提现 | emergency withdraw，burn shares | P0 | ✅ |
| 事件日志 | Deposited/Withdrawn/RewardAdded/EmergencyWithdrawn | P1 | ✅ |
| 部署脚本 | 测试网部署脚本 | P1 | TODO |
| 链下监听 | Go/Python 事件监听脚本 | P2 | TODO |
| 安全分析文档 | SECURITY.md | P0 | TODO |
| 审计报告 | AUDIT_REPORT.md | P0 | TODO |

## 3. 里程碑计划

### Week 1: 基础搭建 ✅

- [x] 初始化 Foundry 项目
- [x] 部署开发环境（本地节点/测试网）
- [x] 编写 `MockERC20.sol` 测试代币
- [x] 编写 `FeeOnTransferToken.sol`、`MaliciousERC20.sol`、`ReentrantReceiver.sol`
- [x] 实现 `ISecureYieldVault.sol` 接口定义
- [x] 学习 Foundry 基本命令：`forge test`、`forge script`

**交付物**: 可编译的空白项目 + MockERC20 ✅

### Week 2: Vault 核心合约开发 ✅

- [x] 实现 `SecureYieldVault.sol`
  - [x] deposit/mint
  - [x] withdraw/redeem
  - [x] totalAssets
  - [x] convertToShares/convertToAssets
  - [x] rounding 策略
- [x] 实现 `RewardDistributor.sol` 奖励分发
- [x] 集成 OpenZeppelin: AccessControl、Pausable、ReentrancyGuard、SafeERC20
- [x] 定义事件: `Deposited`、`Withdrawn`、`RewardAdded`、`EmergencyWithdrawn`、`FeeUpdated`

**Week 2 安全修复:**
- [x] P0: fee-on-transfer 防护 (balanceBefore/balanceAfter)
- [x] P0: withdraw/redeem 奖励扣减修复
- [x] P0: emergencyWithdraw burn shares
- [x] P0: Timelock hash 一致性
- [x] P1: pause/unpause 函数
- [x] P1: Math.mulDiv 溢出防护

**交付物**: 核心合约源码 ✅

### Week 3: 测试覆盖与攻击 PoC

- [x] 单元测试：deposit、mint、withdraw、redeem
- [x] 单元测试：奖励注入后 share 价值变化
- [x] 边界条件测试：零金额、余额不足、暂停状态
- [x] 权限测试：非管理员不能修改参数
- [x] 安全测试：
  - [x] 重入攻击防护
  - [x] 非标准 ERC-20
  - [x] fee-on-transfer token 处理
  - [x] rounding 边界
  - [x] 整数溢出防护
- [ ] Foundry Fuzz Test
- [ ] Attack PoC tests

**交付物**: 完整测试套件 (覆盖率 > 90%) ✅

**额外修复**: emergencyWithdraw 不需要 _spendAllowance（admin 提取自己的资产不需要授权）

### Week 4: 安全分析与链下服务

- [ ] 运行 Slither 安全分析，修复发现的问题
- [ ] Foundry Invariant Test（不变量测试）
  - [ ] sum(userShares) = totalSupply
  - [ ] 用户不能赎回超过份额
  - [ ] totalAssets 与合约余额关系一致
  - [ ] paused 状态下操作必须 revert
- [ ] 编写 `SECURITY.md` 安全分析文档
- [ ] 编写 `AUDIT_REPORT.md`
- [ ] 编写部署脚本 (`script/Deploy.s.sol`)
- [ ] 编写链下事件监听脚本 (Python `web3.py` 或 Go)

**交付物**: Slither 报告 + SECURITY.md + AUDIT_REPORT.md + 部署脚本 + 监听脚本

## 4. 技术栈

| 类别 | 技术 | 状态 |
|------|------|------|
| 智能合约 | Solidity ^0.8.20 | ✅ |
| 测试框架 | Foundry (Forge) | ✅ |
| 库 | OpenZeppelin ^5.0 | ✅ |
| 安全分析 | Slither, Foundry invariant | TODO |
| 链下脚本 | Python web3.py / Go | TODO |
| 数据库 | SQLite (事件存储) | TODO |

## 5. 交付物清单

| 交付物 | 文件路径 | 状态 |
|--------|----------|------|
| 合约源码 | `contracts/*.sol` | ✅ |
| 单元测试 | `test/*.t.sol` | TODO |
| 模糊测试 | `test/*.t.sol` | TODO |
| 安全分析文档 | `docs/SECURITY.md` | TODO |
| 审计报告 | `docs/AUDIT_REPORT.md` | TODO |
| 部署脚本 | `script/Deploy.s.sol` | TODO |
| 链下监听脚本 | `scripts/listener.py` | TODO |
| 项目架构文档 | `docs/ARCHITECTURE.md` | ✅ |
| 部署指南 | `docs/DEPLOYMENT.md` | ✅ |
| README | `README.md` | ✅ |

## 6. 学习补齐（并行进行）

| 周数 | 主题 | 资源 |
|------|------|------|
| Week 1 | Solidity 基础、ERC-20、事件、modifier | Solidity Docs ✅ |
| Week 1 | Foundry 进阶：fuzz/invariant test | Foundry Book ✅ |
| Week 2 | 常见攻击面：重入、权限失控、舍入误差、价格操纵 | SWC Registry ✅ |
| Week 2 | 阅读安全案例：DAO、bZx、Cream、Euler | 公开审计报告 ✅ |
| Week 3-4 | Slither、Echidna 集成 | Slither README |

## 7. 质量标准

- [ ] 所有合约通过 Slither 检测（无 HIGH/CRITICAL 漏洞）
- [ ] 测试覆盖率 > 90%
- [ ] Invariant Test 通过
- [ ] Attack PoC 测试通过
- [ ] shares/assets rounding 有明确测试
- [ ] fee-on-transfer/rebasing token 支持边界写入 SECURITY.md
- [ ] 部署脚本可重复执行（幂等性）

## 8. 当前进度

```
Week 1: ████████████ 100%
Week 2: ████████████ 100% + 安全修复
Week 3: ████████░░░░░ 80% (单元测试完成, 待Fuzz/Invariant)
Week 4: ░░░░░░░░░░░░░ 0%
```

## 9. 合约安全特性清单

| 特性 | 实现 | 状态 |
|------|------|------|
| CEI 模式 | Checks-Effects-Interactions | ✅ |
| 重入防护 | ReentrancyGuard | ✅ |
| fee-on-transfer 防护 | balanceBefore/balanceAfter | ✅ |
| 乘法溢出防护 | OZ Math.mulDiv | ✅ |
| 多角色权限 | AccessControl | ✅ |
| 暂停机制 | Pausable | ✅ |
| 奖励分离会计 | totalAssets + accumulatedRewards | ✅ |
| EmergencyWithdraw 安全 | burn shares + 需暂停 | ✅ |
| Timelock 一致性 | 统一 hash 计算 | ✅ |
