# PRE_DEPLOY.md

> `deploy.sh` 跑**之前**,你要在 1Panel 面板里手动做的事。**所有这些都是纯鼠标/输入框操作**,不需要敲任何命令。
> 完成后把里面提到的占位符(`__XXX__`)替换成实际值,粘到 `deploy.sh` 第一段 `CONFIG` 里。
> 预计耗时:**5-10 分钟**(就装俩数据库 + 建个表)。

---

## 第 0 步:拿到服务器 SSH 入口

两种方式二选一:

- **方式 A:腾讯云控制台** → 轻量服务器 → 实例 → **登录** → "Workbench/VNC 一键登录"
- **方式 B:本机 PowerShell** → `ssh root@101.43.127.178`(需要你提前把本机公钥加到服务器的 `~/.ssh/authorized_keys`)

确认能 SSH 上后,继续。

---

## 第 1 步:在 1Panel 装 PostgreSQL 15(3 分钟)

1. 1Panel 面板(默认 `http://101.43.127.178:10086`)→ 左侧 **应用商店**
2. 搜索 `PostgreSQL` → 选 **15.x** → 点 **安装**
3. 安装参数全部默认,**端口 5432**
4. 等 1-2 分钟装完

> 1Panel 不会自动建数据库和用户,你需要手动建。

## 第 2 步:在 1Panel 建数据库和用户(2 分钟)

1. 1Panel 左侧菜单 → **数据库**
2. 点 **"添加数据库"**:
   - **数据库名**:`asset_platform`
   - **用户名**:`asset_platform`(1Panel 会自动用这个名建 PG 用户,密码你自设)
   - **密码**:`__PG_PASSWORD__` ← **记下来,待会粘到 deploy.sh**
   - **访问权限**:本地访问(`127.0.0.1`)
3. 点确定,等几秒

建完确认:点 "连接信息" 或 "管理" 按钮,1Panel 会显示一段 `psql` 命令,大致是:
```
PGPASSWORD=__PG_PASSWORD__ psql -h 127.0.0.1 -U asset_platform -d asset_platform
```
**把这段命令里真实的 `host` `user` `db` `password` 都记下来**,待会要用。

> ⚠️ **关键**:1Panel 装的 PG 实际监听地址是 `127.0.0.1` 还是 `0.0.0.0`?如果你不需要外网连,1Panel 默认就是 127.0.0.1,这是对的。如果装了之后用 `ss -lntp | grep 5432` 看到 `0.0.0.0:5432`,**必须在 PG 配置文件里加 `listen_addresses = '127.0.0.1'`** 重启。

## 第 3 步:在 1Panel 装 Redis 7(2 分钟)

1. 1Panel 左侧 → **应用商店** → 搜 `Redis` → 装 **7.x**
2. 端口默认 `6379`
3. **密码**:**必填**(`deploy.sh` 跑的时候会 `read -s` 提示你输入,**不会**明文存到任何 git 跟踪文件)
4. 装完不用动,**记下你设的密码**(建议放 1Password / Bitwarden)

## 第 4 步:在 1Panel 装 Docker(1 分钟,NATS 用)

1. 1Panel → **应用商店** → 搜 `Docker` → 装
2. 装完不用启动任何东西,NATS 用 docker run 起

> 如果服务器已经装过 docker,跳过这步。

## 第 5 步:确认 OpenResty 已装(1 分钟)

1. 1Panel → **网站** → 看 "运行环境" 是不是 OpenResty
2. 1Panel 默认装 OpenResty,正常情况下**已经在了**
3. 不用动任何东西

如果没装:
1. → **应用商店** → 搜 `OpenResty` → 装
2. 装完不用建站点,deploy.sh 会自己写 bnb.conf

## 第 6 步:在本机编译好产物并打包成 tar(10 分钟)

SSH 之外的事,在你 Windows 本机 PowerShell 里完成。

```powershell
# ====== 6.1 编译 Go API ======
cd E:\awesomeProject\BNB\web3-blockchain-backend-engineer
$env:GOOS="linux"; $env:GOARCH="amd64"
go build -o api-server-linux-amd64 ./cmd/api-server

# ====== 6.2 打包 frontend dist ======
cd E:\awesomeProject\BNB\portfolio-frontend\app
npm install
npm run build

# ====== 6.3 准备 Rust 源码(在服务器重编) ======
# 这步不需要编译,只需要把源码目录准备好,待会 scp 上传
# 注意:不要传 target/ 和 Cargo.lock,会大幅拖慢上传

# ====== 6.4 打 release 包(给服务器) ======
# 下面 4 个文件/目录打好包,deploy.sh 会从 /opt/bnb/release/ 读
mkdir E:\bnb-release
Copy-Item E:\awesomeProject\BNB\web3-blockchain-backend-engineer\api-server-linux-amd64 E:\bnb-release\
Copy-Item -Recurse E:\awesomeProject\BNB\web3-blockchain-backend-engineer\scripts E:\bnb-release\go-scripts
Copy-Item -Recurse E:\awesomeProject\BNB\portfolio-frontend\app\dist E:\bnb-release\frontend-dist
Copy-Item -Recurse E:\awesomeProject\BNB\protocol-rust-blockchain-engineer\src E:\bnb-release\rust-src
Copy-Item E:\awesomeProject\BNB\protocol-rust-blockchain-engineer\Cargo.toml E:\bnb-release\

# 打成一个 tar
cd E:\
tar -czf bnb-release.tar.gz bnb-release
```

