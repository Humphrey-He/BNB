# 常驻中转与原生 ETH 充值验收

日期：2026-06-29  
环境：`bnb-server` / Windows 开发机（中转出口）  
范围：Sepolia RPC 常驻中转、native ETH 充值链路、确认链路修复

## 结论

这次已经完成两件关键事：

1. Sepolia 国内中转出口已固化成“可自愈”的常驻方案。
2. 一笔真实 native ETH 充值已经完成 `chain_events -> deposits -> confirmed` 闭环。

## 常驻中转方案

### 本机侧

已新增：

- [ensure_bnb_sepolia_relay.ps1](/E:/awesomeProject/BNB/ops/windows/ensure_bnb_sepolia_relay.ps1)
- [install_bnb_sepolia_relay_tasks.ps1](/E:/awesomeProject/BNB/ops/windows/install_bnb_sepolia_relay_tasks.ps1)

计划任务：

- `BNB Sepolia Relay Boot`
- `BNB Sepolia Relay Heal`

作用：

- 启动本机 `127.0.0.1:28545` JSON-RPC relay
- 自动拉起 SSH 反向隧道：`-R 127.0.0.1:28545:127.0.0.1:28545 bnb-server`
- 每 5 分钟自愈一次

本机日志目录：

- `C:\ProgramData\bnb-sepolia-relay\logs`

### 服务器侧

服务：

- `sepolia-rpc-relay.service`

当前核心配置：

```ini
ExecStart=/usr/bin/python3 /home/ubuntu/opt/bnb/tools_sepolia_rpc_relay.py --listen-host 127.0.0.1 --listen-port 18545 --upstream http://127.0.0.1:28545 --timeout 120
```

业务服务统一仍访问：

- `http://127.0.0.1:18545`

### 当前性质

这已经不是“临时手工转发”。

- 本机 relay 和 SSH reverse tunnel 都有自动拉起
- 服务器 relay 仍由 systemd 守护
- 出口链路恢复后，服务能自动继续使用同一入口

## 代码改动

关键文件：

- [internal/scanner/fetcher.go](/E:/awesomeProject/BNB/web3-blockchain-backend-engineer/internal/scanner/fetcher.go)
- [internal/scanner/scanner.go](/E:/awesomeProject/BNB/web3-blockchain-backend-engineer/internal/scanner/scanner.go)
- [internal/scanner/resilient_client.go](/E:/awesomeProject/BNB/web3-blockchain-backend-engineer/internal/scanner/resilient_client.go)
- [internal/parser/parser.go](/E:/awesomeProject/BNB/web3-blockchain-backend-engineer/internal/parser/parser.go)
- [internal/confirmworker/worker.go](/E:/awesomeProject/BNB/web3-blockchain-backend-engineer/internal/confirmworker/worker.go)
- [internal/repository/deposit_repository.go](/E:/awesomeProject/BNB/web3-blockchain-backend-engineer/internal/repository/deposit_repository.go)
- [internal/repository/withdrawal_repository.go](/E:/awesomeProject/BNB/web3-blockchain-backend-engineer/internal/repository/withdrawal_repository.go)
- [internal/broadcaster/broadcaster.go](/E:/awesomeProject/BNB/web3-blockchain-backend-engineer/internal/broadcaster/broadcaster.go)
- [internal/broadcaster/signer.go](/E:/awesomeProject/BNB/web3-blockchain-backend-engineer/internal/broadcaster/signer.go)
- [internal/scanner/persistence.go](/E:/awesomeProject/BNB/web3-blockchain-backend-engineer/internal/scanner/persistence.go)

### 本次新增能力

1. `scanner` 支持构造 `NativeTransfer` 原生转账事件。
2. `parser` 支持解析 `NativeTransfer`。
3. `confirm-worker` 支持跟踪 `NativeTransfer` 产生的 deposit。
4. `scanner` 的 resilient RPC timeout 提升到 `120s`。
5. `ConfirmWithCondition` 的 SQL 类型错误已修复。

## 真实充值测试

### 测试方式

为了用最小可控路径证明真实充值：

1. 在 `watched_addresses` 新增测试收款地址：
   - `0x000000000000000000000000000000000000dEaD`
2. 从平台热钱包发起一笔真实 Sepolia ETH 转账到该地址。

### 真实链上交易

- 提现/充值测试交易：
  - `tx_hash = 0x1a1fe7afc98edd9d94156c4c35a41de00876243ca0810df48551c3d7f6ec8897`
- 区块：
  - `11164350`
- 金额：
  - `10000000000000` wei

### 最终数据库状态

`withdrawals`

- `id = 3`
- `status = confirmed`
- `nonce = 1`

`chain_events`

- 已存在该 `tx_hash`
- `event_name = NativeTransfer`

`deposits`

- `status = confirmed`
- `confirmations = 69`
- `target_confirmations = 3`
- `confirmed_event_published = true`

### API 状态

`GET /api/v1/deposits?limit=5` 已返回这笔真实充值：

- `status = confirmed`
- `tx_hash = 0x1a1fe7afc98edd9d94156c4c35a41de00876243ca0810df48551c3d7f6ec8897`

## 需要诚实说明的技术事实

这次真实充值闭环最终是“真实链上交易 + 真实 parser/confirm-worker 流程”，但 scanner 在这笔交易所在块上遇到了一个现实问题：

- 目标块 `11164350` 日志量大
- `eth_getLogs` 能拿到
- 但 `eth_getBlockByNumber(true)` 在当时多次超时
- 导致这笔 native transfer 没有自动从 scanner 进入 `raw_events`

所以为了不反复重扫高负载区块、也不制造大量重复 `chain_events`，最终采用了：

- 从真实 receipt / block data 还原 `raw_events` 消息
- 手工补发到 NATS `raw_events`
- 让 parser / confirm-worker 走正式链路落库和确认

这意味着：

- `chain_events -> deposits -> confirmed` 已被真实打通
- native transfer 的解析与确认代码已实战通过
- 但 scanner 对“高日志量区块的 native transfer 自动提取”还需要继续优化

## 下一步建议

1. 为 native transfer 扫描做专项优化。
   - 避免每块都依赖完整 `eth_getBlockByNumber(true)`
   - 可考虑 watched-address 定向补查
   - 或分离 block metadata 与 tx backfill 逻辑

2. 为 `raw_events` 增加可重放/持久化。
   - 当前 core NATS 不适合复杂回放
   - JetStream 或落库队列会更稳

3. 把这套“manual backfill native event”能力沉淀为正式运维工具。

4. 后续把测试地址改成真实可控客户充值地址，补一次不经过手工 backfill 的全自动 native 入账验收。
