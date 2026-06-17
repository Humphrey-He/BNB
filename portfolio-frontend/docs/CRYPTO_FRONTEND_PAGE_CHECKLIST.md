# 币圈网站前端主要页面与功能清单

本文档用于规划 `BNB Web3 Career Terminal` 后续页面扩展，也可作为 Web3 / 区块链前端产品设计的功能地图。币圈前端通常不是单纯营销页，而是由行情、资产、交易、钱包、链上数据、安全状态和运营活动共同组成的高信任产品界面。

## 1. 页面总览

| 页面模块 | 常见于 | 主要目标 | 作品集优先级 |
| --- | --- | --- | --- |
| 首页 / Landing | 交易所、钱包、DeFi、基础设施项目 | 建立信任、展示产品定位、引导连接钱包或注册 | 中 |
| Dashboard / Portfolio | 钱包、交易所、资产平台、DeFi | 展示资产、收益、风险、链上活动 | 高 |
| Market / 行情页 | 交易所、行情站、数据平台 | 展示币种价格、涨跌幅、成交量、趋势 | 中 |
| Trading / 交易页 | CEX、DEX、Perp、聚合器 | 下单、换币、查看订单簿和 K 线 | 高，但实现复杂 |
| Swap / Bridge | DEX、钱包、跨链桥 | 代币兑换、跨链转账、报价、滑点控制 | 高 |
| Wallet / 账户资产 | 钱包、交易所、托管系统 | 地址、余额、充值提现、交易历史 | 高 |
| Staking / Earn | DeFi、交易所、LST、收益产品 | 质押、赎回、收益展示、风险说明 | 中 |
| Launchpad / 活动页 | 交易所、项目方、IDO 平台 | 任务、白名单、申购、空投 | 中 |
| Explorer / 链上数据 | 区块浏览器、数据平台、基础设施 | 查询交易、地址、区块、合约、事件 | 高 |
| Risk / Security Center | 钱包、交易所、机构产品 | 风险提示、授权管理、地址风控、审计状态 | 高 |
| Docs / Developer | 协议、基础设施、API 平台 | 开发者文档、API Key、SDK、示例 | 中 |
| Admin / Ops Console | 交易所、钱包后端、链上服务 | 监控、扫描器、确认数、队列、告警 | 很高，适合后端作品集 |

## 2. 核心页面与功能

### 2.1 首页 / Landing

主要功能：
- 展示品牌、产品定位、目标用户和核心价值。
- 引导用户连接钱包、注册、进入 App 或查看文档。
- 展示安全背书，例如审计、资金储备、合作伙伴、链支持。
- 展示核心产品能力，例如 Swap、Earn、Portfolio、API、RPC、Indexer。

关键 UI：
- 顶部导航：Products、Markets、Earn、Docs、Connect Wallet。
- 第一屏：清晰产品名、主 CTA、真实产品预览。
- 链支持区域：Ethereum、BNB Chain、Solana、Arbitrum、Base 等。
- 安全信号：Audit、Bug bounty、Proof of Reserve、Status。

作品集建议：
- 你的项目不是营销站，首页不应太重。
- 可以保留为职业终端入口，但重点放在 Dashboard / Project Evidence。

### 2.2 Dashboard / Portfolio

主要功能：
- 展示总资产、可用余额、锁定余额、未确认充值、风险状态。
- 按链、代币、账户、项目过滤资产。
- 展示最近链上活动、交易状态、充值提现状态。
- 展示收益、风险暴露、未处理告警。

关键 UI：
- Asset cards：Total Balance、Pending Deposits、Confirmed Credits、Risk Alerts。
- Chain filter：All / Ethereum / BNB Chain / Solana / Polygon。
- Activity feed：Detected、Pending Confirmation、Confirmed、Failed、Reorged。
- 状态标签：Success、Pending、Failed、Reorg、Manual Review。

适合你的原因：
- 与 Web3 Backend 项目高度匹配。
- 能展示 scanner、parser、confirm worker、ledger service 的链路。

### 2.3 Market / 行情页

主要功能：
- 展示币种列表、价格、24h 涨跌幅、成交量、市值。
- 支持搜索、收藏、自选、排序。
- 展示热门板块：Layer1、Layer2、DeFi、Meme、AI、RWA。
- 展示简化 K 线或 sparkline。

