# Web3 Backend / Blockchain Backend 分析文档

## 结论

这是你当前最适合、最快能转进区块链业务的方向，也应该作为三个项目里的主线项目。它和传统后端重合度最高，同时能体现 Web3 核心能力：节点 RPC、链上事件、交易确认、链重组、钱包资产、提现广播、幂等、补偿、账本余额和监控。

推荐定位：

> Go / Rust / Python Backend Engineer transitioning into Web3 backend, blockchain infrastructure, on-chain data indexing, wallet and transaction systems.

中文定位：

> 具备 Go/Rust/Python 后端经验，转向 Web3 后端、区块链基础设施、链上数据索引、钱包和交易系统方向。

## 匹配度

匹配度：很高。

优势：

- 3 年后端经验能直接迁移到 API、数据库、缓存、队列、监控、任务系统。
- Go 是 Web3 后端、钱包、交易所、节点工具里非常常见的语言。
- Python 可用于数据脚本、风控任务、ETL、链上分析。
- Rust 可作为高性能组件、交易广播器、索引器或后续协议方向加分项。

短板：

- 需要补链上交易生命周期。
- 需要理解 RPC、ABI、事件日志、确认数、nonce、gas。
- 需要处理链上数据与业务数据库之间的一致性问题。
- 需要能解释提现广播、nonce 冲突、gas replacement、RPC 节点不稳定时的处理策略。

## 推荐作品集项目

项目名：`multi-chain-asset-platform`

项目目标：

做一个贴近真实钱包、交易所、Custody、Crypto Payment 平台的资产后端系统。它不只是 indexer，而是要展示你能把链上数据可靠同步到链下账本，并完成充值、提现、余额、交易历史和监控闭环。

核心功能：

- 监听 Ethereum 或 BSC 的新区块。
- 解析 ERC-20 `Transfer` 事件。
- 识别平台地址充值。
- 根据确认数把交易从 `pending` 变为 `confirmed`。
- 支持链重组检测和回滚补偿。
- 使用账本式余额模型同步用户资产，而不是直接覆盖余额。
- 支持提现申请、状态流转、交易构造、广播和确认。
- 管理 nonce、gas 估算、交易重发和 replacement transaction。
- 支持 RPC provider failover、扫描窗口、限流和指数退避。
- 提供 REST API 查询余额和交易历史。
- 使用 Redis/Kafka/NATS 处理异步任务。
- 暴露 Prometheus 指标。
- 用 Docker Compose 一键启动 PostgreSQL、Redis、服务端。

建议服务拆分：

- `scanner`：扫描区块，拉取 logs。
- `parser`：解析事件，标准化交易。
- `confirm-worker`：处理确认数和状态流转。
- `ledger-service`：维护账本流水和可用/冻结余额。
- `withdrawal-service`：处理提现申请、审批状态、广播任务。
- `broadcaster`：构造并广播链上交易，管理 nonce/gas/retry。
- `api-server`：提供 REST API。
- `reorg-detector`：检测 parent hash 不匹配并触发补偿。
- `rpc-gateway`：管理多 RPC provider、限流、熔断和重试。
- `metrics`：监控区块延迟、处理速度、错误数。

## 数据模型建议

核心表：

- `chains`：链配置。
- `rpc_providers`：RPC 节点配置、权重、健康状态。
- `tokens`：token 合约、精度、symbol。
- `watched_addresses`：平台监听地址。
- `blocks`：已扫描区块和 hash。
- `chain_events`：标准化事件。
- `deposits`：充值记录。
- `withdrawals`：提现申请和链上交易状态。
- `ledger_entries`：账本流水。
- `balances`：用户可用、冻结、总余额快照。
- `nonce_allocations`：地址 nonce 分配记录。
- `scan_checkpoints`：扫描进度。

关键状态：

- `detected`
- `pending_confirmation`
- `confirmed`
- `orphaned`
- `failed`
- `broadcasting`
- `broadcasted`
- `replaced`

## 技术栈建议

主栈：

