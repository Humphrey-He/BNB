# 测试链真实出金接入说明

更新时间：2026-06-23

## 目标

把当前提现链路从“骨架可运行”推进到“测试链可真实出金”。

本次实现已经补上：

- 真实本地私钥 signer
- 真实 EVM RPC broadcaster
- native token 转账签名与广播
- ERC-20 `transfer` 签名与广播
- `PendingNonceAt + nonce_allocations` 组合 nonce 分配
- receipt 轮询确认

## 当前链路

现在提现链路是：

`CreateWithdrawal -> withdrawal_created -> withdrawal-worker -> signing -> broadcaster -> broadcasted -> confirmed/failed`

关键服务：

- `api-server`
- `withdrawal-worker`
- `broadcaster`
- `nats`
- PostgreSQL

## 环境变量

至少要补下面这些变量：

```env
POSTGRES_HOST=127.0.0.1
POSTGRES_PORT=5432
POSTGRES_DB=asset_platform
POSTGRES_USER=platform
POSTGRES_PASSWORD=your_db_password

NATS_URL=nats://127.0.0.1:4222

WITHDRAWAL_HOT_WALLET_ADDRESS=0xYourHotWalletAddress
WITHDRAWAL_MAX_AUTO_APPROVE_AMOUNT=1000000000000000000
WITHDRAWAL_REQUIRE_WHITELIST=true
WITHDRAWAL_ALLOWED_DESTINATIONS=0xYourWhitelistAddress

WITHDRAWAL_SIGNER_PRIVATE_KEY=0xyour_testnet_private_key
```

说明：

- `WITHDRAWAL_SIGNER_PRIVATE_KEY` 只建议先用于测试链热钱包
- `WITHDRAWAL_HOT_WALLET_ADDRESS` 必须和该私钥推导出的地址一致
- `WITHDRAWAL_ALLOWED_DESTINATIONS` 建议先只放你自己的测试地址

## 数据库前置

你需要先保证：

1. `chains` 里有目标测试链
2. `rpc_providers` 里有该测试链的 RPC
3. `tokens` 里有要提现的 token
4. `balances` 里热钱包地址有足够可用余额

### Sepolia 示例

链：

```sql
INSERT INTO chains (chain_id, name, native_symbol, finality_confirmations, is_active)
VALUES (11155111, 'Sepolia', 'ETH', 3, true)
ON CONFLICT (chain_id) DO NOTHING;
```

RPC：

```sql
INSERT INTO rpc_providers (chain_id, name, url, weight, is_active)
VALUES (
  (SELECT id FROM chains WHERE chain_id = 11155111),
  'Sepolia RPC',
  'https://ethereum-sepolia-rpc.publicnode.com',
  100,
  true
);
```

原生 token：

```sql
INSERT INTO tokens (chain_id, contract_address, symbol, decimals, is_native, is_active)
VALUES (
  (SELECT id FROM chains WHERE chain_id = 11155111),
  '0x0000000000000000000000000000000000000000',
  'ETH',
  18,
  true,
  true
)
ON CONFLICT (chain_id, contract_address) DO NOTHING;
```

## 如何发起一笔测试链提现

### 1. 先给热钱包准备余额

确保：

- 链上热钱包有真实测试币用于 gas
- 库里 `balances.available_balance` 也已经有对应资产余额

### 2. 启动服务

至少启动：

- `api-server`
- `withdrawal-worker`
- `broadcaster`

### 3. 创建提现申请

```bash
curl -X POST "http://127.0.0.1:8080/api/v1/withdrawals" \
  -H "Authorization: Bearer ${API_AUTH_TOKEN}" \
  -H "Idempotency-Key: withdrawal-test-001" \
  -H "X-From-Address: 0xYourHotWalletAddress" \
  -H "Content-Type: application/json" \
  -d '{
    "chain_id": 3,
    "token_id": 5,
    "to_address": "0xYourDestinationAddress",
    "amount": "1000000000000000"
  }'
```

注意：

- 这里的 `chain_id` / `token_id` 是你数据库里的主键语义链路在当前项目中的实际使用值
- 不是单纯写 EVM 公链数字就一定生效

## 验证方式

你应该依次看到：

1. `withdrawals.status` 从 `created` 进入 `approved/signing`
2. `broadcaster` 日志打印广播成功
3. `withdrawals.tx_hash` 被写入
4. `withdrawals.status` 更新为 `broadcasted`
5. receipt 轮询后更新为 `confirmed`

可以直接查：

```sql
SELECT id, chain_id, token_id, from_address, to_address, amount, status, tx_hash, nonce, failure_reason
FROM withdrawals
ORDER BY id DESC
LIMIT 20;
```

以及：

```sql
SELECT *
FROM nonce_allocations
ORDER BY id DESC
LIMIT 20;
```

## 当前实现边界

这版已经能用于测试链真实出金 MVP，但还不是最终生产级。

还缺：

- KMS/HSM signer
- EIP-1559 fee 策略
- 更严格的 nonce 串行锁
- replacement transaction
- 更细的广播错误分类
- 提现审批审计字段
- 更完整的对账与回滚策略

## 下一步建议

最优先继续补：

1. `WITHDRAWAL_SIGNER_BACKEND=local|kms` 抽象
2. `eth_feeHistory` / `suggestTipCap` 费率策略
3. `nonce_allocations` 和 withdrawal 强绑定
4. 广播 attempt 表
5. `broadcasted -> confirming -> confirmed` 更细状态机
