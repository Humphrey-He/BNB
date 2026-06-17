# multi-chain-asset-platform 功能架构

## 1. 系统定位

`multi-chain-asset-platform` 是一个面向钱包、交易所、Custody、Crypto Payment 的 Web3 资产后端平台。第一版聚焦 EVM 链，支持 Ethereum / BSC / Polygon 等链的充值监听、确认数、链重组补偿、账本余额、提现广播和 API 查询。

项目重点不是“扫到链上事件”这么简单，而是完整展示后端工程能力：

- 链上数据可靠同步到链下数据库。
- 充值和提现都有明确状态机。
- 用户余额通过账本流水推导，可审计、可回放。
- 链重组、RPC 异常、重复消息、交易卡住都有补偿策略。
- 所有关键路径可观测。

## 2. 系统架构图

```text
External Chains
  Ethereum / BSC / Polygon
        |
        v
+--------------------+       +----------------------+
| rpc-gateway        |       | provider healthcheck |
| failover / retry   |<----->| rate limit / circuit |
+---------+----------+       +----------------------+
          |
          v
+---------+----------+       +----------------------+
| scanner            |-----> | blocks               |
| block/log scanner  |       | scan_checkpoints     |
+---------+----------+       +----------------------+
          |
          v
+---------+----------+       +----------------------+
| parser             |-----> | chain_events         |
| ERC20 log parser   |       | deposits             |
+---------+----------+       +----------------------+
          |
          v
+---------+----------+       +----------------------+
| confirm-worker     |-----> | deposits status      |
| confirmations      |       | ledger_entries       |
+---------+----------+       +----------------------+
          |
          v
+---------+----------+       +----------------------+
| ledger-service     |-----> | balances             |
| available/frozen   |       | audit trail          |
+---------+----------+       +----------------------+

Withdrawal Flow

API -> withdrawal-service -> signer boundary -> broadcaster -> chain
                         |                 |
                         v                 v
                   withdrawals       nonce_allocations
                   ledger_entries    tx_attempts

Observability

All services -> Prometheus metrics -> Grafana dashboards -> alerts
```

## 3. 服务边界

### 3.1 rpc-gateway

职责：

- 管理多个 RPC provider。
- 处理限流、超时、重试、熔断。
- 对同一链提供统一客户端。
- 对关键读操作支持交叉校验，例如最新区块高度、区块 hash。

核心策略：

- 读请求优先使用健康 provider。
- 写请求即交易广播，需要记录 provider、返回 hash、错误码。
- provider 连续失败进入 cool-down。
- 扫描任务使用批量 `eth_getLogs`，窗口大小可配置。

### 3.2 scanner

职责：

- 读取链上新区块。
- 拉取区间内 ERC-20 `Transfer` logs。
- 保存区块 header，用于链重组检测。
- 推送原始 logs 到队列。

关键设计：

- 按链维护 `scan_checkpoints`。
- 扫描窗口不要太大，避免 RPC 超时。
- 每个 log 使用 `chain_id + tx_hash + log_index` 作为天然幂等键。
- 延迟扫描 safe block，减少 reorg 影响；同时仍保存 latest lag 指标。

### 3.3 parser

职责：

- 解析 ERC-20 Transfer 事件。
- 识别平台监听地址的充值。
- 标准化链上事件。
- 写入 `chain_events` 和 `deposits`。

充值识别：

- `to_address` 命中平台地址。
- `token contract` 命中支持的 token。
- `value > 0`。
- `tx_hash + log_index` 未处理过。

### 3.4 confirm-worker

职责：

- 根据最新安全区块高度推进确认数。
- 将充值从 `detected` 推进到 `confirmed`。
- 触发账本入账。
- 处理被 reorg 影响的充值。

状态流转：

```text
detected -> pending_confirmation -> confirmed
      |             |                  |
      v             v                  v
   orphaned      orphaned           orphaned
      |
      v
    failed
```

### 3.5 ledger-service

职责：

- 使用账本流水维护余额。
- 区分 `available`、`frozen`、`total`。
- 支持余额回放和审计。

账本原则：

- 所有余额变化必须来自 `ledger_entries`。
- `balances` 是快照，不是唯一事实来源。
- 充值确认后记一条 credit。
- 提现申请后先冻结余额。
- 提现广播并确认后扣减冻结余额。
- 提现失败或取消后释放冻结余额。

### 3.6 withdrawal-service

职责：

- 创建提现申请。
- 校验余额、token、地址、限额。
- 冻结用户余额。
- 生成广播任务。
- 管理提现状态机。

状态流转：

```text
created -> risk_checking -> approved -> signing -> broadcasting
       -> broadcasted -> confirmed
       -> failed / canceled
```

说明：

