# verifiable-rust-chain-node 协议设计

## 1. 概述

本文档描述 `verifiable-rust-chain-node` 的链上协议规范，包括数据类型定义、状态转换规则、Merkle root 计算方式以及区块验证流程。

## 2. 数据类型

### 2.1 Hash 与 Address

| 类型 | 长度 | 说明 |
|------|------|------|
| `Hash` | 32 bytes | SHA-256 哈希值 |
| `Address` | 20 bytes | 账户地址，由公钥派生 |

### 2.2 SignedTransaction

```rust
pub struct SignedTransaction {
    pub from: Address,           // 发送者地址
    pub to: Option<Address>,      // 接收者地址，None 表示合约创建
    pub value: u128,             // 转账金额
    pub nonce: u64,              // 交易序号
    pub gas_limit: u64,          // Gas 上限（最小 21000）
    pub max_fee_per_gas: u64,    // 最大 Gas 费用
    pub signature: Vec<u8>,       // ECDSA 签名
}
```

**交易 hash 计算**：

```
tx_hash = SHA256(from || to || value.to_le_bytes() || nonce.to_le_bytes()
                  || gas_limit.to_le_bytes() || max_fee_per_gas.to_le_bytes())
```

**签名校验**：验证 `signature` 非空，使用 `from` 从签名中恢复公钥并校验。

### 2.3 Header

```rust
pub struct Header {
    pub parent_hash: Hash,       // 父区块 hash
    pub state_root: Hash,        // 执行后状态 Merkle root
    pub tx_root: Hash,           // 交易列表 Merkle root
    pub receipt_root: Hash,      // 收据列表 Merkle root
    pub timestamp: u64,          // 区块时间戳（Unix 秒）
    pub number: u64,            // 区块高度
    pub gas_used: u64,          // 累计 Gas 消耗
    pub gas_limit: u64,         // Gas 上限
    pub nonce: u64,             // PoA 随机数
    pub proposer: Address,      // 区块提议者地址
    pub extra: Vec<u8>,         // 额外数据（最大 10000 字节）
}
```

**Header hash 计算**：

```
header_hash = SHA256(bincode_serialize(header))
```

### 2.4 Block

```rust
pub struct Block {
    pub header: Header,
    pub transactions: Vec<SignedTransaction>,
}
```

### 2.5 Receipt

```rust
pub struct Receipt {
    pub transaction_hash: Hash,
    pub success: bool,
    pub gas_used: u64,
    pub logs: Vec<Log>,
}
```

**Receipt hash 计算**：

```
receipt_hash = SHA256(bincode_serialize(receipt))
```

## 3. Merkle Root 计算

### 3.1 交易列表 (tx_root)

```
tx_hashes = [transaction_hash(tx) for tx in transactions]
tx_root = merkle_root(tx_hashes)
```

若交易列表为空，`tx_root = ZERO_HASH`（32 字节零值）。

### 3.2 收据列表 (receipt_root)

```
receipt_hashes = [receipt_hash(r) for r in receipts]
receipt_root = merkle_root(receipt_hashes)
```

若收据列表为空，`receipt_root = ZERO_HASH`。

### 3.3 状态树 (state_root)

使用二叉 Merkle tree，key 为 `address`，value 为 `balance || nonce` 的串联字节序。

```
state_entries = sorted([(addr, account.serialize()) for addr, account in state])
state_hashes = [SHA256(entry) for entry in state_entries]
state_root = merkle_root(state_hashes)
```

**Merkle root 算法**：

```
if len(hashes) == 0: return ZERO_HASH
if len(hashes) == 1: return hashes[0]
while len(hashes) > 1:
    if len(hashes) % 2 != 0:
        hashes.append(hashes[-1])  // 奇数时复制最后一个
    next_level = []
    for pair in chunks(hashes, 2):
        combined = pair[0] || pair[1]
        next_level.append(SHA256(combined))
    hashes = next_level
return hashes[0]
```

## 4. 状态转换

### 4.1 账户模型

每个账户包含：

- `balance: u128` — 账户余额（单位：wei）
- `nonce: u64` — 已执行交易计数

新账户初始化：`balance = 0, nonce = 0`。

### 4.2 交易执行

**前置校验**（按顺序）：

1. `tx.nonce >= state.nonce`（非 stale）
2. `tx.gas_limit >= 21000`
3. `tx.max_fee_per_gas > 0`
4. `verify_signature(tx.from, tx.signature, tx_hash) == true`
5. 若存在同一 sender + nonce 的待处理交易，新交易的 `max_fee_per_gas` 必须更高
6. `state.balance(tx.from) >= tx.value + tx.gas_limit * tx.max_fee_per_gas`

**状态更新**：

```
state.decrease_balance(tx.from, tx.value)
if tx.to is Some(addr):
    state.increase_balance(addr, tx.value)
state.increase_nonce(tx.from)
```

**Gas 计算**：

```
gas_used = tx.gas_limit  // 第一版固定消耗
```

### 4.3 区块执行（原子性）

