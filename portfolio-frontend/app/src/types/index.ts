import type { LucideIcon } from 'lucide-react'

export type View =
  | 'dashboard'
  | 'assets'
  | 'ops'
  | 'protocol'
  | 'security'
  | 'project'
  | 'compare'
  | 'interview'
  | 'risk'
  | 'wallet'
  | 'market'
  | 'trading'
export type OpsView = 'health' | 'rpc' | 'queue' | 'reorg' | 'outbox'
export type ProjectId = 'web3-backend' | 'protocol-rust' | 'smart-contract'
export type ProjectKey = 'web3Backend' | 'protocolRust' | 'smartContract'
export type Verification = 'Blocked' | 'Partial' | 'No tests'
export type Status = 'Flagship' | 'Depth' | 'Security'
export type Priority = 'P0' | 'P1' | 'P2'
export type TimeBudget = '3' | '8' | '20'
export type AppIconName = 'dashboard' | 'assets' | 'ops' | 'server' | 'blocks' | 'shield' | 'compare' | 'terminal' | 'wallet' | 'check' | 'alert' | 'arrow' | 'network' | 'dot' | 'chart' | 'trade'
export type ChainFilter = 'All' | 'Ethereum' | 'BNB Chain' | 'Polygon' | 'Solana'
export type DepositStatus = 'detected' | 'pending' | 'confirmed' | 'reorged'
export type ServiceStatus = 'healthy' | 'degraded' | 'down'
export type RpcProviderStatus = 'online' | 'degraded' | 'down'
export type QueueStatus = 'active' | 'backlogged' | 'dead'
export type OutboxStatus = 'pending' | 'published' | 'failed'
export type ReorgSeverity = 'low' | 'medium' | 'high'

export type Finding = {
  priority: Priority
  title: string
  impact: string
  fix: string
}

export type Project = {
  id: ProjectId
  key: ProjectKey
  status: Status
  verification: Verification
  metrics: {
    marketFit: number
    depth: number
    testReadiness: number
    riskControl: number
  }
}

export type ProjectText = {
  name: string
  shortName: string
  role: string
  tagline: string
  repo: string
  stack: string[]
  pipeline: string[]
  wins: string[]
  findings: Finding[]
  hrPitch: string
  seniorPitch: string
  nextMoves: string[]
}

export type IconMap = Record<AppIconName, LucideIcon>

export type Language = {
  code: 'en' | 'zh' | 'ja'
  label: string
  short: string
}

export type BalanceAsset = {
  chain: Exclude<ChainFilter, 'All'>
  token: string
  symbol: string
  available: string
  locked: string
  value: string
  change: string
  contract: string
}

export type DepositRecord = {
  id: string
  chain: Exclude<ChainFilter, 'All'>
  token: string
  amount: string
  balanceDelta: string
  status: DepositStatus
  confirmations: number
  requiredConfirmations: number
  txHash: string
  account: string
  block: string
  timestamp: string
  rawLog: string
  parsedEvent: string
  ledgerStatus: string
}

// Infrastructure Ops Console Types
export type Service = {
  id: string
  name: string
  status: ServiceStatus
  uptime: string
  lastHeartbeat: string
  errorCount: number
  avgLatency: string
  nextAction?: string
}

export type RpcProvider = {
  id: string
  chain: Exclude<ChainFilter, 'All'>
  endpoint: string
  status: RpcProviderStatus
  latency: string
  errorRate: string
  latestBlock: string
  lastUpdate: string
}

export type QueueEntry = {
  id: string
  topic: string
  status: QueueStatus
  backlogCount: number
  retryCount: number
  deadLetterCount: number
  avgProcessTime: string
}

export type ReorgEvent = {
  id: string
  chain: Exclude<ChainFilter, 'All'>
  detectedAt: string
  severity: ReorgSeverity
  depth: number
  affectedBlocks: string[]
  resolvedAt?: string
  compensation: string
}

export type OutboxEntry = {
  id: string
  event: string
  status: OutboxStatus
  createdAt: string
  publishedAt?: string
  retryCount: number
  lastError?: string
}

export type Incident = {
  id: string
  service: string
  severity: 'info' | 'warning' | 'critical'
  message: string
  timestamp: string
  acknowledged: boolean
}

export type RiskStatus = 'open' | 'in_progress' | 'fixed' | 'blocked'

export type RiskCenterRisk = {
  id: string
  projectId: 'web3-backend' | 'protocol-rust' | 'smart-contract'
  priority: 'P0' | 'P1' | 'P2'
  status: RiskStatus
  title: string
  impact: string
  fixPlan: string
  evidenceLinks: string[]
  createdAt: string
  updatedAt: string
  assignee?: string
}

export type WalletAccount = {
  address: string
  type: 'EOA' | 'Contract' | 'Hardware'
  label?: string
  network: string
  isConnected: boolean
}

export type WalletTransaction = {
  id: string
  type: 'send' | 'receive' | 'swap' | 'approve'
  status: 'pending' | 'confirmed' | 'failed'
  amount: string
  token: string
  timestamp: string
  txHash: string
  from: string
  to: string
}

// Market and Trading Types
export type PriceTicker = {
  symbol: string
  name: string
  price: string
  change24h: string
  volume24h: string
  high24h: string
  low24h: string
}

export type OrderBookEntry = {
  price: string
  amount: string
  total: string
}

export type Trade = {
  id: string
  type: 'buy' | 'sell'
  price: string
  amount: string
  timestamp: string
}

export type TradingPair = {
  id: string
  base: string
  quote: string
  symbol: string
}

export type ProtocolTxStatus = 'ready' | 'nonce_gap' | 'stale' | 'overflow_risk'
export type ProtocolValidationStatus = 'pass' | 'fail' | 'warning'

export type ProtocolTransaction = {
  id: string
  sender: string
  nonce: number
  value: string
  fee: string
  status: ProtocolTxStatus
  reason: string
}

export type ProtocolAccountDiff = {
  account: string
  beforeBalance: string
  afterBalance: string
  beforeNonce: number
  afterNonce: number
}

export type ProtocolRootCheck = {
  name: string
  expected: string
  computed: string
  status: ProtocolValidationStatus
  note: string
}

export type ProtocolFinding = {
  priority: 'P1' | 'P2'
  title: string
  file: string
  impact: string
  fix: string
}

export type ForkNode = {
  hash: string
  number: number
  parent: string
  canonical: boolean
}

export type VaultAction = 'deposit' | 'withdraw' | 'redeem'
export type SecuritySeverity = 'critical' | 'high' | 'medium'
export type TestCoverageStatus = 'missing' | 'planned' | 'covered'

export type VaultSimulation = {
  action: VaultAction
  assets: number
  totalAssetsBefore: number
  totalSharesBefore: number
  sharesDelta: number
  totalAssetsAfter: number
  totalSharesAfter: number
  invariant: string
}

export type AttackCase = {
  id: string
  title: string
  severity: SecuritySeverity
  contract: string
  vector: string
  impact: string
  fix: string
}

export type TestCoverageItem = {
  area: string
  unit: TestCoverageStatus
  fuzz: TestCoverageStatus
  invariant: TestCoverageStatus
  attackPoC: TestCoverageStatus
}

export type AuditChecklistItem = {
  item: string
  status: 'open' | 'in_progress' | 'done'
  note: string
}
