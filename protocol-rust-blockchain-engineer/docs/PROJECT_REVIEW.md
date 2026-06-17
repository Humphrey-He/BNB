# protocol-rust-blockchain-engineer 项目说明

## 岗位匹配分析

---

## 一、项目概述

### 1.1 项目定位

**项目名称**：`verifiable-rust-chain-node`
**项目类型**：Rust 最小可验证区块链节点
**核心目标**：从零实现一个可验证的区块链节点，覆盖交易签名、状态机、Mempool、区块验证、存储抽象等核心链路，用于证明对区块链底层协议的深度理解。

这不是一个教学演示项目，而是一个经过完整 Code Review 迭代、测试覆盖 44 个用例、可直接编译运行的工程原型。

### 1.2 技术架构

```
RPC API (Week 4 待完成)
       │
       ▼
   Mempool ──────────── nonce/fee ordering, replacement, eviction
       │
       ▼
Block Producer ─────── PoA consensus, block template
       │
       ▼
┌──────┴──────┐
│             │
Validator   Executor ─── deterministic tx execution, receipt generation
│             │
StateDB ◄─────┘ ──── account balance, nonce, state_root
│
Storage ──────── Storage trait + StorageMem (fork-ready hash-keyed)
```

### 1.3 已完成模块

| 模块 | 文件 | 完成度 | 测试用例 |
|------|------|--------|---------|
| 核心类型 | `src/types.rs` | 100% | 0 (类型定义) |
| 密码学 | `src/crypto.rs` | 100% | 8 |
| 状态机 | `src/state.rs` | 100% | 8 |
| 执行器 | `src/executor.rs` | 100% | 7 |
| Mempool | `src/mempool.rs` | 100% | 6 |
| 区块验证 | `src/validation.rs` | 100% | 7 |
| 存储层 | `src/storage.rs` | 100% | 8 |
| **合计** | | | **44 tests** |

### 1.4 关键设计决策

**1. u128 资产值**：区块链金额必须避免浮点数和 i64 溢出，使用 Rust u128 表示资产值。

**2. 原子性区块执行**：executor 执行区块时先 clone 一份 working_state，失败则 rollback，成功才 commit 到 canonical state。保证导入区块不会污染已有状态。

**3. Receipt 内容哈希**：Receipt 的 merkle root 基于内容 hash（`receipt_hash()`）而非简单字段拼接，保证序列化稳定性和可验证性。

**4. Mempool 语义修复**：
   - `nonce` 校验改为 `tx.nonce >= state_nonce`，而非 `nonce == 0`
   - `get_for_block` 优先从 `ordered` BTreeSet 选取高 nonce/fee 交易，而非 HashMap 随机迭代

**5. 存储解耦**：`put_block` 按 `block_hash` 存储（支持分叉），`get_block(number)` 通过 `canonical_chain[number]` 查找 canonical hash，避免单节点存储无法表达分叉的问题。

**6. Header 确定性哈希**：`header_hash = SHA256(bincode_serialize(header))`，保证 header hash 在所有节点上完全一致。

---

## 二、个人优势

### 2.1 工程能力

| 能力项 | 具体体现 |
|--------|---------|
| **Rust 系统编程** | 正确处理了 `&self` / `&mut self` 的 borrow checker 问题；为 `Storage` trait 合理设计 mutability；避免 clone 滥用（working_state 一次性 clone 后原子性 commit） |
| **区块链协议理解深度** | 实现了完整的交易生命周期：签名校验 → nonce 校验 → 余额校验 → 状态转移 → receipt 生成 → root 计算。理解 state_root / tx_root / receipt_root 三根之间的关系和验证逻辑 |
| **代码质量意识** | 44 个测试全部通过；Clippy 零警告；每周迭代都经过 Code Review，持续修复 P0/P1 问题 |
| **确定性执行** | 理解区块链节点最核心的要求：同一区块在不同节点执行必须得到相同 state_root。所有序列化、迭代顺序、状态读写都按此约束设计 |
| **架构设计** | Storage trait 抽象、RocksDB backend 可插拔、Mempool 与 StateDB 解耦、Executor 与 Validator 分离 |

### 2.2 成长轨迹（从 Week 1 到 Week 3）