- 第一版不做真实私钥托管，可以用 signer interface 或本地 dev key。
- 文档中明确生产环境应接 HSM、MPC 或独立签名服务。

### 3.7 broadcaster

职责：

- 构造 EVM 交易。
- 分配 nonce。
- 估算 gas 和 fee。
- 广播交易。
- 查询 receipt。
- 处理卡住交易和 replacement transaction。

关键策略：

- 按 `chain_id + from_address` 串行分配 nonce。
- `nonce_allocations` 记录每次分配。
- 广播失败可重试，但不能重复分配 nonce。
- 交易长时间 pending 时使用更高 gas fee 替换。
- receipt 成功后推进提现状态。

### 3.8 reorg-detector

职责：

- 检测本地区块 hash 与链上 canonical hash 是否不一致。
- 标记孤块。
- 回滚受影响的 deposits 和 ledger entries。
- 从分叉点重新扫描。

补偿原则：

- 不直接删除业务记录，优先标记 `orphaned`。
- 所有余额修正通过反向账本流水完成。
- 重扫新链后重新产生 canonical 事件。

### 3.9 api-server

主要 API：

| Method | Path | 说明 |
|---|---|---|
| GET | `/api/v1/chains` | 链配置 |
| GET | `/api/v1/tokens` | Token 配置 |
| GET | `/api/v1/addresses/{addr}/balances` | 地址余额 |
| GET | `/api/v1/addresses/{addr}/transactions` | 交易历史 |
| GET | `/api/v1/deposits` | 充值列表 |
| GET | `/api/v1/withdrawals` | 提现列表 |
| POST | `/api/v1/withdrawals` | 创建提现 |
| GET | `/api/v1/chain-status` | 扫描高度和延迟 |
| GET | `/healthz` | 健康检查 |
| GET | `/metrics` | Prometheus 指标 |

## 4. 数据模型

### 4.1 核心表

```sql
CREATE TABLE chains (
    id BIGSERIAL PRIMARY KEY,
    chain_id BIGINT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    native_symbol TEXT NOT NULL,
    finality_confirmations INT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE rpc_providers (
    id BIGSERIAL PRIMARY KEY,
    chain_id BIGINT NOT NULL REFERENCES chains(id),
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    weight INT NOT NULL DEFAULT 100,
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_error TEXT,
    last_checked_at TIMESTAMPTZ
);

CREATE TABLE tokens (
    id BIGSERIAL PRIMARY KEY,
    chain_id BIGINT NOT NULL REFERENCES chains(id),
    contract_address TEXT NOT NULL,
    symbol TEXT NOT NULL,
    decimals INT NOT NULL,
    is_native BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    UNIQUE(chain_id, contract_address)
);

CREATE TABLE watched_addresses (
    id BIGSERIAL PRIMARY KEY,
    chain_id BIGINT NOT NULL REFERENCES chains(id),
    address TEXT NOT NULL,
    owner_ref TEXT,
    label TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    UNIQUE(chain_id, address)
);

CREATE TABLE blocks (
    id BIGSERIAL PRIMARY KEY,
    chain_id BIGINT NOT NULL REFERENCES chains(id),
    block_number BIGINT NOT NULL,
    block_hash TEXT NOT NULL,
    parent_hash TEXT NOT NULL,
    block_time TIMESTAMPTZ,
    is_orphaned BOOLEAN NOT NULL DEFAULT false,
    scanned_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(chain_id, block_number, block_hash)
);

CREATE TABLE chain_events (
    id BIGSERIAL PRIMARY KEY,
    chain_id BIGINT NOT NULL REFERENCES chains(id),
    tx_hash TEXT NOT NULL,
    log_index INT NOT NULL,
    block_number BIGINT NOT NULL,
    block_hash TEXT NOT NULL,
    contract_address TEXT NOT NULL,
    event_name TEXT NOT NULL,
    from_address TEXT,
    to_address TEXT,
    amount NUMERIC(78, 0),
    is_orphaned BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(chain_id, tx_hash, log_index)
);

CREATE TABLE deposits (
    id BIGSERIAL PRIMARY KEY,
    chain_id BIGINT NOT NULL REFERENCES chains(id),
    token_id BIGINT NOT NULL REFERENCES tokens(id),
    tx_hash TEXT NOT NULL,
    log_index INT NOT NULL,
    from_address TEXT NOT NULL,
    to_address TEXT NOT NULL,
    amount NUMERIC(78, 0) NOT NULL,
    block_number BIGINT NOT NULL,
    status TEXT NOT NULL,
    confirmations INT NOT NULL DEFAULT 0,
    idempotency_key TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE withdrawals (
    id BIGSERIAL PRIMARY KEY,
    chain_id BIGINT NOT NULL REFERENCES chains(id),
    token_id BIGINT NOT NULL REFERENCES tokens(id),
    from_address TEXT NOT NULL,
    to_address TEXT NOT NULL,
    amount NUMERIC(78, 0) NOT NULL,
    status TEXT NOT NULL,
    tx_hash TEXT,
    nonce BIGINT,
    idempotency_key TEXT NOT NULL UNIQUE,
    failure_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE ledger_entries (
    id BIGSERIAL PRIMARY KEY,
    account_address TEXT NOT NULL,
    chain_id BIGINT NOT NULL REFERENCES chains(id),
    token_id BIGINT NOT NULL REFERENCES tokens(id),
    direction TEXT NOT NULL,
    amount NUMERIC(78, 0) NOT NULL,
    entry_type TEXT NOT NULL,
    reference_type TEXT NOT NULL,
    reference_id BIGINT NOT NULL,
    reversal_of BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(reference_type, reference_id, entry_type)
);

CREATE TABLE balances (
    id BIGSERIAL PRIMARY KEY,
    account_address TEXT NOT NULL,
    chain_id BIGINT NOT NULL REFERENCES chains(id),
    token_id BIGINT NOT NULL REFERENCES tokens(id),
    available NUMERIC(78, 0) NOT NULL DEFAULT 0,
    frozen NUMERIC(78, 0) NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(account_address, chain_id, token_id)
);

CREATE TABLE nonce_allocations (
    id BIGSERIAL PRIMARY KEY,
    chain_id BIGINT NOT NULL REFERENCES chains(id),
    from_address TEXT NOT NULL,
    nonce BIGINT NOT NULL,
    withdrawal_id BIGINT REFERENCES withdrawals(id),
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(chain_id, from_address, nonce)
);
```

