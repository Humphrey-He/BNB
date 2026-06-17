# 资料调研与设计依据

调研日期：2026-05-06

## 1. 主流币圈产品 UI 规律

### 1.1 暗色、高对比、数据优先

币圈产品常见暗色界面，尤其是交易、钱包、DeFi、资产监控、链上数据工具。暗色背景有三个作用：

- 降低长时间看盘和看数据的视觉疲劳。
- 让价格、状态、风险、交易动作更突出。
- 营造专业金融工具而非普通营销站的感觉。

设计参考：

- Binance 风格分析中提到 dark-mode、高对比、黄色 CTA、紧凑 8px spacing 这类特征，适合交易和市场类产品。
- Webflow 的 DeFi dashboard 模板描述了 dark-mode、live trading interface、modular dashboard、portfolio tracking、wallet connection 等常见组合。

落地到本项目：

- 不做浅色个人主页。
- 首页做“工程资产/项目状态仪表盘”。
- 主色用黑色/深灰作为底，黄色作为关键行动，青绿色/蓝色作为链路状态和工程可信度。

### 1.2 组件系统必须清晰和可访问

Coinbase Design System 明确强调组件、设计 tokens、可访问性、focus 状态、色彩对比和键盘支持。它不是单纯视觉规范，而是 production-ready crypto UI 的工程化方式。

落地到本项目：

- 使用 tokens 管理颜色、间距、圆角、阴影、字体。
- 所有标签页、按钮、筛选器、项目卡、风险状态都有一致状态。
- 即使视觉偏币圈，也不能牺牲可读性。

### 1.3 钱包连接是 Web3 心智，但本项目不需要真连接

RainbowKit 和 WalletConnect 代表了用户熟悉的 Web3 wallet 交互：Connect Wallet、链切换、账户状态、session、QR/一键连接。

本项目是求职展示，不需要真的连钱包。建议做“模拟 Web3 Shell”：

- 顶部显示 `Demo Wallet` / `Read-only Portfolio`。
- 提供一个非真实交易的 `Connect Demo` 状态切换。
- 用钱包交互的视觉语言表达 Web3 native，但避免引入真实钱包依赖造成实现和安全复杂度。

如果后续要做真实 dApp demo，可再引入 RainbowKit + wagmi + viem。

## 2. 资料来源

### Coinbase Design System

来源：

- [Coinbase Design System](https://cds.coinbase.com/)
- [Coinbase open-sourced its design system](https://www.coinbase.com/en-gb/blog/Coinbase-has-open-sourced-its-design-system)

关键启发：

- Coinbase CDS 是面向 crypto products 的开源组件系统。
- 它强调 layout、typography、inputs、cards、data display、feedback、navigation、charts 等组件覆盖。
- 官方博客强调可访问性、focus indication、色彩对比、键盘支持、design tokens 和 themeability。

应用方式：

- 本前端采用 token-driven design。
- 组件要覆盖：导航、指标卡、项目详情、风险列表、状态反馈、图表、命令面板。

### RainbowKit

来源：

- [RainbowKit](https://v0.rainbowkit.com/)

关键启发：

- Web3 用户熟悉 `Connect Wallet`、chain、account、wallet modal 等交互。
- RainbowKit 强调 easy installation、custom themes、light/dark mode、custom chains、custom connect button。

应用方式：

- 首版做模拟 wallet shell，不接真实钱包。
- 后续扩展可用 RainbowKit/wagmi/viem 增加真实连接。

### WalletConnect

来源：

- [WalletConnect Wallet SDK Web Usage](https://docs.walletconnect.network/wallet-sdk/web/usage)

关键启发：

- WalletConnect 文档强调 session proposal、session requests、URI、QR code、one-click connection。
- 对 Web3 用户来说，连接、授权、会话和网络状态是关键 UX。

应用方式：

- 在交互文档里保留 wallet/session 状态位。
- 项目详情页可用“transaction lifecycle”方式展示链上/链下流程。

### DeFi Dashboard / Portfolio UI

来源：

- [Walletco DeFi Portfolio Dashboard Webflow template](https://webflow.com/made-in-webflow/website/walletco)

关键启发：

- 常见模块包括 portfolio tracking、multi-token visibility、live trading interface、yield farming、top gainers、liquidity pools、ROI insights、historical charts、multi-wallet tracking。

应用方式：

- 本项目不是展示真实资产，而是把三个项目当成“工程资产组合”。
- 使用类似 dashboard 模块：项目健康度、验证状态、风险敞口、技术深度、岗位匹配。

### OKX Web3 Security

来源：

- [OKX Web3 Security White Paper](https://web3.okx.com/cdn/oksupport/common/okx-web3-security-white-paper.fa6993b5b2919b.pdf)

关键启发：

- Web3 产品需要把安全能力显性化。
- 对合约和资产系统，风险、权限、审计、交易安全不是附属信息，而是核心产品信息。

应用方式：

- 三个项目都展示 `Risk Register`。
- Smart Contract 项目必须突出测试缺口和安全路线，而不是只展示合约功能。

## 3. 设计关键词

推荐关键词：

- Crypto-native
- Terminal dashboard
- Asset infrastructure
- Protocol depth
- Security-aware
- High contrast
- Dense but readable
- Trust through evidence
- Interview-ready

避免关键词：

- Generic portfolio
- Personal blog
- Landing-page-heavy
- Decorative Web3 glow
- NFT cartoon style
- Over-purple gradient
- Empty hero slogan

## 4. 对本项目的设计判断

最合适的形式不是“个人主页 + 三张项目卡”，而是：

> 一个类似交易所/链上监控台的工程项目仪表盘。

原因：

- Web3 Backend 项目本身就是资产链路、确认数、账本、RPC、重组。
- Rust Protocol 项目可被表达成 block execution、state root、mempool、storage pipeline。
- Smart Contract 项目可被表达成 vault risk surface、share/asset accounting、attack tests。
- 这种形式既符合币圈视觉，又能展示真实工程能力。

