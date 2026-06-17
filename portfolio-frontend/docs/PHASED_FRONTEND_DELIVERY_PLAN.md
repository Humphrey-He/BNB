# 币圈作品集前端阶段性开发排期表

本文档基于 `CRYPTO_FRONTEND_PAGE_CHECKLIST.md`，把币圈网站常见页面能力转化为 `BNB Web3 Career Terminal` 的阶段性开发计划。目标不是一次性做成“大而全交易所”，而是围绕你的三个求职项目，逐步交付能被 HR 看懂、能被资深区块链工程师追问的前端功能。

## 总体周期

建议周期：8 周。

开发节奏：
- 每周交付一个可演示版本。
- 每个阶段都必须有可点击页面、明确状态、模拟数据和验收标准。
- 优先做 Web3 Backend / Infrastructure Dashboard，再补 Protocol Visualizer 和 Smart Contract Security Lab。

## 阶段总览

| 阶段 | 周期 | 阶段目标 | 主要页面 | 交付优先级 |
| --- | --- | --- | --- | --- |
| Phase 0 | 第 0 周 | 设计系统和信息架构固化 | Docs / Design System | 必做 |
| Phase 1 | 第 1 周 | 完成三项目职业终端 MVP | Career Dashboard、Project Detail、Risk Register | 已基本完成，继续增强 |
| Phase 2 | 第 2 周 | 做 Web3 Backend 资产与链路展示 | Asset Dashboard、Deposit Lifecycle | 最高 |
| Phase 3 | 第 3 周 | 做后端基础设施运维视图 | Service Health、Queue Monitor、Reorg Monitor | 最高 |
| Phase 4 | 第 4 周 | 做链上数据与事件浏览器 | Chain Event Explorer、Transaction Detail | 高 |
| Phase 5 | 第 5 周 | 做 Rust Protocol 可视化 | Mempool、Block Builder、State Transition | 高 |
| Phase 6 | 第 6 周 | 做 Smart Contract 安全实验室 | Vault Overview、Attack Cases、Test Coverage | 中高 |
| Phase 7 | 第 7 周 | 做统一 Risk Center 和 Interview Mode 强化 | Risk Center、Interview Flow | 高 |
| Phase 8 | 第 8 周 | polish、响应式、多语言、部署与演示脚本 | 全站 | 必做 |

## Phase 0：设计系统与架构固化

周期：第 0 周，1-2 天。

目标：
- 固化币圈产品风格、页面信息架构和组件体系。
- 避免后续页面各自长成不同风格。

开发内容：
- 整理页面路由结构。
- 拆分 `App.tsx` 中的数据、组件和 i18n 文案。
- 确定 mock data schema。
- 建立通用组件规范。

主要组件：
- `AppShell`
- `SideRail`
- `Topbar`
- `StatusPill`
- `MetricBar`
- `RiskItem`
- `DataTable`
- `Timeline`
- `ChainSelector`
- `CopyableHash`
- `EmptyState`
- `LoadingState`

验收标准：
- 所有页面使用统一 dark crypto dashboard 风格。
- 英文、中文、日语切换可用。
- mock 数据集中管理，不散落在组件中。
- `npm run build` 和 `npm run lint` 通过。

## Phase 1：Career Dashboard MVP

周期：第 1 周。

目标：
- 完成当前作品集入口，让用户能理解三个项目的定位、风险和优先级。

已实现基础：
- Dashboard。
- Project Detail。
- Compare。
- Interview Mode。
- Risk Register。
- Architecture Map。
- i18n 英文 / 中文 / 日语。

继续增强内容：
- 首页增加“目标岗位匹配度”区域。
- Project Detail 增加“代码证据入口”，例如 README、测试状态、关键模块。
- Interview Mode 增加“HR 版 / Senior 版讲解脚本”切换。
- Risk Register 增加“已修复 / 未修复 / 阻塞”状态。

主要页面：
- `/dashboard`
- `/projects/:projectId`
- `/compare`
- `/interview`

主要功能：
- 三项目卡片。
- 项目健康矩阵。
- 风险筛选。
- 架构节点 hover / selected。
- 多语言切换。

