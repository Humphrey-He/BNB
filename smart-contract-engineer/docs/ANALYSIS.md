# Smart Contract Engineer 分析文档

## 结论

这个方向可以做，但不适合作为你转 Web3 的唯一主线。你当前最强资产是 3 年后端经验、Go/Rust/Python、系统设计和工程落地能力；而纯 Solidity 合约岗更看重合约安全、审计经验、DeFi 攻击面和链上业务抽象。

更合理的定位是：

> Backend-oriented Smart Contract Engineer / Rust Smart Contract Engineer / Solidity + Web3 Backend Engineer

也就是：你不只会写合约，还能把合约、索引器、API、交易生命周期、监控和脚本串成一个完整系统。

## 匹配度

匹配度：中等。

优势：

- Rust 背景可以迁移到 Solana、CosmWasm、Sui/Aptos Move 生态。
- 后端经验能补足很多纯合约工程师欠缺的链下服务能力。
- Python/Go 可以用于部署脚本、事件监听、安全工具集成、数据校验。

短板：

- Solidity 安全知识需要系统补齐。
- DeFi 合约对经济模型、权限边界、预言机、重入、签名、升级模式要求很高。
- 如果没有审计、CTF、Foundry/Hardhat 测试经验，直接投中高级合约岗难度较大。

## 推荐作品集项目

项目名：`secure-yield-vault`

项目目标：

做一个安全导向的收益金库。不要只写一个简单 staking 合约，而是要用 ERC-4626 风格的 shares/assets 模型展示 DeFi 合约设计、安全测试、审计意识和链下协作能力。

核心功能：

- 用户 deposit ERC-20 asset，收到 vault shares。
- 用户 redeem/withdraw 时按 shares/assets 汇率取回资产。
- 支持奖励注入和收益累计。
- 支持领取奖励、退出金库、紧急暂停。
- 管理员调整参数必须经过 timelock 或多角色权限边界。
- 支持 emergency withdraw，并明确牺牲收益或仅提本金的边界。
- 支持事件日志，便于 indexer 解析。
- 提供部署脚本和测试网部署说明。
- 提供 Go 或 Python 事件监听脚本，读取 `Deposited`、`Withdrawn`、`RewardAdded`、`EmergencyWithdrawn` 事件。

建议合约模块：

- `SecureYieldVault.sol`
- `RewardDistributor.sol`
- `TimelockController` 或 OpenZeppelin timelock 集成
- `MockERC20.sol`
- `MaliciousERC20.sol`
- `ReentrantReceiver.sol`
- `interfaces/ISecureYieldVault.sol`

建议测试模块：

- 单元测试：正常 deposit、reward injection、withdraw/redeem、边界条件。
- 权限测试：非管理员不能改参数。
- 安全测试：重入、重复领取、零金额、溢出、暂停状态、舍入误差、fee-on-transfer token 限制。
- 攻击 PoC：恶意 ERC-20、重入 receiver、权限误配。
- 不变量测试：totalAssets 与 shares 关系一致，用户不能赎回超过份额，奖励不会凭空增加。

## 技术栈建议

主栈：

- Solidity
- Foundry
- OpenZeppelin

安全工具：

- Slither
- Foundry fuzz test
- Foundry invariant test
- 可选 Echidna
- OWASP Smart Contract Security Verification Standard / SCSVS 检查清单

链下辅助：

- Python `web3.py` 或 Go `go-ethereum`
- PostgreSQL 或 SQLite 记录事件
- 简单 REST API 查询用户 vault 操作历史

## 学习补齐路线

第 1 周：

- Solidity 基础、ERC-20、事件、modifier、错误处理。
- Foundry：`forge test`、`forge script`、fuzz test。
- OpenZeppelin：Ownable、AccessControl、Pausable、ReentrancyGuard。

第 2 周：

- 常见攻击面：重入、权限失控、价格操纵、签名重放、升级合约风险。
- 阅读典型安全案例：DAO、bZx、Cream、Euler、Nomad。
- 给项目补安全检查清单。

第 3-4 周：

- 完成 `secure-yield-vault`。
- 加上 Slither 报告。
- 写一份 `SECURITY.md`，说明攻击面和缓解方式。
- 写一份 `AUDIT_REPORT.md`，包含 findings、severity、修复说明和 known limitations。

## 简历表达

可以写：

> Built a security-focused ERC-4626-style yield vault with Solidity, Foundry, OpenZeppelin, role-based access control, timelock-protected admin actions, emergency withdrawal, fuzz tests, invariant tests, Slither analysis, audit report, and off-chain event indexing.

中文表达：

> 设计并实现安全导向的 ERC-4626 风格收益金库，覆盖 shares/assets 模型、奖励结算、角色权限、timelock、暂停机制、紧急提现、事件索引、Foundry 单元测试、模糊测试、不变量测试和基础审计报告。

## 求职策略

优先投：

- Solidity + Backend Engineer
- Smart Contract Engineer, Rust preferred
- Solana Smart Contract Engineer
- CosmWasm Engineer
- Web3 Protocol Backend Engineer

谨慎投：

- Senior Solidity Auditor
- DeFi Smart Contract Security Engineer
- 纯合约经济模型设计岗

原因是这些岗位通常要求更深审计经验或真实资金规模项目经验。

## 面试准备重点

你需要能讲清：

- 为什么使用 Checks-Effects-Interactions。
- `ReentrancyGuard` 能防什么，不能防什么。
- `delegatecall` 和升级代理的风险。
- ERC-20 返回值不标准时怎么处理。
- 奖励结算如何避免重复领取。
- 合约事件如何支持链下 indexer。
- 如果链下监听服务漏块或链重组，如何补偿。
- ERC-4626 中 shares/assets 的舍入风险。
- timelock 为什么能降低管理员风险。
- 为什么 fee-on-transfer token、rebasing token 需要特殊处理或明确不支持。
- invariant test 和 fuzz test 的区别。

## 当前市场判断

智能合约岗位仍存在，但对安全和真实项目经验要求较高。2026 年合约相关岗位更强调 Foundry/Hardhat、审计协作、安全设计、DeFi 攻击面、形式化/不变量测试和合约与后端/API 的协作。对你的背景来说，最佳包装不是“只会写 Solidity”，而是“能把合约、安全测试、链下索引和后端服务一起交付”。

参考：

- [Web3.career Rust Smart Contract Jobs](https://web3.career/rust%2Bsmart-contract-jobs)
- [Web3.career Backend Jobs](https://web3.career/backend-jobs)
