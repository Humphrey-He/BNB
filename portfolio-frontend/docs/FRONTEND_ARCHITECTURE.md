# 前端产品架构文档

产品名：`BNB Web3 Career Terminal`

## 1. 产品形态

这是一个单页应用，但不是传统 landing page。整体形态是 dashboard + project deep dive：

- 首页：三项目组合仪表盘。
- 项目页：每个项目的架构、实现、风险、验证状态、岗位匹配。
- 对比页：三个项目从 HR / 资深开发角度对比。
- 面试模式：按 3 分钟、8 分钟、20 分钟生成讲解路径。

## 2. 推荐技术栈

首版推荐：

- React
- Vite
- TypeScript
- Tailwind CSS
- Framer Motion
- Recharts 或 Visx
- Lucide React
- Mermaid 渲染或自定义 SVG 流程图

可选扩展：

- RainbowKit + wagmi + viem：如果后续要加入真实 wallet connect。
- MDX：如果希望直接把项目 docs 变成页面内容。
- Zustand：如果交互状态复杂到需要全局 store。

不建议首版引入：

- Next.js：除非要 SEO、服务端渲染、博客系统或部署到 Vercel 做内容站。
- 真链上读写：当前目标是项目展示，不是 dApp。
- 复杂 3D：币圈审美不等于到处放 3D token。

## 3. 信息架构

```text
/
  Dashboard
    - Hero command center
    - Portfolio overview
    - Project health matrix
    - Role fit radar
    - Current blockers
    - 30-day upgrade roadmap

/projects/web3-backend
  - System pipeline
  - Implementation modules
  - Data model
  - Reliability risks
  - HR pitch
  - Senior dev deep dive

/projects/protocol-rust
  - Node architecture
  - Execution pipeline
  - State/root validation
  - Mempool/storage limitations
  - Test status
  - Next protocol milestones

/projects/smart-contract
  - Vault model
  - Contract modules
  - Risk surface
  - Test gap
  - Security roadmap
  - Solidity interview talking points

/compare
  - Project comparison table
  - Market fit matrix
  - Technical depth score
  - Offer strategy

/interview
  - 3-minute HR mode
  - 8-minute technical screen mode
  - 20-minute onsite deep dive mode
```

## 4. 组件架构

```text
src/
  app/
    App.tsx
    routes.tsx

  data/
    projects.ts
    metrics.ts
    findings.ts
    roadmap.ts

  components/
    layout/
      AppShell.tsx
      TopNav.tsx
      SideRail.tsx
      MobileNav.tsx

    primitives/
      Button.tsx
      IconButton.tsx
      Badge.tsx
      StatusDot.tsx
      MetricCard.tsx
      SectionHeader.tsx
      SegmentedControl.tsx

    crypto/
      ChainStatusBar.tsx
      WalletDemoButton.tsx
      VerificationBadge.tsx
      RiskSeverityPill.tsx
      TechStackTicker.tsx

    project/
      ProjectCard.tsx
      ProjectScorePanel.tsx
      ProjectTimeline.tsx
      ProjectArchitectureMap.tsx
      ReviewFindingList.tsx
      RoleFitPanel.tsx

    charts/
      RadarChart.tsx
      HealthMatrix.tsx
      RoadmapGantt.tsx
      RiskHeatmap.tsx

  pages/
    DashboardPage.tsx
    ProjectDetailPage.tsx
    ComparePage.tsx
    InterviewModePage.tsx

  styles/
    tokens.css
    globals.css
```

## 5. 数据模型

### Project

```ts
type Project = {
  id: "web3-backend" | "protocol-rust" | "smart-contract";
  name: string;
  subtitle: string;
  primaryRole: string;
  priority: 1 | 2 | 3;
  repoPath: string;
  status: "strong" | "promising" | "needs-tests";
  stack: string[];
  pitch: {
    hr: string;
    seniorDev: string;
  };
  metrics: {
    implementation: number;
    technicalDepth: number;
    marketFit: number;
    testReadiness: number;
    riskControl: number;
  };
  verification: {
    command: string;
    result: "pass" | "partial" | "blocked" | "missing";
    note: string;
  };
  risks: ReviewFinding[];
  roadmap: RoadmapItem[];
};
```

### ReviewFinding

```ts
type ReviewFinding = {
  id: string;
  priority: "P0" | "P1" | "P2";
  title: string;
  summary: string;
  file?: string;
  status: "open" | "mitigated" | "accepted";
  audience: "hr" | "senior-dev" | "both";
};
```

## 6. 页面数据来源

首版使用静态 TypeScript 数据，直接从当前项目评估文档整理：

- `BLOCKCHAIN_PROJECTS_IMPLEMENTATION_REVIEW.md`
- 三个项目的 `docs/PROJECT_PLAN.md`
- 三个项目的 `docs/ARCHITECTURE.md`
- 三个项目的实现文件和测试状态

后续可扩展为读取 Markdown/JSON：

- 构建时把 Markdown 转成 JSON。
- 显示最近一次验证命令和结果。
- 从 review findings 生成风险列表。

## 7. 交互状态

全局状态：

- `mode`: `hr` | `senior-dev`
- `selectedProject`
- `timeBudget`: `3min` | `8min` | `20min`
- `demoWalletConnected`: boolean
- `riskFilter`: `all` | `P0` | `P1` | `P2`

设计意义：

- HR mode：减少代码细节，突出岗位、业务场景、技术栈、结果。
- Senior Dev mode：展示风险、验证、架构和修复路线。
- Time budget：直接服务面试讲解。

## 8. 响应式策略

Desktop：

- 左侧项目导航 rail。
- 中间主内容。
- 右侧 context panel，展示当前项目关键指标和风险。

Tablet：

- 左侧 rail 收缩为 icon rail。
- 右侧 context panel 移到内容顶部或底部。

Mobile：

- 顶部 sticky nav。
- 项目切换使用 segmented control。
- 图表改成横向滚动或 compact cards。
- 架构图改成 stepper timeline。

## 9. 性能策略

- 首屏数据全部静态内联，不依赖 API。
- 图表组件按页面 lazy load。
- 代码高亮和 Mermaid 图按需加载。
- 动画限制在 opacity/transform，不做昂贵布局动画。
- 尊重 `prefers-reduced-motion`。

## 10. 可访问性

- 所有按钮和 tab 支持 keyboard focus。
- P0/P1/P2 不只靠颜色区分，同时显示文本。
- 图表必须有文字摘要。
- 色彩对比按深色金融产品标准处理。
- 移动端不出现横向正文溢出。