验收标准：
- 第一次打开页面能在 10 秒内看懂主投方向。
- 三个项目都能进入详情页。
- 风险项可筛选、可切换影响和修复计划。
- 移动端不出现文字溢出。

## Phase 2：Web3 Asset Dashboard

周期：第 2 周。

目标：
- 把 `blockchain-backend` 项目转换为可视化资产后端产品。
- 展示链上数据到链下资产账户的业务闭环。

新增页面：
- Web3 Asset Dashboard。
- Deposit Lifecycle。
- Account Balance Detail。

主要功能：
- 展示总资产、可用余额、冻结余额、待确认充值。
- 按链筛选：All、Ethereum、BNB Chain、Polygon、Solana。
- 展示 token balance table。
- 展示最近 deposit / withdraw / ledger activity。
- 展示充值确认数进度。
- 展示 hash/address copy 和 explorer link 占位。

核心组件：
- `AssetSummaryCards`
- `ChainFilter`
- `TokenBalanceTable`
- `DepositTimeline`
- `TransactionStatusStepper`
- `CopyableAddress`

模拟数据：
- 账户地址。
- token 列表。
- 链 ID。
- 余额。
- 充值记录。
- 确认数。
- ledger entry。

交互重点：
- 点击某个 deposit 后，在右侧详情面板展示：
  - raw log。
  - parsed event。
  - confirmations。
  - ledger status。
  - balance delta。
- 切换链后，表格和活动流同步过滤。
- 点击 hash 复制并显示 toast。

验收标准：
- 面试官可以清楚看到 scanner -> parser -> confirmation -> ledger 的价值。
- 至少有 3 条不同状态的 deposit：detected、pending、confirmed。
- 充值确认进度不是静态文案，而是 stepper / progress。
- 所有资金状态有明确颜色、图标和解释。

## Phase 3：Infrastructure Ops Console

周期：第 3 周。

目标：
- 展示区块链基础设施后端能力，包括 RPC、队列、reorg、worker 健康状态。

新增页面：
- Service Health。
- Queue Monitor。
- Reorg Monitor。
- RPC Provider Status。

主要功能：
- 展示 scanner、parser、confirm worker、ledger service 状态。
- 展示 RPC provider latency、error rate、latest block。
- 展示 queue backlog、retry count、dead letter。
- 展示 reorg 检测 timeline。
- 展示 outbox / inbox 状态。

核心组件：
- `ServiceHealthGrid`
- `RpcProviderTable`
- `QueueBacklogChart`
- `ReorgTimeline`
- `OutboxStatusPanel`
- `IncidentBanner`

模拟数据：
- 服务状态：healthy、degraded、down。
- RPC 延迟。
- 当前扫描高度。
- 队列堆积数量。
- reorg 事件。
- outbox 待发布事件。

交互重点：
- 点击服务卡片查看服务详情。
- 队列支持按 topic 过滤。
- reorg timeline 可展开查看 affected blocks。
- outbox 事件可切换 pending / published / failed。

验收标准：
- 页面能体现你理解生产级 Web3 后端的可靠性问题。
- 明确展示 reorg、RPC lag、outbox retry 这些高价值概念。
- 所有异常状态有修复建议或 next action。

## Phase 4：Chain Event Explorer

周期：第 4 周。

目标：
- 展示链上数据解析、索引器和事件浏览能力。

新增页面：
- Chain Event Explorer。
- Transaction Detail。
- Address Activity。
- Raw Event Detail。

主要功能：
- 搜索 tx hash、address、block number。
- 展示事件列表：Transfer、Deposit、Withdraw、Swap。
- 展示事件原始字段和解析字段。
- 展示区块高度、block hash、tx hash、log index。
- 展示 parsed status、save status、matched account。

核心组件：
- `GlobalChainSearch`
- `EventTable`
- `EventDetailDrawer`
- `RawJsonViewer`
- `DecodedEventPanel`
- `AddressActivityList`

模拟数据：
- ERC-20 Transfer logs。
- parsed deposits。
- failed parse events。
- unknown contract events。
- duplicate event examples。

