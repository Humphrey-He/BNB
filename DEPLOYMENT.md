# BNB Web3 Career Projects — 1Panel 服务器部署文档

> 适用:把 `BNB/` 下的 4 个项目同时部署到一台腾讯云服务器(101.43.127.178)。
> 目标受众:候选人本人 / HR 通过 IP 直接访问作品集。
> 服务器:腾讯云 CVM / Lighthouse(任意 Linux,本文按 1Panel 官方支持的 Debian 12 / Ubuntu 22 写)。
> 时间:2026-06-12 评审。

---

## 1. 部署总览

### 1.1 四个项目对外暴露形态

| 项目 | 进程类型 | 默认端口(代码内) | 部署后对外端口 | 暴露形态 |
|---|---|---:|---:|---|
| **portfolio-frontend** | 静态文件 | — (Vite dev 5173) | **80/443** | OpenResty 直接服务,无进程 |
| **web3-blockchain-backend-engineer** | Go 单进程 | 8080(`PORT` env 可改) | **8080**(内网) | OpenResty 反代到 `/api/*` |
| **protocol-rust-blockchain-engineer** | Rust 单进程 | **8080 写死**(需改源码) | **8081**(内网) | OpenResty 反代到 `/node/*` |
| **smart-contract-engineer** | Foundry 工具链 | — | **不在服务器跑** | 本地 anvil / 测试网即可 |

> ⚠️ **端口冲突点**:Go 和 Rust 默认都 8080。**Rust 是写死的**(在 `protocol-rust-blockchain-engineer/src/main.rs:51`),要改端口必须改源码 → 见 §6 步骤 4。

### 1.2 中间件清单

| 中间件 | 项目内 | 你服务器已有 | 本次部署 | 用途 | 端口 | 部署方式 |
|---|---|---|---|---|---|---|
| **OpenResty** | ❌ | ✅(1Panel 默认) | ✅ 复用 | 反代 + 静态托管 + HTTPS | 80 / 443 | 1Panel 已有,改 conf.d |
| **PostgreSQL 15+** | ✅(主存) | ✅ pgsql | ✅ 复用 | 资产/账本/事件/区块/出块状态 | 5432 | 1Panel 应用商店装 |
| **Redis 7+** | ✅(缓存/锁/限流/nonce) | ✅ redis | ✅ 复用 | 分布式锁 / 速率限制 / nonce 窗口 | 6379 | 1Panel 应用商店装 |
| **NATS 2.10+ (JetStream)** | ✅(确认数/提现/补偿事件) | ❌ 你只有 Kafka | ✅ **新装** | 异步消息、deposit_confirmed 投递 | 4222 / 8222 | Docker 跑 |
| **Prometheus** | ✅(可选) | ❌ | ⚪ 可选 | 服务指标采集 | 9090 | 暂不装 |
| **Kafka** | ❌ | ✅ 你有 | ❌ **不动** | 本项目用 NATS,语义不同,API 不兼容 | — | 不部署,不删除 |
| **MySQL / MongoDB** | ❌ | — | ❌ | 没用 | — | — |

### 1.3 公网 vs 内网端口开放策略

| 端口 | 监听地址 | 谁能访问 | 理由 |
|---|---|---|---|
| 80 / 443 | 0.0.0.0(OpenResty) | 公网 | HR / 面试官 |
| 8080 | 127.0.0.1(Go API) | 仅本机 + OpenResty | API 走反代,不直暴露 |
| 8081 | 127.0.0.1(Rust node) | 仅本机 + OpenResty | 同上 |
| 4222 | 127.0.0.1(NATS client) | 仅本机内部 | NATS 永远不暴露公网 |
| 8222 | 127.0.0.1(NATS monitor) | 仅本机 | 监控接口 |
| 5432 | 127.0.0.1(PG) | 仅本机 | **强烈建议不要开公网** |
| 6379 | 127.0.0.1(Redis) | 仅本机 | 同上 |

> **腾讯云安全组规则**:入站只放 `80/443` 给 `0.0.0.0/0`,放 `22` 给你固定 IP(或关掉改用 1Panel 终端)。其它一律不开。

---

## 2. 服务器前置准备

### 2.1 操作系统与 1Panel

- 1Panel 官方支持:Debian 11+/12、Ubuntu 22.04、CentOS 7+/8(本文以 Debian 12 为例)。
- 1Panel 装好默认带 OpenResty,默认站点在 `/usr/local/openresty/nginx/conf/conf.d/default.conf`,**不要直接改这个**。

### 2.2 服务器初始账号与密钥

- 1Panel 安装时设的 admin 密码
- SSH 端口(腾讯云默认 22)
- 域名(本方案暂不需要)

