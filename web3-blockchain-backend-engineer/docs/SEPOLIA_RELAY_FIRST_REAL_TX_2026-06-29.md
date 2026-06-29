# Sepolia 中转 RPC 与第一笔真实提现验收

日期：2026-06-29  
环境：`bnb-server` (`ubuntu@101.43.127.178`)  
范围：`scanner + broadcaster + sepolia relay + withdrawal flow`

## 这次完成了什么

1. 修复了 `broadcaster` 的链 ID 语义错误。
   - 之前签名时错误使用了数据库 `chains.id`
   - 现已改为使用真实 EVM `chains.chain_id`
   - 仍保留数据库主键作为 RPC provider / 业务侧关联键

2. 修复了 `scanner` 的 checkpoint 初始化容错。
   - `EnsureCheckpoint` 现在只在 `sql.ErrNoRows` 时创建新 checkpoint
   - 避免把其他数据库错误误判成“没有 checkpoint”

3. 为服务器切换了一条可控的国内中转出口。
   - 本机 Windows 起本地 relay：`127.0.0.1:28545`
   - 通过 SSH 反向隧道暴露到服务器：`127.0.0.1:28545`
   - 服务器本地 `sepolia-rpc-relay.service` 再转发到该隧道
   - 业务服务统一继续访问 `http://127.0.0.1:18545`

4. 成功打出第一笔真实 Sepolia 提现交易并确认上链。

## 本次落地的代码改动

- [internal/broadcaster/broadcaster.go](/E:/awesomeProject/BNB/web3-blockchain-backend-engineer/internal/broadcaster/broadcaster.go)
- [internal/broadcaster/signer.go](/E:/awesomeProject/BNB/web3-blockchain-backend-engineer/internal/broadcaster/signer.go)
- [internal/scanner/persistence.go](/E:/awesomeProject/BNB/web3-blockchain-backend-engineer/internal/scanner/persistence.go)
- [internal/repository/withdrawal_repository.go](/E:/awesomeProject/BNB/web3-blockchain-backend-engineer/internal/repository/withdrawal_repository.go)

补充说明：

- `withdrawal_repository.go` 额外修了 `nonce=0` 不落库的问题
- 这次第一笔成功交易正好使用了 `nonce=0`

## 服务器侧改动

### 1. relay 服务

`/etc/systemd/system/sepolia-rpc-relay.service`

当前核心配置：

```ini
ExecStart=/usr/bin/python3 /home/ubuntu/opt/bnb/tools_sepolia_rpc_relay.py --listen-host 127.0.0.1 --listen-port 18545 --upstream http://127.0.0.1:28545 --timeout 30
```

### 2. rpc_providers

数据库已切换为：

```text
name = Local Sepolia Relay
url  = http://127.0.0.1:18545
```

### 3. 已重启服务

- `sepolia-rpc-relay.service`
- `asset-platform-scanner.service`
- `asset-platform-broadcaster.service`

## 真实验收结果

### 提现结果

- `withdrawal_id = 2`
- `status = confirmed`
- `tx_hash = 0x774cc554e8b774afede1a2ccee919e2bc11a8a89fb856070162ed7de0dc3fad5`
- `block_number = 11164251`
- `nonce = 0`

### scanner 结果

- `scan_checkpoints.chain_id = 1`
- `scan_checkpoints.last_scanned_block` 已从 `0` 推进到 `14`
- 说明：
  - `scanner -> relay -> Sepolia RPC` 已经打通
  - 当前会从低块位开始回扫

### API 结果

- `/healthz` 正常
- `/api/v1/chains` 正常
- `/api/v1/withdrawals?limit=5` 可返回最新确认的提现
- `/api/v1/deposits?limit=5` 当前仍为空数组

## 当前真相

这次改造已经让项目发生了性质变化：

- 提现链路不再只是展示代码
- 已经具备“真实签名 -> 真实广播 -> 真实回执确认”的闭环
- `scanner` 也已经恢复真实扫链能力

但充值侧还没有完成“真实充值事件 -> deposits 入库”的最终验收。  
当前只证明了：

- `scan_checkpoints` 不再卡死
- `scanner` 能持续通过 relay 拉链上区块

还没证明的部分：

- 是否已经存在被监听的真实充值地址
- parser 是否已经在命中目标地址时稳定落 `chain_events/deposits`
- confirm-worker 是否完成充值确认推进

## 现阶段风险

1. 当前 RPC 出口依赖本机 Windows relay + SSH 反向隧道。
   - 这是“可落地最小方案”
   - 不是长期生产方案

2. `scanner` 现在从很低的 checkpoint 回扫，速度较慢。
   - 更适合把 checkpoint 调整到当前头部附近后再做真实充值验收

3. 充值侧虽然链路恢复，但还没完成一次真实 monitored address 入账证明。

## 下一步建议

1. 把国内可控 RPC 出口做成常驻服务。
   - 例如自建轻代理机 / 香港跳板 / 专用 RPC relay
   - 避免依赖开发机在线

2. 为充值侧准备一个明确受监控的测试地址。
   - 然后在 Sepolia 做一笔真实入账
   - 验证 `chain_events -> deposits -> confirmations`

3. 将 `scan_checkpoints` 直接提升到当前链头附近。
   - 避免从创世块附近慢速回扫
   - 更适合真实验收

4. 继续补生产化能力。
   - signer/HSM 边界
   - 提现风控与审批
   - 正式 RPC 多活
   - 监控告警
   - CI/CD 与安全扫描
