# verifiable-rust-chain-node - 项目规划

## 1. 项目概述

**项目名称**: verifiable-rust-chain-node
**项目类型**: Rust 最小可验证区块链节点
**目标**: 用 Rust 实现最小可验证区块链节点，覆盖签名交易、mempool、区块验证、状态机、Merkle root、P2P 同步、RPC 查询、持久化和性能测试

## 2. 核心功能范围

| 功能模块 | 描述 | 优先级 |
|---------|------|--------|
| 类型定义 | SignedTransaction、Block、Header、Receipt、Account | P0 |
| Crypto | hash、签名校验、地址派生 | P0 |
| 状态机 | 账户余额、nonce、状态转移、state root | P0 |
| Mempool | 交易校验、去重、nonce ordering、fee priority | P0 |
| 交易执行器 | 执行交易并生成 receipt | P0 |
| 区块验证 | header、parent、tx root、state root 验证 | P0 |
| 持久化存储 | 区块、状态、交易索引持久化 (RocksDB) | P0 |
| RPC API | 查询区块、交易、账户状态，提交交易 | P0 |
| 单节点出块 | PoA 共识，单节点定时出块 | P0 |
| P2P 网络 | headers-first sync、block request、peer scoring | P1 |
| 分叉处理 | 分叉选择规则、链重组 | P1 |
| 性能 Benchmark | Criterion 基准测试 | P2 |
| 指标监控 | 出块耗时、交易执行耗时、mempool 大小 | P2 |

## 3. 里程碑计划

### Week 1: 基础搭建

- [x] 初始化 Rust 项目，配置 Cargo.toml 依赖
- [x] 实现核心类型：`SignedTransaction`、`Block`、`Header`、`Receipt`、`Account`
- [x] 实现基础序列化 (serde)
- [x] 实现 hash 和交易 hash
- [x] 搭建项目目录结构
- [x] 配置 tracing 日志

**交付物**: 基础项目骨架 + 核心类型定义 ✅

### Week 2: 签名、状态机与执行器

- [x] 实现 `state` 模块：账户余额、nonce 存储
- [x] 实现 state root 计算
- [x] 实现状态转移逻辑
- [x] 实现 `executor` 模块：交易执行并生成 receipt
- [x] 实现 tx root 和 receipt root

**交付物**: 状态机 + 执行器源码 ✅

### Week 3: Mempool、验证与持久化

- [x] 实现 `storage` 模块：Storage trait + StorageMem 内存实现
- [x] 实现区块、状态、交易索引持久化
- [x] 实现基础 mempool：签名校验、去重、nonce ordering、fee priority
- [x] 实现交易替换和 mempool eviction
- [x] 实现 `validation` 模块：header、parent、root 验证

**交付物**: Mempool + Validation + Storage ✅

### Week 4: RPC 与单节点运行

- [ ] 实现 `rpc` 模块：axum 或 tonic
- [ ] RPC 接口：get_balance、get_block、send_transaction、get_transaction
- [ ] 实现单节点出块 (PoA)
- [ ] 实现交易池与出块流程集成
- [ ] 端到端测试：提交交易 → 打包 → 执行 → 查询
- [ ] 编写单元测试和集成测试

**交付物**: 可运行的单节点区块链

### Week 5-6: P2P 同步

- [ ] 实现 `p2p` 模块：libp2p
- [ ] 实现 Status、GetHeaders、Headers、GetBlocks、Blocks 消息
- [ ] 实现 headers-first sync
- [ ] 实现 block request 和区块导入
- [ ] 实现 peer scoring
- [ ] 启动两个节点并验证同步

**交付物**: 双节点同步运行

### Week 7-8: 分叉处理与性能

- [ ] 实现分叉选择规则
- [ ] 实现链重组逻辑
- [ ] 编写 Criterion benchmark
  - [ ] tx execution throughput
  - [ ] block import latency
  - [ ] mempool insert/select latency
  - [ ] storage read/write latency
- [ ] Profiling 和性能优化
- [ ] 编写 `PROTOCOL.md` 协议设计文档

**交付物**: 完整协议文档 + Benchmark 报告

## 4. 技术栈

| 类别 | 技术 |
|------|------|
| 语言 | Rust |
| 异步运行时 | Tokio |
| 网络 | libp2p |
| 存储 | RocksDB |
| 序列化 | serde / bincode |
| RPC 框架 | axum / tonic |
| 性能测试 | Criterion |
| 日志追踪 | tracing |
| 属性测试 | proptest |

## 5. 交付物清单

| 交付物 | 文件路径 | 状态 |
|--------|----------|------|
| 项目源码 | `src/` | DONE |
| 核心类型 | `src/types.rs` | DONE |
| Crypto 模块 (hash/merkle) | `src/crypto.rs` | DONE (sig pending) |
| 错误类型 | `src/error.rs` | DONE |
| 性能测试 | `benches/bench.rs` | DONE |
| 状态机 | `src/state.rs` | DONE |
| 执行器 | `src/executor.rs` | DONE |
| Mempool | `src/mempool.rs` | DONE |
| 存储层 | `src/storage.rs` | DONE |
| 验证模块 | `src/validation.rs` | DONE |
| P2P 模块 | `src/p2p.rs` | TODO |
| RPC 服务 | `src/rpc.rs` | TODO |
| 共识模块 | `src/consensus.rs` | TODO |
| 协议文档 | `docs/PROTOCOL.md` | TODO |
| 架构文档 | `docs/ARCHITECTURE.md` | DONE |
| 项目规划 | `docs/PROJECT_PLAN.md` | DONE |

## 6. 第一版功能边界 (MVP)

- 单机启动一个节点
- 通过 RPC 提交签名转账交易
- 节点打包交易并生成区块
- 导入区块时重新执行交易并验证 root
- 状态持久化
- 查询账户余额、区块高度、交易详情

## 7. 第二版扩展

- 启动两个节点并同步区块
- 增加分叉选择规则
- 增加 benchmark
- 增加 tracing 日志
- 增加 peer scoring
- 增加 reorg 集成测试

## 8. 质量标准

- [ ] 核心模块有完整单元测试
- [ ] 集成测试覆盖主要流程
- [ ] Criterion benchmark 报告生成
- [ ] 代码通过 `cargo clippy` 检查
- [ ] 协议设计文档完整
- [ ] property test 覆盖状态执行关键不变量
- [ ] 双节点同步测试通过
- [ ] 分叉选择测试通过
