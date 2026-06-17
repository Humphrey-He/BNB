# Smart Contract Engineer Project

目标：建立一个安全导向的智能合约作品集项目，用来证明你不仅能写 Solidity，还理解 vault 设计、权限边界、测试策略、审计流程和链下协作。

推荐项目名：

`secure-yield-vault`

一句话定位：

> 一个 ERC-4626 风格的安全收益金库，覆盖 shares/assets 模型、存取款、奖励结算、角色权限、timelock、暂停、emergency withdraw、Foundry fuzz/invariant test、Slither 检查和审计报告。

为什么从普通 staking 升级成 vault：

- 2026 年合约岗位更关注 DeFi、安全、审计和测试深度，普通 staking 太容易显得像教程项目。
- ERC-4626 风格的 shares/assets 模型更接近真实 DeFi vault、收益聚合器和资产管理协议。
- 你的目标不是把自己包装成纯 Solidity 工程师，而是展示“后端工程师 + 安全意识 + 链上/链下协作”的组合优势。

建议技术栈：

- Solidity ^0.8.20
- Foundry
- OpenZeppelin
- Slither
- Foundry invariant test
- Echidna 可选
- Python/Go 事件监听和数据校验脚本

核心产物：

- Vault 合约源码
- 单元测试、fuzz test、invariant test
- 攻击 PoC 测试：重入、权限、价格操纵占位、舍入误差
- `SECURITY.md`
- `AUDIT_REPORT.md`
- Slither 报告
- 部署脚本
- 链下事件监听脚本
- README 中的 threat model 和 known limitations

详细分析见 [ANALYSIS.md](./ANALYSIS.md)。
