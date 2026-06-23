# 真实出金链路规划

更新时间：2026-06-18

## 目标

把当前仓库里已经存在的提现骨架，推进成可接真实链路、可控风险、可重试、可回滚、可观测的生产级出金系统。

本轮聚焦 4 个核心主题：

1. 真正的签名与广播
2. 出金审批与风控
3. nonce 管理
4. 失败重试与回滚

## 当前基线

仓库里已经具备可继续演进的基础，不是从零开始：

- 提现创建接口已支持冻结余额，并写入 `freeze` 账本分录
- 手工审批 / 驳回接口已存在，驳回时会执行 `unfreeze`
- `withdrawalservice` 已有风控骨架，支持 `manual_review`
- `broadcaster` 已有广播、回执轮询、失败补偿骨架
- `nonce_allocations` 表和 repository 已存在
- receipt 成功 / reverted / 广播失败后的账本补偿函数已存在

但距离真实生产链路还有几个关键缺口：

- 交易签名仍是占位实现，不是真实 EVM 签名
- 广播数据仍是占位字符串，不是真实原始交易
- receipt adapter 只实现了查 receipt，`GasPrice / SendRawTransaction / NonceAt` 还未接通
- nonce 分配还不是生产级，没有链上 pending nonce 对齐、没有串行锁定策略
- 缺少 withdrawal worker / broadcaster 的独立启动入口和部署路径
- 表结构还不够支撑广播尝试、replacement、回执时间线、审计字段
- 错误分类与重试策略还比较粗，无法区分“可重试 / 不确定 / 必须人工介入”

## 建议的最终链路

目标链路建议收敛为：

`API -> withdrawal-service -> risk/manual-review -> signer -> broadcaster -> receipt-poller -> settlement/reversal`

对应账务动作：

1. 创建提现：`available -> frozen`
2. 广播成功并确认上链：`frozen -> debit settled`
3. 广播失败且未上链：`frozen -> available`
4. 已广播但链上 reverted：`debit reversal -> available`

## 分阶段实施

### P0：把骨架补成“可跑的真实链 MVP”

目标：

- 支持真实 EVM native transfer 或 ERC-20 transfer 签名
- 能真正发到测试链
- 能基于 tx hash 查 receipt 并推进状态

本阶段必须完成：

- 抽象 `Signer` 接口，替换当前占位 `signTransaction`
- 抽象真实 `EVM RPC client`，实现：
  - `NonceAt`
  - `GasPrice` 或 EIP-1559 fee 建议
  - `SendRawTransaction`
  - `GetTransactionReceipt`
- 引入 `cmd/withdrawal-worker`
- 引入 `cmd/broadcaster`
- self-hosted 部署流程把这两个服务纳入 systemd
- 在测试链跑通：
  - 创建提现
  - 审批
  - 广播
  - receipt confirmed
  - 账本完成结算

验收标准：

- 可以在测试链上看到真实 tx hash
- `withdrawals.status` 从 `created` 走到 `confirmed`
- `balances.frozen_balance` 和 `ledger_entries` 结果正确

### P1：把审批与风控从骨架升级为业务可控

目标：

- 自动审批和人工审批边界清晰
- 能阻断高风险目标地址和异常金额

本阶段必须完成：

- 风控规则拆成可配置策略，而不是只在代码里硬编码
- 增加审批审计字段：
  - `review_status`
  - `review_reason`
  - `reviewed_by`
  - `reviewed_at`
- 增加提现策略：
  - 白名单目标地址
  - 单笔限额
  - 单地址日限额
  - 热钱包一致性校验
  - Token / chain 维度开关
- 管理后台或内部 API 支持：
  - approve
  - reject
  - force-fail
  - retry

验收标准：

- 超限提现自动进入 `manual_review`
- 审批行为可追踪到操作人和原因
- 驳回后余额与账本自动回滚

### P2：把 nonce 管理做成生产级

目标：

- 并发提现不冲突
- replacement transaction 可控
- 服务重启后 nonce 不乱