### 2.3 防火墙 & 安全组

1. **腾讯云控制台 → 安全组**:入站规则只保留 `22`、`80`、`443`,源 `0.0.0.0/0`(22 建议改为你固定 IP)
2. **服务器本地**:`1Panel → 主机 → 防火墙` 默认状态下不需要再开,1Panel 内部已经管理

---

## 3. 阶段 A:中间件准备

### 3.1 PostgreSQL 15(1Panel 应用商店)

1. `1Panel → 应用商店` 搜索 `PostgreSQL` 装 15.x
2. 安装完在 `数据库` 菜单里:
   - 建库 `asset_platform`
   - 记下密码(后续 `.env` 用)
3. 跑迁移:
   ```bash
   # 在 1Panel 终端里执行
   cd /opt/bnb/api
   PGPASSWORD=你的密码 psql -h 127.0.0.1 -U postgres -d asset_platform \
     -f scripts/migrations/000001_init_schema_up.sql
   ```

### 3.2 Redis 7(1Panel 应用商店)

- 1Panel 装 Redis 7,**确认监听 127.0.0.1:6379**(默认就是)
- 不要开公网

### 3.3 NATS 2.10(1Panel 没有,直接 docker run)

1Panel 应用商店没 NATS,直接用 Docker:

```bash
docker run -d \
  --name nats \
  --restart unless-stopped \
  -p 127.0.0.1:4222:4222 \
  -p 127.0.0.1:8222:8222 \
  -v /opt/bnb/nats-data:/data \
  nats:2.10-alpine \
  -js -sd /data
```

验证:`curl http://127.0.0.1:8222/varz` 返回 JSON。

**安全要求**:
- NATS 永远不暴露公网(`-p 127.0.0.1:...`)
- 没启用 auth(JetStream 走单租户本机,不需要)
- 长期建议加 `--user` `--pass`,但本项目属内网 demo,可选

### 3.4 不要碰的东西

- ❌ **Kafka**:不用,API 跟项目里的 NATS 客户端不兼容,要改代码
- ❌ **OpenResty 主配置**:`/usr/local/openresty/nginx/conf/nginx.conf` 不动
- ❌ **default.conf**:1Panel 自带默认站点不动
- ❌ **Prometheus**:本方案不监控,可后续按需加

---

## 4. 阶段 B:本地编译产物

> 编译分两条线:**Go 跨平台友好,Rust 必须用服务器架构在服务器上重编**(不能传 Windows .exe)。

### 4.1 本地 Windows 编译 Go API

```powershell
# 在 PowerShell / Git Bash
cd E:\awesomeProject\BNB\web3-blockchain-backend-engineer
$env:GOOS="linux"; $env:GOARCH="amd64"
go build -o api-server-linux-amd64 ./cmd/api-server
```

把 `api-server-linux-amd64` 传到服务器 `/opt/bnb/api/`。

### 4.2 服务器本地编译 Rust 节点

**不能**传 Windows .exe 上 Linux 服务器,GLIBC 不兼容。需要在服务器上重编:

```bash
# 服务器装 rust(如果还没)
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
source $HOME/.cargo/env

# 同步代码(用 git 或 scp 上传 src/ 目录)
# 假设源码已放 /opt/bnb/node
cd /opt/bnb/node
cargo build --release
# 产物:target/release/verifiable-rust-chain-node
```

### 4.3 本地编译 frontend

```powershell
cd E:\awesomeProject\BNB\portfolio-frontend\app
npm install
npm run build
```

把 `dist/` 全部上传到 `/opt/bnb/frontend/`。

### 4.4 合约不编译上传

`smart-contract-engineer` 是 Foundry 项目,**不在服务器上跑**(不需要 24h 在线的合约,部署是单次操作,演示用本地 anvil 即可)。

---

## 5. 阶段 C:Go 后端部署

### 5.1 目录结构

```
/opt/bnb/api/
├── api-server-linux-amd64          # 主二进制
├── .env                            # 环境变量
├── configs/
├── scripts/migrations/000001_*.sql # 已经在阶段 A 跑过
└── README.md
```

### 5.2 写 .env

```bash
cat > /opt/bnb/api/.env <<'EOF'
POSTGRES_HOST=127.0.0.1
POSTGRES_PORT=5432
POSTGRES_DB=asset_platform
POSTGRES_USER=postgres
POSTGRES_PASSWORD=__改成实际密码__
REDIS_HOST=127.0.0.1
REDIS_PORT=6379
NATS_URL=nats://127.0.0.1:4222
APP_ENV=production
LOG_LEVEL=info
PORT=8080
EOF
chmod 600 /opt/bnb/api/.env
```

