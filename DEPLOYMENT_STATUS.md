# BNB 部署状态

> 更新日期：2026-06-17
> 服务器：`101.43.127.178`
> SSH 用户：`ubuntu`
> 项目根目录：`/home/ubuntu/opt/bnb`
> 对外子域名：`bnb.shuhong.icu`
> 主站：`https://www.shuhong.icu/`（当前仍建议保留在 Cloudflare Pages）

## 当前结论

BNB 项目已经完成一版可运行部署，当前建议的线上结构是：

- 主站 `www.shuhong.icu` 继续放在 Cloudflare Pages
- BNB 项目通过 `bnb.shuhong.icu` 挂到云服务器
- 云服务器由 1Panel + OpenResty 承载前端静态文件与反向代理

当前线上服务已可在服务器侧正常运行，域名解析也已切到云服务器；但 `1Panel HTTPS` 证书自动签发还没有最终完成，因此应视为：

- HTTP 链路已打通
- DNS 已切换完成
- HTTPS 仍待在 1Panel 内完成证书申请或导入

## 已完成项

### 1. 服务部署

- 前端静态文件已部署到：`/opt/1panel/www/bnb/frontend`
- Go API 已部署到：`/home/ubuntu/opt/bnb/api`
- Rust 节点已部署到：`/home/ubuntu/opt/bnb/node`

### 2. systemd 服务

当前远端服务状态：

- `asset-platform-api`: `active`
- `chain-node`: `active`

说明：

- Go API 由 `asset-platform-api.service` 托管
- Rust 节点由 `chain-node.service` 托管
- Rust RPC 目前绑定 `0.0.0.0:8081`，目的是让 1Panel 的 OpenResty 容器可通过宿主机 bridge 访问

### 3. OpenResty 反向代理

当前 1Panel/OpenResty 使用路径：

- 站点配置：`/opt/1panel/www/conf.d/bnb.conf`
- 前端目录：`/opt/1panel/www/bnb/frontend`

当前代理关系：

- `/` -> 前端静态站点
- `/api/` -> `http://172.17.0.1:8080`
- `/node/` -> `http://172.17.0.1:8081/rpc/`

### 4. 域名与 DNS

当前推荐使用：

- 主站：`www.shuhong.icu`
- 项目：`bnb.shuhong.icu`

DNS 现状：

- `bnb.shuhong.icu` 已解析到云服务器
- 用户本机曾出现 `198.18.0.x`，这是 Clash/Mihomo fake-ip，本身不代表公网解析错误
- 公共 DNS 实测曾返回：
  - 橙云代理时：Cloudflare Anycast IP
  - 灰云直连时：`101.43.127.178`

当前更适合的部署策略：

- 主站继续保留 Cloudflare Pages，成本低、稳定、无需迁移
- BNB 项目作为子域名走云服务器，便于运行 Go/Rust 服务与 OpenResty 反代

### 5. 云安全组

已确认腾讯云 Lighthouse 安全组当前对公网放行的关键端口包括：

- `22`
- `80`
- `443`
- `8089`
- `8090`

未直接放行：

- `8080`
- `8081`
- `5432`
- `6379`
- `4222`
- `8222`

结论：

- API/RPC/数据库端口虽然宿主机存在监听，但当前没有直接经安全组暴露到公网
- `8089`、`8090` 仍属于管理面暴露，后续建议只允许固定办公 IP 或直接关闭公网访问

### 6. 自动化发布基础

GitHub Actions 自动发布基础文件已经写入仓库：

- `.github/workflows/deploy-bnb.yml`
- `scripts/deploy_remote_ci.sh`
- `GITHUB_ACTIONS_AUTODEPLOY.md`

当前状态：

- 工作流已推送到远程 `main`
- GitHub Actions SSH 部署密钥已在本地生成，并验证可登录服务器
- 但 GitHub 仓库 Secrets / Variables 还需要补齐，补完后才能真正一键发布

## 当前未完成项

### 1. HTTPS 证书

还未完成：

- 在 1Panel 中为 `bnb.shuhong.icu` 成功签发证书

原因：

- 你当前域名托管链路不是 1Panel 内置 DNS 账号即开即用模式
- 如果继续走 Cloudflare，推荐在 1Panel 中配置 Cloudflare API Token 作为 DNS 账号

替代方案：

- 在 Cloudflare 上保持 DNS 托管
- 在 1Panel 里新增 Cloudflare DNS 供应商账号
- 通过 DNS Challenge 为 `bnb.shuhong.icu` 申请证书

### 2. GitHub 正式主仓整合

当前本地项目已经开始向单一正式主仓收口，但还在整理阶段。

已完成：

- 远程仓库：`https://github.com/Humphrey-He/BNB`
- 自动部署工作流已推到远程
- 嵌套子仓库 `.git` 已移除，当前只保留根仓库 `.git`

风险提示：

- 之前做过一次嵌套 `.git` 备份，但备份过程中出现过 `Copy-Item` 路径报错，因此那份备份不应视为绝对完整

### 3. GitHub Actions 可运行性

当前 workflow 设计目标是：

- push 到 `main`
- 自动构建前端
- 自动构建 Go API
- 自动构建 Rust 节点
- 自动上传服务器
- 自动重启服务并校验健康状态

真正启用前，还需要在 GitHub `Settings -> Secrets and variables -> Actions` 中补这些值：

- `DEPLOY_SSH_PRIVATE_KEY`
- `REMOTE_SUDO_PASSWORD`
- `POSTGRES_DB`
- `POSTGRES_USER`
- `POSTGRES_PASSWORD`
- `REDIS_PASSWORD`（可为空）

可选 Variables：

- `DEPLOY_HOST`
- `DEPLOY_USER`
- `DEPLOY_ROOT`
- `FRONTEND_WEB_ROOT`
- `NGINX_CONF_PATH`
- `BNB_SERVER_NAME`

## 建议的正式线上架构

推荐长期保持：

- `www.shuhong.icu` -> Cloudflare Pages
- `bnb.shuhong.icu` -> 腾讯云服务器

优点：

- 主站和项目站职责分离
- 主站继续享受 Pages/CDN/静态托管便利
- BNB 服务端保留服务器可控性，适合 Go/Rust/API/RPC
- 后续新增 `api.`、`admin.`、`demo.` 等子域名也更清晰

不建议现在把主站强行整体迁到同一台云服务器，除非你后续有统一运维、统一日志、统一反代入口的强需求。

## 后续操作顺序

建议下一步按这个顺序继续：

1. 完成 GitHub 正式主仓整合并推送源码
2. 在 GitHub 仓库中补齐 Actions Secrets / Variables
3. 在 1Panel 中配置 Cloudflare DNS 账号并申请 `bnb.shuhong.icu` 证书
4. 验证 `https://bnb.shuhong.icu`
5. 收紧 Lighthouse 安全组中的 `22`、`8089`、`8090`

## 一句话总结

BNB 项目已经完成服务器部署、子域名路线明确、自动发布骨架已接好；当前最关键的剩余工作是：正式主仓收口、GitHub Secrets 补齐、以及 `bnb.shuhong.icu` 的 HTTPS 证书完成签发。
