# BNB GitHub Actions 自动发布说明

## 当前方案

当前项目已经从：

- GitHub 公共 runner 构建
- 通过 SSH / `scp` 上传服务器

切换为：

- GitHub Actions 触发
- 服务器上的 self-hosted runner 本机构建
- 服务器上的 self-hosted runner 本机部署

这是因为：

- 你的服务器本地 SSH 正常
- 但 GitHub 公共 runner 到国内云主机 `22` 端口链路不稳定
- 因此长期最稳的做法是让服务器自己跑 runner

## 仓库里现在有哪些文件

- workflow：`.github/workflows/deploy-bnb.yml`
- 本机部署脚本：`scripts/deploy_self_hosted.sh`
- runner 安装说明：`SELF_HOSTED_RUNNER_SETUP.md`

## 触发逻辑

支持两种方式：

1. push 到 `main`
2. GitHub Actions 页面手动点 `Run workflow`

## 当前 workflow 做什么

在服务器本机上依次执行：

1. 校验变量
2. 校验本机工具链
3. 构建前端
4. 构建 Go API
5. 构建 Rust 节点
6. 调用 `scripts/deploy_self_hosted.sh`
7. 覆盖前端、二进制、systemd、OpenResty 配置
8. 重启服务并做健康检查

## 不再需要的旧配置

由于不再通过 SSH 从 GitHub 公共 runner 上传，所以这几个配置可以删除：

### 不再需要的 Secret

- `DEPLOY_SSH_PRIVATE_KEY`

### 不再需要的 Variable

- `DEPLOY_HOST`
- `DEPLOY_USER`

## 仍然需要的 Secrets

- `REMOTE_SUDO_PASSWORD`
- `POSTGRES_DB`
- `POSTGRES_USER`
- `POSTGRES_PASSWORD`
- `REDIS_PASSWORD`（没有可留空）

## 仍然需要的 Variables

- `DEPLOY_ROOT=/home/ubuntu/opt/bnb`
- `FRONTEND_WEB_ROOT=/opt/1panel/www/bnb/frontend`
- `NGINX_CONF_PATH=/opt/1panel/www/conf.d/bnb.conf`
- `BNB_SERVER_NAME=bnb.shuhong.icu 101.43.127.178`
- `GO_API_PORT=8080`
- `RUST_RPC_PORT=8081`
- `NATS_URL=nats://127.0.0.1:4222`
- `APP_ENV=production`
- `LOG_LEVEL=info`
- `POSTGRES_HOST=127.0.0.1`
- `POSTGRES_PORT=5432`
- `REDIS_HOST=127.0.0.1`
- `REDIS_PORT=6379`

## 你现在需要做的事

最关键的是先把 self-hosted runner 接上：

1. 去 GitHub 仓库 `Settings -> Actions -> Runners`
2. 新建 Linux x64 self-hosted runner
3. 在服务器执行 `SELF_HOSTED_RUNNER_SETUP.md` 里的命令
4. 确认 runner 状态变成 `Idle`
5. 手动运行一次 `Deploy BNB`

## 首次执行前的服务器要求

服务器本机需要具备：

- `node`
- `npm`
- `go`
- `cargo`
- `docker`
- `sudo`

因为这版构建不再在 GitHub 公共 runner 上做，而是在你自己的云服务器上本地做。

## 推荐顺序

1. 安装 self-hosted runner
2. 检查本机工具链
3. 确认 GitHub Secrets / Variables
4. 手动跑一次 workflow
5. 验证：
   - `http://bnb.shuhong.icu/health`
   - `http://bnb.shuhong.icu/api/v1/chain-status`
   - `http://bnb.shuhong.icu/node/get_block_number`

## 一句话总结

这版自动发布已经不再依赖 GitHub 公共 runner SSH 进服务器，而是改成服务器自己跑 GitHub Actions，这会更适合你当前的腾讯云部署环境。