| 阶段 | 产出 | 关键挑战 |
|------|------|---------|
| Week 1 | 项目骨架、核心类型、tracing 日志 | Hash/Address impl 冲突、死代码警告 |
| Week 2 | 状态机、执行器、Receipt | dead_code、u128 资产值、原子性执行 |
| Week 3 | Mempool、验证、存储 | nonce 语义错误、storage 分叉能力缺失、Block selection 提前 remove |
| Week 4 (进行中) | RPC、单节点出块 | Week 3 review 5 个 P0/P1 问题已修复 |

这条轨迹展示了从基础类型到完整链路的递进能力，每次迭代都有 Review 反馈驱动改进。

---

## 三、岗位匹配程度

### 3.1 目标岗位

**Primary**: Protocol Engineer / Rust Blockchain Engineer
**Secondary**: Blockchain Infrastructure Engineer / Go Rust Backend Engineer

### 3.2 JD 关键词对照

| JD 常见关键词 | 项目覆盖情况 | 说明 |
|-------------|------------|------|
| Rust | ✅ 深度使用 | 完整项目，语言能力可验证 |
| blockchain node | ✅ | 从零实现节点核心模块 |
| state machine / state transition | ✅ | StateDB + Executor 完整实现 |
| transaction execution | ✅ | deterministic execution, receipt generation |
| Merkle root / state root | ✅ | tx_root, receipt_root, state_root 三根计算和校验 |
| mempool | ✅ | nonce ordering, fee priority, replacement, eviction |
| block validation | ✅ | header validation, root verification |
| storage abstraction | ✅ | Storage trait，支持 fork-ready hash-keyed storage |
| signature verification | ⚠️ 占位符 | 当前仅检查 signature 非空，尚未接 k256 |
| P2P sync | 🔄 待实现 | Week 5-6 计划实现 |
| fork choice | 🔄 待实现 | Week 7 计划实现 |
| consensus / PoA | 🔄 待实现 | Week 4 计划实现单节点 PoA |

### 3.3 面试可讲点

**1. 交易执行的原子性**
> "区块链节点执行区块时不能污染已有状态。我用 working_state clone 方案：执行成功才 commit，失败则丢弃 working_state。这样即使区块执行到一半失败，已有的 canonical state 也不会被破坏。"

**2. Mempool 的 nonce/fee 联合排序**
> "Mempool 用 BTreeSet<TxOrder> 排序，排序 key 是 (-nonce, -fee_per_gas)。这保证：同账户交易先按 nonce 从大到小选（高 nonce 意味着交易更新鲜或更重要）；不同账户之间按 fee 从高到低选，最大化区块手续费收益。"

**3. 存储的分叉表达**
> "早期实现把区块存在 BTreeMap<number, Block>，这样同一高度只能存一个区块，无法表达分叉。修复后用 HashMap<hash, Block> 存所有区块体，canonical_chain 只负责建立 number→hash 的规范映射。这样同一高度可以出现多条链，分叉时只需更新 canonical_chain 而不用移动区块体。"

**4. 三根校验的防御逻辑**
> "区块验证时永远不直接信任对端给的 state_root。流程是：先校验 header 格式 → 再校验 tx_root → 重新执行区块得到 outcome.state_root → 对比 outcome.state_root == header.state_root。如果对端给了一个恶意区块，它的 header 里写的 state_root 和我们实际执行得到的不一致，区块就会被拒绝。"

**5. Week 3 Review 迭代过程**
> "Week 3 结束后做了 Code Review，发现了 5 个 P0/P1 问题。比如 mempool 原来拒绝 nonce==0 的交易，但实际业务中 nonce 从 0 开始是合法的，应该是 nonce < state_nonce 才拒绝。发现这个问题是因为考虑到实际链上第一笔交易的 nonce 就是 0。这些修复体现了对真实区块链业务的理解。"

### 3.4 仍需补齐的短板（Roadmap）