### 5.3 systemd 托管

```bash
cat > /etc/systemd/system/asset-platform-api.service <<'EOF'
[Unit]
Description=BNB Web3 Asset Platform API
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/bnb/api
EnvironmentFile=/opt/bnb/api/.env
ExecStart=/opt/bnb/api/api-server-linux-amd64
Restart=on-failure
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable --now asset-platform-api
systemctl status asset-platform-api
```

### 5.4 启动后验证

```bash
curl http://127.0.0.1:8080/healthz
# 应返回 {"status":"ok"} 或类似
curl http://127.0.0.1:8080/api/v1/chain-status
```

### 5.5 scanner 子模块的可选启动

Go 后端的 `internal/scanner` 会**真连公网 RPC**(eth_getLogs 等)。
**强烈建议**:本次部署**先不启用 scanner**,只跑 API 层。

如果你后续想给面试官演示"真在跑链上监听",再加 systemd 单元,先把代码里的 RPC provider 配好,见 §10。

---

## 6. 阶段 D:Rust 节点部署

### 6.1 端口冲突解决(必读)

Rust 代码里 RPC 端口**写死 8080**(`src/main.rs:51`),跟 Go 后端冲突。
**两个解法**:
- **方案 A(推荐)**:改源码为 `8081`,重编
  ```rust
  // protocol-rust-blockchain-engineer/src/main.rs:51
  if let Err(e) = start_rpc_server(rpc_state, 8081).await {
  ```
- **方案 B**:把 Go API 改成 8081,Rust 保持 8080 — 不推荐,因为 8080 数字更主流

### 6.2 改源码 → 服务器重编

```bash
# 上传修改后的 src/main.rs 到服务器 /opt/bnb/node/src/main.rs
cd /opt/bnb/node
cargo build --release
```

### 6.3 systemd 托管

```bash
cat > /etc/systemd/system/chain-node.service <<'EOF'
[Unit]
Description=BNB Verifiable Rust Chain Node
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/bnb/node
ExecStart=/opt/bnb/node/target/release/verifiable-rust-chain-node
Restart=on-failure
RestartSec=5
Environment=RUST_LOG=info

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable --now chain-node
systemctl status chain-node
```

### 6.4 验证

```bash
curl -X POST http://127.0.0.1:8081/rpc \
  -H "Content-Type: application/json" \
  -d '{"method":"get_mempool_status","params":[],"id":1}'
```

---

## 7. 阶段 E:Frontend 部署(纯静态)

### 7.1 目录结构

```
/opt/bnb/frontend/
├── index.html
├── favicon.svg
├── icons.svg
└── assets/
    ├── index-0N4p21yw.css
    └── index-CbuvN4dG.js
```

### 7.2 不需要任何进程

OpenResty 直接服务这个目录即可。**没有 npm run preview 之类的服务进程**,省内存。

### 7.3 SPA 路由注意

`portfolio-frontend` 是 Vite SPA,所有路由都是客户端 history 模式(?)
**风险点**:如果用户在 HR 发来的链接末尾加 `/risk` `/protocol` 这种 hash 或路径,刷新会 404。
**OpenResty 必须配** `try_files $uri $uri/ /index.html;`(见 §8)

---

## 8. 阶段 F:OpenResty 反代配置

### 8.1 不动默认配置

`/usr/local/openresty/nginx/conf/conf.d/default.conf` 是 1Panel 自己的默认站,**不要碰**。

### 8.2 新建专属配置

```bash
cat > /usr/local/openresty/nginx/conf/conf.d/bnb.conf <<'EOF'
server {
    listen 80 default_server;
    server_name 101.43.127.178 _;

    # 前端静态
    root /opt/bnb/frontend;
    index index.html;

    # SPA fallback
    location / {
        try_files $uri $uri/ /index.html;
    }

    # 静态资源缓存
    location /assets/ {
        expires 30d;
        add_header Cache-Control "public, immutable";
    }

    # Go 后端 API
    location /api/ {
        proxy_pass http://127.0.0.1:8080/api/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 60s;
    }

    # Rust 节点 RPC
    location /node/ {
        proxy_pass http://127.0.0.1:8081/rpc;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # health 探针
    location = /health {
        access_log off;
        return 200 "ok\n";
        add_header Content-Type text/plain;
    }
}
EOF
```

### 8.3 关闭 1Panel 默认站冲突(如果需要)

`/usr/local/openresty/nginx/conf/conf.d/default.conf` 默认 listen 80 会跟上面 `default_server` 冲突。**两种处理方式二选一**:

