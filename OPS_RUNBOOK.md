# BNB 上线与运维入口

更新时间：2026-06-23

这份文档的目标很直接：

- 把 BNB 项目所有常驻服务整理成一套统一上线入口
- 让 `systemd`、环境变量、部署脚本、验证脚本有固定位置
- 降低“服务虽然写了，但线上不好拉起”的运维摩擦

## 1. 当前推荐的服务拓扑

### 对外入口

- `OpenResty / 1Panel`
- `bnb.shuhong.icu`

### 应用服务

- `asset-platform-api.service`
- `asset-scanner.service`
- `asset-parser.service`
- `asset-confirm-worker.service`
- `asset-ledger.service`
- `asset-withdrawal-worker.service`
- `asset-broadcaster.service`
- `chain-node.service`

### 中间件

- `PostgreSQL`
- `Redis`
- `NATS`

## 2. 仓库内固定入口

### 部署脚本

- 自托管 runner 发布脚本：`scripts/deploy_self_hosted.sh`
- 手工部署脚本：`scripts/deploy.sh`
- 回滚脚本：`scripts/rollback.sh`
- 服务验证脚本：`scripts/verify.sh`
- 安装 systemd 模板：`scripts/install_systemd_units.sh`

### systemd 模板

目录：`ops/systemd`

包含：

- `asset-platform-api.service`
- `asset-scanner.service`
- `asset-parser.service`
- `asset-confirm-worker.service`
- `asset-ledger.service`
- `asset-withdrawal-worker.service`
- `asset-broadcaster.service`
- `chain-node.service`

### 环境变量模板

目录：`ops/env`

包含：

- `api.env.example`
- `workers.env.example`
- `node.env.example`

## 3. 服务器推荐目录

以当前腾讯云服务器为例：

```text
/home/ubuntu/opt/bnb/
├── api/
│   ├── api-server-linux-amd64
│   └── .env
├── workers/
│   ├── scanner-linux-amd64
│   ├── parser-linux-amd64
│   ├── confirm-worker-linux-amd64
│   ├── ledger-service-linux-amd64
│   ├── withdrawal-worker-linux-amd64
│   ├── broadcaster-linux-amd64
│   ├── rpc-health-linux-amd64
│   └── .env
├── node/
│   ├── target/release/verifiable-rust-chain-node
│   └── .env
└── releases/
```

前端目录推荐继续放在：

```text
/opt/1panel/www/bnb/frontend
```

OpenResty 站点配置推荐继续放在：

```text
/opt/1panel/www/conf.d/bnb.conf
```

## 4. 一次性安装步骤

### 4.1 写环境变量

从这些模板复制：

- [ops/env/api.env.example](/E:/awesomeProject/BNB/ops/env/api.env.example)
- [ops/env/workers.env.example](/E:/awesomeProject/BNB/ops/env/workers.env.example)
- [ops/env/node.env.example](/E:/awesomeProject/BNB/ops/env/node.env.example)

落到服务器：

- `/home/ubuntu/opt/bnb/api/.env`
- `/home/ubuntu/opt/bnb/workers/.env`
- `/home/ubuntu/opt/bnb/node/.env`

### 4.2 安装 systemd unit

把仓库同步到服务器后执行：

```bash
sudo bash scripts/install_systemd_units.sh
```

### 4.3 启动服务

```bash
sudo systemctl enable --now \
  asset-platform-api \
  asset-scanner \
  asset-parser \
  asset-confirm-worker \
  asset-ledger \
  asset-withdrawal-worker \
  asset-broadcaster \
  chain-node
```

## 5. 发布方式

### 方式 A：推荐

使用 self-hosted runner：

- workflow：`.github/workflows/deploy-bnb.yml`
- 发布脚本：`scripts/deploy_self_hosted.sh`

适合当前你的腾讯云环境，因为不依赖 GitHub 公共 runner SSH 进服务器。

### 方式 B：手工发布

当 workflow 还没完全稳定时：

1. 本地构建产物
2. 上传到服务器
3. 执行 `scripts/deploy.sh`
4. 执行 `scripts/verify.sh`

## 6. 服务健康检查

统一验证脚本：

```bash
bash scripts/verify.sh
```

建议至少检查：

- `asset-platform-api`
- `asset-scanner`
- `asset-parser`
- `asset-confirm-worker`
- `asset-ledger`
- `asset-withdrawal-worker`
- `asset-broadcaster`
- `chain-node`
- `NATS`
- OpenResty `/`
- OpenResty `/api/`
- OpenResty `/node/`

## 7. 常用运维命令

### 看状态

```bash
systemctl status asset-platform-api --no-pager
systemctl status asset-broadcaster --no-pager
systemctl status asset-withdrawal-worker --no-pager
```

### 看日志

```bash
journalctl -u asset-platform-api -f
journalctl -u asset-scanner -f
journalctl -u asset-ledger -f
journalctl -u asset-withdrawal-worker -f
journalctl -u asset-broadcaster -f
journalctl -u chain-node -f
```

### 重启单个服务

```bash
sudo systemctl restart asset-platform-api
sudo systemctl restart asset-broadcaster
```

### 整体重启

```bash
sudo systemctl restart \
  asset-platform-api \
  asset-scanner \
  asset-parser \
  asset-confirm-worker \
  asset-ledger \
  asset-withdrawal-worker \
  asset-broadcaster \
  chain-node
```

## 8. 当前最值得强调的上线注意事项

- `api-server` 必须只监听 `127.0.0.1`
- `NATS / PostgreSQL / Redis` 不要暴露公网
- `broadcaster` 真实广播依赖 `WITHDRAWAL_SIGNER_PRIVATE_KEY`
- `workers/.env` 需要和数据库中的 `chains / tokens / rpc_providers` 一致
- 服务器若出站 `443` 仍有问题，`scanner` 和 `broadcaster` 的真实链访问会受影响

## 9. 这份运维收口带来的结果

到这一步，项目上线不再是零散知识点，而是被收口成三层固定入口：

1. 文档入口：`OPS_RUNBOOK.md`
2. 配置入口：`ops/env` 和 `ops/systemd`
3. 执行入口：`scripts/deploy_self_hosted.sh`、`scripts/install_systemd_units.sh`、`scripts/verify.sh`

后面你如果继续补真实充值闭环、真实出金广播、人工审核后台，这套入口可以继续沿用，不需要再重做一套运维结构。
