# 真实出金链路技术设计

更新时间：2026-06-18

## 设计目标

设计一条面向 EVM 测试链 / 主网的真实出金链路，满足：

- 可真实签名和广播
- 可审批、可风控、可人工介入
- 并发提现下 nonce 不冲突
- 失败后余额与账本可恢复
- 所有关键动作可审计

## 当前实现审计

### 已有能力

- `internal/api/handlers/withdrawal_service.go`
  - 创建提现时冻结余额
  - 写入 `withdrawals`
  - 写入 `freeze` 账本
  - manual review 驳回时做 `unfreeze`
- `internal/withdrawalservice`
  - 有风险检查流程
  - 有 `manual_review / approved / signing` 状态推进
  - 有 NATS 事件
- `internal/broadcaster`
  - 有广播主循环
  - 有回执检查
  - 有失败释放 / reverted 补偿
- `internal/repository/nonce_allocation_repository.go`
  - 有 nonce allocation 表和 repository

### 明确缺口

- `signTransaction` 还是占位实现
- `buildTransactionData` 还是字符串占位，不是链上交易
- `ReceiptRPCAdapter` 只实现了 receipt 查询
- `broadcaster` 里虽然声明了 `NonceRepository`，但还没和真正的 repository + 链上 nonce 对齐逻辑打通
- 缺少独立 `cmd/withdrawal-worker` 和 `cmd/broadcaster`
- 当前提现状态机过于粗，不足以表达 submitted / dropped / replaced / reverted

## 目标状态机

建议将 `withdrawals.status` 演进为：

- `created`
- `risk_checking`
- `manual_review`
- `approved`
- `signing`
- `signed`
- `broadcasting`
- `broadcasted`
- `confirming`
- `confirmed`
- `failed`
- `canceled`
- `replaced`

状态解释：

- `created`: 已创建并冻结余额
- `manual_review`: 等待人工审批
- `approved`: 已通过审批，等待签名
- `signing`: 开始生成交易 payload
- `signed`: 已生成带 nonce 的待广播交易
- `broadcasted`: 节点已返回 tx hash
- `confirming`: 正在等待 receipt / 确认数
- `confirmed`: 链上成功，账务完成结算
- `failed`: 确认不会成功，需要释放或补偿
- `replaced`: 被相同 nonce 的更高 fee 交易替代

## 建议的服务拆分

### 1. Withdrawal API

职责：

- 创建提现请求
- 冻结余额
- 幂等控制
- 提供内部审批 / 驳回 / 重试接口

输入：

- `chain_id`
- `token_id`
- `from_address`
- `to_address`
- `amount`
- `idempotency_key`

输出：

- `withdrawal_id`
- `status`

### 2. withdrawal-worker

职责：

- 拉取 `created` 状态提现
- 执行风险检查
- 决定：
  - auto approve
  - manual review
  - fail
- 发布 `withdrawal_approved`

### 3. signer

职责：

- 根据 token 类型构造真实 EVM 交易
- 绑定 chain id
- 绑定 nonce
- 绑定 gas 参数
- 从安全密钥源完成签名

建议抽象：

```go
type Signer interface {
    SignNativeTransfer(ctx context.Context, req NativeTransferSignRequest) (*SignedTx, error)
    SignERC20Transfer(ctx context.Context, req ERC20TransferSignRequest) (*SignedTx, error)
}
```

建议的 signer 后端：

1. `LocalHexKeySigner`
2. `KMSKeySigner`
3. `HSMKeySigner`

本项目建议先做：

- `LocalHexKeySigner` 跑测试链
- 接口层提前按 KMS/HSM 设计，后续可替换

### 4. broadcaster

职责：

- 获取或分配 nonce
- 获取 gas 参数
- 调 signer 生成 signed tx
- 发送 raw transaction
- 保存 attempt
- 推进 withdrawal 状态

### 5. receipt-poller / reconciliation

职责：

- 扫描 `broadcasted / confirming`
- 查询 receipt
- 处理：
  - success
  - reverted
  - not found
  - dropped
  - replaced

## 真实交易构造

### Native transfer

适用于：

- 原生币提现，如 BNB / ETH

必要字段：

- `to`
- `value`
- `nonce`
- `gas_limit`
- `chain_id`
- `max_fee_per_gas`
- `max_priority_fee_per_gas`

### ERC-20 transfer

适用于：

- Token 提现

必要字段：

- `token_contract`
- `data = transfer(to, amount)`
- `value = 0`
- `nonce`
- `gas_limit`
- `chain_id`
- `fee params`

建议增加 token 元数据支持：

- `is_native`
- `contract_address`
- `decimals`

## RPC 设计

建议把当前 broadcaster 的 RPC interface 落成真实 EVM 版本：

```go
type EVMRPCClient interface {
    PendingNonceAt(ctx context.Context, chainID int64, address string) (uint64, error)
    SuggestGasPrice(ctx context.Context, chainID int64) (*big.Int, error)
    SuggestGasTipCap(ctx context.Context, chainID int64) (*big.Int, error)
    EstimateGas(ctx context.Context, chainID int64, call ethereum.CallMsg) (uint64, error)
    SendRawTransaction(ctx context.Context, chainID int64, rawTx []byte) (string, error)
    TransactionReceipt(ctx context.Context, chainID int64, txHash string) (*types.Receipt, error)
}
```

建议优先支持：

- EIP-1559
- provider failover
- timeout / retry
- provider health logging

## nonce 管理设计

### 核心原则

- 同一 `chain_id + from_address` 必须串行分配
- replacement transaction 复用旧 nonce
- nonce 既参考链上，也参考本地

### 推荐分配算法

