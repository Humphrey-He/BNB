# BNB 项目面试讲解稿

## 1. 使用方式

这份文档给面试时直接口述使用，分为两部分：

- 自我介绍版：适合 1 到 3 分钟开场
- 项目深挖版：适合面试官继续追问项目细节时展开

建议表达原则：

- 先讲项目定位，再讲核心闭环
- 少讲“我写了多少代码”，多讲“我解决了什么可靠性问题”
- 少讲零散模块，多讲资金链路和工程取舍

---

## 2. 自我介绍版

### 2.1 60 秒版本

大家好，我主要做的是 Go 和 Web3 后端方向。这个阶段我重点做了一个多链资产后端平台项目，核心目标不是单纯监听链上事件，而是把充值、确认数、账本入账、提现申请、审批、广播、回执确认这些资金链路做成一个完整闭环。

这个项目第一版聚焦 EVM 链，技术上用的是 Go、PostgreSQL、NATS 和 go-ethereum。充值侧我做了 `scanner`、`parser`、`confirm-worker`、`ledger-service` 这条链路，把链上 ERC-20 事件可靠地同步到链下，并沉淀成可审计余额。提现侧我做了从创建申请、冻结余额、风险检查、审批、签名广播到 receipt 确认的状态机，并重点处理了幂等、nonce 分配、RPC failover、失败补偿这些生产里真正容易出问题的点。

这个项目对我帮助最大的是，我不再只是把 Web3 当成“调 RPC”或者“扫链”，而是开始站在资金系统可靠性的角度去设计后端。

### 2.2 90 秒版本

大家好，我目前主要专注在 Go 后端和 Web3 资产系统方向。最近我在做的核心项目，是一个多链资产后端平台，面向钱包、交易所或者 Custody 这一类场景。这个项目第一版聚焦 EVM 链，覆盖了充值监听、确认数处理、链重组补偿、账本余额、提现申请与广播、链路健康监控这些核心能力。

我比较想强调的是，这个项目不是简单的 indexer，也不是只做一个提现接口，而是围绕“资金闭环”来设计的。充值链路上，我做了 `scanner -> parser -> confirm-worker -> ledger-service`，把链上 Transfer 事件经过确认数推进后，可靠地入账到链下系统。提现链路上，我做了 `CreateWithdrawal -> risk_checking -> approved -> signing -> broadcasting -> broadcasted -> confirmed/failed` 的状态机，把余额冻结、风险审批、nonce 管理、交易广播和失败补偿串起来。

这里面我比较有代表性的工作有几个：第一，做了账本式余额模型，`ledger_entries` 是事实来源，`balances` 是快照，所以系统可回放、可审计。第二，做了 RPC provider failover、重试和熔断，避免单节点不稳定把核心业务拖挂。第三，最近我把提现审批后的事件路径做了统一，避免 API 直接绕过 worker 推广播，同时修了 nonce allocation 的脏记录问题，并把 `GetChainStatus` 改成了真实链路健康数据。

所以如果从能力上总结，我这个项目比较能体现的是：我不仅会写 Go 服务，也能从资金系统的视角去思考幂等性、一致性、补偿机制和可运维性。

### 2.3 可以突出给面试官的关键词

- 多链资产后端
- 充值和提现完整闭环
- 账本式余额
- 幂等和事务一致性
- nonce 管理和广播重试
- RPC failover 和链路健康
- 面向生产问题设计，不只是 happy path

---

## 3. 项目深挖版

### 3.1 项目定位

如果面试官问“这个项目本质是什么”，可以这样回答：

这个项目本质上是一个多链资产后端平台，面向交易所、钱包、Custody 或 Crypto Payment 场景。第一版先聚焦 EVM 生态，目标不是做一个浏览器式的扫链工具，而是把链上资金流可靠地接入链下业务系统，实现充值入账、提现出账、余额审计和链路容错。

### 3.2 为什么这个项目有含金量

可以这样讲：

普通的 Web3 项目很多只做到“拉到链上数据”或者“能发一笔交易”，但资金系统真正难的点在于：

- 重复消息不能重复入账
- 链重组后不能把错误余额留在系统里
- 提现不能因为重试而重复广播
- RPC 节点不稳定时业务不能直接失效
- 用户余额不能靠一张简单的 balance 表硬改

我的项目重点就在这些点上做了工程化设计。

### 3.3 充值核心链路

充值链路可以按下面的结构展开：

#### 第一步：scanner 扫链

`scanner` 负责按链和区块范围去拉取 ERC-20 `Transfer` 日志，同时维护 `scan_checkpoints`，记录每条链扫描到哪里。这里的重点不是“会调 `eth_getLogs`”，而是：

- 扫描窗口要可配置，避免大批量日志导致 RPC 超时
- 要记录区块 header，为后续 reorg 检测准备数据
- 要把扫描进度持久化，服务重启后可以续跑

对应入口在 [scanner main](E:/awesomeProject/BNB/web3-blockchain-backend-engineer/cmd/scanner/main.go) 和 [scanner](E:/awesomeProject/BNB/web3-blockchain-backend-engineer/internal/scanner/scanner.go)。

#### 第二步：parser 解析事件

`parser` 消费原始链上日志，把 ERC-20 `Transfer` 标准化，并识别是不是平台关心的充值事件。这里最关键的是幂等：

- 用 `chain_id + tx_hash + log_index` 作为天然唯一键
- 即使消息重复投递，也不会重复落库和重复入账

对应实现可参考 [parser main](E:/awesomeProject/BNB/web3-blockchain-backend-engineer/cmd/parser/main.go) 和 [parser](E:/awesomeProject/BNB/web3-blockchain-backend-engineer/internal/parser/parser.go)。

