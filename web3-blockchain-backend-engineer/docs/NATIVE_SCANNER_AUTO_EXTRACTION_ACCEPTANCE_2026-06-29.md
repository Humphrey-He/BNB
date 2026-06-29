# Native Scanner 自动提取验收

日期：2026-06-29  
环境：Sepolia / `bnb-server`

## 验收目标

把 scanner 的 native block 提取改成不怕“大块超时”的版本，让下一笔真实 native ETH 充值不再需要手工回填 `raw_events`。

## 本次实现

### scanner 侧改造

1. 不再依赖 `eth_getBlockByNumber(true)` 拉整块全量交易详情。
2. 改为先拉轻量区块元信息：
   - `eth_getBlockByNumber(false)`
   - 只拿区块头和交易哈希列表
3. 再按交易哈希批量拉交易详情：
   - `eth_getTransactionByHash`
   - 每批 100 笔
4. 只对 `watched_addresses` 命中的 `to_address` 生成 `NativeTransfer` 原始事件。
5. 如果区块元信息或交易详情提取失败，scanner 本轮直接失败，不推进 checkpoint，避免静默漏单。
6. scanner 自身 RPC timeout 提升到 `120s`，与 relay 保持一致。

### 相关文件

- [internal/scanner/fetcher.go](/E:/awesomeProject/BNB/web3-blockchain-backend-engineer/internal/scanner/fetcher.go)
- [internal/scanner/resilient_client.go](/E:/awesomeProject/BNB/web3-blockchain-backend-engineer/internal/scanner/resilient_client.go)
- [internal/scanner/scanner.go](/E:/awesomeProject/BNB/web3-blockchain-backend-engineer/internal/scanner/scanner.go)
- [cmd/scanner/main.go](/E:/awesomeProject/BNB/web3-blockchain-backend-engineer/cmd/scanner/main.go)

## 真实自动化验证

### 测试 watched 地址

- `0x000000000000000000000000000000000000bEEF`

### 真实链上交易

- `withdrawal id = 4`
- `tx_hash = 0xf0d98679937901f3c160552d401733ac37f67e93aa5164281ee568130e8cb3a2`
- `block_number = 11164476`
- `amount = 10000000000000` wei

### 关键要求

这次验证 **没有** 手工回填 `raw_events`。

只做了两件事：

1. 把 `scan_checkpoints` 调到目标块前一格，强制 scanner 重扫目标块。
2. 启动新的 scanner 二进制。

### 最终结果

`chain_events`

- 已自动出现该交易：
  - `event_name = NativeTransfer`
  - `to_address = 0x000000000000000000000000000000000000bEEF`

`deposits`

- 已自动出现并推进：
  - `status = confirmed`
  - `confirmations = 17`
  - `target_confirmations = 3`
  - `confirmed_event_published = true`

这说明：

- scanner 自动 native 提取已成功
- parser 自动入库已成功
- confirm-worker 自动确认已成功
- 下一次真实 native ETH 充值不再依赖手工补发 `raw_events`

## 和上一版的差异

上一笔 native 测试交易：

- `0x1a1fe7afc98edd9d94156c4c35a41de00876243ca0810df48551c3d7f6ec8897`

当时的问题是：

- scanner 能扫到目标块的 ERC-20 logs
- 但在高日志量区块上，native transfer 所需的整块交易详情提取不稳定
- 最终需要手工补发 `raw_events`

这一版的改造后：

- 第二笔真实交易 `0xf0d986...` 已经通过 scanner 自动提取完成闭环

## 当前结论

native ETH 充值路径现在具备了真实可用的自动化能力：

- 可自动扫块
- 可自动识别 watched address 的 native 转账
- 可自动进入 `chain_events`
- 可自动形成 `deposits`
- 可自动推进到 `confirmed`

剩余优化方向不是“能不能用”，而是“怎样更稳、更快、更生产化”：

1. 为 `GetTransactionsByHashes` 加更细粒度重试与批次降级。
2. 为 scanner 增加 watched address 缓存与刷新机制。
3. 为 native transfer 增加专门指标：
   - block tx count
   - native tx extracted count
   - watched native hit count
   - transaction detail fetch latency
4. 把“定向重扫某个目标块”沉淀成正式运维工具。
