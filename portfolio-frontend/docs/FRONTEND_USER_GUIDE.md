# BNB Web3 Career Terminal 使用文档

## 1. 快速启动

进入前端应用目录：

```bash
cd E:\awesomeProject\BNB\portfolio-frontend\app
```

安装依赖：

```bash
npm install
```

启动开发服务：

```bash
npm run dev -- --host 127.0.0.1
```

浏览器访问：

```text
http://127.0.0.1:5173/
```

构建生产版本：

```bash
npm run build
```

运行 lint：

```bash
npm run lint
```

## 2. 顶部栏使用

顶部栏包含三类控制：

### 2.1 语言切换

可切换：

- `EN`
- `中`
- `日`

用于展示英文、中文、日语三种面试和作品集语境。

### 2.2 视角切换

可切换：

- `HR View`
- `Senior Dev`

用途：

- HR View：更强调岗位匹配、项目定位和易理解表达。
- Senior Dev：更强调架构、风险、实现细节和技术追问点。

### 2.3 Read-only Portfolio / Demo Connected

这是一个演示状态按钮，不连接真实钱包。

用途：

- 展示钱包连接意识。
- 表明当前系统是 read-only portfolio demo。
- 避免被误解为真实资金操作页面。

## 3. 侧边栏导航

侧边栏通过图标进入不同页面。

| 页面 | 用途 |
| --- | --- |
| Dashboard | 三项目职业终端总览 |
| Assets | Web3 资产后端 Dashboard |
| Ops Console | 区块链基础设施运维视图 |
| Protocol | Rust Protocol Visualizer |
| Security Lab | Smart Contract Security Lab |
| Risk Center | 三项目统一风险中心 |
| Projects | 三个项目详情 |
| Compare | 三项目对比 |
| Interview | 面试讲解模式 |
| Wallet | 钱包与多链资产演示 |
| Market | 行情展示 |
| Trading | 只读交易终端 |

## 4. 推荐演示路线

### 4.1 3 分钟 HR 路线

适合招聘初筛。

路线：

1. 打开 `Dashboard`。
2. 说明主线是 Web3 Backend / Blockchain Backend。
3. 打开 `Assets`。
4. 展示充值从链上事件到账本入账。
5. 打开 `Risk Center`。
6. 说明你知道项目当前风险，并有修复计划。

重点表达：

- 你不是只会写页面，而是在展示链上/链下结合的后端能力。
- Web3 Backend 是主项目。
- Protocol 和 Smart Contract 是深度与安全意识补充。

### 4.2 8 分钟技术路线

适合技术面第一轮。

路线：

1. `Dashboard`：说明三项目结构。
2. `Assets`：讲 scanner -> parser -> confirmation -> ledger。
3. `Ops Console`：讲 RPC、queue、reorg、outbox。
4. `Protocol`：讲 mempool、state transition、root validation。
5. `Security Lab`：讲合约风险和测试覆盖。
6. `Risk Center`：讲 P0/P1/P2 和修复优先级。

重点表达：

- Web3 系统的关键不是页面，而是状态正确性和资金安全。
- 你理解确认数、reorg、幂等、outbox、账本一致性。
- 你能把 review findings 转成工程路线图。

### 4.3 20 分钟深度路线

适合资深开发或架构面。

路线：

1. `Assets`
   - 讲充值状态机。
   - 点击不同 deposit 状态：detected、pending、confirmed、reorged。
   - 展示 tx hash、account、raw event、parsed event、ledger status。

2. `Ops Console`
   - 讲服务健康状态。
   - 展示 RPC provider 状态。
   - 展示 queue backlog。
   - 展示 reorg timeline。
   - 展示 outbox retry。

3. `Protocol`
   - 讲 mempool 中 nonce=0、nonce gap、stale tx。
   - 切换 valid / invalid block scenario。
   - 展示 atomic commit 和 rollback required。
   - 展示 receipt_root mismatch。
   - 展示 fork-aware storage。

4. `Security Lab`
   - 切换 deposit / withdraw / redeem。
   - 点击 fee-on-transfer、reward accounting、emergencyWithdraw 等攻击案例。
   - 展示测试覆盖矩阵。
   - 展示 audit checklist。

5. `Risk Center`
   - 按项目、优先级、状态筛选风险。
   - 说明下一步修复顺序。

## 5. 页面使用说明

## 5.1 Dashboard

用途：

- 快速说明三项目定位。
- 展示市场匹配度和技术深度。
- 展示当前 blocker 和 30 天 roadmap。

可操作项：

- 点击项目卡片进入项目详情。
- 点击主按钮进入 Web3 Backend 项目。
- 切换 HR / Senior Dev 影响讲解视角。

## 5.2 Assets

用途：

- 展示 Web3 Backend 的资产系统能力。
- 重点讲链上事件如何变成链下余额。

可操作项：

- 使用链筛选：All、Ethereum、BNB Chain、Polygon、Solana。
- 查看 Token Balance Table。
- 点击 Deposit Lifecycle 中不同充值记录。
- 查看右侧详情面板。
- 点击 hash 或 address 复制。

讲解重点：

- detected：scanner 已发现事件。
- pending：确认数不足，不能入账。
- confirmed：确认数达标，可以进入 ledger。
- reorged：链重组导致事件失效，必须补偿。