- **方式 A**(推荐):把上面 bnb.conf 设成 `default_server`,把 default.conf 里的 `default_server` 关键字删掉(保留 listen 80 但不抢 default)
- **方式 B**:把 default.conf 改 listen 到别的端口(比如 8088),让 1Panel 内部管理用

### 8.4 校验与重载

```bash
/usr/local/openresty/bin/openresty -t
/usr/local/openresty/bin/openresty -s reload
```

---

## 9. 部署后的端到端验证

| 验证项 | 命令 | 预期 |
|---|---|---|
| OpenResty 配置语法 | `openresty -t` | `syntax is ok` |
| OpenResty 80 端口 | `curl -I http://101.43.127.178/` | `200 OK`,`Content-Type: text/html` |
| Frontend JS 加载 | `curl -I http://101.43.127.178/assets/index-*.js` | `200 OK` |
| Go API health | `curl http://101.43.127.178/api/healthz` | `{"status":"ok"}` |
| Go API chain status | `curl http://101.43.127.178/api/v1/chain-status` | 返回 chain 列表 |
| Rust node RPC | `curl -X POST http://101.43.127.178/node -d '{"method":"get_mempool_status","params":[],"id":1}' -H 'Content-Type: application/json'` | 返回 JSON-RPC 响应 |
| NATS monitor(本机) | `curl http://127.0.0.1:8222/varz` | 返回 nats varz |
| PostgreSQL(本机) | `PGPASSWORD=... psql -h 127.0.0.1 -U postgres -d asset_platform -c '\dt'` | 列出表 |

---

## 10. 注意事项与已知风险

### 10.1 安全风险(P0 必看)

1. **不要把 PG/Redis 暴露公网**:1Panel 默认装好是 bind 0.0.0.0,需要在 `postgresql.conf` 和 `redis.conf` 加 `bind 127.0.0.1`,或用 1Panel 防火墙拒绝 5432/6379 来自公网的连接。
2. **不要把 NATS 4222 暴露公网**:NATS 没 auth,任何能连上的人就能推消息进你的事件流。
3. **不要把 Go/Rust 进程直暴露 8080/8081**:必须经 OpenResty,避免 API 服务直接对公网(`gin` 默认是 0.0.0.0,需要 `r.Run("127.0.0.1:8080")`)。
4. **Rust 进程当前绑定 0.0.0.0**(`src/rpc.rs:36` 是 `0.0.0.0:{}`),如果忘了 OpenResty 反代就是裸奔。**必须**先用反代,或者改源码绑 127.0.0.1。
5. **.env 文件**:`chmod 600`,不要 commit。

### 10.2 资源风险

- **服务器最低配置**:2C4G 够 demo(Go ~50MB + Rust ~80MB + NATS ~30MB + PG ~200MB + Redis ~30MB ≈ 400MB,剩余给 OS 和 OpenResty)
- **磁盘**:`/opt/bnb/nats-data` 持续增长,定期清理旧 stream;PG 的 deposit/ledger 表会膨胀,需要加清理 job(项目目前没有)

### 10.3 项目已知缺陷(来自 `BLOCKCHAIN_PROJECTS_IMPLEMENTATION_REVIEW.md`)

| 缺陷 | 部署影响 |
|---|---|
| Go 后端用 core NATS,不是 JetStream durable consumer | 进程崩了 `deposit_confirmed` 可能丢 —— 但你不跑 scanner,这条没有实际数据流入,可忽略 |
| Go ledger 用 `int64` 解析 NUMERIC(78,0) | 大额资产会算错 —— 但你不接真链,余额是 mock,可忽略 |
| Rust 签名是 placeholder | 仅做协议演示,可忽略 |
| `cargo fmt --check` 失败 | 不影响运行,后续再修 |
| Foundry 合约没测 | 不影响本部署 |

### 10.4 防火墙 & 安全组(Tencent Cloud 一次性配置)

| 协议 | 端口 | 源 | 状态 |
|---|---|---|---|
| TCP | 22 | 你的固定 IP / 1Panel 终端 | 放行 |
| TCP | 80 | 0.0.0.0/0 | 放行 |
| TCP | 443 | 0.0.0.0/0 | 放行(以后上 HTTPS 用) |
| TCP | 5432 / 6379 / 4222 / 8222 / 8080 / 8081 | 0.0.0.0/0 | **拒绝** |

> 腾讯云安全组是**第一道墙**,服务器本地 iptables / firewalld 是第二道。本项目以 1Panel + 安全组为主,服务器本地 iptables 保持 1Panel 默认即可。

### 10.5 HTTPS / 域名(本方案暂不做)