1. Clone working state
2. 按顺序执行每笔交易，更新 working state
3. 若任何交易执行失败，整个区块 revert
4. 执行成功后，将 working state commit 到 canonical state
5. 计算 `state_root` 并与 `header.state_root` 对比验证

## 5. 区块验证

### 5.1 Header 校验

| 字段 | 校验规则 |
|------|---------|
| `timestamp` | `> 0` |
| `gas_limit` | `> 0` |
| `gas_used` | `<= gas_limit` |
| `extra` | `len <= 10000` |
| `parent_hash` | 必须匹配父区块 hash |

### 5.2 Block 校验

1. 校验 `header`（见上表）
2. 校验 `tx_root == merkle_root([tx_hash for tx in transactions])`
3. 校验每笔交易签名
4. 重新执行区块，校验 `outcome.state_root == header.state_root`
5. 校验 `outcome.receipt_root == header.receipt_root`
6. 校验 `outcome.gas_used == header.gas_used`

## 6. Mempool 策略

### 6.1 交易验证

进入 mempool 的交易必须通过：

- `tx.nonce >= state_nonce`
- `tx.gas_limit >= 21000`
- `tx.max_fee_per_gas > 0`
- 签名校验
- 非重复交易

### 6.2 打包顺序

按以下优先级选择：

1. 按 sender 分组，同一 sender 按 nonce 升序选择（保证连续性）
2. 同 nonce 的多笔交易，选 `max_fee_per_gas` 最高的一笔
3. 不同 sender 之间，按 `max_fee_per_gas` 降序选择

### 6.3 替换规则

若存在相同 sender + nonce 的待处理交易，只有新交易的 `max_fee_per_gas` 严格更高时才替换。

### 6.4 驱逐策略

当 mempool 超过容量上限时，优先驱逐：

- `max_fee_per_gas` 最低的交易
- 同 fee 时驱逐 nonce 最大的

## 7. P2P 消息协议

### 7.1 消息类型

| 消息 | 方向 | 说明 |
|------|------|------|
| `Status` | 双向 | 交换 chain_id、best_height、best_hash |
| `GetHeaders` | 请求 | 从指定 height 起请求最多 200 条 header |
| `Headers` | 响应 | 返回 header 列表 |
| `GetBlocks` | 请求 | 按 hash 请求指定 header 对应的 block body |
| `Blocks` | 响应 | 返回 block 列表 |
| `NewTransaction` | 广播 | 广播新交易 |
| `NewBlock` | 广播 | 广播新区块 |

### 7.2 Headers-First 同步

```
1. 交换 Status，获取对方 best_height
2. 发送 GetHeaders，从 local_height + 1 起
3. 收到 Headers，验证每条 header（parent link、gas、timestamp 递增）
4. 对连续的 header 链发送 GetBlocks
5. 收到 Blocks，验证并执行每条区块
6. 更新 canonical chain
```

### 7.3 Peer Scoring

| 行为 | 分数变化 |
|------|---------|
| 返回有效 Headers | +1 |
| 返回有效 Blocks | +2 |
| 返回无效 header | -10 |
| 返回无效 block | -20 |
| 请求超时 | -5 |
| 分数低于 -50 | 断开连接 |

## 8. 分叉选择规则

采用最长链规则 + deterministic tiebreak：

1. 选择累计区块高度最大的链
2. 高度相同则选择 `header.state_root` 字典序更小的（deterministic tiebreak）

**链重组（Reorg）**：

1. 找到分叉点（首个不一致的区块）
2. 从 canonical chain 回滚分叉点之后的区块状态
3. 从新链导入分叉点之后的区块并重新执行
4. 更新 canonical chain 和 block_hashes 映射

## 9. 存储布局

### 9.1 Storage Trait 接口

```rust
pub trait Storage: Send + Sync {
    fn put_block(&mut self, number: u64, block: &Block) -> Result<()>;
    fn get_block(&self, number: u64) -> Result<Option<Block>>;
    fn put_account_state(&mut self, addr: &Address, account: &Account) -> Result<()>;
    fn get_account_state(&self, addr: &Address) -> Result<Option<Account>>;
    fn put_tx_index(&mut self, tx_hash: &Hash, location: (u64, usize)) -> Result<()>;
    fn get_tx_index(&self, tx_hash: &Hash) -> Result<Option<(u64, usize)>>;
    fn put_receipt(&mut self, tx_hash: &Hash, receipt: &Receipt) -> Result<()>;
    fn get_receipt(&self, tx_hash: &Hash) -> Result<Option<Receipt>>;
    fn put_block_hash(&mut self, number: u64, hash: &Hash) -> Result<()>;
    fn get_block_hash(&self, number: u64) -> Result<Option<Hash>>;
    fn update_canonical_chain(&mut self, blocks: &[(u64, Hash)]) -> Result<()>;
}
```

### 9.2 区块存储语义

- `put_block` 按 `block_hash` 存储区块体，支持同一高度出现不同 hash（分叉场景）
- `get_block(number)` 通过 `canonical_chain[number]` 查找 canonical hash，再获取区块体
- `update_canonical_chain` 更新 `canonical_chain` 和 `block_hashes` 映射