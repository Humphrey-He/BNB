# CMD 与服务入口落地清单

更新时间：2026-06-23

## 目的

这份文档只回答一个非常实际的问题：

当前 BNB 项目还缺哪些 `cmd/` 和运行入口，补完之后项目才算“真正能跑起来”。

这里强调的是：

- 以当前仓库真实代码为准
- 区分“内部 package 已经存在”和“可执行服务入口已存在”
- 以最小可运行闭环优先，不先追求完美架构

## 当前现状

当前 `web3-blockchain-backend-engineer/cmd` 下已有：

- `api-server`
- `scanner`
- `parser`
- `confirm-worker`
- `rpc-health`

当前 `internal/` 下已经有、但还没有独立 `cmd/` 入口的服务：

- `internal/ledgerservice`
- `internal/withdrawalservice`
- `internal/broadcaster`

当前部署与 workflow 也只明确构建和部署了：

- Go `api-server`
- Rust `chain-node`
- 前端

也就是说，当前项目最大的问题不是“没有业务逻辑”，而是：

- 有些关键服务代码已经写了
- 但没有独立二进制入口
- 没有纳入 systemd / workflow / deploy 脚本

## 先补哪些入口，项目就能真正跑起来

建议按下面顺序补。

### 第一优先级：把充值闭环真正跑起来

这一组是最先必须完整托管的服务：

1. `cmd/scanner`
2. `cmd/parser`
3. `cmd/confirm-worker`
4. `cmd/ledger-service`

原因：

- `scanner` 负责扫链
- `parser` 负责写 `chain_events / deposits`
- `confirm-worker` 负责推进确认数
- `ledger-service` 负责把 confirmed deposit 真正入账到 `balances / ledger_entries`

当前缺口：

- `scanner / parser / confirm-worker` 已经有 `cmd`
- `ledger-service` 只有 `internal/ledgerservice`，没有 `cmd/ledger-service`

结论：

要让“充值闭环”完整跑起来，第一件事就是补：

- `web3-blockchain-backend-engineer/cmd/ledger-service/main.go`

### 第二优先级：把提现审批骨架真正跑起来

这一组是提现链路最小服务入口：

1. `cmd/withdrawal-worker`
2. `cmd/broadcaster`

原因：

- `withdrawal-worker` 负责风险检查、manual review、approved -> signing 推进
- `broadcaster` 负责广播、retry、receipt、失败补偿

当前缺口：

- `internal/withdrawalservice` 已存在
- `internal/broadcaster` 已存在
- 但都没有独立 `cmd`

结论：

要让“提现骨架链路”真正跑起来，必须新增：

- `web3-blockchain-backend-engineer/cmd/withdrawal-worker/main.go`
- `web3-blockchain-backend-engineer/cmd/broadcaster/main.go`

### 第三优先级：把部署入口补齐

就算 `cmd` 补完了，如果不纳入构建和 systemd，线上还是跑不起来。

必须补的部署层内容：

1. `.github/workflows/deploy-bnb.yml`
   - 构建 `scanner`
   - 构建 `parser`
   - 构建 `confirm-worker`
   - 构建 `ledger-service`
   - 构建 `withdrawal-worker`
   - 构建 `broadcaster`

2. `scripts/deploy_self_hosted.sh`
   - 安装上述所有 Go worker 二进制
   - 写对应 `.env`
   - 写 systemd unit
   - `enable + restart`

3. `scripts/verify.sh`
   - 增加 worker 进程健康检查
   - 检查 NATS / API / scanner / parser / confirm-worker / ledger-service / broadcaster

## 推荐的最小运行分层

### A. 先跑“充值主链路”

最小需要：

- PostgreSQL
- NATS
- `api-server`
- `scanner`
- `parser`
- `confirm-worker`
- `ledger-service`

这时你就能跑通：

`链上事件 -> deposits -> confirmed -> balances`

### B. 再跑“提现骨架链路”

再补：

- `withdrawal-worker`
- `broadcaster`

这时你就能跑通：

`创建提现 -> freeze -> risk -> approve -> broadcast skeleton -> receipt / compensation`

