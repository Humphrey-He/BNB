# ops 目录说明

这个目录专门放 BNB 项目的上线与运维固定资产。

## 目录结构

- `env/`
  - 环境变量模板
- `systemd/`
  - systemd unit 模板

## 推荐阅读顺序

1. [OPS_RUNBOOK.md](/E:/awesomeProject/BNB/OPS_RUNBOOK.md)
2. [ops/env/api.env.example](/E:/awesomeProject/BNB/ops/env/api.env.example)
3. [ops/env/workers.env.example](/E:/awesomeProject/BNB/ops/env/workers.env.example)
4. [ops/systemd/asset-platform-api.service](/E:/awesomeProject/BNB/ops/systemd/asset-platform-api.service)

## 推荐执行顺序

1. 写好 `/home/ubuntu/opt/bnb/api/.env`
2. 写好 `/home/ubuntu/opt/bnb/workers/.env`
3. 写好 `/home/ubuntu/opt/bnb/node/.env`
4. 上传二进制
5. 执行 `sudo bash scripts/install_systemd_units.sh`
6. 执行 `sudo systemctl enable --now ...`
7. 执行 `bash scripts/verify.sh`