关键 UI：
- Market table：Symbol、Price、24h Change、Volume、Market Cap。
- Watchlist：用户自选。
- Filters：Spot、Perp、DeFi、New Listing。
- 状态颜色：涨绿色/跌红色，注意地区习惯可配置。

作品集建议：
- 可作为展示页，但不应成为核心。
- 如果实现，应使用模拟数据或公开行情 API，并明确数据来源。

### 2.4 Trading / 交易页

主要功能：
- 展示 K 线、订单簿、成交记录、买卖盘。
- 支持限价、市价、止盈止损、杠杆、保证金模式。
- 展示当前仓位、订单、历史成交、风险率。
- 提供价格提醒、深度图、交易对切换。

关键 UI：
- 左侧交易对列表。
- 中央 K 线图。
- 右侧订单簿和最新成交。
- 下方订单/仓位表格。
- 下单面板：Buy / Sell、Limit / Market、Amount、Slider、Submit。

实现难点：
- 实时行情、WebSocket、撮合状态、精度处理。
- 风险极高，不适合作为早期 MVP 的完整实现。

作品集建议：
- 如果要做，只做 read-only trading terminal。
- 不模拟真实下单，避免产品边界混乱。

### 2.5 Swap / Bridge

主要功能：
- 选择 From token / To token，输入金额。
- 展示报价、价格影响、滑点、手续费、路由。
- 钱包连接、余额检查、Approve、Swap、交易跟踪。
- 跨链场景还要展示源链、目标链、预计到账时间和桥状态。

关键 UI：
- Token selector。
- Chain selector。
- Quote panel：Rate、Slippage、Price Impact、Gas、Route。
- Transaction state：Review、Approve、Pending、Confirmed、Failed。

安全重点：
- 高滑点警告。
- 代币授权额度提示。
- 假币和未知合约风险提示。
- 跨链桥延迟和失败补偿说明。

作品集建议：
- 适合作为交互模块。
- 可用于展示交易生命周期和风控意识。

### 2.6 Wallet / 资产账户页

主要功能：
- 展示地址、链、代币余额、法币估值。
- 支持充值、提现、转账、复制地址、二维码。
- 展示交易历史、充值确认数、提现状态。
- 支持地址白名单、多签、MPC、审批流状态。

关键 UI：
- Account overview。
- Token balance table。
- Deposit modal：地址、二维码、确认数要求。
- Withdraw modal：地址、数量、手续费、风控提示。
- Transaction history：hash、chain、status、confirmations。

适合你的原因：
- 和 Web3 Backend / Wallet Backend / Custody Backend 岗位最贴近。
- 可以把后端中的幂等、确认数、账本、reorg 解释成可视化状态。

### 2.7 Staking / Earn

主要功能：
- 展示可质押资产、APY、锁定期、风险等级。
- 支持 Stake、Unstake、Claim Rewards。
- 展示收益曲线、奖励记录、赎回倒计时。
- 显示合约风险、流动性风险、惩罚规则。

关键 UI：
- Earn product cards。
- Stake form。
- Position table。
- Reward history。
- Risk disclosure。

作品集建议：
- 对 Smart Contract 项目有帮助。
- 必须强调风险披露和测试状态，不要包装成真实收益产品。

### 2.8 Launchpad / Campaign / Airdrop

主要功能：
- 展示项目介绍、活动时间、资格条件。
- 钱包连接后展示任务完成状态。
- 支持白名单、积分、申购、领取空投。
- 展示规则、锁仓、vesting、claim schedule。

关键 UI：
- Campaign timeline。
- Eligibility checklist。
- Task cards。
- Allocation panel。
- Claim button。

常见交互：
- Connect wallet。
- Verify task。
- Claim。
- Share referral。

作品集建议：
- 适合展示产品设计能力。
- 与后端转区块链主线关系中等，不是最高优先级。

### 2.9 Explorer / 链上数据页

主要功能：
- 查询交易 hash、地址、区块、合约。
- 展示交易详情、事件日志、内部调用、状态变化。
- 展示地址余额、代币转账、NFT、合约交互。
- 展示区块列表、交易列表、gas、validator / miner。

