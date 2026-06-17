# 页面与交互设计方案

## 1. 全局交互

### Mode Switch

模式：

- `HR View`
- `Senior Dev View`

HR View：

- 强调岗位匹配、技术栈、项目结果、验证状态。
- 隐藏过多文件路径和 P0 细节。
- 文案更像简历和面试介绍。

Senior Dev View：

- 展示架构图、风险列表、代码路径、测试命令、修复路线。
- 默认展开 P0/P1 findings。

### Time Budget Switch

模式：

- `3 min`
- `8 min`
- `20 min`

用途：

- 面试时根据时间快速切换讲解深度。
- 每种模式生成不同讲解顺序。

### Demo Wallet

状态：

- `Read-only`
- `Demo Connected`

用途：

- 保留币圈产品心智。
- 不触发真实钱包连接。
- 不要求用户签名。

## 2. Dashboard Page

### 目标

让用户在 30 秒内理解：

- 你不是单一合约开发，而是 Web3 infrastructure/backend 候选人。
- 三个项目的优先级和用途不同。
- 当前最强主项目是 Web3 Backend。

### 首屏模块

1. Top Navigation
   - Brand：`BNB Web3 Career Terminal`
   - Nav：`Dashboard`、`Projects`、`Compare`、`Interview`
   - Mode switch
   - Demo Wallet button

2. Portfolio Overview
   - 三张项目卡。
   - 每张项目卡显示岗位、栈、分数、状态。

3. Offer Strategy Panel
   - Primary target：Web3 Backend / Blockchain Backend。
   - Secondary：Protocol / Rust Blockchain。
   - Supporting：Smart Contract Security Awareness。

4. Current Blockers
   - Web3：Go version blocks tests。
   - Rust：fmt check fails / protocol semantics pending。
   - Solidity：no tests found。

### 第二屏模块

1. Role Fit Radar
2. Project Health Matrix
3. 30-day Roadmap
4. Evidence Strip：测试命令、文档、核心文件数量。

## 3. Project Detail Page

### 通用结构

```text
Project Header
  Role target
  One-line pitch
  Tech stack
  Verification status

Architecture Map
  Visual pipeline

Implementation Evidence
  Modules
  Key files
  What works now

Risk Register
  P0/P1/P2
  Impact
  Fix plan

Interview Pitch
  HR version
  Senior engineer version
```

## 4. Web3 Backend Detail

### 页面标题

```text
Multi-chain Asset Platform
```

### 一句话

```text
Go-based blockchain asset backend covering scanner, parser, confirmation worker, ledger service, RPC gateway, and reorg handling.
```

### 架构图节点

```text
RPC Providers
  -> Scanner
  -> Raw Events
  -> Parser
  -> Deposits
  -> Confirm Worker
  -> Deposit Confirmed
  -> Ledger Service
  -> Balances / Ledger Entries
```

### 核心展示点

- RPC provider selection and circuit breaker。
- ERC-20 Transfer event scanning。
- watched address deposit detection。
- confirmation count state machine。
- ledger and balance update。
- reorg detector and compensator。

### 风险展示

默认展示：

- Go version blocks test verification。
- deposit confirmed publish retry needs outbox。
- NATS core ack is not durable。
- ledger amount calculation must use big.Int/NUMERIC。

## 5. Protocol Rust Detail

### 页面标题

```text
Verifiable Rust Chain Node
```

### 一句话

```text
Minimal Rust blockchain execution engine with state transition, mempool, block validation, Merkle roots, and storage abstraction.
```

### 架构图节点

```text
Signed Transaction
  -> Mempool
  -> Block Builder
  -> Executor
  -> StateDB
  -> Receipt Root / State Root
  -> Validator
  -> Storage
```

### 核心展示点

- Account model。
- nonce validation。
- u128 value transfer。
- atomic block execution。
- receipt content hash。
- tx root/state root/receipt root validation。
- storage trait。

### 风险展示

- `cargo fmt --check` fails。
- signature verification is placeholder。
- mempool selection should not remove before execution。
- fork-aware storage still pending。

## 6. Smart Contract Detail

### 页面标题

```text
Secure Yield Vault
```

### 一句话

```text
Solidity vault prototype showing share/asset accounting, reward injection, role-based control, pause support, and security test planning.
```

### 架构图节点

```text
User Deposit
  -> Asset Transfer
  -> Share Mint
  -> Reward Injection
  -> Withdraw / Redeem
  -> Emergency Path
  -> Security Tests
```

### 核心展示点

- SafeERC20。
- AccessControl。
- ReentrancyGuard。
- Pausable。
- fee-on-transfer handling。
- reward accounting。
- timelock-style reward distributor。

### 风险展示

- No tests found。
- invariant tests missing。
- audit/security docs missing。
- ERC-4626 compatibility needs clarification。

## 7. Compare Page

### 比较维度

- Market Fit
- Engineering Depth
- Test Readiness
- Interview Strength
- Current Risk
- Best Job Target

### 默认结论

```text
Use Web3 Backend as the flagship project.
Use Rust Protocol as the depth project.
Use Smart Contract as the security-awareness project until tests are complete.
```

## 8. Interview Mode Page

### 3-minute mode

结构：

1. Background：backend + Go/Rust/Python。
2. Main target：Web3 backend infrastructure。
3. Project 1：multi-chain asset backend。
4. Project 2：Rust protocol fundamentals。
5. Project 3：Solidity security awareness。
6. Closing：next improvement focus。

### 8-minute mode

增加：

- Web3 backend pipeline。
- Rust executor/state validation。
- Smart contract risk surface。
- Testing status honesty。

### 20-minute mode

增加：

- 具体 P0/P1 风险。
- 修复路线。
- 架构取舍。
- 为什么不把 Solidity 作为唯一主线。

## 9. 空状态和错误状态

### No Tests

显示：

```text
No executable tests found yet.
This is a known readiness gap, not hidden.
```

### Blocked Verification

显示：

```text
Verification blocked by local toolchain mismatch.
Required: Go 1.25.0
Local: Go 1.24.4
```

### Risk Open

显示：

```text
Open risk
Needs fix before production-grade claim.
```

## 10. 文案风格

语气：

- 专业。
- 直接。
- 不夸大。
- 能承认限制。
- 像资深工程师在讲 trade-off。

避免：

- “革命性”
- “下一代”
- “生产级”除非有测试和验证。
- “已审计”除非真有审计报告。

