# 单公网 IP + 单域名部署方案

> 适用场景：一台服务器、一个域名，部署博客 (Mak's Blog)、BNB 前端、Golang 后端、Rust Node

---

## 一、架构概览

```
                        DNS 解析
                             │
              ┌──────────────┼──────────────┐
              ▼              ▼              ▼
        blog.xxx.com   bnb.xxx.com    node.xxx.com
              │              │              │
              └──────────────┼──────────────┘
                             ▼
                    ┌─────────────────┐
                    │   Nginx (80/443) │
                    └─────────────────┘
                             │
          ┌──────────────────┼──────────────────┐
          ▼                  ▼                  ▼
    Next.js (3000)    Golang (8080)     Rust Node HTTP (26657)
                                              │
                                        Rust Node P2P (26656)
                                        (直接暴露，不走 Nginx)
```

---

## 二、DNS 解析配置

在域名服务商控制台添加以下 A 记录：

| 记录类型 | 主机记录 | 值 (公网IP) | 说明 |
|---------|---------|-------------|------|
| A | blog | 你的公网IP | 博客 |
| A | bnb | 你的公网IP | BNB 项目 |
| A | node | 你的公网IP | Rust Node |

> **提示**：如果域名在国内备案，记得先完成备案才能使用 80/443 端口

---

## 三、安装依赖

### 3.1 安装 Nginx

```bash
# Ubuntu / Debian
sudo apt update
sudo apt install nginx -y

# CentOS / RHEL
sudo yum install nginx -y
```

### 3.2 安装 Certbot (SSL 证书)

```bash
# Ubuntu / Debian
sudo apt install certbot python3-certbot-nginx -y

# CentOS
sudo yum install certbot python3-certbot-nginx -y
```

---

## 四、Nginx 配置

### 4.1 主配置文件

```bash
sudo nano /etc/nginx/sites-available/blog-aggregated
```

写入以下内容：

```nginx
# ============================================
# 博客 - Mak's Blog (Next.js)
# ============================================
server {
    listen 80;
    server_name blog.你的域名.com;

    # Next.js 静态资源缓存
    location /_next/static {
        proxy_pass http://127.0.0.1:3000;
        proxy_cache_valid 200 60m;
        add_header Cache-Control "public, max-age=31536000, immutable";
    }

    # 主要代理
    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }

    # 上传文件大小限制
    client_max_body_size 100M;
}

# ============================================
# BNB 项目 - 前端 + Golang 后端
# ============================================
server {
    listen 80;
    server_name bnb.你的域名.com;

    # 静态资源
    location /static/ {
        proxy_pass http://127.0.0.1:3001;
        proxy_cache_valid 200 60m;
        add_header Cache-Control "public, max-age=31536000";
    }

    # API 请求代理到 Golang 后端
    location /api/ {
        proxy_pass http://127.0.0.1:8080/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # 前端页面
    location / {
        proxy_pass http://127.0.0.1:3001;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}

# ============================================
# Rust Node - HTTP API (可选)
# ============================================
server {
    listen 80;
    server_name node.你的域名.com;

    location / {
        proxy_pass http://127.0.0.1:26657;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### 4.2 启用配置

```bash
# 创建软链接
sudo ln -s /etc/nginx/sites-available/blog-aggregated /etc/nginx/sites-enabled/

# 删除默认配置（避免冲突）
sudo rm /etc/nginx/sites-enabled/default

# 测试配置
sudo nginx -t

# 重载 Nginx
sudo systemctl reload nginx
```

---

## 五、配置 SSL 证书

```bash
sudo certbot --nginx \
  -d blog.你的域名.com \
  -d bnb.你的域名.com \
  -d node.你的域名.com
```

按提示完成配置，Certbot 会自动修改 Nginx 配置启用 HTTPS。

### 证书自动续期验证

```bash
# 测试续期
sudo certbot renew --dry-run
```

---

## 六、防火墙配置

### 6.1 UFW (Ubuntu)

```bash
# 开放必要端口
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS

# Rust Node P2P 端口 (直接暴露，不走 Nginx)
sudo ufw allow 26656/tcp # P2P 通信

# 启用防火墙
sudo ufw enable
sudo ufw status
```

### 6.2 云服务器安全组

如果你使用云服务器 (阿里云/腾讯云/AWS)，还需要在控制台开放：

| 端口 | 协议 | 用途 |
|-----|------|-----|
| 80 | TCP | HTTP |
| 443 | TCP | HTTPS |
| 26656 | TCP | Rust Node P2P |

---

## 七、启动服务

### 7.1 启动脚本 (start-all.sh)

```bash
#!/bin/bash

# 博客 (Next.js)
echo "Starting Mak's Blog..."
cd /path/to/Mak-s-Bolg-remote
npm run build
npm start &
BLOG_PID=$!

# BNB 前端 (端口 3001)
echo "Starting BNB Frontend..."
cd /path/to/bnb-frontend
npm run build
npm start &
BNB_FRONTEND_PID=$!

# Golang 后端 (端口 8080)
echo "Starting BNB Backend..."
cd /path/to/golang-backend
go run main.go &
BACKEND_PID=$!

# Rust Node
echo "Starting Rust Node..."
cd /path/to/rust-node
cargo run --release &
RUST_PID=$!

echo "All services started!"
echo "Blog PID: $BLOG_PID"
echo "BNB Frontend PID: $BNB_FRONTEND_PID"
echo "Backend PID: $BACKEND_PID"
echo "Rust Node PID: $RUST_PID"

# 等待所有进程
wait
```

### 7.2 使用 systemd 管理 (推荐)

#### 创建博客服务

```bash
sudo nano /etc/systemd/system/mak-blog.service
```

```ini
[Unit]
Description=Mak's Blog (Next.js)
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/path/to/Mak-s-Bolg-remote
ExecStart=/usr/bin/node .next/standalone/server.js
Restart=on-failure
RestartSec=10
Environment="NODE_ENV=production"

[Install]
WantedBy=multi-user.target
```

#### 创建 BNB 后端服务

```bash
sudo nano /etc/systemd/system/bnb-backend.service
```

```ini
[Unit]
Description=BNB Golang Backend
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/path/to/golang-backend
ExecStart=/path/to/golang-backend/main
Restart=on-failure
RestartSec=10
Environment="RUST_LOG=info"

[Install]
WantedBy=multi-user.target
```

#### 启动服务

```bash
# 重载 systemd
sudo systemctl daemon-reload

# 启用服务
sudo systemctl enable mak-blog
sudo systemctl enable bnb-backend

# 启动服务
sudo systemctl start mak-blog
sudo systemctl start bnb-backend

# 查看状态
sudo systemctl status mak-blog
sudo systemctl status bnb-backend
```

---

## 八、服务端口汇总

| 服务 | 内部端口 | 对外方式 | 域名 |
|-----|---------|---------|------|
| Mak's Blog | 3000 | Nginx 代理 | blog.xxx.com |
| BNB 前端 | 3001 | Nginx 代理 | bnb.xxx.com |
| Golang 后端 | 8080 | Nginx 代理 | bnb.xxx.com/api |
| Rust Node HTTP | 26657 | Nginx 代理 | node.xxx.com |
| Rust Node P2P | 26656 | 直接暴露 | 无域名，节点间直连 |

---

## 九、验证部署

### 9.1 本地测试

```bash
# 测试 HTTP
curl http://localhost:3000
curl http://localhost:8080/health
curl http://localhost:26657/health

# 测试 Nginx 代理
curl -I http://blog.你的域名.com
curl -I http://bnb.你的域名.com/api/health
```

### 9.2 检查日志

```bash
# Nginx 日志
sudo tail -f /var/log/nginx/access.log
sudo tail -f /var/log/nginx/error.log

# 服务日志
sudo journalctl -u mak-blog -f
sudo journalctl -u bnb-backend -f
```

---

## 十、故障排查

### 常见问题

1. **502 Bad Gateway**
   - 检查后端服务是否启动：`curl http://localhost:3000`
   - 检查 Nginx 错误日志：`sudo tail /var/log/nginx/error.log`

2. **连接被拒绝**
   - 检查防火墙：`sudo ufw status`
   - 检查云服务器安全组

3. **SSL 证书失效**
   - 续期：`sudo certbot renew`
   - 查看续期定时任务：`systemctl list-timers`

---

## 十一、后续维护

### 证书自动续期

Certbot 会自动创建定时任务，但可以手动验证：

```bash
sudo certbot renew
```

### 更新服务

```bash
# 更新博客
cd /path/to/Mak-s-Bolg-remote
git pull
npm install
npm run build
sudo systemctl restart mak-blog

# 更新后端
cd /path/to/golang-backend
git pull
go build -o main
sudo systemctl restart bnb-backend
```

---

## 十二、简化版配置 (仅保留博客)

如果暂时只部署博客，可用简化配置：

```nginx
server {
    listen 80;
    server_name blog.你的域名.com;

    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }

    client_max_body_size 100M;
}
```

---

*文档更新时间：2026-06-16*