关键 UI：
- Global search。
- Latest blocks。
- Latest transactions。
- Address detail。
- Transaction detail。
- Event logs。

适合你的原因：
- 和 Indexer / Blockchain Infrastructure / Protocol 项目高度匹配。
- 可以展示你理解链上数据结构和索引器设计。

### 2.10 Risk / Security Center

主要功能：
- 展示钱包授权、风险地址、异常交易、合约审计状态。
- 支持 revoke approval、地址标记、风险评分。
- 展示系统告警：reorg、RPC lag、failed withdrawal、manual review。
- 展示安全建议和操作记录。

关键 UI：
- Risk score。
- Approval list。
- Suspicious activity。
- Contract audit status。
- Incident timeline。

适合你的原因：
- 对 HR 和资深工程师都很有价值。
- 能证明你不只会写功能，还理解资金安全和故障边界。

### 2.11 Developer / Docs / API

主要功能：
- 展示 API 文档、SDK、RPC endpoint、Webhook。
- 提供 API Key 管理、调用量、错误率、限流状态。
- 提供示例代码、事件 schema、回调重试规则。

关键 UI：
- Docs sidebar。
- Endpoint list。
- Code examples。
- API key table。
- Usage metrics。

作品集建议：
- 如果你主投 Web3 Backend / Infrastructure，可以加。
- 对后端岗位非常加分，尤其是展示 API contract 和事件 schema。

### 2.12 Admin / Ops Console

主要功能：
- 展示 scanner、parser、confirm worker、ledger service 状态。
- 展示区块高度、RPC 延迟、队列积压、失败任务。
- 支持补扫、重试、人工确认、冻结地址。
- 展示 reorg 检测、ledger 对账、outbox 状态。

关键 UI：
- Service health grid。
- Chain sync status。
- Queue monitor。
- Failed job table。
- Reorg timeline。
- Ledger reconciliation panel。

适合你的原因：
- 这是最适合你当前三个项目的扩展方向。
- 它能把后端工程能力、链上基础设施和安全意识一起展示出来。

## 3. 币圈前端的通用功能组件

### 3.1 钱包连接

核心状态：
- Not connected。
- Connecting。
- Connected。
- Wrong network。
- Signature required。
- Rejected。
- Session expired。

关键功能：
- Connect wallet。
- Switch network。
- Copy address。
- Disconnect。
- Read-only mode。

### 3.2 链与代币选择器

核心功能：
- 选择链。
- 搜索代币。
- 显示余额。
- 显示合约地址。
- 标记 unknown token。

设计重点：
- 合约地址要可复制。
- 风险代币要明确提示。
- 常用 token 要优先展示。

### 3.3 交易状态组件

常见状态：
- Draft。
- Awaiting approval。
- Awaiting signature。
- Submitted。
- Pending confirmation。
- Confirmed。
- Failed。
- Reorged。

关键功能：
- 显示 hash。
- 跳转 explorer。
- 显示确认数。
- 提供 retry / dismiss。

### 3.4 风险提示组件

常见风险：
- High slippage。
- Unknown token。
- Contract not audited。
- Approval too large。
- Bridge delay。
- Reorg risk。
- RPC degraded。
- Manual review required。

设计重点：
- P0/P1/P2 优先级。
- 不要只用颜色，必须有文案和 icon。
- 高风险操作要二次确认。

### 3.5 数据表格

常见表格：
- Asset balances。
- Transactions。
- Deposits / Withdrawals。
- Orders。
- Positions。
- Events。
- Alerts。
- API calls。

关键功能：
- 搜索、过滤、排序、分页。
- 状态标签。
- hash/address 缩写与复制。
- 时间格式。
- 空状态和加载状态。

## 4. 按产品类型的页面组合

### 4.1 交易所 / CEX

必备页面：
- Home。
- Markets。
- Spot Trading。
- Futures Trading。
- Wallet / Assets。
- Deposit / Withdraw。
- Orders。
- KYC / Security。
- Earn。
- Support / Announcements。

后台或机构版：
- Risk console。
- Withdrawal approval。
- Treasury dashboard。
- Ledger reconciliation。

### 4.2 DEX / DeFi 协议

