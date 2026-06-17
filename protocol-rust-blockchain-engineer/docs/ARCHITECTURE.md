# verifiable-rust-chain-node 功能架构

## 1. 系统定位

`verifiable-rust-chain-node` 是一个最小可验证区块链节点项目。它不是完整公链，不实现复杂 BFT，也不追求 EVM 兼容。它的目标是展示 Protocol Engineer / Rust Blockchain Engineer 岗位最关心的底层能力：

- 交易是否可验证。
- 区块是否可验证。
- 状态执行是否确定。
- 节点同步是否有明确协议。
- mempool 是否考虑 nonce、fee、容量和驱逐。
- 存储和索引是否能支持区块导入和查询。
- 性能瓶颈是否可观测。

## 2. 总体架构

```text
                   +----------------------+
                   | RPC API              |
                   | send_tx / query      |
                   +----------+-----------+
                              |
                              v
                   +----------+-----------+
                   | Mempool              |
                   | nonce / fee / evict  |
                   +----------+-----------+
                              |
                              v
                   +----------+-----------+
                   | Block Producer       |
                   | PoA / block template |
                   +----------+-----------+
                              |
                              v
+------------------+----------+-----------+------------------+
| Validation       | Executor             | State            |
| header / tx/root | deterministic exec   | accounts/root    |
+------------------+----------+-----------+------------------+
                              |
                              v
                   +----------+-----------+
                   | Storage              |
                   | blocks/state/indexes |
                   +----------+-----------+
                              |
                              v
                   +----------+-----------+
                   | P2P Sync             |
                   | headers -> blocks    |
                   +----------------------+
```

## 3. 核心模块

### 3.1 types

核心数据结构：

```rust
pub struct SignedTransaction {
    pub from: Address,
    pub to: Address,
    pub value: u128,
    pub nonce: u64,
    pub gas_limit: u64,
    pub max_fee_per_gas: u64,
    pub signature: Signature,
}

pub struct Header {
    pub parent_hash: Hash,
    pub number: u64,
    pub timestamp: u64,
    pub tx_root: Hash,
    pub receipt_root: Hash,
    pub state_root: Hash,
    pub gas_used: u64,
    pub gas_limit: u64,
    pub proposer: Address,
}

pub struct Block {
    pub header: Header,
    pub transactions: Vec<SignedTransaction>,
}

pub struct Receipt {
    pub tx_hash: Hash,
    pub success: bool,
    pub gas_used: u64,
    pub error: Option<String>,
}

pub struct Account {
    pub balance: u128,
    pub nonce: u64,
}
```

### 3.2 crypto

职责：

- hash 计算。
- 签名校验。
- 地址派生。
- 交易 hash 和 block hash 计算。

第一版可以使用成熟 crate，不手写密码学算法。

### 3.3 state

职责：

- 管理账户余额和 nonce。
- 执行状态变更。
- 生成 state root。
- 支持 snapshot 和 rollback。

确定性要求：

- 同一个区块在不同节点执行后必须得到相同 state root。
- 状态读写顺序要稳定。
- 序列化格式要稳定。

### 3.4 executor

职责：

- 执行单笔交易。
- 校验余额、nonce、gas limit。
- 应用余额转移。
- 生成 receipt。
- 执行整个区块并返回新 state root。

执行流程：

```text
for tx in block.transactions:
    verify signature
    verify nonce
    verify balance
    apply transfer
    increment nonce
    generate receipt
commit state
compute roots
```

### 3.5 mempool

职责：

- 接收 RPC 和 P2P 广播的交易。
- 校验签名和基础字段。
- 按账户 nonce 排队。
- 按 fee priority 选择可打包交易。
- 控制容量并驱逐低价值交易。

策略：

- 同一账户必须按 nonce 连续打包。
- nonce gap 交易进入 pending queue。
- 相同 nonce 的更高 fee 交易可以替换旧交易。
- mempool 容量超限时驱逐低 fee 或过期交易。

### 3.6 validation

职责：

