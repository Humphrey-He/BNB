# BNB 正式主仓整合说明

## 当前结论

如果这个 GitHub 仓库要作为 **正式主仓**，现在还不能直接把所有目录一次性 `git add .` 推上去。

原因不是代码有问题，而是当前本地目录里存在 **嵌套 Git 仓库**：

- `protocol-rust-blockchain-engineer/.git`
- `smart-contract-engineer/.git`
- `web3-blockchain-backend-engineer/.git`

这会导致主仓在 `git add` 这些目录时，把它们当成 **embedded repository / gitlink**，而不是普通源码目录。

如果直接提交，结果通常不是你想要的“单一正式主仓源码”，而是：

- 主仓里只记录一个 gitlink 指针
- 子项目源码不会完整进入主仓
- GitHub 上也不会变成一个真正可浏览的统一代码仓

---

## 推荐方案

正式主仓采用：

- **一个主仓**
- **完整收录源码**
- **不保留子项目独立 `.git` 目录**

也就是：

- 保留目录结构
- 去掉子项目里的嵌套 Git 元数据
- 让主仓统一管理版本

这是最适合你当前诉求的方案，因为你要的是：

- 一个正式 GitHub 主仓
- GitHub Actions 直接从主仓构建和部署
- 不是多仓库分裂，也不是 submodule 管理

---

## 现在已经完成的安全准备

主仓里已经完成：

1. 自动化发布工作流已推到远程 `main`
2. 根 `.gitignore` 已补强，避免把这些内容误收入主仓：
   - 前端 `node_modules/`、`dist/`
   - Rust `target/`
   - Go / 脚本二进制
   - 运行日志
   - `bnb-release/` 和发布压缩包

这意味着：

- 下一步整合源码时，不会把明显的构建产物和发布垃圾一并推上去

---

## 应该纳入正式主仓的内容

推荐纳入：

- `portfolio-frontend/`
- `web3-blockchain-backend-engineer/`
- `protocol-rust-blockchain-engineer/`
- `smart-contract-engineer/`
- `scripts/`
- `.github/`
- `README.md`
- `DEPLOYMENT.md`
- `DEPLOYMENT_STATUS.md`
- 架构/部署/验收类文档

建议不纳入：

- `node_modules/`
- `dist/`
- `target/`
- `out/`
- `cache/`
- `broadcast/`
- 本地编译二进制
- 临时 relay / probe 日志
- 上传服务器用的 release 目录和压缩包

---

## 整合前必须确认的一件事

### 是否保留子项目自己的 Git 历史

当前三个子项目各自有独立 `.git`。

如果你继续走“正式主仓整合”方案，就要接受：

- 主仓会统一管理这些子项目
- 子项目自己的独立本地 Git 历史不再继续作为独立仓结构使用

源码还在，目录还在，但它们的“独立仓身份”会被移除。

所以这一步虽然不是复杂技术问题，但属于 **结构性变更**。

---

## 最稳的正式整合步骤

### 第 1 步：先备份子项目 Git 元数据

建议先把这几个目录各自备份一份：

- `protocol-rust-blockchain-engineer/.git`
- `smart-contract-engineer/.git`
- `web3-blockchain-backend-engineer/.git`

你可以单独打包，或者复制到别的目录保存。

### 第 2 步：移除嵌套 `.git`

移除后，这些目录就会变成主仓里的普通子目录源码。

### 第 3 步：在主仓检查状态

执行：

```bash
git status --short
```

这时你看到的应当是普通文件新增，而不是 embedded repository 警告。

### 第 4 步：分批提交

不要第一次就 `git add .`

建议按模块分批：

1. `portfolio-frontend`
2. `web3-blockchain-backend-engineer`
3. `protocol-rust-blockchain-engineer`
4. `smart-contract-engineer`
5. `scripts` 和文档

### 第 5 步：推到远程 `main`

等每一批确认无误后再推送。

---

## 为什么不推荐直接用 submodule

你当前目标是：

- 一个正式主仓
- 一个仓库里统一展示项目能力
- 一个仓库里跑 GitHub Actions 自动部署

如果改成 submodule，会带来这些额外复杂度：

- 克隆后还要初始化 submodule
- GitHub Actions 也要额外处理 submodule
- 仓库浏览体验更碎
- 对“求职作品主仓”不够友好

所以不推荐。

---

## 对 GitHub Actions 的影响

当前已经推到远程的自动化工作流默认假设：

- 三个项目目录都在 **同一个主仓** 里
- GitHub Runner 能直接读取源码

所以如果你不先完成正式主仓整合，自动化工作流虽然存在，但后续价值会被打折扣。

---

## 下一步建议

最合理的下一步是：

1. 先备份三个子项目的 `.git`
2. 然后移除嵌套 `.git`
3. 再由主仓统一纳管

---

## 当前状态一句话总结

主仓自动化发布能力已经接上远程仓库，但要让这个 GitHub 仓库真正成为完整正式主仓，下一步必须先处理三个子项目的嵌套 `.git` 结构。