#### 第三步：confirm-worker 推确认数

链上事件不是扫到就能算充值成功，还要经过确认数。`confirm-worker` 会根据最新安全区块高度推进状态，比如从 `detected` 到 `pending_confirmation` 再到 `confirmed`。这一步体现的是对链 finality 的理解，而不是简单 CRUD。

对应入口在 [confirm-worker main](E:/awesomeProject/BNB/web3-blockchain-backend-engineer/cmd/confirm-worker/main.go)。

#### 第四步：ledger-service 入账

当充值确认完成后，`ledger-service` 才把金额写入账本和余额快照。这里我会重点讲：

- `ledger_entries` 是唯一事实来源
- `balances` 只是当前快照
- 所有余额变化都来源于账本流水
- 这样系统支持审计、回放和补偿

这也是这个项目和普通“加减余额表”项目最大的差异之一。入口在 [ledger-service main](E:/awesomeProject/BNB/web3-blockchain-backend-engineer/cmd/ledger-service/main.go)。

### 3.4 提现核心链路

提现链路更适合展示工程能力，因为它包含更多状态管理和异常场景。

#### 第一步：CreateWithdrawal

用户创建提现申请时，系统不会直接广播，而是先做严格校验并冻结余额。我最近补强了这一层校验，主要包括：

- `from_address` 和 `to_address` 必须是合法 EVM 地址
- `from` 和 `to` 不能相同
- `amount` 必须是正整数
- `chain_id` 必须存在且激活
- `token_id` 必须存在、激活，并且属于对应链

对应代码在 [CreateWithdrawal](E:/awesomeProject/BNB/web3-blockchain-backend-engineer/internal/api/handlers/handlers.go:349)。

#### 第二步：withdrawal-worker 做风险和审批流转

提现进入 `created` 后，不是直接发交易，而是先进入 `risk_checking`，再进入 `approved` 或 `manual_review`。这里的价值在于把“资金申请”和“链上执行”解耦，使系统支持人工审核、风控规则和后续审批扩展。

现在审批后的路径我已经统一成：

`approved -> withdrawal_approved event -> signing -> withdrawal_broadcast`

这样做的好处是，审批动作和广播动作之间有清晰的事件边界，不会由 API 直接跨层推动状态。

对应实现见 [withdrawal worker](E:/awesomeProject/BNB/web3-blockchain-backend-engineer/internal/withdrawalservice/worker.go:110)。

#### 第三步：broadcaster 负责真实交易广播

`broadcaster` 做的是最接近链的一层，包括：

- 构造原生币或 ERC-20 转账交易
- 获取 gas price 和估算 gas limit
- 分配 nonce
- 调 signer 进行签名
- 广播 raw transaction
- 轮询 receipt 并推进提现状态

这里我最近重点修了 nonce 逻辑。原来的问题是：本地 nonce allocation 和链上 pending nonce 可能不一致，导致系统拿到的本地 nonce 比链上 nonce 小，最终留下脏记录。现在修复后的策略是：

- 先读链上 pending nonce
- 本地做 `AllocateAtLeast`
- 广播成功后 `MarkUsed`
- 广播失败则 `Release`
- 过期记录不再参与后续最大 nonce 计算

这点非常适合面试说，因为它很能体现你对真实 EVM 广播问题的理解。对应代码在 [broadcaster](E:/awesomeProject/BNB/web3-blockchain-backend-engineer/internal/broadcaster/broadcaster.go:99) 和 [nonce adapter](E:/awesomeProject/BNB/web3-blockchain-backend-engineer/internal/broadcaster/nonce_adapter.go:27)。

### 3.5 RPC 和可运维能力

这部分是你和只会写业务接口的候选人拉开差距的地方。

项目里我做了 RPC provider 管理能力：

- 支持多 provider
- 支持 failover
- 支持超时和重试
- 支持熔断
- 支持 provider health inspection

我最近还把 `GetChainStatus` 从“静态扫描高度接口”改成了真实链路健康接口，现在返回的不只是 `last_scanned_block`，还包括：

- `latest_block`
- `scan_lag`
- `rpc_healthy`
- `rpc_provider`
- `rpc_error`
- `provider_count`

这样运维层面能更快判断问题到底出在扫描、数据库还是 RPC 出口。

对应代码在 [GetChainStatus](E:/awesomeProject/BNB/web3-blockchain-backend-engineer/internal/api/handlers/handlers.go:753)。

### 3.6 你可以主动强调的设计取舍

如果面试官追问“为什么这么设计”，建议突出这几个思路：

- 资金系统优先考虑正确性，再考虑吞吐
- 余额不能直接靠业务表改，要靠账本
- 提现不能同步一把梭，要有状态机和补偿点
- RPC 访问不能依赖单节点，否则非常脆弱
- 业务状态推进要尽量通过事件解耦，减少强耦合调用

### 3.7 如果面试官问“做到哪里了”

建议诚实但有力度地回答：

现在充值链路已经具备比较完整的工程骨架，提现链路也已经具备从申请、审批到广播、receipt 跟踪的主流程。最近完成了审批后事件路径统一、nonce allocation 脏记录修复、CreateWithdrawal 严格校验和真实链路健康检查。合约工程依赖也恢复到了可编译状态。后续如果继续增强，我会优先补人工审核接口、失败重试可视化、真实测试链闭环验证和部署自动化。

这样的回答比“全做完了”更可信，也更像一个真正理解工程演进的人。

---

## 4. 一句话总结

如果最后要一句话总结这个项目，可以说：

这是一个围绕“资金安全和链路可靠性”设计的多链资产后端项目，我做的不只是扫链和发交易，而是把充值入账、提现出账、账本一致性、RPC 容灾和可运维性串成了一个完整的工程闭环。