- 验证区块 header。
- 验证 parent hash。
- 验证 block number。
- 验证 gas limit。
- 验证 tx root、receipt root、state root。
- 验证 proposer 是否在 PoA validator set。

区块导入原则：

- 永远不要只信任对端给出的 state root。
- 导入区块时必须重新执行交易。
- 执行结果 root 与 header 不一致则拒绝区块。

### 3.7 storage

推荐 RocksDB。

Column families：

- `blocks_by_hash`
- `canonical_hash_by_number`
- `headers`
- `transactions`
- `receipts`
- `account_state`
- `metadata`

关键索引：

- `block_number -> canonical_hash`
- `block_hash -> block`
- `tx_hash -> tx location`
- `tx_hash -> receipt`
- `address -> account`

### 3.8 consensus

第一版使用 PoA：

- 固定 validator set。
- 单节点或多节点轮流出块。
- 验证 proposer 是否合法。

fork choice：

- 优先选择累计高度更高的链。
- 高度相同时选择 hash 更小或先到达的链，规则必须 deterministic。
- reorg 时回滚旧 canonical chain 并导入新 chain。

### 3.9 p2p

推荐 libp2p。

协议：

- `Status`：交换 chain id、best height、best hash。
- `GetHeaders` / `Headers`：headers-first sync。
- `GetBlocks` / `Blocks`：按 header 请求 block bodies。
- `NewBlock`：广播新区块。
- `NewTransaction`：广播交易。

同步流程：

```text
connect peer
exchange Status
request headers from local height + 1
validate headers
request missing blocks
execute and validate blocks
update canonical chain
```

peer scoring：

- 返回无效 header 降分。
- 返回无效 block 降分。
- 超时降分。
- 连续有效响应加分。
- 低于阈值断开。

### 3.10 rpc

API：

| Method | 说明 |
|---|---|
| `send_transaction` | 提交签名交易 |
| `get_transaction` | 查询交易 |
| `get_receipt` | 查询 receipt |
| `get_block_by_number` | 查询区块 |
| `get_block_by_hash` | 查询区块 |
| `get_balance` | 查询余额 |
| `get_nonce` | 查询 nonce |
| `get_mempool_status` | 查询 mempool |
| `get_sync_status` | 查询同步状态 |

## 4. 数据流

### 4.1 交易提交流程

```text
RPC send_transaction
  -> decode SignedTransaction
  -> crypto.verify_signature
  -> mempool.validate_basic
  -> mempool.insert
  -> return tx_hash
```

### 4.2 出块流程

```text
tick block time
  -> mempool.select_transactions
  -> executor.execute_block
  -> compute tx_root / receipt_root / state_root
  -> build header
  -> validation.verify_local_block
  -> storage.commit_block
  -> p2p.broadcast NewBlock
```

### 4.3 区块导入流程

```text
receive block
  -> validation.verify_header
  -> validation.verify_parent
  -> executor.reexecute_block
  -> compare roots
  -> storage.commit_block
  -> update fork choice
```

## 5. Benchmark

必须有的 benchmark：

- 单笔交易执行耗时。
- 100 / 1,000 / 10,000 笔交易区块执行耗时。
- block import latency。
- mempool insert latency。
- mempool select latency。
- RocksDB read/write latency。

## 6. 目录结构

```text
verifiable-rust-chain-node/
├── Cargo.toml
├── src/
│   ├── main.rs
│   ├── lib.rs
│   ├── types/
│   ├── crypto/
│   ├── state/
│   ├── executor/
│   ├── mempool/
│   ├── validation/
│   ├── storage/
│   ├── consensus/
│   ├── p2p/
│   ├── rpc/
│   └── metrics/
├── benches/
│   ├── executor_bench.rs
│   ├── mempool_bench.rs
│   └── storage_bench.rs
├── tests/
│   ├── block_import_test.rs
│   ├── fork_choice_test.rs
│   └── sync_test.rs
├── docs/
│   ├── ARCHITECTURE.md
│   ├── PROJECT_PLAN.md
│   └── PROTOCOL.md
└── README.md
```