必备页面：
- Swap。
- Liquidity。
- Pool detail。
- Portfolio。
- Rewards。
- Governance。
- Analytics。
- Docs。

安全增强：
- Slippage warning。
- Price impact warning。
- Contract risk disclosure。
- Approval manager。

### 4.3 钱包 / Custody

必备页面：
- Portfolio。
- Token detail。
- Send / Receive。
- Swap / Bridge。
- Transaction history。
- Address book。
- Security center。
- Approval manager。

机构版：
- Multi-sig workflow。
- MPC policy。
- Withdrawal limits。
- Audit logs。

### 4.4 区块链基础设施 / RPC / Indexer

必备页面：
- Dashboard。
- API Keys。
- RPC endpoints。
- Usage metrics。
- Logs。
- Webhooks。
- Indexer status。
- Docs。
- Billing。

适合你的扩展页面：
- Chain sync monitor。
- Event pipeline。
- Reorg detector。
- Outbox / retry queue。

### 4.5 区块浏览器 / 链上数据平台

必备页面：
- Search。
- Blocks。
- Transactions。
- Address detail。
- Token detail。
- Contract detail。
- Event logs。
- Charts。
- Labels / risk tags。

## 5. 你的作品集建议实现路线

### 第一优先级：Backend / Infrastructure Dashboard

建议新增页面：
- Asset Dashboard。
- Deposit Lifecycle。
- Chain Event Explorer。
- Service Health。
- Risk Center。

原因：
- 最贴合 Web3 Backend / Blockchain Backend 岗位。
- 能展示链上/链下结合、确认数、账本、reorg、消息可靠性。

### 第二优先级：Protocol Visualizer

建议新增页面：
- Mempool。
- Block Builder。
- State Transition。
- Receipt Root / State Root。
- Fork-aware Storage。

原因：
- 对 Protocol Engineer / Rust Blockchain Engineer 有说服力。
- 可以把 Rust 项目的底层逻辑可视化。

### 第三优先级：Smart Contract Security Lab

建议新增页面：
- Vault overview。
- Deposit / Withdraw simulation。
- Attack cases。
- Test coverage。
- Audit checklist。

原因：
- 适合作为 Smart Contract Engineer 的辅助证明。
- 重点展示安全意识，而不是假装已生产可用。

## 6. 推荐的 MVP 页面清单

如果只做 4-6 个页面，建议如下：

| 页面 | 目标 | 对应项目 |
| --- | --- | --- |
| Career Dashboard | 展示三个项目总体定位、风险和路线图 | 三项目总览 |
| Web3 Asset Backend | 展示 scanner -> parser -> confirm -> ledger | blockchain-backend |
| Deposit Lifecycle | 展示充值从链上日志到账本入账的状态机 | blockchain-backend |
| Protocol Visualizer | 展示 mempool、block、state root、validation | protocol-rust |
| Smart Contract Security | 展示 vault 风险、测试计划、攻击面 | smart-contract |
| Risk Center | 汇总 P0/P1/P2、修复计划、验证状态 | 三项目总览 |

## 7. 前端交互规范清单

必须有：
- 明确的 loading / empty / error 状态。
- 钱包连接或只读模式状态。
- Hash / address copy。
- 状态标签和风险等级。
- 筛选、排序、搜索。
- 交易或任务的 stepper。
- 可解释的失败原因。
- 移动端可读布局。

建议有：
- WebSocket-like live indicator。
- Chain selector。
- Event timeline。
- Toast / notification。
- Command palette。
- Keyboard-friendly navigation。

不建议早期做：
- 真实下单交易。
- 高复杂 K 线交易页。
- 假的收益承诺。
- 没有安全边界的 Claim / Airdrop 页面。

## 8. 结论

币圈网站的核心不是“酷炫视觉”，而是让用户在高风险、高波动、高金额的环境里快速判断：资产在哪里、交易到哪一步、风险是什么、系统是否可靠。

对你的三个项目来说，最有价值的前端方向不是通用 Landing Page，而是：

1. Web3 Backend 运维与资产 Dashboard。
2. Deposit / Ledger / Reorg 可视化。
3. Rust Protocol 执行流程可视化。
4. Smart Contract Security Lab。
5. 三项目统一 Risk Center。

这条路线最能同时打动 HR 和资深区块链工程师。
