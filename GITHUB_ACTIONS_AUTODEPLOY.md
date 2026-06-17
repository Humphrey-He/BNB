# BNB GitHub Actions 自动化发布说明

## 目标

把当前“本地编译/打包 -> 手工 `scp` -> 远端重启服务”的流程，改成：

1. `push` 代码到 GitHub
2. GitHub Actions 自动构建
3. 自动上传服务器
4. 自动刷新 `systemd` 和 OpenResty

---

## 当前落地状态

项目里已经新增：

- GitHub Actions 工作流: `.github/workflows/deploy-bnb.yml`
- 远端无交互部署脚本: `scripts/deploy_remote_ci.sh`

这两部分已经把自动化主流程准备好了。

当前还**不能直接自动执行**的原因只有两个：

1. 当前仓库还没有配置 Git remote
2. GitHub 仓库 Secrets / Variables 还没有接入

也就是说：

- 自动化方案已经在项目里
- 你只差把代码推到 GitHub，并在仓库里填好 Secrets

---

## 自动化发布流程设计

### 构建发生在 GitHub Actions

GitHub Actions 会在 Ubuntu Runner 上做这些事情：

- 构建前端 `portfolio-frontend/app`
- 构建 Go API `web3-blockchain-backend-engineer`
- 构建 Rust 节点 `protocol-rust-blockchain-engineer`
- 组装一个发布包 `bnb-release-<sha>.tar.gz`

### 部署发生在服务器

GitHub Actions 会把发布包和部署脚本传到服务器，然后远端脚本会：

- 更新前端静态文件到 `/opt/1panel/www/bnb/frontend`
- 更新 Go 二进制到 `/home/ubuntu/opt/bnb/api`
- 更新 Rust 二进制到 `/home/ubuntu/opt/bnb/node/target/release`
- 重写 `.env`
- 刷新 `systemd`
- 重载 1Panel OpenResty
- 运行基础健康检查

---

## 你需要在 GitHub 做的接入

### 1. 把仓库推到 GitHub

当前本地仓库没有 Git remote。

你需要先创建 GitHub 仓库，然后在本地执行：

```bash
git remote add origin <你的 GitHub 仓库地址>
git push -u origin main
```

如果默认分支不是 `main`，记得同步修改 `.github/workflows/deploy-bnb.yml` 里的分支名。

### 2. 配置 GitHub Secrets

进入 GitHub 仓库：

- `Settings`
- `Secrets and variables`
- `Actions`

添加这些 **Secrets**：

| 名称 | 说明 |
|------|------|
| `DEPLOY_SSH_PRIVATE_KEY` | 用于登录服务器的 SSH 私钥 |
| `REMOTE_SUDO_PASSWORD` | 服务器 `ubuntu` 账号的 sudo 密码 |
| `POSTGRES_DB` | 数据库名，当前一般是 `asset_platform` |
| `POSTGRES_USER` | PostgreSQL 用户 |
| `POSTGRES_PASSWORD` | PostgreSQL 密码 |
| `REDIS_PASSWORD` | Redis 密码；如果没有可留空 |

### 3. 配置 GitHub Variables

同一个页面里添加这些 **Variables**：

| 名称 | 推荐值 |
|------|--------|
| `DEPLOY_HOST` | `101.43.127.178` |
| `DEPLOY_USER` | `ubuntu` |
| `DEPLOY_ROOT` | `/home/ubuntu/opt/bnb` |
| `FRONTEND_WEB_ROOT` | `/opt/1panel/www/bnb/frontend` |
| `NGINX_CONF_PATH` | `/opt/1panel/www/conf.d/bnb.conf` |
| `BNB_SERVER_NAME` | `bnb.shuhong.icu 101.43.127.178` |
| `GO_API_PORT` | `8080` |
| `RUST_RPC_PORT` | `8081` |
| `NATS_URL` | `nats://127.0.0.1:4222` |
| `APP_ENV` | `production` |
| `LOG_LEVEL` | `info` |
| `POSTGRES_HOST` | `127.0.0.1` |
| `POSTGRES_PORT` | `5432` |
| `REDIS_HOST` | `127.0.0.1` |
| `REDIS_PORT` | `6379` |