本阶段必须完成：

- nonce 分配基于数据库行锁或 advisory lock 串行化
- 分配 nonce 时同时参考：
  - 链上 latest nonce
  - 链上 pending nonce
  - 本地 `nonce_allocations`
- 扩展 `nonce_allocations.status`，建议包括：
  - `allocated`
  - `signed`
  - `submitted`
  - `mined`
  - `released`
  - `replaced`
  - `expired`
- withdrawal 与 nonce allocation 强绑定
- replacement tx 不重新分配 nonce，只提升 fee

验收标准：

- 同一热钱包并发 20 笔提现，无 nonce 冲突
- broadcaster 重启后不会重复占用 nonce
- 同 nonce replacement 能被完整审计

### P3：把失败重试与回滚做成可运营

目标：

- 明确知道哪些错误该重试，哪些必须人工介入
- 失败后资金状态可恢复

本阶段必须完成：

- 错误分类：
  - pre-broadcast hard fail
  - pre-broadcast retryable
  - submitted but unknown
  - on-chain reverted
  - receipt timeout
- 引入重试策略：
  - 指数退避
  - 最大重试次数
  - replacement fee bump
- 引入 reconciliation 任务：
  - 扫描 `broadcasted` 且长时间无 receipt 的提现
  - 基于 tx hash 回查链上最终状态
  - 对 orphan / replaced / dropped 交易做补偿
- 扩展账本补偿场景：
  - release before send
  - reversal after reverted
  - replacement after dropped

验收标准：

- 广播失败不会卡死 frozen balance
- reverted 会自动反向补偿
- unknown 状态提现最终都能收敛到 `confirmed / failed / manual_review`

## 需要新增的表字段

建议在 `withdrawals` 增加：

- `fee_limit`
- `gas_limit`
- `gas_price`
- `max_fee_per_gas`
- `max_priority_fee_per_gas`
- `raw_tx_hash`
- `signed_payload_ref`
- `broadcast_attempts`
- `last_broadcast_at`
- `first_broadcast_at`
- `confirmed_at`
- `failed_at`
- `reviewed_by`
- `reviewed_at`
- `review_reason`
- `rpc_provider_id`
- `replacement_of`

建议在 `nonce_allocations` 增加：

- `tx_hash`
- `last_seen_onchain_at`
- `released_at`
- `replaced_by_tx_hash`

建议新增 `withdrawal_attempts` 表：

- 一笔 withdrawal 多次签名 / 广播 / replacement 的历史
- 记录：
  - attempt_no
  - nonce
  - gas params
  - tx_hash
  - provider
  - send_error
  - receipt_status
  - created_at

## 代码实施顺序

建议按下面顺序开发，阻力最小：

1. `internal/broadcaster`
   - 先做真实 RPC client 和真实 signer
2. `cmd/withdrawal-worker`
   - 把现有 worker 变成独立可部署服务
3. `cmd/broadcaster`
   - 把 broadcaster 变成独立可部署服务
4. `internal/repository`
   - 扩展 withdrawals / nonce_allocations / attempts
5. `internal/withdrawalservice`
   - 风控策略和 manual review 审计增强
6. `internal/api/handlers`
   - 增强审批、重试、失败介入 API
7. `scripts/migrations`
   - 落库新增字段和新表
8. 部署脚本
   - 把提现 worker 和 broadcaster 纳入 systemd / workflow

## 这轮建议优先做什么

如果以“最短时间跑通真实测试链出金”为目标，优先级建议如下：

1. 真实 EVM signer + broadcaster
2. 独立 `withdrawal-worker` / `broadcaster` 服务入口
3. receipt 轮询与最终状态推进
4. nonce 串行分配
5. approval / manual review 审计增强
6. replacement / retry / rollback 完整化

## 一句话结论

当前仓库已经有真实出金链路的半成品骨架，最关键的下一步不是再补概念，而是把“真实签名 + 真实广播 + 生产级 nonce + 可恢复重试”四件事接通。
