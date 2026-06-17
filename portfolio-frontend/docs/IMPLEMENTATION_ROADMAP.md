# 实现路线与验收标准

## 1. 阶段规划

### Phase 1：静态前端 MVP

目标：完成可浏览、可演示的三项目 dashboard。

范围：

- React + Vite + TypeScript 初始化。
- Tailwind token system。
- Dashboard Page。
- 三个 Project Detail Page。
- Compare Page。
- Interview Mode Page。
- 静态数据来自 `data/projects.ts`。

验收：

- 本地可启动。
- 桌面和移动端无明显溢出。
- 每个项目至少有：岗位、技术栈、验证状态、风险、修复路线。
- HR View 和 Senior Dev View 能切换。

### Phase 2：交互增强

目标：让前端更像币圈产品和面试工具。

范围：

- 风险过滤。
- 时间预算切换。
- 架构图节点 hover。
- Roadmap 图表。
- Role Fit Radar。
- Demo Wallet 状态切换。

验收：

- 所有控件有真实 UI 状态。
- 图表有文字摘要。
- 动效不影响阅读。

### Phase 3：内容自动化

目标：减少手动维护。

范围：

- 从 Markdown 文档构建 JSON 数据。
- 读取验证命令结果。
- 把 review findings 转成风险列表。

验收：

- 文档更新后前端数据能同步。
- 风险列表可追踪 open/mitigated 状态。

### Phase 4：可选 Web3 集成

目标：如果需要更强 Web3 native 观感，引入真实 wallet。

范围：

- RainbowKit + wagmi + viem。
- read-only chain status。
- 不做交易签名。

验收：

- wallet connect 不影响核心展示。
- 未连接状态也能完整浏览。

## 2. 首版页面优先级

P0：

1. Dashboard。
2. Web3 Backend Project Detail。
3. Compare Page。
4. Interview Mode。

P1：

1. Rust Protocol Detail。
2. Smart Contract Detail。
3. Roadmap chart。
4. Risk filters。

P2：

1. Demo Wallet。
2. Animated architecture path。
3. Markdown auto-ingestion。

## 3. 实现前检查清单

开始写代码前确认：

- 是否采用 React + Vite + TypeScript。
- 是否只做静态展示，不接真实钱包。
- 是否允许使用 lucide、framer-motion、recharts。
- 是否需要中英文双语。
- 是否部署到 Vercel / GitHub Pages / 本地演示。

默认假设：

- 使用 React + Vite + TypeScript。
- 首版静态展示。
- 中文主文案，英文岗位关键词保留。
- 不接真实钱包。
- 本地演示优先。

## 4. 验收标准

视觉：

- 暗色金融 dashboard 观感明确。
- 黄色主行动色使用克制。
- 不出现普通模板化个人主页。
- 不使用大面积紫色渐变。
- 移动端不挤压、不遮挡。

内容：

- 三个项目定位清晰。
- Web3 Backend 明确是主项目。
- Rust Protocol 是技术深度项目。
- Smart Contract 是安全意识补充项目。
- 风险和限制清楚展示。

交互：

- HR / Senior Dev 模式有效。
- 风险过滤有效。
- 项目切换有效。
- 面试模式能按时间预算显示内容。

工程：

- `npm install` 后可启动。
- `npm run build` 通过。
- 组件拆分清晰。
- 静态数据集中维护。

## 5. 后续可加入的高级功能

- 面试讲稿导出 Markdown。
- 项目风险燃尽图。
- 代码文件 deep link。
- 测试结果快照上传。
- Mermaid 架构图自动渲染。
- Resume bullet generator。
- GitHub commit/status badge。

## 6. 不做事项

首版不做：

- 用户登录。
- 真实交易。
- 真实钱包签名。
- 后端 API。
- CMS。
- 博客系统。
- 动态行情。

理由：

这个前端的目标是求职展示和面试讲解，不是构建一个真实交易产品。首版应把三项目讲清楚，而不是引入无关复杂度。

