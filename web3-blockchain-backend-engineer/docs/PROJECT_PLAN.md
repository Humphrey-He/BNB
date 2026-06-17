# multi-chain-asset-platform 项目规划

## 项目概述

**项目名称**：multi-chain-asset-platform
**项目定位**：多链资产监听、账本余额、充值提现、交易历史 API 和监控告警系统
**目标用户**：Web3 钱包平台、交易所、Custody 服务商

## 技术栈

| 类别 | 技术选型 |
|------|----------|
| 主语言 | Go |
| Web 框架 | Gin / chi |
| 区块链交互 | go-ethereum |
| 数据库 | PostgreSQL |
| 缓存 | Redis |
| 消息队列 | NATS / Kafka |
| Python 辅助 | 数据校验、余额修复、运营脚本 |
| 容器化 | Docker Compose |
| 监控 | Prometheus / Grafana |

## 服务拆分

| 服务 | 职责 |
|------|------|
| `scanner` | 扫描 EVM 区块，拉取 logs |
| `parser` | 解析 ERC-20 事件，标准化交易 |
| `confirm-worker` | 处理确认数，实现状态流转 |
| `ledger-service` | 维护账本流水和用户资产余额 |
| `withdrawal-service` | 创建提现申请、冻结余额、推进提现状态 |
| `broadcaster` | 构造和广播交易，管理 nonce/gas/retry |
| `rpc-gateway` | RPC 节点容灾、限流、重试、熔断 |
| `api-server` | 提供 REST API 查询 |
| `reorg-detector` | 检测链重组并触发补偿 |
| `metrics` | 暴露 Prometheus 监控指标 |

## 开发阶段规划

### 第一阶段：核心数据模型与本地环境 (第 1 周)

- [ ] 设计并创建 PostgreSQL 表结构
  - [ ] `chains` - 链配置表
  - [ ] `rpc_providers` - RPC provider 配置表
  - [ ] `tokens` - Token 信息表
  - [ ] `watched_addresses` - 监听地址表
  - [ ] `blocks` - 区块记录表
  - [ ] `chain_events` - 标准化事件表
  - [ ] `deposits` - 充值记录表
  - [ ] `withdrawals` - 提现记录表
  - [ ] `ledger_entries` - 账本流水表
  - [ ] `balances` - 余额表
  - [ ] `nonce_allocations` - nonce 分配表
  - [ ] `scan_checkpoints` - 扫描进度表
- [ ] 实现数据库迁移脚本
- [ ] 编写基础 CRUD DAO 层
- [ ] 编写 Docker Compose：PostgreSQL、Redis、NATS/Kafka
- [ ] 加入 dev seed 数据：链、token、监听地址

### 第二阶段：区块扫描与事件解析 (第 2-3 周)

- [ ] 实现 `rpc-gateway`
  - [ ] 多 provider 配置
  - [ ] 超时、重试、熔断
  - [ ] 批量 `eth_getLogs` 窗口配置
- [ ] 实现 `scanner` 服务
  - [ ] 连接 Ethereum/BSC RPC 节点
  - [ ] 区块遍历与最新高度追踪
  - [ ] 获取区块 logs（ERC-20 Transfer 事件）
  - [ ] 保存 block hash 和 parent hash
  - [ ] 扫描进度持久化
- [ ] 实现 `parser` 服务
  - [ ] ERC-20 Transfer 事件解析
  - [ ] 充值识别逻辑
  - [ ] 事件去重与幂等写入
- [ ] 实现 `reorg-detector` 服务
  - [ ] 链重组检测（parent hash 不匹配）
  - [ ] 触发补偿任务

### 第三阶段：确认数处理与状态流转 (第 4 周)

- [ ] 实现 `confirm-worker` 服务
  - [ ] 待确认交易队列管理
  - [ ] 确认数递增逻辑
  - [ ] 状态流转：`detected` → `pending_confirmation` → `confirmed`
  - [ ] 超时处理与失败重试
- [ ] 实现 `ledger-service`
  - [ ] 充值确认后生成 credit 账本流水
  - [ ] 从账本流水更新 `balances`
  - [ ] 支持可用余额和冻结余额
  - [ ] 支持余额回放和校验脚本

### 第四阶段：提现与交易广播 (第 5 周)

