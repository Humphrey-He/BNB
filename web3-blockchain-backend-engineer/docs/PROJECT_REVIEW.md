# multi-chain-asset-platform 项目说明

## 一、项目概述

**项目名称**：multi-chain-asset-platform
**技术栈**：Go + PostgreSQL + NATS + go-ethereum + Prometheus
**定位**：多链 Web3 资产后端平台，覆盖链上充值监听、确认数处理、链重组补偿、账本余额、提现广播和 API 查询

### 核心功能

| 模块 | 职责 |
|------|------|
| RPC Gateway | 多 Provider 容灾、加权随机选择、熔断限流 |
| Scanner | 区块遍历、ERC-20 Transfer 事件拉取、扫描进度持久化 |
| Parser | 链上事件解析、充值识别、幂等写入 |
| Confirm-Worker | 确认数跟踪、状态流转（detected→pending_confirmation→confirmed） |
| Ledger-Service | 账本流水、可用/冻结余额管理 |
| Reorg-Detector | 链重组检测、 orphaned 标记、余额补偿 |

### 数据流

```
链上事件 → Scanner → NATS(raw_events) → Parser → deposits(detected)
                                                       ↓
                            Confirm-Worker ← NATS(parsed_events)
                                   ↓
                         deposits(confirmed) → Ledger-Service
                                   ↓
                            balances + ledger_entries
```

---

## 二、技术亮点（与普通 indexer 项目的差异）

### 1. 资金链路可靠性
- **幂等性保证**：每个充值事件用 `chain_id + tx_hash + log_index` 作为唯一键，重复消息不会重复入账
- **原子操作**：账本流水和余额更新在同一数据库事务内，防止数据不一致
- **可靠发布**：确认数达标后发布事件，失败自动重试，直到成功才从 tracking 移除
- **条件更新**：`ConfirmWithCondition` 使用 `WHERE status = 'pending_confirmation' AND confirmations >= target` 防止状态竞态

### 2. 链重组补偿
- 保存完整区块 header（block_hash、parent_hash）用于 reorg 检测
- 检测到 reorg 时标记 orphaned 区块，创建反向账本流水修正余额
- 发布 reorg 事件供其他服务消费

### 3. RPC Provider 管理
- 加权随机选择（`math/rand` 实现，权重越高选中概率越大）
- 熔断器模式：连续失败进入 cooldown，恢复后逐步启用
- 批量 `eth_getLogs` 窗口可配置，避免超时

### 4. 账本式余额
- 所有余额变化来自 `ledger_entries`，`balances` 是快照而非唯一事实来源
- 支持 `available_balance` 和 `frozen_balance` 分离
- 每条账本记录包含 `balance_before` 和 `balance_after`，可完整回放和审计

### 5. 监控可观测性
- Prometheus metrics：扫描延迟、错误率、确认数、队列积压
- 每个服务独立日志字段（chain_id、tx_hash、block_number）

---

## 三、个人优势

### 1. 完整闭环能力
不只是"扫到链上事件"，而是从**链上事件 → 确认 → 入账 → 余额**的完整闭环。理解资金系统不允许丢钱的要求，每个环节都有幂等和补偿机制。

### 2. 工程严谨性
- P0/P1 问题分级处理，不为了快速上线牺牲资金安全
- 数据库事务保证原子性，唯一约束兜底幂等
- 条件更新避免状态竞态，不依赖锁或乐观锁的复杂实现

### 3. 后端综合能力
- **Go 基建**：scanner、worker、API 服务、数据库 CRUD
- **分布式系统**：NATS 消息队列、异步任务、重试机制
- **数据库**：PostgreSQL 事务、唯一约束、UPSERT、查询优化
- **区块链**：EVM、ERC-20 事件解析、RPC 调用、交易构造

### 4. 问题驱动学习
Week 1-4 迭代过程中主动发现并修复了：
- JSON 字段名不匹配（PascalCase vs snake_case）
- 确认数只递增不覆盖
- 发布失败不重试
- Credit 后 Ledger 失败导致重复入账
- 并发创建余额可能丢失数据

这些问题的发现和修复体现了对资金系统可靠性的敏感度。

---

## 四、岗位匹配度分析

### 目标岗位类型

| 岗位类型 | 匹配度 | 说明 |
|----------|--------|------|
| Web3/Blockchain 后端工程师 | ★★★★★ | 核心技术栈完全匹配，完整项目闭环 |
| 交易所/钱包后端工程师 | ★★★★★ | 充值监听、确认数、账本余额、提现广播完整覆盖 |
| DeFi/Custody 后端工程师 | ★★★★☆ | 链上数据同步可靠，余额可审计 |
| 传统 Go 后端工程师 | ★★★☆☆ | 分布式系统、数据库、消息队列能力可迁移 |

### 面试高频问题准备

| 问题 | 项目可展示的答案 |
|------|-----------------|
| ERC-20 充值监听原理 | topic0 = keccak256("Transfer")，scanner 批量拉取 logs |
| 幂等性保证机制 | chain_id+tx_hash+log_index 唯一键，数据库约束兜底 |
| 确认数的作用 | 链重组概率随确认数指数下降，12 确认=大概率 finality |
| 链重组时的数据补偿 | 标记 orphaned，创建反向账本流水，不硬删数据 |
| 余额一致性保证 | ledger_entries 是唯一来源，balances 是快照，可回放校验 |
| 交易广播与 nonce 处理 | nonce_allocations 表，按 chain_id+from_address 串行分配 |
| RPC Provider 容灾 | 加权随机选择 + 熔断器 + 失败 cooldown |
| 死信队列和人工补偿 | 消息处理失败重试N次后进入死信，人工介入 |

### 技术广度覆盖

- **链上数据工程**：scanner、parser、block 存储
- **状态机设计**：充值状态流转、提现状态流转
- **账本系统**：双记账、balance_before/after、可回放审计
- **分布式任务**：NATS 异步消息、重试机制
- **监控告警**：Prometheus metrics、SLO 监控
- **数据库工程**：事务、索引、UPSERT、查询优化

---

## 五、项目待完善（面试可主动提及）

以下模块是架构设计中有但尚未实现的，是后续迭代方向：

| 模块 | 状态 | 说明 |
|------|------|------|
| withdrawal-service | 待开发 | 提现申请、冻结余额、风险检查 |
| broadcaster | 待开发 | 交易构造、nonce 管理、gas 估算、replacement tx |
| api-server | 待开发 | REST API：余额/充值/提现/交易历史 |
| JetStream | 待迁移 | 当前用 core NATS，生产环境应迁移到 JetStream durable consumer |
| HSM/MPC signer | 设计说明 | 真实签名应接 HSM/MPC，本项目 signer interface 作为占位 |

面试时主动说明这些"未完成但有设计"的部分，反而能体现架构思维和后续学习方向。

---

## 六、代码质量自评

- **构建通过**：`go build ./...` 无错误
- **架构清晰**：服务边界明确，依赖通过接口注入
- **命名规范**：Go 命名惯例，结构体字段语义明确
- **错误处理**：关键路径有日志记录和错误传播
- **事务安全**：资金链路关键操作在事务内执行
- **配置外化**：无硬编码，魔数上常量或配置文件

---

*文档版本：v1.0 | 2026-05-06*
