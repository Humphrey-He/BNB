# verifiable-rust-chain-node

一个用 Rust 实现的最小可验证区块链节点，覆盖签名交易、mempool、区块验证、状态机、Merkle root、P2P 同步、RPC 查询和持久化。

## 项目状态

| 模块 | 状态 |
|------|------|
| 核心类型 (types) | ✅ Done |
| Crypto (hash/merkle) | ✅ Done |
| 状态机 (state) | ✅ Done |
| 执行器 (executor) | ✅ Done |
| Mempool | ✅ Done |
| 存储层 (storage) | ✅ Done |
| 验证模块 (validation) | ✅ Done |
| P2P 模块 | 🔄 Week 5 |
| RPC 服务 | 🔄 Week 4 |
| 共识模块 | 🔄 Week 4 |
| 分叉处理 | 🔄 Week 7 |
| 性能 Benchmark | 🔄 Week 8 |

当前处于 **Week 4** 开发阶段：RPC 与单节点运行。

## 快速开始

### 环境要求

- Rust 1.75+
- Cargo (stable)

### 编译

```bash
# Debug 模式
cargo build

# Release 模式
cargo build --release
```

### 运行

```bash
cargo run
```

### 测试

```bash
# 单元测试
cargo test --lib

# 全部测试（含集成测试）
cargo test

# Clippy 检查
cargo clippy --all-targets
```

### 性能基准

```bash
cargo bench
# 报告位于 target/criterion/report/index.html
```

## 项目结构

```
src/
├── crypto.rs       # 密码学工具：hash、签名校验、Merkle root
├── types.rs        # 核心类型：Transaction、Block、Header、Receipt、Account
├── state.rs        # 状态机：账户余额、nonce、state root
├── executor.rs     # 交易执行器：执行交易并生成 receipt
├── mempool.rs      # 交易池：签名校验、去重、nonce ordering、fee priority
├── validation.rs   # 区块验证：header、parent、tx root、state root 校验
├── storage.rs      # 存储抽象：Storage trait + StorageMem 实现
└── lib.rs          # 库入口，导出所有公开接口

docs/
├── ARCHITECTURE.md  # 功能架构文档
├── PROJECT_PLAN.md  # 项目规划（Milestones）
├── PROTOCOL.md      # 协议设计规范
└── DEPLOYMENT.md    # 部署与使用指南
```

## 核心设计

### 确定性执行

同一区块在不同节点执行后必须得到完全相同的状态根（state_root）。这要求：
- 交易执行顺序确定性
- 状态读写顺序确定性
- 序列化格式确定性

### 区块验证原则

永远不要只信任对端给出的 state_root。导入区块时必须重新执行交易，执行结果 root 与 header 不一致则拒绝区块。

### Mempool 策略

- 按 `nonce` 升序、 `max_fee_per_gas` 降序组织交易
- 相同 nonce 的更高 fee 交易可替换旧交易
- 容量超限时驱逐低 fee 交易

### 分叉处理

采用最长链规则，高度相同时选择 `state_root` 字典序更小的链（deterministic tiebreak）。

## RPC API

启动节点后可通过 HTTP 调用：

| Method | 说明 |
|--------|------|
| `send_transaction` | 提交签名交易 |
| `get_transaction` | 查询交易详情 |
| `get_receipt` | 查询交易收据 |
| `get_block_by_number` | 按高度查询区块 |
| `get_block_by_hash` | 按 hash 查询区块 |
| `get_balance` | 查询账户余额 |
| `get_nonce` | 查询账户 nonce |
| `get_mempool_status` | 查询 mempool 状态 |

详情见 [DEPLOYMENT.md](docs/DEPLOYMENT.md)。

## 技术栈

| 类别 | 技术 |
|------|------|
| 语言 | Rust |
| 异步运行时 | Tokio |
| 网络 | libp2p |
| 存储 | RocksDB |
| 序列化 | serde / bincode |
| RPC 框架 | axum |
| 性能测试 | Criterion |
| 日志追踪 | tracing |
| 属性测试 | proptest |

## License

MIT