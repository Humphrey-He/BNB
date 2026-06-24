# BNB 发布检查清单

## 1. 版本定位

当前建议版本：`v0.1.0-beta`

适用范围：

- 项目演示
- 测试环境部署
- 面试展示
- 小范围功能联调

不建议直接视为正式生产版，除非完成本文档中的真实链路验收项。

---

## 2. 代码与仓库检查

发布前确认：

- `git status` 干净，没有未提交的业务代码
- 当前分支是预期发布分支
- 关键改动已经提交：
  - 提现链路修复
  - 合约依赖恢复
  - 部署文档与运维模板
- 发布文档已补齐：
  - `INTERVIEW_TALK_TRACK.md`
  - `RELEASE_CHECKLIST.md`

建议执行：

```bash
git status
git log --oneline -10
git tag --list
```

---

## 3. Go 后端检查

确认以下服务代码可以正常构建或测试通过：

- `api-server`
- `scanner`
- `parser`
- `confirm-worker`
- `ledger-service`
- `withdrawal-worker`
- `broadcaster`
- `rpc-health`

建议执行：

```bash
go test ./internal/... ./cmd/...
```

发布通过标准：

- Go 测试通过
- 没有关键编译错误
- 新增提现链路修复已纳入测试通过版本

---

## 4. 合约工程检查

确认合约依赖完整，Foundry 可以编译和测试：

- `openzeppelin-contracts`
- `forge-std`

建议执行：

```bash
forge test
```

发布通过标准：

- 合约编译通过
- 测试通过
- `remappings.txt` 与 `foundry.lock` 状态一致

---

## 5. 配置与依赖检查

确认基础依赖已就绪：

- PostgreSQL 可连接
- NATS 可连接
- Redis 如有依赖则可连接
- RPC provider 已配置
- 热钱包地址和 signer 私钥在目标环境正确配置

重点环境变量检查：

- `POSTGRES_HOST`
- `POSTGRES_PORT`
- `NATS_URL`
- `WITHDRAWAL_SIGNER_PRIVATE_KEY`
- `WITHDRAWAL_HOT_WALLET_ADDRESS`
- `API_AUTH_TOKEN`

发布通过标准：

- 关键依赖存在
- 配置不为空
- 没有把密码、私钥、token 写入仓库

---

## 6. 充值链路验收

目标链路：

`scanner -> parser -> confirm-worker -> ledger-service`

至少确认以下项目：

- scanner 能持续推进 `scan_checkpoints`
- parser 能写入 `deposits`
- confirm-worker 能推进确认数
- ledger-service 能在充值确认后正确入账
- `balances` 与 `ledger_entries` 一致

建议检查：

- `deposits.status`
- `ledger_entries`
- `balances.available_balance`
- `scan_checkpoints.last_scanned_block`

发布通过标准：

- 至少完成一笔真实或测试链充值闭环
- 没有重复入账
- 没有确认后状态卡死

---

## 7. 提现链路验收

目标链路：

`CreateWithdrawal -> risk_checking -> approved -> signing -> broadcasting -> broadcasted -> confirmed/failed`

至少确认以下项目：

- `CreateWithdrawal` 严格校验生效
- 提现创建后余额先冻结
- 审批通过后走统一事件路径
- broadcaster 能正确分配 nonce
- 广播成功后写入 tx hash
- receipt 成功后正确结算
- receipt reverted 或广播失败时能补偿

建议检查：

- `withdrawals.status`
- `withdrawals.tx_hash`
- `nonce_allocations`
- `balances.available_balance`
- `balances.frozen_balance`
- `ledger_entries`

发布通过标准：

- 至少完成一笔真实测试链提现闭环
- 没有重复广播
- 没有错误 nonce 分配
- 成功和失败两类账务结果都正确

---

## 8. RPC 与健康检查

确认链路健康接口不是静态假数据：

- `GetChainStatus` 能返回真实 `latest_block`
- 返回 `scan_lag`
- 返回 `rpc_healthy`
- 返回 `provider_count`
- RPC failover 路径可用

建议检查接口：

- `GET /api/v1/chain-status`
- `cmd/rpc-health`

发布通过标准：

- 能区分扫描延迟和 RPC 问题
- 单一 provider 异常不会立即让服务整体失效

---

## 9. 部署与运行检查

确认部署入口和运行方式明确：

- `api-server` 有固定启动方式
- `withdrawal-worker` 有固定启动方式
- `broadcaster` 有固定启动方式
- scanner / parser / confirm-worker / ledger-service 启动顺序清晰
- systemd 或部署脚本可复用

建议至少验证：

- 服务能启动
- 服务能重启
- 服务日志可读
- 健康检查接口正常

发布通过标准：

- 不是“本地手工跑起来一次”，而是可重复启动
- 部署文档足够让第二个人接手

---

## 10. Beta 发布结论

满足以下条件时，可以发布 `v0.1.0-beta`：

- Go 测试通过
- 合约测试通过
- 核心服务入口明确
- 充值链路主流程可运行
- 提现链路主流程可运行
- 提现关键修复已合入
- 发布文档已沉淀

如果要继续升级到正式版，建议追加完成：

- 真实测试链充值闭环留档
- 真实测试链提现闭环留档
- 失败补偿演练
- 部署自动化全链路验收
- Release note 与运维回滚方案

---

## 11. 本次版本建议话术

`v0.1.0-beta` 表示：

- 核心架构和关键链路已经成型
- 具备演示、联调和测试部署价值
- 提现和广播能力已进入可验证阶段
- 仍保留正式生产前的验收和加固空间