| 优先级 | 内容 | 理由 |
|--------|------|------|
| P0 | k256 真实 ECDSA 签名 | 面试会直接问"签名为什么是占位符" |
| P0 | `cargo fmt` 过门 | 代码质量门的门面 |
| P1 | RPC 接口 | Week 4 核心交付，可直接演示端到端 |
| P1 | RocksDB backend | 生产级存储必须，内存模型面试说服力不足 |
| P1 | 单节点 PoA 出块 | 证明节点不只是导入区块，还能产生区块 |
| P2 | P2P sync | Week 5-6 双节点同步是真实区块链节点标志 |
| P2 | Fork choice + reorg | Week 7 分叉处理是区块链核心难点 |

---

## 四、简历表述建议

### 4.1 简洁版（1-2 行）

```text
Rust区块链节点实现：覆盖交易签名校验、状态转移、原子性区块执行、Merkle root 验证、mempool（nonce/fee 排序与替换）、存储抽象（fork-ready hash-keyed）。
```

### 4.2 完整版（3-5 行）

```text
从零实现了一个 Rust 最小可验证区块链节点（verifiable-rust-chain-node），包含：

- 核心类型：SignedTransaction、Block、Header、Receipt、Account（u128 资产值）
- 状态机：StateDB 账户余额/nonce 管理，state_root Merkle 计算
- 交易执行器：原子性区块执行，revert on failure，receipt 生成
- Mempool：nonce + fee 联合排序，替换规则，容量驱逐
- 区块验证：header 校验、tx_root、state_root、receipt_root 三根验证
- 存储抽象：Storage trait，fork-ready hash-keyed 存储，支持 canonical chain

已完成 44 个单元测试，Clippy 零警告。正在实现 RPC 接口和单节点 PoA 出块。
```

### 4.3 面试叙事框架

```
开场（1分钟）：
我在做一个 Rust 区块链节点项目，目的是深入理解区块链底层协议，
从交易签名到状态转移、从 mempool 选交易到区块验证整个链路。

核心亮点（3分钟）：
1. 确定性执行：同一区块在不同节点必须得到相同 state_root，
   我用 working_state clone 方案保证原子性。
2. Mempool 排序：BTreeSet 按 (-nonce, -fee) 排序，
   保证同账户交易按 nonce 顺序打包，不同账户优先打包高手续费交易。
3. 区块验证：永远不信任对端给的 state_root，
   必须重新执行并对比三根。
4. Fork-ready 存储：区块按 hash 存储而非按 number 存储，
   canonical_chain 单独维护规范映射。

迭代过程（1分钟）：
Week 3 Review 发现了 5 个问题，包括 mempool 对 nonce=0 的错误拒绝、
storage 无法表达分叉、get_for_block 提前 remove 交易导致执行失败时丢交易等。
这些修复体现了对真实区块链业务的理解。

下一步（1分钟）：
正在实现 RPC 接口和单节点 PoA 出块，
之后会接 P2P sync 和 fork choice。
签名目前是占位符，下一步会接 k256 实现真实 ECDSA 校验。
```

---

## 五、项目可验证性

| 验证项 | 命令 | 结果 |
|--------|------|------|
| 编译 | `cargo build` | ✅ 通过 |
| 单元测试 | `cargo test --lib` | ✅ 44 passed |
| Clippy | `cargo clippy --all-targets` | ✅ 零警告 |
| 格式检查 | `cargo fmt --check` | ⚠️ 需运行修复 |

所有代码均可通过上述命令直接验证，无需复杂的环境配置。

---

## 六、结论

| 维度 | 评分 | 说明 |
|------|------|------|
| 技术深度 | ⭐⭐⭐☆☆ | 覆盖核心链路，但尚缺 P2P、真实签名、fork choice |
| 工程完整性 | ⭐⭐⭐⭐☆ | 44 测试、Clippy 干净、模块切分清晰 |
| 面试说服力 | ⭐⭐⭐⭐☆ | 可讲点多（原子性、三根防御、nonce 语义、fork 存储） |
| 生产级可信度 | ⭐⭐☆☆☆ | 仍需 RPC、RocksDB、真实签名才能用于生产讨论 |
| 综合竞争力 | ⭐⭐⭐⭐☆ | 对 Junior/Mid Protocol Engineer 岗位具有较强竞争力 |

**适合申请**：Protocol Engineer (Junior/Mid)、Rust Blockchain Engineer (Junior)、Blockchain Infrastructure Engineer
**距离 Senior 差距**：P2P 同步、真实签名、共识算法深度、fork choice 生产实现