## 5.3 Ops Console

用途：

- 展示区块链基础设施运维能力。

可操作项：

- 切换 Services、RPC Providers、Queue Monitor、Reorg Events、Outbox Status。
- 查看服务状态、错误数、延迟、队列积压和重试次数。

讲解重点：

- RPC 延迟会影响 scanner。
- 队列积压会影响 parser 和 ledger。
- reorg 会影响已检测事件。
- outbox 用来避免 confirmed 事件丢失。

## 5.4 Protocol

用途：

- 展示 Rust Protocol 项目的底层理解。

可操作项：

- 点击 mempool 交易查看详情。
- 切换 Valid / Invalid scenario。
- 查看 block candidate。
- 查看 state transition。
- 查看 root validation。
- 查看 review findings。

讲解重点：

- nonce=0 首笔交易不能被 mempool 拒绝。
- nonce gap 交易不能提前执行。
- stale transaction 应被过滤。
- u128 value 不能被截断到 u64。
- block execution 必须 atomic。
- receipt root 必须覆盖 receipt 内容。
- storage 需要支持 fork，而不是只按高度存 block。

## 5.5 Security Lab

用途：

- 展示 Smart Contract 项目的安全意识。

可操作项：

- 切换 deposit、withdraw、redeem。
- 查看 accounting invariant。
- 点击不同 attack case。
- 查看 test coverage matrix。
- 查看 audit checklist。

讲解重点：

- 该合约目前不应包装成生产收益产品。
- 重点是展示你知道风险在哪里、如何测试、如何修复。
- fee-on-transfer、reward accounting、emergencyWithdraw、timelock hash mismatch 都是高价值追问点。

## 5.6 Risk Center

用途：

- 汇总所有项目风险。
- 把问题转成工程修复计划。

可操作项：

- 按项目筛选。
- 按 P0/P1/P2 筛选。
- 按状态筛选：open、in progress、fixed、blocked。
- 点击风险查看详情。

讲解重点：

- 不隐藏问题。
- 用风险优先级体现工程判断。
- 用修复计划体现可执行能力。

## 5.7 Wallet

用途：

- 展示钱包集成意识和多链资产展示。

可操作项：

- 查看 demo wallet 状态。
- 查看多链余额。
- 查看最近交易活动。

说明：

- 当前不连接真实钱包。
- 当前不签名、不发交易、不请求真实资产。

## 5.8 Market

用途：

- 展示币圈行情页基础能力。

可操作项：

- 查看行情卡片和市场概览。

说明：

- 当前为 mock 行情数据。
- 不作为核心求职主线。

## 5.9 Trading

用途：

- 展示只读交易终端界面。

可操作项：

- 切换交易对。
- 查看订单簿。
- 查看最近成交。

说明：

- 当前不支持真实下单。
- 不模拟资金风险操作。
- 适合作为交易系统界面意识展示。

## 6. 开发者说明

### 6.1 添加新页面

步骤：

1. 在 `src/types/index.ts` 的 `View` 中添加新 view。
2. 在 `src/pages` 中新增页面组件。
3. 在 `src/pages/index.ts` 导出页面。
4. 在 `src/App.tsx` 中添加 view 渲染分支。
5. 在 `src/components/SideRail.tsx` 中添加导航入口。
6. 在 `src/i18n.ts` 中添加导航文案和页面文案。
7. 在 `src/index.css` 中补页面样式。

### 6.2 添加模拟数据

建议位置：

```text
src/data/<feature>.ts
```

要求：

- 数据类型先写入 `src/types/index.ts`。
- 页面只消费 data，不在组件内硬编码大量业务数据。
- Hash、address、block number、status 要贴近真实 Web3 场景。

### 6.3 添加国际化文案

位置：

```text
src/i18n.ts
```

要求：

- 英文文案优先准确。
- 中文文案用于自我复盘和中文面试。
- 日语文案用于展示多语言能力。
- 技术名词可以保留英文，例如 mempool、outbox、reorg、ledger。

## 7. 质量检查

每次修改后运行：

```bash
npm run lint
npm run build
```

当前已知提示：

- Vite 构建时提示部分 chunk 超过 500KB。
- 这是性能优化提示，不是构建失败。
- 后续建议用 `React.lazy` 做页面级 code splitting。

## 8. 面试时的注意事项

不要把页面说成真实生产系统。

推荐说法：

> 这是一个基于我三个区块链项目整理出来的前端作品集终端，数据是 mock 的，但页面状态、风险模型和系统链路都对应真实 Web3 后端、协议和合约安全场景。

避免说法：

- “这是一个真实交易所。”
- “这里可以真实连接钱包。”
- “这个 Vault 已经可以生产使用。”
- “这些收益是真实收益。”

## 9. 最佳展示顺序

推荐默认展示顺序：

```text
Dashboard
-> Assets
-> Ops Console
-> Protocol
-> Security Lab
-> Risk Center
-> Interview
```

这条路线最能说明：

- 你主线是 Web3 Backend。
- 你懂区块链基础设施的真实问题。
- 你有 Rust 协议深度。
- 你知道智能合约安全边界。
- 你能诚实识别风险并规划修复。
