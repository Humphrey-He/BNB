# 部署指南

## 1. 环境要求

| 依赖 | 版本要求 |
|------|---------|
| Rust | 1.75+ |
| Cargo | stable |
| 操作系统 | Linux / macOS / Windows |

安装 Rust：

```bash
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
rustup update stable
```

## 2. 编译

```bash
# 克隆项目
git clone https://github.com/your-org/verifiable-rust-chain-node.git
cd verifiable-rust-chain-node

# Debug 模式编译
cargo build

# Release 模式编译（生产环境推荐）
cargo build --release
```

编译产物位于 `target/debug/` 或 `target/release/` 目录。

## 3. 运行

### 3.1 单节点启动（开发模式）

```bash
# 使用默认配置启动
cargo run

# 或使用 release 构建
./target/release/verifiable-rust-chain-node
```

启动后会输出类似日志：

```
2026-05-06T10:00:00Z  INFO verifiable_rust_chain_node: Starting node...
2026-05-06T10:00:00Z  INFO verifiable_rust_chain_node: Listening on 0.0.0.0:8080
2026-05-06T10:00:00Z  INFO verifiable_rust_chain_node: RPC endpoint: http://127.0.0.1:8080
2026-05-06T10:00:00Z  INFO verifiable_rust_chain_node: Mempool capacity: 10000
2026-05-06T10:00:01Z  INFO verifiable_rust_chain_node: Produced block #1 hash=0x...
```

### 3.2 配置

通过环境变量或配置文件（`config.toml`）调整参数：

```toml
[server]
rpc_host = "0.0.0.0"
rpc_port = 8080

[chain]
chain_id = 1337
block_time = 12  # 出块间隔（秒），0 表示不自动出块

[mempool]
max_size = 10000
eviction_threshold = 8000

[storage]
path = "./data"  # RocksDB 存储路径
```

## 4. RPC API

服务启动后可通过 HTTP 调用 RPC 接口。

### 4.1 提交交易

```bash
curl -X POST http://127.0.0.1:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "method": "send_transaction",
    "params": [{
      "from": "0x1234567890123456789012345678901234567890",
      "to": "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
      "value": "1000000000000000000",
      "nonce": 0,
      "gas_limit": 21000,
      "max_fee_per_gas": "1000000000",
      "signature": "0x..."
    }],
    "id": 1
  }'
```

### 4.2 查询余额

```bash
curl -X POST http://127.0.0.1:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "method": "get_balance",
    "params": ["0x1234567890123456789012345678901234567890"],
    "id": 2
  }'
```

### 4.3 查询区块

```bash
# 按高度查询
curl -X POST http://127.0.0.1:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "method": "get_block_by_number",
    "params": [1],
    "id": 3
  }'

# 按 hash 查询
curl -X POST http://127.0.0.1:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "method": "get_block_by_hash",
    "params": ["0xabc123..."],
    "id": 4
  }'
```

### 4.4 查询交易

```bash
curl -X POST http://127.0.0.1:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "method": "get_transaction",
    "params": ["0xdef456..."],
    "id": 5
  }'
```

### 4.5 查询收据

```bash
curl -X POST http://127.0.0.1:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "method": "get_receipt",
    "params": ["0xdef456..."],
    "id": 6
  }'
```

### 4.6 查询 Nonce

```bash
curl -X POST http://127.0.0.1:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "method": "get_nonce",
    "params": ["0x1234567890123456789012345678901234567890"],
    "id": 7
  }'
```

### 4.7 查询 Mempool 状态

```bash
curl -X POST http://127.0.0.1:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "method": "get_mempool_status",
    "params": [],
    "id": 8
  }'
```

## 5. 日志

日志通过 `tracing` 记录，支持以下环境变量配置：

```bash
# 全部日志
RUST_LOG=trace

# 仅 info 级别以上
RUST_LOG=info

# 仅警告和错误
RUST_LOG=warn

# 输出到文件
RUST_LOG=info,verifiable_rust_chain_node=debug
```

运行时指定输出：

```bash
RUST_LOG=info cargo run 2>&1 | tee node.log
```

## 6. 测试

### 6.1 单元测试

```bash
# 运行所有单元测试
cargo test --lib

# 运行特定模块测试
cargo test --lib mempool
cargo test --lib validation
cargo test --lib executor
```

### 6.2 集成测试

```bash
cargo test
```

### 6.3 性能基准测试

```bash
# 运行 Criterion benchmarks
cargo bench

# 查看报告（HTML）
open target/criterion/report/index.html
```

## 7. Docker 部署（可选）

### 7.1 构建镜像

```bash
docker build -t verifiable-rust-chain-node .
```

### 7.2 运行容器

```bash
docker run -d \
  --name chain-node \
  -p 8080:8080 \
  -v chain-data:/data \
  -e RUST_LOG=info \
  verifiable-rust-chain-node
```

### 7.3 Docker Compose

```yaml
version: "3.8"
services:
  node:
    image: verifiable-rust-chain-node:latest
    ports:
      - "8080:8080"
    volumes:
      - chain-data:/data
    environment:
      - RUST_LOG=info
    restart: unless-stopped

volumes:
  chain-data:
```

```bash
docker-compose up -d
```

## 8. 目录结构

运行后会生成以下目录结构：

```
.
├── data/                    # 区块链数据（RocksDB）
│   ├── blocks/             # 区块数据
│   ├── state/             # 状态数据
│   └── indices/           # 索引数据
├── logs/                   # 日志文件（可选）
├── config.toml             # 配置文件（可选）
└── verifiable-rust-chain-node  # 可执行文件
```

## 9. 常见问题

### 编译报错：找不到 `sha2`

确保安装了 Rust stable 工具链：

```bash
rustup default stable
cargo update
cargo build
```

### 测试失败：mempool 或 validation 错误

检查是否引入了 Week 3 review 后的新语义变更。确保所有模块使用一致的 `state_nonce` 参数。

### RPC 无响应

确认服务已启动并监听正确端口（默认 8080）。检查防火墙配置。

## 10. 生产环境注意事项

1. **数据备份**：定期备份 `data/` 目录
2. **签名校验**：当前实现使用占位符签名校验，生产环境需替换为真实 ECDSA（k256 crate）
3. **P2P 安全**：生产部署建议启用 TLS 和节点认证
4. **监控**：建议接入 Prometheus + Grafana 监控关键指标
5. **资源限制**：合理配置 RocksDB 缓存大小（`--cache-size`）