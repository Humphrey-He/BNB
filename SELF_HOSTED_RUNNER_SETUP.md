# BNB Self-Hosted Runner 接入说明

> 适用服务器：`101.43.127.178`
> 系统：Ubuntu 22.04
> 推荐标签：`self-hosted`, `linux`, `x64`, `bnb-prod`

## 目标

把当前 GitHub Actions 从 `GitHub-hosted runner + SSH 上传` 改成：

1. GitHub 触发 workflow
2. 云服务器上的 self-hosted runner 直接拉取仓库
3. 在服务器本机完成构建
4. 在服务器本机完成部署

这样可以避开 GitHub 公共 runner 到国内云主机 `22` 端口不稳定的问题。

## 这版 workflow 的变化

当前仓库已经切换为：

- workflow 文件：`.github/workflows/deploy-bnb.yml`
- 本机部署脚本：`scripts/deploy_self_hosted.sh`

这意味着：

- 不再需要 `DEPLOY_SSH_PRIVATE_KEY`
- 不再需要 `DEPLOY_HOST`
- 不再需要 `DEPLOY_USER`
- 不再依赖 `scp` / `ssh` 上传发布包

保留必须项：

- `REMOTE_SUDO_PASSWORD`
- `POSTGRES_DB`
- `POSTGRES_USER`
- `POSTGRES_PASSWORD`
- `REDIS_PASSWORD`（可空）

## GitHub 页面先做什么

打开仓库：

- `Settings`
- `Actions`
- `Runners`
- `New self-hosted runner`

选择：

- `Linux`
- `x64`

GitHub 会给你一组命令，其中包含：

- 下载 runner 压缩包命令
- `./config.sh --url ... --token ...`

> 注意：`token` 有时效，请在打开页面后尽快执行。

## 服务器安装命令

先 SSH 到服务器，然后执行下面这组命令。

### 1. 创建 runner 目录

```bash
mkdir -p /home/ubuntu/actions-runner
cd /home/ubuntu/actions-runner
```

### 2. 下载 runner

> 版本号以 GitHub 页面实时给出的命令为准，下面只是模板。

```bash
curl -o actions-runner-linux-x64.tar.gz -L https://github.com/actions/runner/releases/download/v2.328.0/actions-runner-linux-x64-2.328.0.tar.gz
tar xzf ./actions-runner-linux-x64.tar.gz
```

### 3. 安装依赖

```bash
sudo ./bin/installdependencies.sh
```

### 4. 配置 runner

把 GitHub 页面给出的 `config.sh` 命令粘进去执行，并加上标签：

```bash
./config.sh \
  --url https://github.com/Humphrey-He/BNB \
  --token <你在 GitHub 页面复制的临时 token> \
  --name bnb-prod-runner \
  --labels bnb-prod \
  --work _work \
  --unattended \
  --replace
```

## 注册成系统服务

```bash
sudo ./svc.sh install ubuntu
sudo ./svc.sh start
```

检查状态：

```bash
sudo ./svc.sh status
```

正常时，你在 GitHub 仓库的 runner 页面会看到：

- 状态 `Idle`
- 标签包含：
  - `self-hosted`
  - `linux`
  - `x64`
  - `bnb-prod`

## 服务器需要具备的本机工具

这版 workflow 会在服务器本机运行这些命令：

- `node`
- `npm`
- `go`
- `cargo`
- `sudo`
- `docker`

快速检查：

```bash
node -v
npm -v
go version
cargo --version
docker --version
```

如果缺少 Node 22，建议补安装：

```bash
curl -fsSL https://deb.nodesource.com/setup_22.x | sudo -E bash -
sudo apt-get install -y nodejs
```

如果缺少 Rust：

```bash
curl https://sh.rustup.rs -sSf | sh -s -- -y
source "$HOME/.cargo/env"
rustup default stable
```

如果缺少 Go，建议安装与仓库 `go.mod` 兼容的版本。

## 变量与 Secret 调整

### 可以删除的 Secrets

- `DEPLOY_SSH_PRIVATE_KEY`

### 可以删除的 Variables

- `DEPLOY_HOST`
- `DEPLOY_USER`

### 仍然需要的 Secrets

- `REMOTE_SUDO_PASSWORD`
- `POSTGRES_DB`
- `POSTGRES_USER`
- `POSTGRES_PASSWORD`
- `REDIS_PASSWORD`

### 仍然需要的 Variables

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

## 首次验证方式

Runner 变成 `Idle` 后：

1. 打开 GitHub 仓库 `Actions`
2. 进入 `Deploy BNB`
3. 点击 `Run workflow`
4. 观察是否顺序通过：
   - `Validate deploy configuration`
   - `Verify local toolchain`
   - `Build frontend`
   - `Build Go API`
   - `Build Rust node`
   - `Run local deployment`

成功后验证：

```bash
curl http://127.0.0.1:8080/healthz
curl http://127.0.0.1:8080/api/v1/chain-status
curl -X POST http://127.0.0.1:8081/rpc/get_block_number
curl http://bnb.shuhong.icu/health
```

## 一句话总结

这版 self-hosted runner 方案的本质是：让服务器自己执行 GitHub Actions，不再依赖 GitHub 公共 runner SSH 进来，因此会比现在稳定很多。
