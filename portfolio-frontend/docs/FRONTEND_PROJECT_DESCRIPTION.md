# BNB Web3 Career Terminal 前端项目说明文档

## 1. 项目定位

`BNB Web3 Career Terminal` 是一个面向区块链求职和项目展示的前端作品集应用。它不是普通 Landing Page，而是围绕三个核心求职项目构建的 Web3 工程展示终端：

- Web3 / Blockchain Backend：链上事件、资产、充值确认、账本和基础设施运维。
- Protocol Engineer / Rust Blockchain Engineer：mempool、区块构建、状态转换、root 校验和 fork-aware storage。
- Smart Contract Engineer：Vault accounting、安全风险、测试覆盖和审计 readiness。

项目目标是让 HR 能快速理解岗位匹配度，让资深区块链工程师能进一步追问系统设计、风险边界和实现深度。

## 2. 技术栈

| 分类 | 技术 |
| --- | --- |
| 构建工具 | Vite |
| 前端框架 | React 19 |
| 类型系统 | TypeScript |
| 国际化 | i18next、react-i18next |
| 图标 | lucide-react |
| 图表/数据展示 | Recharts |
| 样式 | 原生 CSS，暗色 crypto dashboard 设计系统 |
| 质量检查 | ESLint、TypeScript build |

## 3. 当前实现状态

最近检查结果：

- `npm run build` 通过。
- `npm run lint` 通过。
- 本地服务 `http://127.0.0.1:5173/` 返回 `HTTP 200`。
- Vite 有 chunk size warning：当前 JS bundle 超过 500KB，后续建议做页面级懒加载分包。

## 4. 页面清单

| 页面 | 入口 View | 主要作用 |
| --- | --- | --- |
| Career Dashboard | `dashboard` | 三项目总览、健康矩阵、主风险、路线图 |
| Web3 Asset Dashboard | `assets` | 展示余额、充值生命周期、确认数、账本状态 |
| Infrastructure Ops Console | `ops` | 展示服务健康、RPC、队列、reorg、outbox |
| Rust Protocol Visualizer | `protocol` | 展示 mempool、block builder、state transition、root validation |
| Smart Contract Security Lab | `security-lab` | 展示 Vault accounting、攻击案例、测试覆盖和审计 checklist |
| Project Detail | `project` | 三项目详情、架构图、实现证据、下一步修复 |
| Project Comparison | `compare` | 三项目岗位定位与对比 |
| Interview Mode | `interview` | 3/8/20 分钟面试讲解路线 |
| Unified Risk Center | `risk` | 汇总三项目 P0/P1/P2 风险与修复状态 |
| Wallet | `wallet` | 钱包连接意识、多链资产和活动历史展示 |
| Market | `market` | 行情页展示，作为币圈前端扩展能力 |
| Trading | `trading` | 只读交易终端，展示订单簿、成交和交易对 |

## 5. 目录结构

```text
portfolio-frontend/app
├─ src
│  ├─ App.tsx
│  ├─ main.tsx
│  ├─ i18n.ts
│  ├─ index.css
│  ├─ components
│  ├─ data
│  ├─ pages
│  └─ types
└─ package.json
```

### 5.1 `components`

通用 UI 组件，负责可复用展示和交互：

- `AppShell`
- `SideRail`
- `Topbar`
- `AppIcon`
- `StatusPill`
- `Metric`
- `MetricBar`
- `SectionTitle`
- `ChainSelector`
- `CopyableHash`
- `ServiceHealthGrid`
- `RpcProviderTable`
- `QueueMonitor`
- `ReorgMonitor`
- `OutboxStatusPanel`
- `RiskBoard`
- `RiskDetailDrawer`
- `WalletOverview`
- `WalletBalances`
- `WalletActivity`
- `PriceTicker`
- `MarketOverview`
- `OrderBook`
- `RecentTrades`
- `TradingPairSelector`

### 5.2 `data`

模拟数据层，用来支撑当前作品集演示：

- `projects.ts`：三项目基础信息、图标映射。
- `assets.ts`：资产余额、充值记录、确认数、账本状态。
- `ops.ts`：服务健康、RPC、队列、reorg、outbox。
- `protocol.ts`：mempool、区块候选、状态 diff、root 校验、协议 review findings。
- `securityLab.ts`：Vault 模拟、攻击案例、测试覆盖、审计 checklist。
- `riskCenter.ts`：统一风险中心数据。
- `wallet.ts`：钱包和交易历史数据。
- `market.ts`：行情、订单簿、成交数据。

