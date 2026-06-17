# Protocol Engineer / Rust Blockchain Engineer 分析文档

## 结论

这是三个方向里技术壁垒最高、长期价值最好的方向之一。它比业务后端更难，但和你的 Rust、Go、后端系统经验有明显连接点。你不需要实现完整公链，也不要把项目包装成“自研公链”；更好的做法是做一个最小可验证节点，证明你理解协议层核心问题。

推荐定位：

> Rust Blockchain Engineer / Protocol Backend Engineer / Node Infrastructure Engineer

## 匹配度

匹配度：高，但学习曲线陡。

优势：

- Rust 能直接匹配协议、节点、Solana、Substrate、reth、Lighthouse、Nearcore 等方向。
- 后端经验可以迁移到网络服务、并发、存储、监控、RPC、稳定性。
- Go 背景有助于理解 geth、Cosmos SDK、CometBFT 等生态。

短板：

- 需要补区块链底层：共识、P2P、状态机、交易池、存储结构。
- 面试会更强调系统设计、性能、并发、安全边界。
- 很多岗位期望你读过或改过真实节点代码。

## 推荐作品集项目

项目名：`verifiable-rust-chain-node`

项目目标：

用 Rust 做一个最小可运行、可验证、可讲清楚边界的区块链节点。它不追求复杂共识或 EVM 兼容，而是聚焦协议工程能力：签名校验、mempool、状态执行、区块验证、持久化、headers-first 同步、fork choice 和性能观测。

核心模块：

- `types`：SignedTransaction、Block、Header、Receipt、Account。
- `crypto`：哈希、签名校验、地址派生。
- `state`：账户余额、nonce、状态转移、state root。
- `mempool`：交易校验、去重、nonce ordering、fee priority、eviction。
- `executor`：执行交易并生成 receipt。
- `validation`：区块头、交易、状态根、父区块、gas limit 验证。
- `storage`：区块、状态、交易索引、receipt、canonical chain 持久化。
- `p2p`：peer 管理、headers-first sync、block request、peer scoring。
- `rpc`：查询区块、交易、账户状态，提交交易。
- `consensus`：PoA 出块和 fork choice，不做复杂 BFT。
- `metrics`：出块耗时、交易执行耗时、mempool 大小。

第一版功能边界：

- 单机启动一个节点。
- 通过 RPC 提交转账交易。
- 交易包含签名、nonce、fee。
- 节点打包交易并生成区块。
- 区块包含 tx root、receipt root、state root。
- 节点导入区块时重新执行并验证 root。
- 状态持久化。
- 查询账户余额、区块高度、交易详情。

第二版扩展：

- 启动两个节点并同步区块。
- 增加分叉选择规则。
- 增加交易签名校验。
- 增加 benchmark。
- 增加 tracing 日志。
- 增加 peer score 和同步错误处理。

## 技术栈建议

主栈：

- Rust
- Tokio
- serde
- axum 或 tonic
- RocksDB
- tracing
- Criterion

可选：

- libp2p：更贴近真实协议岗位。
- prost/tonic：如果想展示 gRPC。
- Prometheus exporter：展示基础设施意识。

## 学习补齐路线

第 1-2 周：

- Rust async、Tokio、channel、task、错误处理。
- 区块链核心结构：block、signed transaction、state、receipt、mempool、Merkle root。
- 读 geth 或 reth 的高层架构文档。

第 3-4 周：

- 实现单节点状态机。
- 实现 RPC。
- 实现持久化。
- 写单元测试和集成测试。

第 5-8 周：

- 加 P2P 同步。
- 加分叉处理。
- 加 benchmark 和 profiling。
- 写协议设计文档。

## 简历表达

英文：

> Built a verifiable Rust blockchain node with signed transactions, nonce-aware mempool, block production, deterministic state execution, Merkle roots, persistent storage, RPC APIs, headers-first peer synchronization, fork choice, and Criterion benchmarks.

中文：

> 使用 Rust 实现最小可验证区块链节点，覆盖签名交易、nonce-aware mempool、出块、确定性状态执行、Merkle root、持久化存储、RPC 查询、headers-first P2P 同步、分叉选择和 Criterion 性能测试。

## 求职策略

优先投：

- Rust Blockchain Engineer
- Protocol Engineer
- Node Engineer
- Blockchain Infrastructure Engineer
- Core Blockchain Engineer
- Solana/Reth/Substrate Engineer

岗位关键词：

- Rust
- P2P
- Consensus
- Mempool
- State machine
- Node
- Validator
- RPC
- Performance
- Distributed systems

## 面试准备重点

你需要能讲清：

- 交易从提交到上链的生命周期。
- mempool 如何校验、去重、排序。
- nonce gap、低 fee 交易、mempool eviction 怎么处理。
- 节点如何处理分叉和链重组。
- 状态机执行如何保证确定性。
- 区块导入时为什么要重新执行交易并验证 state root。
- 为什么区块链节点需要持久化和索引。
- Rust async 在节点网络层中的优势和陷阱。
- 如何做性能 profiling 和 backpressure。

## 当前市场判断

公开岗位中 Rust + blockchain 的组合仍集中在节点、协议、平台基础设施、DeFi 执行系统、MEV、链上数据处理等方向。部分岗位明确要求 Rust 生产经验，也接受有 Go/C++ 背景并愿意深入 Rust 的工程师。

参考：

- [Web3Jobs Rust Backend Engineer, Protocol](https://web3jobs.io/jobs/126766398-rust-backend-engineer-protocol)
- [Goremotejob Rust Blockchain Platform role](https://goremotejob.com/remote-jobs/software-engineer-rust-blockchain-platform)
- [ZipRecruiter Protocol Engineer Rust Jobs](https://www.ziprecruiter.com/Jobs/Protocol-Engineer-Rust)
