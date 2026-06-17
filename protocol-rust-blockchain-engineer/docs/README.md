# Protocol Engineer / Rust Blockchain Engineer Project

目标：建立一个能证明 Rust 系统编程、协议设计、节点执行、P2P 同步、mempool、状态存储和性能分析能力的技术壁垒项目。

推荐项目名：

`verifiable-rust-chain-node`

一句话定位：

> 一个用 Rust 实现的最小可验证区块链节点，覆盖签名交易、mempool、区块验证、状态机、Merkle root、headers-first 同步、RPC 查询、存储和 benchmark。

为什么从 `mini node` 升级成 `verifiable node`：

- 市场上的 Rust Protocol / Node 岗位更看重“节点如何验证数据”，而不只是能跑一个玩具链。
- 面试高频追问点是交易有效性、状态确定性、区块导入、同步协议、分叉选择、mempool 策略和性能瓶颈。
- 这个项目不冒充完整公链，但要把每个协议边界讲清楚。

建议技术栈：

- Rust
- Tokio
- libp2p
- RocksDB
- serde / bincode
- axum 或 tonic
- tracing
- Criterion
- proptest

核心产物：

- 签名交易和账户状态
- block/header/receipt 数据结构
- mempool nonce ordering 和 fee priority
- 状态执行和 state root / tx root
- 区块验证 pipeline
- headers-first P2P 同步
- fork choice 和 reorg 基础策略
- RPC API
- 性能 benchmark 和 profiling 记录
- `PROTOCOL.md` 协议设计文档

详细分析见 [ANALYSIS.md](ANALYSIS.md)。