交互重点：
- 搜索后高亮目标事件。
- 点击事件打开详情抽屉。
- raw / decoded 两个视图切换。
- 支持只看 failed parse。

验收标准：
- 能体现 indexer / parser 工程能力。
- 能展示原始链上数据和业务数据之间的映射。
- 至少包含一个解析失败案例和修复提示。

## Phase 5：Rust Protocol Visualizer

周期：第 5 周。

目标：
- 把 `protocol-rust-blockchain` 项目从代码工程变成可解释的协议执行流程。

新增页面：
- Mempool。
- Block Builder。
- State Transition。
- Root Verification。
- Fork-aware Storage Preview。

主要功能：
- 展示交易池中的交易，按 fee、nonce、sender 排序。
- 展示 block builder 如何选择交易。
- 展示交易执行前后账户状态变化。
- 展示 state root、receipt root、tx root。
- 展示 block validation 结果。
- 展示 fork storage 的设计草图。

核心组件：
- `MempoolTable`
- `BlockCandidatePanel`
- `StateDiffViewer`
- `RootHashPanel`
- `ValidationChecklist`
- `ForkGraph`

模拟数据：
- pending transactions。
- nonce gap。
- stale transaction。
- block candidate。
- execution result。
- validation error。

交互重点：
- 点击交易加入 block candidate。
- 执行区块后展示 state diff。
- 切换 valid / invalid block 示例。
- 展示 root mismatch 的错误解释。

验收标准：
- 面试官能直观看到你理解 nonce、mempool、state transition、root validation。
- 至少有一个失败区块示例。
- 不能只做静态流程图，必须有交互状态。

## Phase 6：Smart Contract Security Lab

周期：第 6 周。

目标：
- 把 `smart-contract` 项目定位为安全意识展示，而不是未完成合约包装。

新增页面：
- Vault Overview。
- Deposit / Withdraw Simulation。
- Attack Cases。
- Test Coverage。
- Audit Checklist。

主要功能：
- 展示 Vault asset / share accounting。
- 模拟 deposit、withdraw、redeem 的状态变化。
- 展示 fee-on-transfer、reentrancy、rounding、timelock hash mismatch 等攻击案例。
- 展示测试覆盖：unit、fuzz、invariant、PoC。
- 展示审计 checklist 和未修复风险。

核心组件：
- `VaultAccountingPanel`
- `ShareAssetSimulator`
- `AttackCaseList`
- `TestCoverageMatrix`
- `AuditChecklist`
- `SecurityFindingDetail`

模拟数据：
- deposit amount。
- shares minted。
- reward injected。
- withdraw assets。
- attack scenario。
- test status。

交互重点：
- 输入 deposit amount，模拟 shares 变化。
- 点击 attack case 查看影响和修复方案。
- 切换 tested / missing / planned。

验收标准：
- 明确展示安全边界，不把未测试合约伪装成生产可用。
- 至少展示 4 个安全风险案例。
- 测试覆盖矩阵可读、可筛选。

## Phase 7：Unified Risk Center + Interview Flow

周期：第 7 周。

目标：
- 将三个项目的风险、修复计划和面试叙事统一管理。

新增页面：
- Unified Risk Center。
- Interview Script Builder。
- Offer Target Matrix。

主要功能：
- 汇总所有 P0/P1/P2 风险。
- 支持按项目、优先级、状态过滤。
- 展示修复进度和证据链接。
- 生成 HR / Senior / Protocol / Backend 四种讲解路线。
- 展示岗位匹配矩阵。

核心组件：
- `RiskBoard`
- `FindingDrawer`
- `FixProgressTimeline`
- `InterviewScriptTabs`
- `RoleFitMatrix`

模拟数据：
- 所有项目 findings。
- fix status：open、in progress、fixed、blocked。
- evidence links。
- role mapping。

交互重点：
- 点击风险进入详情。
- 切换面试对象自动调整讲解顺序。
- 标记“主讲 / 可追问 / 暂不展开”。

验收标准：
- 页面能支持真实面试前 5 分钟快速复盘。
- 风险不是隐藏项，而是变成专业可信的工程路线图。
- HR 视角和资深开发视角文案有明显差异。