### 5.3 `pages`

页面级组件：

- `Dashboard.tsx`
- `AssetDashboard.tsx`
- `OpsConsole.tsx`
- `ProtocolVisualizer.tsx`
- `SecurityLab.tsx`
- `ProjectDetail.tsx`
- `Compare.tsx`
- `InterviewMode.tsx`
- `RiskCenter.tsx`
- `Wallet.tsx`
- `Market.tsx`
- `Trading.tsx`

### 5.4 `types`

统一 TypeScript 类型定义，覆盖：

- 页面路由 view。
- 项目类型。
- 资产、充值、链筛选。
- Ops console 状态。
- Risk center 状态。
- Wallet / market / trading 数据。
- Protocol visualizer 数据。
- Smart contract security lab 数据。

## 6. 核心产品能力

### 6.1 Web3 Backend 展示能力

前端已展示：

- 多链资产余额。
- 充值状态：detected、pending、confirmed、reorged。
- 确认数进度。
- scanner -> parser -> confirm worker -> ledger 链路。
- RPC、队列、reorg、outbox 运维状态。
- hash / address 复制。

这些内容对应 Web3 Backend / Blockchain Backend / Wallet Backend 岗位中常见的资产服务、充值提现、链上监听、幂等、重试和账本一致性问题。

### 6.2 Rust Protocol 展示能力

前端已展示：

- mempool 交易列表。
- nonce=0 首笔交易。
- nonce gap。
- stale transaction。
- u128 overflow risk。
- block candidate。
- atomic commit / rollback required。
- state transition。
- state_root / receipt_root / tx_root 校验。
- fork-aware storage。
- review findings。

这些内容对应 Protocol Engineer / Rust Blockchain Engineer 的协议执行、状态机、区块验证和存储设计能力。

### 6.3 Smart Contract Security 展示能力

前端已展示：

- Vault deposit / withdraw / redeem accounting。
- totalAssets / totalShares / shares delta。
- accounting invariant。
- fee-on-transfer attack。
- reward accounting bug。
- emergencyWithdraw shares bug。
- timelock hash mismatch。
- pause controls 缺失。
- mulDiv 精度与溢出风险。
- Foundry unit / fuzz / invariant / attack PoC 覆盖矩阵。
- audit checklist。

这些内容把智能合约项目定位为“安全意识和测试路线展示”，而不是未验证的生产收益产品。

## 7. 国际化

项目使用 `i18next` 和 `react-i18next`。

支持语言：

- English
- 中文
- 日本語

语言切换入口在顶部栏。主要导航、页面标题、说明、状态标签和核心交互文案已接入 i18n。部分模拟数据中的 hash、合约文件名、技术名词保持英文，这是为了贴近真实 Web3 工程语境。

## 8. 当前限制与后续建议

### 8.1 当前限制

- 当前是前端作品集演示，使用 mock 数据，没有接真实后端 API。
- 页面状态不会持久化到数据库。
- Trading 页面是只读终端，不做真实交易。
- Wallet 页面是 read-only portfolio demo，不接真实钱包。
- Build 存在 chunk size warning。

### 8.2 建议优化

优先级从高到低：

1. 页面级 code splitting：使用 `React.lazy` 和 `Suspense` 降低首包体积。
2. 将大型 CSS 拆分为页面局部样式或 CSS modules。
3. 为核心页面增加 Playwright smoke tests。
4. 将 mock data schema 对齐未来真实 API response。
5. 为 Asset Dashboard 和 Ops Console 预留后端 API adapter。
6. 增加 URL router，让页面可直接通过路径访问。
7. 增加截图和 README 演示路径。

## 9. 面试定位

推荐讲解顺序：

1. Career Dashboard：说明三个项目定位。
2. Web3 Asset Dashboard：主讲 Web3 后端资产链路。
3. Ops Console：展示生产可靠性意识。
4. Protocol Visualizer：展示 Rust 协议深度。
5. Security Lab：展示合约安全意识。
6. Risk Center：用诚实风险和修复计划收尾。

一句话总结：

> 这个前端不是单纯 UI 展示，而是把三个区块链后端 / 协议 / 合约项目转成可解释、可追问、可演示的工程证据终端。
