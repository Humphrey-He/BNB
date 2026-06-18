# BNB Self-Hosted Runner 服务器预检

> 检查日期：2026-06-18
> 目标服务器：`101.43.127.178`
> SSH 别名：`bnb-server`

## 结论

服务器已经具备运行 `self-hosted runner` 的一部分条件，但当前还不能直接跑这版 GitHub Actions workflow。

### 已满足
- SSH 正常
- `sudo` 正常
- `docker` 已安装且服务为 `active`
- `cargo` / `rustc` 已安装
- `/home/ubuntu/opt/bnb` 已存在
- `/opt/1panel/www/bnb/frontend` 已存在
- 当前线上服务：
  - `asset-platform-api` = `active`
  - `chain-node` = `active`

### 当前缺失

这 3 项当前没有安装：

- `node`
- `npm`
- `go`

说明：
- 现在的 workflow 已调整为不强制要求预装 Node / Go
- `actions/setup-node` 和 `actions/setup-go` 会在 runner 执行时动态准备对应版本
- 因此它们不是阻塞 self-hosted runner 上线的前置条件

## 实测结果

### 系统与资源
- 系统：Ubuntu 22.04.5 LTS
- 内核：`5.15.0-113-generic`
- 磁盘：`40G`，剩余约 `27G`
- 内存：`3.3Gi`
- Swap：`0`

### 已安装工具
- `cargo 1.75.0`
- `rustc 1.75.0`
- `docker 29.5.3`
- `sudo 1.9.9`
- `python3 3.10.12`

### 缺失工具

- `node`
- `npm`
- `go`

## 可选补装

### 1. 安装 Node 22 与 npm

```bash
curl -fsSL https://deb.nodesource.com/setup_22.x | sudo -E bash -
sudo apt-get install -y nodejs
node -v
npm -v
```

### 2. 安装 Go

如果你想走系统包，可以先看版本是否满足仓库要求；更稳的是手动装官方二进制。

示例：
```bash
cd /tmp
curl -LO https://go.dev/dl/go1.25.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.25.0.linux-amd64.tar.gz
echo 'export PATH=/usr/local/go/bin:$PATH' >> ~/.profile
export PATH=/usr/local/go/bin:$PATH
go version
```

> 如果你后面实际看到仓库要求的 Go 版本不同，以 `web3-blockchain-backend-engineer/go.mod` 为准。

## 已确认可继续的项目
- 可以继续安装 self-hosted runner
- 可以继续保留当前 systemd 与 OpenResty 结构
- 可以继续使用当前部署路径与域名变量

## 下一步建议顺序
1. 先按 `SELF_HOSTED_RUNNER_SETUP.md` 安装 runner
2. 在 GitHub 页面确认 runner 状态为 `Idle`
3. 手动运行一次 `Deploy BNB`
4. 如果后续发现网络无法从 GitHub 下载 Node / Go，再考虑补装本机工具链

## 一句话总结

你的服务器已经具备 self-hosted runner 的核心条件；Node.js 和 Go 当前未预装，但新 workflow 已改为可由 GitHub Actions 在运行时自动准备。
