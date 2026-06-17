# UI 设计系统方案

## 1. 视觉方向

关键词：

- 深色金融终端
- 高对比
- 数据密集但可读
- 工程可信
- 链上基础设施感
- Binance 式黄色行动色 + Coinbase 式组件清晰度

不采用：

- 大面积紫蓝渐变
- NFT 卡通风
- 过度玻璃拟态
- 普通个人品牌博客
- 大量无意义发光装饰

## 2. 色彩系统

### 背景

```css
--bg-base: #07080b;
--bg-elevated: #0d1117;
--bg-panel: #111722;
--bg-panel-soft: #151c29;
```

### 文本

```css
--text-primary: #f4f7fb;
--text-secondary: #a8b0bf;
--text-muted: #6f7a8c;
```

### 边框

```css
--border-subtle: #202838;
--border-strong: #2f3a4f;
```

### 品牌与语义色

```css
--accent-yellow: #f0b90b;
--accent-yellow-soft: #3b2c08;
--accent-cyan: #15d4c8;
--accent-blue: #377dff;
--success: #16c784;
--warning: #f0b90b;
--danger: #f6465d;
--purple: #8b5cf6;
```

使用规则：

- 黄色只用于主 CTA、当前选中、关键机会点。
- 绿色用于 pass/healthy/confirmed。
- 红色用于 P0、blocked、critical。
- 蓝色/青色用于 protocol/data flow/technical signal。
- 紫色只少量用于 smart contract/security，不做主色。

## 3. 字体系统

推荐：

- UI 字体：`Inter`, `SF Pro Display`, `Segoe UI`, sans-serif。
- 数字和代码：`JetBrains Mono`, `SF Mono`, monospace。

字号：

```css
--font-xs: 12px;
--font-sm: 13px;
--font-md: 14px;
--font-lg: 16px;
--font-xl: 20px;
--font-2xl: 28px;
--font-3xl: 40px;
```

规则：

- Dashboard 内部不使用夸张 hero 字号。
- 指标数字使用 mono，增强金融终端感。
- 控件文字不使用浏览器默认 16px，统一 13-14px。
- 字间距保持 0。

## 4. 间距与布局

基础单位：4px。

主要 spacing：

```css
--space-1: 4px;
--space-2: 8px;
--space-3: 12px;
--space-4: 16px;
--space-5: 20px;
--space-6: 24px;
--space-8: 32px;
--space-10: 40px;
--space-12: 48px;
```

布局规则：

- 卡片圆角不超过 8px。
- Dashboard 区块密度偏紧，但每组信息之间留 24-32px。
- 不使用卡套卡。
- 项目卡可以是卡片，但页面 section 不做漂浮大卡。

## 5. 组件风格

### Button

Primary：

- 背景：黄色。
- 文本：黑色。
- 高度：36px 或 40px。
- 圆角：6px。
- 用途：进入项目、打开面试模式、查看修复路线。

Secondary：

- 背景：透明或 panel。
- 边框：subtle。
- 文本：primary。

IconButton：

- 用 lucide 图标。
- 适合复制路径、切换视图、展开详情、过滤风险。

### Status Badge

状态：

- `PASS`
- `BLOCKED`
- `PARTIAL`
- `NO TESTS`
- `P0`
- `P1`
- `P2`

规则：

- badge 不能只靠颜色表达。
- P0 badge 颜色红，P1 黄色，P2 蓝灰。

### Project Card

每张项目卡展示：

- 项目名。
- 对应岗位。
- 技术栈。
- 当前状态。
- HR 匹配度。
- 技术深度。
- 最大风险。
- CTA：`Open Deep Dive`。

卡片结构：

```text
[status dot] Web3 Backend
Blockchain asset backend
Go · PostgreSQL · NATS · Ethereum RPC

Market Fit  92
Depth       84
Tests       Blocked

Main Risk: Go version blocks tests
[Open Deep Dive]
```

### Risk List

风险列表像交易风控台：

```text
P0  deposit_confirmed publish retry gap
P0  ledger transaction consistency
P1  core NATS ack is not durable
```

交互：

- 可按 P0/P1/P2 过滤。
- 可切换 `Impact` / `Fix Plan`。
- 点击展开文件路径和修复建议。

### Architecture Map

展示方式：

- Web3 Backend：event pipeline。
- Protocol Rust：block execution pipeline。
- Smart Contract：vault asset/share lifecycle。

视觉：

- 横向流程图。
- 节点用矩形，不用过度圆角。
- 成功路径用青色线。
- 风险节点用红色边框。

## 6. 图表类型

### Role Fit Radar

维度：

- Backend Fit
- Protocol Depth
- Smart Contract Readiness
- Test Readiness
- Market Value

### Health Matrix

横轴：市场匹配度。

纵轴：技术完成度。

点：

- Web3 Backend：右上偏强，但 testing blocked。
- Rust Protocol：中上，技术深但产品闭环不足。
- Smart Contract：中间偏低，测试缺口明显。

### Roadmap Gantt

展示 30 天优化路线：

- Week 1：Web3 Backend 可验证。
- Week 2：Rust Protocol 深挖。
- Week 3：Smart Contract 安全测试。
- Week 4：统一展示和面试材料。

## 7. 动效规范

允许：

- 页面进入时轻微 fade/slide。
- hover 显示边框强化。
- 流程图节点按顺序高亮。
- 面试模式切换时内容平滑过渡。

避免：

- 大面积粒子背景。
- 过度霓虹闪烁。
- 影响阅读的循环动画。
- 模拟交易涨跌但和项目无关。

## 8. 图标规范

使用 lucide：

- Server：Web3 Backend。
- Blocks：Protocol。
- ShieldCheck：Smart Contract/Security。
- GitBranch：reorg/fork。
- Database：ledger/storage。
- AlertTriangle：risk。
- CheckCircle：pass。
- XCircle：blocked。
- Terminal：interview/demo mode。

图标尺寸：

- 导航：18px。
- 卡片标题：20px。
- 小状态：14px。

## 9. 首屏视觉方案

首屏名称：`Career Command Center`

布局：

- 顶部：品牌名、模式切换、Demo Wallet 状态。
- 左侧：项目导航。
- 中间：三项目 Portfolio Overview。
- 右侧：Offer Strategy / Current Blockers。
- 下方露出下一屏：Project Health Matrix。

首屏文案：

```text
BNB Web3 Career Terminal
Backend → On-chain Infrastructure → Protocol Depth
```

主 CTA：

```text
Open Web3 Backend
```

次 CTA：

```text
Start 8-min Interview Mode
```

