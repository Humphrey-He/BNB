import type { AttackCase, AuditChecklistItem, TestCoverageItem, VaultSimulation } from '../types'

export const vaultSimulations: VaultSimulation[] = [
  {
    action: 'deposit',
    assets: 1000,
    totalAssetsBefore: 10000,
    totalSharesBefore: 10000,
    sharesDelta: 1000,
    totalAssetsAfter: 11000,
    totalSharesAfter: 11000,
    invariant: 'shares minted must equal received assets when exchange rate is 1:1',
  },
  {
    action: 'withdraw',
    assets: 450,
    totalAssetsBefore: 11200,
    totalSharesBefore: 11000,
    sharesDelta: -442,
    totalAssetsAfter: 10750,
    totalSharesAfter: 10558,
    invariant: 'reward-aware withdraw must burn proportional shares and update asset accounting',
  },
  {
    action: 'redeem',
    assets: 728,
    totalAssetsBefore: 10750,
    totalSharesBefore: 10558,
    sharesDelta: -715,
    totalAssetsAfter: 10022,
    totalSharesAfter: 9843,
    invariant: 'redeem must preserve totalAssets / totalSupply consistency after rewards',
  },
]

export const attackCases: AttackCase[] = [
  {
    id: 'fee-token',
    title: 'Fee-on-transfer token desynchronizes shares',
    severity: 'critical',
    contract: 'SecureYieldVault.sol:105-116',
    vector: 'User deposits declared assets, but the vault receives less after transfer fee.',
    impact: 'Attacker can receive full shares while the vault holds fewer real assets.',
    fix: 'Use balanceBefore / balanceAfter and require received == assets, or explicitly reject fee-on-transfer tokens.',
  },
  {
    id: 'reward-accounting',
    title: 'Reward injection breaks withdraw accounting',
    severity: 'critical',
    contract: 'SecureYieldVault.sol:191-205',
    vector: 'Conversion includes accumulatedRewards, but withdraw only subtracts totalAssets.',
    impact: 'Withdraw can underflow or block users from redeeming earned assets.',
    fix: 'Use token balance as total assets or split principal / reward accounting and subtract proportionally.',
  },
  {
    id: 'emergency-shares',
    title: 'Emergency withdraw does not burn shares',
    severity: 'critical',
    contract: 'SecureYieldVault.sol:213-226',
    vector: 'Admin can withdraw assets while retaining shares.',
    impact: 'The same shares can claim assets twice, breaking vault solvency.',
    fix: 'Emergency path must burn shares or implement a paused principal-only redeem with clear accounting.',
  },
  {
    id: 'timelock-hash',
    title: 'Timelock operation hash mismatch',
    severity: 'high',
    contract: 'RewardDistributor.sol:161-181',
    vector: 'Schedule hash includes eta, but execute / cancel hash does not.',
    impact: 'Queued timelock operations cannot be executed or cancelled reliably.',
    fix: 'Use the exact same operation id parameters or pass eta explicitly to execute and cancel.',
  },
  {
    id: 'pause-controls',
    title: 'Pausable inherited without admin entrypoints',
    severity: 'medium',
    contract: 'SecureYieldVault.sol:98-155',
    vector: 'deposit / withdraw use whenNotPaused, but no pause / unpause functions exist.',
    impact: 'Admin cannot stop vault activity during an incident.',
    fix: 'Expose access-controlled pause and unpause functions, ideally with granular operation pause flags.',
  },
  {
    id: 'muldiv',
    title: 'Share conversion lacks full-precision mulDiv',
    severity: 'medium',
    contract: 'SecureYieldVault.sol:306-329',
    vector: 'assets * supply and shares * assets can overflow before division.',
    impact: 'Extreme values can revert conversions and create denial-of-service edges.',
    fix: 'Use OpenZeppelin Math.mulDiv with explicit rounding directions.',
  },
]

export const testCoverage: TestCoverageItem[] = [
  { area: 'Deposit / mint', unit: 'planned', fuzz: 'missing', invariant: 'missing', attackPoC: 'planned' },
  { area: 'Withdraw / redeem', unit: 'planned', fuzz: 'missing', invariant: 'missing', attackPoC: 'planned' },
  { area: 'Reward accounting', unit: 'missing', fuzz: 'missing', invariant: 'planned', attackPoC: 'missing' },
  { area: 'Emergency path', unit: 'missing', fuzz: 'missing', invariant: 'missing', attackPoC: 'planned' },
  { area: 'Timelock operations', unit: 'planned', fuzz: 'missing', invariant: 'missing', attackPoC: 'planned' },
  { area: 'Access control / pause', unit: 'planned', fuzz: 'missing', invariant: 'missing', attackPoC: 'missing' },
]

export const auditChecklist: AuditChecklistItem[] = [
  {
    item: 'SECURITY.md risk boundaries',
    status: 'open',
    note: 'Document supported tokens, admin powers, upgrade assumptions and known exclusions.',
  },
  {
    item: 'Foundry unit tests',
    status: 'in_progress',
    note: 'Start with deposit, withdraw, redeem, reward injection and pause controls.',
  },
  {
    item: 'Attack PoC suite',
    status: 'open',
    note: 'Implement malicious ERC20, fee-on-transfer, reentrancy and timelock mismatch cases.',
  },
  {
    item: 'Invariant tests',
    status: 'open',
    note: 'Assert assets, shares, rewards and user balances stay consistent across random operations.',
  },
  {
    item: 'Static analysis',
    status: 'open',
    note: 'Run Slither and document accepted warnings before presenting as contract work.',
  },
]