- Go
- Gin 或 chi
- go-ethereum
- PostgreSQL
- Redis
- NATS 或 Kafka
- Docker Compose
- Prometheus

Python 辅助：

- FastAPI 管理后台原型。
- web3.py 验证脚本。
- 数据修复和补偿脚本。

Rust 加分：

- 用 Rust 写高性能 log parser。
- 用 Rust 写独立 transaction broadcaster。
- 用 Rust 写链上数据校验 CLI。

## 学习补齐路线

第 1 周：

- Ethereum/BSC RPC。
- 区块、交易、receipt、log、topic。
- ERC-20 Transfer ABI。
- 确认数、链重组、nonce、gas。

第 2-3 周：

- Go 实现区块扫描和事件解析。
- PostgreSQL 建模。
- API 查询余额和交易历史。
- 实现幂等写入。
- 实现账本式余额模型。

第 4 周：

- 加确认数 worker。
- 加链重组处理。
- 加 Redis/Kafka/NATS 异步任务。
- 加 Prometheus 指标。

第 5 周：

- 实现提现申请和状态机。
- 实现交易广播器。
- 实现 nonce 分配、gas 估算和重发策略。

第 6 周：

- 补 Docker Compose。
- 补集成测试。
- 写架构文档。
- 写简历项目描述。

## 简历表达

英文：

> Built a Go-based multi-chain asset backend platform that scans EVM blocks, parses ERC-20 Transfer logs, detects deposits, handles confirmations and chain reorgs, maintains ledger-based balances, broadcasts withdrawals with nonce/gas management, and exposes REST APIs with PostgreSQL, Redis, NATS/Kafka, and Prometheus metrics.

中文：

> 使用 Go 构建多链资产后端平台，支持 EVM 区块扫描、ERC-20 Transfer 事件解析、充值识别、确认数处理、链重组补偿、账本式余额、提现广播、nonce/gas 管理、REST API、异步任务和 Prometheus 监控。

## 求职策略

优先投：

- Web3 Backend Engineer
- Blockchain Backend Engineer
- Wallet Backend Engineer
- Custody Backend Engineer
- Blockchain Infrastructure Engineer
- Indexer Engineer
- On-chain Data Engineer
- Crypto Payment Engineer

岗位关键词：

- Go
- Python
- Rust
- REST API
- gRPC
- PostgreSQL
- Redis
- Kafka
- RPC
- Indexer
- Wallet
- Transaction lifecycle
- Blockchain data
- Monitoring

## 面试准备重点

你需要能讲清：

- 如何监听 ERC-20 充值。
- 如何保证事件重复处理时幂等。
- 充值为什么需要确认数。
- 链重组发生时数据库如何回滚或修正。
- API 查询余额时如何避免链上和链下不一致。
- 交易广播失败如何重试。
- nonce 冲突如何处理。
- 钱包系统如何区分热钱包、冷钱包、多签、MPC。
- 为什么资产系统应该使用账本流水而不是直接改余额。
- RPC provider 不稳定、限流或返回不一致时如何容灾。
- 提现交易被卡住时如何使用 replacement transaction。

## 当前市场判断

2026 年公开招聘中，Web3 后端岗位仍然高频要求 Go/Rust/Python、节点交互、API、数据库、链上数据、实时系统和基础设施能力。钱包、Custody、交易所、支付和 DeFi 后端尤其关注交易生命周期、RPC/indexer、账本一致性和高可用。相比纯合约岗，这条路线和你的已有经验重合度最高，也最容易快速产出能讲清楚的项目。

参考：

- [Web3.career Backend Jobs, May 2026](https://web3.career/backend-jobs)
- [Web3.career Backend + Rust Jobs](https://web3.career/backend%2Brust-jobs)
- [Web3Vacancy Blockchain Backend Developer Jobs](https://web3vacancy.com/jobs/backend-developer)
- [Goremotejob Rust Blockchain Platform role](https://goremotejob.com/remote-jobs/software-engineer-rust-blockchain-platform)
