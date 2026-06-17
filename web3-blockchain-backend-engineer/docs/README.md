# Web3 Backend / Blockchain Backend Project

目标：做一个最贴近 2026 年 Web3 后端、钱包后端、交易所资产后端、链上数据工程岗位的主线作品集项目。

推荐项目名：

`multi-chain-asset-platform`

一句话定位：

> 一个以 Go 为主的多链资产后端平台，覆盖链上充值监听、确认数处理、链重组补偿、账本式余额、提现交易广播、nonce/gas 管理、交易历史 API、异步任务和监控告警。

为什么从 `indexer` 升级到 `asset-platform`：

- 单纯 indexer 能证明链上数据解析能力，但岗位里更常见的是“链上数据 + 资产状态 + 交易生命周期 + 钱包/支付/风控/API”组合能力。
- 钱包、Custody、交易所、支付、DeFi 后端都需要充值和提现闭环，而不是只监听 `Transfer`。
- 账本式余额、幂等、重试、补偿、RPC 容灾、监控 SLO 是后端工程师最容易和纯 Web3 初学者拉开差距的地方。

建议技术栈：

- Go：核心服务、API、scanner、worker、交易广播
- PostgreSQL：业务状态、账本、扫描进度、交易记录
- Redis：缓存、分布式锁、限流、nonce 窗口
- NATS 或 Kafka：异步任务、确认数、提现广播、补偿事件
- go-ethereum：EVM RPC、logs、receipt、交易构造
- Docker Compose：本地开发环境
- Prometheus / Grafana：服务指标和告警面板
- Python：数据修复、链上校验、运营脚本
- Rust：可选高性能 log parser 或交易广播 CLI 加分模块

核心产物：

- 多链配置和 RPC provider failover
- EVM 区块扫描器和 ERC-20 Transfer 事件解析
- 充值识别、确认数处理、链重组补偿
- 账本式余额系统，而不是简单覆盖余额
- 提现申请、审批状态、交易构造、签名占位、广播、确认
- nonce 管理、gas 估算、replacement transaction 策略
- REST API：余额、充值、提现、交易历史、链状态
- 异步任务队列、重试、死信队列
- Prometheus 指标和 Grafana dashboard
- 架构文档、数据一致性说明、面试讲解稿

详细分析见 [ANALYSIS.md](ANALYSIS.md)。
