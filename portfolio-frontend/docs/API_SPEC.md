# API 接口文档

本文档定义前端 `BNB Web3 Career Terminal` 与后端对接所需的 RESTful API 接口。

---

## 1. 项目接口

### GET `/api/projects`

获取所有项目列表（Dashboard 用）。

**响应**

```json
{
  "projects": [
    {
      "id": "web3-backend",
      "name": "Multi-chain Asset Platform",
      "shortName": "Web3 Backend",
      "role": "Web3 / Blockchain Backend Engineer",
      "tagline": "Go-based asset backend...",
      "repo": "web3-blockchain-backend-engineer",
      "status": "Flagship",
      "verification": "Blocked",
      "stack": ["Go", "PostgreSQL", "NATS", "Ethereum RPC", "Ledger"],
      "metrics": {
        "marketFit": 92,
        "depth": 84,
        "testReadiness": 42,
        "riskControl": 68
      }
    }
  ]
}
```

---

### GET `/api/projects/:id`

获取单个项目完整详情（Project Detail 页用）。

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 项目 ID：`web3-backend`、`protocol-rust`、`smart-contract` |

**响应**

```json
{
  "id": "web3-backend",
  "name": "Multi-chain Asset Platform",
  "shortName": "Web3 Backend",
  "role": "Web3 / Blockchain Backend Engineer",
  "tagline": "Go-based asset backend covering RPC gateway...",
  "repo": "web3-blockchain-backend-engineer",
  "status": "Flagship",
  "verification": "Blocked",
  "stack": ["Go", "PostgreSQL", "NATS", "Ethereum RPC", "Ledger"],
  "metrics": {
    "marketFit": 92,
    "depth": 84,
    "testReadiness": 42,
    "riskControl": 68
  },
  "pipeline": [
    "RPC Providers", "Scanner", "Raw Events", "Parser",
    "Deposits", "Confirm Worker", "Ledger Service", "Balances"
  ],
  "wins": [
    "Covers the real deposit lifecycle from on-chain logs to account balance.",
    "Includes scanner, parser, confirmation state machine, ledger and reorg modules."
  ],
  "findings": [
    {
      "priority": "P0",
      "title": "Go toolchain blocks verification",
      "impact": "go.mod requires Go 1.25.0...",
      "fix": "Align local and CI toolchain..."
    }
  ],
  "hrPitch": "The flagship project: a Web3 backend asset platform...",
  "seniorPitch": "The strongest engineering story is the chain-to-ledger pipeline...",
  "nextMoves": [
    "Make go test ./... pass on a reproducible toolchain.",
    "Add outbox/inbox and ledger replay checks."
  ]
}
```

---

## 2. 指标接口

### GET `/api/metrics/matrix`

获取项目健康矩阵数据（Dashboard 散点图用）。

**响应**

```json
{
  "projects": [
    {
      "id": "web3-backend",
      "shortName": "Web3 Backend",
      "marketFit": 92,
      "depth": 84
    },
    {
      "id": "protocol-rust",
      "shortName": "Rust Protocol",
      "marketFit": 78,
      "depth": 86
    },
    {
      "id": "smart-contract",
      "shortName": "Smart Contract",
      "marketFit": 58,
      "depth": 62
    }
  ]
}
```

---

## 3. 风险接口

### GET `/api/findings`

获取所有风险项，可按严重度过滤。

**查询参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| severity | string | 可选：`P0`、`P1`、`P2`、`all` |
| projectId | string | 可选：按项目过滤 |

**响应**

```json
{
  "findings": [
    {
      "id": "f001",
      "projectId": "web3-backend",
      "priority": "P0",
      "title": "Go toolchain blocks verification",
      "impact": "go.mod requires Go 1.25.0...",
      "fix": "Align local and CI toolchain..."
    }
  ]
}
```

---

## 4. 路线图接口

### GET `/api/roadmap`

获取 30 天升级路线图。

**响应**

```json
{
  "roadmap": [
    ["Week 1", "Web3 Backend", "Fix Go toolchain, add end-to-end pipeline tests and outbox."],
    ["Week 2", "Rust Protocol", "Add real signatures, format gate and block import integration."],
    ["Week 3", "Smart Contract", "Add Foundry unit, fuzz, invariant and attack PoC tests."],
    ["Week 4", "Portfolio", "Polish README, diagrams, interview scripts and frontend evidence."]
  ]
}
```

---

## 5. 对比接口

### GET `/api/compare`

获取项目对比数据。

**响应**

```json
{
  "projects": [
    {
      "id": "web3-backend",
      "shortName": "Web3 Backend",
      "role": "Web3 / Blockchain Backend Engineer",
      "marketFit": 92,
      "testReadiness": 42,
      "hrPitch": "The flagship project: a Web3 backend asset platform..."
    },
    {
      "id": "protocol-rust",
      "shortName": "Rust Protocol",
      "role": "Protocol Engineer / Rust Blockchain Engineer",
      "marketFit": 78,
      "testReadiness": 76,
      "hrPitch": "A technical depth project that shows Rust..."
    },
    {
      "id": "smart-contract",
      "shortName": "Smart Contract",
      "role": "Smart Contract Engineer",
      "marketFit": 58,
      "testReadiness": 12,
      "hrPitch": "A supporting project that proves Solidity exposure..."
    }
  ],
  "recommendation": "Lead with the Web3 backend project..."
}
```

---

## 6. 面试模式接口

### GET `/api/interview/:duration`

获取指定时长的面试内容。

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| duration | string | 时长：`3`、`8`、`20`（分钟） |

**查询参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| mode | string | `hr` 或 `senior` |

**响应（3分钟）**

```json
{
  "duration": "3",
  "mode": "senior",
  "steps": [
    "Backend engineer moving into Web3 infrastructure.",
    "Flagship project: Go multi-chain asset backend.",
    "Depth project: Rust execution and validation engine.",
    "Supporting project: Solidity vault security awareness."
  ]
}
```

**响应（8分钟）**

```json
{
  "duration": "8",
  "mode": "senior",
  "steps": [
    "Explain the chain-to-ledger lifecycle in the Web3 backend.",
    "Show Rust protocol fundamentals: state, nonce, roots and validation.",
    "Explain why Solidity is a supporting track until tests are complete.",
    "Close with the 30-day roadmap and honest verification status."
  ]
}
```

**响应（20分钟）**

```json
{
  "duration": "20",
  "mode": "senior",
  "steps": [
    "Walk through scanner -> parser -> confirm worker -> ledger service.",
    "Discuss outbox, durable messaging, reorg compensation and ledger replay.",
    "Deep dive Rust executor, state root and storage limitations.",
    "Review Solidity vault risk surface and missing Foundry tests.",
    "Map all projects to target roles and interview questions."
  ]
}
```

---

## 错误响应格式

所有接口错误时返回：

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Project not found"
  }
}
```

| HTTP 状态码 | 说明 |
|-------------|------|
| 200 | 成功 |
| 400 | 参数错误 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

---

## 备注

- 当前前端为纯静态应用（数据硬编码在 App.tsx），本文档定义的是**待实现的 API 接口**
- 建议后端使用 Node.js/Express 或 Go/Gin 搭建
- 可先返回静态 JSON 数据，后续替换为真实数据库查询