- 当前:HTTP,浏览器标"不安全"
- 后续要加:域名解析到 101.43.127.178 → 1Panel 网站 → 申请 Let's Encrypt → 强制 HTTPS
- 不影响本次 HR 演示

### 10.6 Foundry 合约不上服务器

合约部署是**一次性操作**,演示用本地 anvil 即可。**不要把 foundry 装服务器当常驻进程**。

---

## 11. 运维 Runbook

### 11.1 看日志

```bash
journalctl -u asset-platform-api -f
journalctl -u chain-node -f
tail -f /opt/bnb/nats-data/nats.log  # 如果开了 log file
```

### 11.2 重启服务

```bash
systemctl restart asset-platform-api
systemctl restart chain-node
```

### 11.3 重新加载 OpenResty

```bash
/usr/local/openresty/bin/openresty -s reload
```

### 11.4 备份

- PostgreSQL:`pg_dump -U postgres asset_platform > backup_$(date +%F).sql`
- NATS 数据:`tar czf nats-backup-$(date +%F).tar.gz /opt/bnb/nats-data`
- Frontend 是纯静态,不需要备份,丢了 `npm run build` 重出

### 11.5 回滚

- 保留上一版二进制:`mv api-server-linux-amd64 api-server-linux-amd64.bak`
- 出问题:`cp api-server-linux-amd64.bak api-server-linux-amd64 && systemctl restart asset-platform-api`

---

## 12. 部署后给 HR 的链接(暂时只有 IP)

```
http://101.43.127.178/
```

按 OpenResty 配的 `default_server`,直接 IP 访问就是 frontend 首页。
- 浏览器建议:Chrome / Edge / Safari,手机端能正常看(响应式)

---

## 13. 端口与服务快速速查

| 端口 | 服务 | 监听 | 谁能连 |
|---|---|---|---|
| 80 | OpenResty | 0.0.0.0 | 公网 |
| 443 | OpenResty(HTTPS 占位) | 0.0.0.0 | 公网(暂未签证书) |
| 22 | SSH | 0.0.0.0 | 限 IP |
| 8080 | Go API(内部) | **必须改 127.0.0.1** | 仅 OpenResty |
| 8081 | Rust node(内部) | 当前 0.0.0.0,**必须改** | 仅 OpenResty |
| 4222 | NATS client | **必须 127.0.0.1** | 仅本机 |
| 8222 | NATS monitor | **必须 127.0.0.1** | 仅本机 |
| 5432 | PostgreSQL | **必须 127.0.0.1** | 仅本机 |
| 6379 | Redis | **必须 127.0.0.1** | 仅本机 |

---

## 14. 部署检查清单

- [ ] 1Panel 已装好,OpenResty 跑着
- [ ] PostgreSQL 15 装好,数据库 `asset_platform` 已建,迁移已跑
- [ ] Redis 7 装好,监听 127.0.0.1
- [ ] NATS 2.10 用 docker run 启好,`-js -sd /data`
- [ ] Go API 跨平台编译产物已传到 `/opt/bnb/api/`,`.env` 已写
- [ ] Rust 源码已上传(改过端口),`cargo build --release` 已编
- [ ] `frontend/dist` 已传到 `/opt/bnb/frontend/`
- [ ] `asset-platform-api.service` 和 `chain-node.service` 已 enable
- [ ] `bnb.conf` 已放 `conf.d/`,default_server 冲突已处理
- [ ] `openresty -t` 通过,`openresty -s reload` 完成
- [ ] 9 项端到端验证全过
- [ ] 腾讯云安全组只放 22/80/443
- [ ] PG/Redis/NATS 确认 bind 127.0.0.1
- [ ] `http://101.43.127.178/` 在浏览器能正常打开
- [ ] journalctl 三个服务都无 ERROR

---

## 15. 后续优化路线(可选,不影响本次 demo)

1. **域名 + HTTPS**:1Panel 网站 → 申请证书 → 强制 HTTPS
2. **scanner 真实链监听**:补 ETH/BSC RPC key,启用 scanner systemd 单元
3. **合约 Sepolia 部署**:`forge script` 部署到测试网,把合约地址写进 frontend
4. **Prometheus 监控**:加 `/metrics` 抓取
5. **CI/CD**:GitHub Actions 编译 → 推送到服务器 → systemd restart
6. **outbox 模式**:`deposit_confirmed` 走 outbox 表,解决 NATS 投递可靠性

---

**版本**:v1.0(2026-06-12 评审)
**前置评审文档**:`BLOCKCHAIN_PROJECTS_IMPLEMENTATION_REVIEW.md`
**评审视角**:候选人本人 / 1Panel 运维 / 面试官演示