1. 对 `chain_id + from_address` 加数据库锁
2. 获取链上 `pending nonce`
3. 查询本地 `nonce_allocations` 最大未结束 nonce
4. `next_nonce = max(chain_pending_nonce, local_next_nonce)`
5. 插入 `nonce_allocations`
6. 将 allocation id 绑定到 withdrawal / attempt

### 为什么当前实现不够

当前 `GetNextAvailableNonce()` 只是：

- 查本地 `MAX(nonce) + 1`

这会导致：

- 不知道链上 pending nonce 已推进到哪里
- 多实例部署时容易竞争
- 重启后难以识别 dropped / replaced 情况

### nonce allocation 推荐状态

- `allocated`
- `signed`
- `submitted`
- `mined`
- `released`
- `replaced`
- `expired`

## 审批与风控设计

当前已有风控规则：

- 热钱包地址一致性
- 白名单地址
- 自动审批金额上限
- 源余额充足性

建议继续增强：

- 每日累计出金额度
- 每地址 24h 次数限制
- 黑名单地址
- 链 / token 单独开关
- 目的地址标签校验
- 运营侧强制 manual review

建议新增表或配置源：

- `withdrawal_policies`
- `withdrawal_destination_whitelist`
- `withdrawal_blacklist`

建议审批动作带审计字段：

- `reviewed_by`
- `reviewed_at`
- `review_reason`
- `review_decision`

## 广播尝试与重试模型

建议新增 `withdrawal_attempts` 表。

推荐字段：

- `id`
- `withdrawal_id`
- `attempt_no`
- `nonce`
- `tx_hash`
- `rpc_provider_id`
- `raw_tx_hash`
- `max_fee_per_gas`
- `max_priority_fee_per_gas`
- `gas_limit`
- `send_status`
- `send_error`
- `receipt_status`
- `created_at`

这样做的价值：

- 一笔 withdrawal 可以记录多次广播
- replacement transaction 有独立 attempt
- 便于排查“到底有没有真正发出去”

## 失败处理矩阵

### 场景 A：签名前失败

例如：

- nonce 获取失败
- gas 估算失败
- signer 不可用

处理：

- 不扣减 frozen
- withdrawal 留在 `approved` 或进入 `failed`
- 可自动重试

### 场景 B：发送请求失败，但不确定是否已入池

例如：

- RPC timeout
- connection reset
- upstream 502/503

处理：

- 不要立即释放 frozen
- 标记为 `broadcasted?unknown`
- 继续通过 tx hash 或 nonce 回查
- 若无 tx hash，可按 nonce 查询链上 pending / latest 变化并人工介入

### 场景 C：明确未广播成功

例如：

- 本地签名失败
- RPC 明确返回 invalid tx 且未接收

处理：

- 自动释放 frozen
- 记录失败原因

### 场景 D：链上 reverted

处理：

- withdrawal -> `failed`
- 写 reversal 账本
- available balance 补回

### 场景 E：同 nonce replacement

处理：

- 原 attempt 标为 `replaced`
- 新 attempt 复用原 nonce
- withdrawal 指向最新 tx hash

## 账务一致性要求

### 创建提现

- balance:
  - `available -= amount`
  - `frozen += amount`
- ledger:
  - `freeze`

### 广播并成功确认

- balance:
  - `frozen -= amount`
- ledger:
  - `withdrawal`

### 驳回或广播前失败

- balance:
  - `available += amount`
  - `frozen -= amount`
- ledger:
  - `unfreeze`

### 已广播但链上 reverted

- balance:
  - `available += amount`
- ledger:
  - `reversal`

## 推荐新增命令入口

建议新增：

- `cmd/withdrawal-worker/main.go`
- `cmd/broadcaster/main.go`

环境变量建议：

- `WITHDRAWAL_HOT_WALLET_ADDRESS`
- `WITHDRAWAL_SIGNER_BACKEND`
- `WITHDRAWAL_SIGNER_PRIVATE_KEY`
- `WITHDRAWAL_KMS_KEY_ID`
- `WITHDRAWAL_MAX_AUTO_APPROVE_AMOUNT`
- `WITHDRAWAL_REQUIRE_WHITELIST`
- `WITHDRAWAL_ALLOWED_DESTINATIONS`
- `WITHDRAWAL_BROADCAST_MAX_RETRIES`
- `WITHDRAWAL_RECEIPT_POLL_INTERVAL`

## 推荐迁移清单

### migration 1

- 扩展 `withdrawals`
- 扩展 `nonce_allocations`

### migration 2

- 新增 `withdrawal_attempts`

### migration 3

- 新增风控策略表

## 开发落点建议

优先改这些文件：

- `internal/broadcaster/broadcaster.go`
- `internal/broadcaster/rpc_receipt.go`
- `internal/repository/nonce_allocation_repository.go`
- `internal/repository/withdrawal_repository.go`
- `internal/withdrawalservice/worker.go`
- `internal/api/handlers/withdrawal_service.go`
- `scripts/migrations/*.sql`

新增这些文件：

- `cmd/withdrawal-worker/main.go`
- `cmd/broadcaster/main.go`
- `internal/broadcaster/evm_rpc.go`
- `internal/broadcaster/evm_signer.go`
- `internal/broadcaster/attempt_repository.go` 或 repository 层对应文件

## 最小可行实施顺序

1. 真 signer
2. 真 RPC send + nonce + gas
3. 独立 broadcaster 命令
4. 独立 withdrawal-worker 命令
5. `withdrawal_attempts`
6. nonce 串行化
7. replacement 和 retry 分类
8. manual review 审计增强

## 结论

当前仓库已经有真实出金系统 60% 左右的结构基础，最缺的不是业务流程图，而是把“签名、广播、nonce、补偿”四条关键技术路径从占位实现替换成真实实现。