## Phase 8：Polish、响应式、部署与演示脚本

周期：第 8 周。

目标：
- 让项目达到可以投递简历和现场演示的完成度。

开发内容：
- 全站响应式检查。
- 多语言文案润色。
- 页面空状态、错误状态、加载状态补齐。
- 增加演示数据 reset。
- 增加 README 截图和功能说明。
- 部署到 Vercel / Render / Netlify。
- 准备 3 分钟、8 分钟、20 分钟演示脚本。

验收标准：
- `npm run build` 通过。
- `npm run lint` 通过。
- 桌面端和移动端核心页面可用。
- 英文 / 中文 / 日语主要页面无明显溢出。
- README 能解释页面和三个后端项目的关系。
- 面试时可以按固定路径演示，不需要临时找页面。

## 推荐开发顺序

如果时间有限，优先顺序如下：

1. Web3 Asset Dashboard。
2. Deposit Lifecycle。
3. Infrastructure Ops Console。
4. Unified Risk Center。
5. Rust Protocol Visualizer。
6. Smart Contract Security Lab。
7. Chain Event Explorer。
8. Market / Trading / Launchpad 类页面。

理由：
- 前四项最贴合你的主投方向：Web3 Backend / Blockchain Backend。
- Rust Protocol 能补技术深度。
- Smart Contract Security Lab 适合作为辅助证明。
- Market / Trading / Launchpad 更偏产品展示，不是当前最高收益。

## 两周冲刺版

如果你只想快速做出可投递版本，可以压缩为两周：

### Week 1

目标：强化 Web3 Backend 主项目。

开发内容：
- Asset Dashboard。
- Deposit Lifecycle。
- Service Health。
- Risk Center 基础版。

验收：
- 能完整讲清充值从链上日志到账本入账。
- 能展示 outbox、reorg、confirmations 的工程价值。

### Week 2

目标：补技术深度和安全意识。

开发内容：
- Protocol Visualizer 基础版。
- Smart Contract Security Lab 基础版。
- Interview Mode 强化。
- README 和演示脚本。

验收：
- 能分别给 HR 和资深开发演示。
- 能解释为什么 Web3 Backend 是主线，Rust 和 Smart Contract 是支撑线。

## 页面与项目映射

| 前端页面 | 展示能力 | 对应后端项目 | 面试价值 |
| --- | --- | --- | --- |
| Asset Dashboard | 资产、余额、账户系统 | blockchain-backend | 很高 |
| Deposit Lifecycle | 事件监听、确认数、幂等、账本 | blockchain-backend | 很高 |
| Service Health | RPC、worker、队列、监控 | blockchain-backend | 很高 |
| Chain Event Explorer | indexer、parser、链上数据 | blockchain-backend | 高 |
| Protocol Visualizer | mempool、executor、root validation | protocol-rust | 高 |
| Security Lab | 合约风险、测试、攻击面 | smart-contract | 中高 |
| Risk Center | 工程判断、诚实风险、修复计划 | 三项目 | 很高 |
| Interview Flow | 求职表达、岗位匹配、项目包装 | 三项目 | 高 |

## 每阶段验收清单

每个阶段完成前都要检查：

- 页面是否有明确用户目标。
- 是否有真实币圈状态：pending、confirmed、failed、reorged、degraded。
- 是否有可点击交互，而不是静态展示。
- 是否有 loading、empty、error 状态。
- 是否支持英文、中文、日语。
- 是否移动端可读。
- 是否能映射到一个具体面试问题。
- 是否能解释背后的后端或协议设计。
- 是否有风险边界说明。
- 是否通过 `npm run build` 和 `npm run lint`。

## 最终交付物

第 8 周结束时，建议形成：

- 一个可在线访问的前端作品集。
- 三个项目的可视化入口。
- 一份英文 README。
- 一份中文项目说明。
- 一套面试演示脚本。
- 一套截图：Dashboard、Deposit Lifecycle、Ops Console、Protocol Visualizer、Security Lab、Risk Center。

最终目标：让页面本身成为你的技术叙事工具，而不是单纯展示“我做了三个项目”。