最后产出:`E:\bnb-release.tar.gz`(约 5-10MB,不含 node_modules / target)

把这个 tar 传上服务器(下面第 7 步)。

## 第 7 步:把 release 包传到服务器(1 分钟)

在**本机 PowerShell**:

```powershell
scp E:\bnb-release.tar.gz root@101.43.127.178:/opt/bnb/release.tar.gz
```

> 没有 scp?装个 Git for Windows 用 `C:\Program Files\Git\usr\bin\scp.exe`。
> 不想配 SSH key?用 1Panel 终端 + base64:
> 1. PowerShell:`[Convert]::ToBase64String([IO.File]::ReadAllBytes("E:\bnb-release.tar.gz")) | Out-File b64.txt`
> 2. 把 b64.txt 内容贴到 1Panel 终端
> 3. 服务器:`base64 -d > /opt/bnb/release.tar.gz`(粘贴后 Ctrl+D 结束)

---

## 第 8 步:确认信息、填到 deploy.sh

现在你应该收集到这些信息了:

| 占位符 | 含义 | 来源 |
|---|---|---|
| `__PG_HOST__` | 127.0.0.1 | 1Panel 数据库面板 |
| `__PG_PORT__` | 5432 | 默认 |
| `__PG_DB__` | asset_platform | 第 2 步设的 |
| `__PG_USER__` | asset_platform | 第 2 步设的 |
| `__PG_PASSWORD__` | (你设的密码) | 第 2 步设的 |
| `__REDIS_HOST__` | 127.0.0.1 | 1Panel Redis |
| `__REDIS_PORT__` | 6379 | 默认 |
| `__SERVER_IP__` | 101.43.127.178 | 你的服务器 |
| `__RELEASE_TAR__` | /opt/bnb/release.tar.gz | 第 7 步传的 |

---

## 第 9 步:腾讯云安全组(1 分钟)

1. 腾讯云控制台 → 轻量应用服务器 → 你的实例 → **防火墙** → **添加规则**
2. 加 3 条:

| 协议 | 端口 | 来源 | 备注 |
|---|---|---|---|
| TCP | 80 | 0.0.0.0/0 | HTTP |
| TCP | 443 | 0.0.0.0/0 | HTTPS(以后用) |
| TCP | 22 | 你的固定 IP | SSH(临时可写 0.0.0.0/0) |

> 其它端口 5432/6379/4222/8080/8081 一律不开。

---

## 第 10 步:SSH 上去跑 deploy.sh

```bash
ssh root@101.43.127.178
cd /opt/bnb
tar -xzf release.tar.gz
# 此时 /opt/bnb/release/ 下有 release 的产物
# 把 deploy.sh 脚本先 scp 上来
# 然后
bash deploy.sh
```

`deploy.sh` 会提示你输入 `__PG_PASSWORD__`,**用 read -s(隐藏输入)**。

预计 10-15 分钟,跑完会输出:
```
✅ deploy finished
下一步:bash /opt/bnb/verify.sh
```

---

## 第 11 步:浏览器验证

打开 `http://101.43.127.178/`,应看到 frontend 首页。

如果看不到,跑:
```bash
bash /opt/bnb/verify.sh
```
它会逐项检测并告诉你哪一项挂了。

---

## 常见 PRE_DEPLOY 阶段问题

| 问题 | 解决 |
|---|---|
| 1Panel 装 PG 失败 | 检查磁盘 `df -h`,1Panel 默认数据盘可能没挂载 |
| 1Panel 装 Redis 失败 | 同上 |
| SSH 上不去 | 腾讯云控制台 → 防火墙 → 放行 22;或 1Panel 的"主机 → 终端" |
| scp 报密码错 | 服务器重设 root 密码;或换 1Panel 终端 + base64 |
| tar 解压报 "Cannot open: Permission denied" | 服务器上 `chown -R root:root /opt/bnb` |
| Rust 编译 `cargo: command not found` | 服务器上没装 Rust,见 deploy.sh 步骤 7 |

---

**PRE_DEPLOY 完成,继续 deploy.sh。**