- [ ] 实现 `withdrawal-service`
  - [ ] 创建提现申请
  - [ ] 校验余额、token、地址、限额
  - [ ] 冻结余额
  - [ ] 发送广播任务
- [ ] 实现 `broadcaster`
  - [ ] nonce 分配
  - [ ] gas 估算
  - [ ] EVM 交易构造
  - [ ] 广播与 receipt 查询
  - [ ] replacement transaction 策略
- [ ] 明确 signer 边界
  - [ ] dev 环境使用本地私钥
  - [ ] 生产设计说明 HSM/MPC/独立 signer

### 第五阶段：API 服务 (第 5 周)

- [ ] 实现 `api-server` 服务
  - [ ] 查询地址余额 API
  - [ ] 查询交易历史 API
  - [ ] 查询充值记录 API
  - [ ] 创建和查询提现 API
  - [ ] 查询链扫描状态 API
  - [ ] 分页与过滤支持
- [ ] 实现幂等写入
  - [ ] 唯一约束防止重复
  - [ ] 幂等键设计

### 第六阶段：异步任务与补偿机制 (第 6 周)

- [ ] 集成 NATS / Kafka
  - [ ] 区块扫描任务入队
  - [ ] 确认数处理任务入队
  - [ ] 补偿任务入队
  - [ ] 提现广播任务入队
- [ ] 实现消费者 worker
  - [ ] 并发处理控制
  - [ ] 失败重试与死信队列
- [ ] 实现链重组补偿闭环
  - [ ] 标记 orphaned 区块和事件
  - [ ] 反向账本流水修正余额
  - [ ] 从分叉点重新扫描

### 第七阶段：监控、测试与部署 (第 7 周)

- [ ] 实现 `metrics` 服务
  - [ ] 区块扫描延迟指标
  - [ ] 处理速度指标
  - [ ] 错误率指标
  - [ ] 队列积压指标
  - [ ] 提现广播成功率
  - [ ] nonce 冲突数
  - [ ] RPC provider 错误率
- [ ] 配置 Prometheus 抓取
- [ ] 配置 Grafana Dashboard
- [ ] 编写 Docker Compose 一键启动
- [ ] 编写集成测试

### 第八阶段：文档与交付 (第 8 周)

- [ ] 编写 README
- [ ] 编写架构文档
- [ ] 编写部署文档
- [ ] 编写数据一致性说明
- [ ] 编写故障处理 RUNBOOK
- [ ] 更新简历项目描述

## 交易状态流转

```
detected → pending_confirmation → confirmed
    ↓            ↓                    ↓
    ↓         orphaned            orphaned
    ↓            ↓                    ↓
    └──────── failed ←←←←←←←←←←←←←
```

| 状态 | 说明 |
|------|------|
| `detected` | 检测到充值事件 |
| `pending_confirmation` | 等待确认数达标 |
| `confirmed` | 确认数达标，充值成功 |
| `orphaned` | 被链重组回滚 |
| `failed` | 处理失败 |

## 链重组处理策略

1. **检测**：对比本地区块 hash 与链上区块 hash
2. **标记**：将受影响区块状态标记为 `orphaned`
3. **回滚**：修正 deposits、balances 等业务数据
4. **重扫**：从分叉点重新扫描新链

## 面试要点

- [ ] 能讲清 ERC-20 充值监听原理
- [ ] 能讲清幂等性保证机制
- [ ] 能讲清确认数的作用
- [ ] 能讲清链重组时的数据补偿方案
- [ ] 能讲清余额一致性保证
- [ ] 能讲清交易广播与 nonce 处理
- [ ] 能讲清 gas replacement transaction
- [ ] 能讲清为什么用账本流水维护余额
- [ ] 能讲清 RPC provider failover 和扫描窗口控制
- [ ] 能讲清死信队列和人工补偿入口

## 质量标准

- [ ] 充值重复消息不会重复入账
- [ ] 链重组后余额能通过反向账本修正
- [ ] 提现请求重复提交不会重复冻结
- [ ] nonce 分配在并发提现下不冲突
- [ ] RPC provider 故障时服务能切换到备用 provider
- [ ] 集成测试覆盖充值、确认、reorg、提现、广播失败
- [ ] Prometheus 暴露核心指标