---

## 服务器侧需要的前置条件

自动化发布默认依赖这些条件已经存在：

- 服务器 SSH 可登录
- `ubuntu` 用户有 sudo 权限
- 1Panel OpenResty 容器已经安装并正常运行
- PostgreSQL / Redis / NATS 已经准备好
- `/home/ubuntu/opt/bnb/` 已存在

如果你后面改了服务器路径，记得同步更新 GitHub Variables。

---

## SSH 私钥怎么准备

如果你还没有专门给 GitHub Actions 用的密钥，推荐单独生成一把：

```bash
ssh-keygen -t ed25519 -C "github-actions-bnb-deploy" -f ~/.ssh/bnb_github_actions
```

然后：

- 把公钥 `bnb_github_actions.pub` 追加到服务器 `~/.ssh/authorized_keys`
- 把私钥内容填进 GitHub Secret `DEPLOY_SSH_PRIVATE_KEY`

---

## 触发方式

当前工作流支持两种触发：

1. 推送到 `main`
2. GitHub Actions 页面手动点击 `Run workflow`

---

## 当前工作流会发布哪些内容

### 前端

来源：

- `portfolio-frontend/app`

产物：

- `dist/`

部署到：

- `/opt/1panel/www/bnb/frontend`

### Go API

来源：

- `web3-blockchain-backend-engineer`

产物：

- `api-server-linux-amd64`
- `scripts/migrations/`

部署到：

- `/home/ubuntu/opt/bnb/api`

### Rust 节点

来源：

- `protocol-rust-blockchain-engineer`

产物：

- `verifiable-rust-chain-node`

部署到：

- `/home/ubuntu/opt/bnb/node/target/release`

---

## 和你当前手工发布方式的区别

### 当前手工方式

- 本地 build
- 手工 `scp`
- 远端手工执行脚本

### 新自动化方式

- `push` 代码
- GitHub 自动 build
- 自动上传
- 自动刷新服务

收益：

- 减少漏文件、漏重启
- 不再依赖本机 Windows 构建环境
- Rust / Go / Frontend 都统一在 Linux Runner 构建

---

## 注意事项

### 1. 远端脚本默认会重写 `.env`

所以数据库和 Redis 配置应当通过 GitHub Secrets / Variables 管理。

### 2. Rust 节点仍然会监听 `0.0.0.0:8081`

这是为了兼容 1Panel OpenResty 容器反代。

因此：

- `8081` 不要在云防火墙里对公网开放

### 3. Go API 当前仍监听 `*:8080`

因此：

- `8080` 同样不要在云防火墙里开放

### 4. 第一次启用前建议先手动跑一次 workflow_dispatch

这样更容易观察问题，不建议第一次就靠 `push` 直接发生产。

---

## 建议的启用顺序

1. 创建 GitHub 仓库并推送代码
2. 配置 SSH 密钥
3. 配置 GitHub Secrets / Variables
4. 手动运行一次 `Deploy BNB`
5. 看 Actions 日志和远端服务状态
6. 验证：
   - `https://bnb.shuhong.icu/`
   - `https://bnb.shuhong.icu/api/v1/chain-status`
   - `https://bnb.shuhong.icu/node/get_block_number`
7. 确认无误后，再把它当正式发布流程使用

---

## 当前还没替你完成的部分

由于这部分依赖外部平台接入，我现在还不能替你直接完成：

1. 创建 GitHub 远程仓库
2. 把本地仓库 push 到 GitHub
3. 在 GitHub 后台填写 Secrets / Variables
4. 生成并安装 GitHub Actions 专用 SSH 公钥

但这些前置一旦补齐，项目里的工作流和远端部署脚本就可以直接接上。

---

## 你下一步最小动作

先做这两件事：

1. 创建 GitHub 仓库并把当前代码 push 上去
2. 告诉我你是否要我继续为你整理一份“GitHub Secrets/Variables 填写清单”

如果你愿意，我下一步可以继续直接给你一份：

- **逐项照抄的 GitHub Secrets / Variables 填表模板**
- **GitHub Actions 首次启用检查清单**