## 5. 队列设计

| Topic | 生产者 | 消费者 | 用途 |
|---|---|---|---|
| `raw_logs` | scanner | parser | 原始链上 logs |
| `parsed_events` | parser | confirm-worker | 标准化事件 |
| `deposit_confirmed` | confirm-worker | ledger-service | 充值入账 |
| `reorg_detected` | reorg-detector | confirm-worker, ledger-service | 链重组补偿 |
| `withdrawal_requested` | withdrawal-service | broadcaster | 提现广播 |
| `tx_receipt_checked` | broadcaster | withdrawal-service | 提现确认 |
| `dead_letter` | all workers | ops/manual | 失败消息 |

## 6. 幂等与一致性

关键幂等键：

- 充值事件：`chain_id + tx_hash + log_index`
- 提现申请：客户端传入 `Idempotency-Key`
- 账本流水：`reference_type + reference_id + entry_type`
- nonce：`chain_id + from_address + nonce`

一致性策略：

- 数据库唯一约束兜底。
- 消费者按 at-least-once 设计，重复消费必须无副作用。
- 余额变化只通过账本流水产生。
- 链重组使用反向流水修正，不硬删历史。
- 提现先冻结，再广播，确认后扣减冻结。

## 7. 监控指标

| 指标 | 类型 | 说明 |
|---|---|---|
| `scanner_block_lag` | Gauge | 当前扫描落后区块数 |
| `scanner_rpc_errors_total` | Counter | RPC 错误数 |
| `scanner_logs_scanned_total` | Counter | 扫描 logs 数 |
| `parser_events_total` | Counter | 解析事件数 |
| `deposits_confirmed_total` | Counter | 确认充值数 |
| `deposits_orphaned_total` | Counter | reorg 充值数 |
| `withdrawals_broadcasted_total` | Counter | 已广播提现数 |
| `withdrawals_failed_total` | Counter | 失败提现数 |
| `nonce_allocation_conflicts_total` | Counter | nonce 冲突数 |
| `queue_consumer_lag` | Gauge | 队列积压 |
| `api_request_duration_seconds` | Histogram | API 延迟 |

## 8. 本地开发目录

```text
multi-chain-asset-platform/
├── cmd/
│   ├── scanner/
│   ├── parser/
│   ├── confirm-worker/
│   ├── ledger-service/
│   ├── withdrawal-service/
│   ├── broadcaster/
│   ├── api-server/
│   └── reorg-detector/
├── internal/
│   ├── chain/
│   ├── config/
│   ├── db/
│   ├── ledger/
│   ├── models/
│   ├── queue/
│   ├── rpcgateway/
│   ├── service/
│   └── metrics/
├── migrations/
├── scripts/
│   ├── seed-dev-data.py
│   └── reconcile-balances.py
├── docker-compose.yml
├── README.md
├── docs/
│   ├── ARCHITECTURE.md
│   ├── PROJECT_PLAN.md
│   └── RUNBOOK.md
└── tests/
    └── integration/
```