但这里仍然只是“骨架可运行”，不等于真实出金已经可用。

### C. 最后才是“真实出金链路”

在 `cmd/withdrawal-worker` 和 `cmd/broadcaster` 运行起来之后，再继续补：

- 真实 EVM signer
- 真实 raw transaction 构造
- 真实 nonce 分配
- replacement / retry / rollback 完整策略

## 建议新增的 cmd 列表

### 现在就应该新增

1. `cmd/ledger-service/main.go`
2. `cmd/withdrawal-worker/main.go`
3. `cmd/broadcaster/main.go`

### 之后可选新增

1. `cmd/reorg-detector/main.go`
2. `cmd/withdrawal-reconciler/main.go`
3. `cmd/nonce-reconciler/main.go`

说明：

- `reorg-detector` 当前 `internal/reorgdetector` 在你的工作区里还不是主运行重点
- 当前第一目标不是把所有理想服务补齐，而是先让主业务闭环真正跑起来

## 每个入口建议注入的依赖

### `cmd/ledger-service`

依赖：

- `app.OpenPostgres`
- `app.ConnectNATS`
- `repository.NewLedgerEntryRepository`
- `repository.NewBalanceRepository`
- `repository.NewDepositRepository`
- `ledgerservice.NewLedgerService`

### `cmd/withdrawal-worker`

依赖：

- `app.OpenPostgres`
- `app.ConnectNATS`
- `repository.NewWithdrawalRepository`
- `repository.NewBalanceRepository`
- `withdrawalservice.NewWithdrawalWorker`

### `cmd/broadcaster`

依赖：

- `app.OpenPostgres`
- `app.ConnectNATS`
- `repository.NewWithdrawalRepository`
- `repository.NewNonceAllocationRepository`
- 真实或占位 `RPCClient`
- `broadcaster.NewBroadcaster`

注意：

- 这里真正的阻塞点不是 `cmd` 本身，而是 `broadcaster` 目前的 RPC / signer 仍是半成品
- 但先把入口补齐，至少能把提现链路从“代码存在”推进到“服务可启动”

## 环境变量建议

### 充值主链路

- `POSTGRES_HOST`
- `POSTGRES_PORT`
- `POSTGRES_DB`
- `POSTGRES_USER`
- `NATS_URL`
- `SCAN_CHAIN_DB_ID`
- `SCAN_BATCH_SIZE`
- `SCAN_RPC_URL`

### 提现链路

- `WITHDRAWAL_HOT_WALLET_ADDRESS`
- `WITHDRAWAL_MAX_AUTO_APPROVE_AMOUNT`
- `WITHDRAWAL_REQUIRE_WHITELIST`
- `WITHDRAWAL_ALLOWED_DESTINATIONS`
- `WITHDRAWAL_SIGNER_BACKEND`
- 由密钥管理系统或部署平台注入的 signer 配置
- `WITHDRAWAL_BROADCAST_MAX_RETRIES`

## 推荐实施顺序

### 第 1 步

补：

- `cmd/ledger-service/main.go`

目标：

- 把充值闭环补完整

### 第 2 步

补：

- `cmd/withdrawal-worker/main.go`
- `cmd/broadcaster/main.go`

目标：

- 把提现骨架链路补到可启动

### 第 3 步

改：

- `.github/workflows/deploy-bnb.yml`
- `scripts/deploy_self_hosted.sh`

目标：

- 让 worker 真正能上服务器

### 第 4 步

改：

- `internal/broadcaster`
- `internal/repository/nonce_allocation_repository.go`

目标：

- 把 broadcaster 从“伪广播”补成“真实广播”

## 最终建议

如果你的目标是“项目真正跑起来”，而不是继续停留在模块完成度层面，那么最短路径是：

1. 先补 `cmd/ledger-service`
2. 再补 `cmd/withdrawal-worker`
3. 再补 `cmd/broadcaster`
4. 然后把这三个服务纳入 workflow 和 deploy 脚本

做到这一步，你的项目才会从“有内部包代码”变成“有完整可运行服务拓扑”。
