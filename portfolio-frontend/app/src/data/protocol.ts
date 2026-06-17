import type {
  ForkNode,
  ProtocolAccountDiff,
  ProtocolFinding,
  ProtocolRootCheck,
  ProtocolTransaction,
} from '../types'

export const protocolTransactions: ProtocolTransaction[] = [
  {
    id: 'tx-001',
    sender: '0xA17c...90F1',
    nonce: 0,
    value: '42',
    fee: '18 gwei',
    status: 'ready',
    reason: 'New account first transaction. Mempool must accept nonce=0.',
  },
  {
    id: 'tx-002',
    sender: '0xA17c...90F1',
    nonce: 1,
    value: '9',
    fee: '22 gwei',
    status: 'ready',
    reason: 'Sequential nonce after tx-001, safe to include in the same block.',
  },
  {
    id: 'tx-003',
    sender: '0xB42e...71C0',
    nonce: 4,
    value: '7',
    fee: '31 gwei',
    status: 'nonce_gap',
    reason: 'Account state nonce is 2, so nonce 4 must wait for nonce 2 and 3.',
  },
  {
    id: 'tx-004',
    sender: '0xC91d...4AA2',
    nonce: 2,
    value: '340282366920938463463374607431768211455',
    fee: '11 gwei',
    status: 'overflow_risk',
    reason: 'u128 value exceeds u64 and must not be truncated during execution.',
  },
  {
    id: 'tx-005',
    sender: '0xD8f2...0F22',
    nonce: 1,
    value: '3',
    fee: '7 gwei',
    status: 'stale',
    reason: 'Account state nonce is already 2, so this transaction is stale.',
  },
]

export const blockCandidate = ['tx-001', 'tx-002']

export const accountDiffs: ProtocolAccountDiff[] = [
  {
    account: '0xA17c...90F1',
    beforeBalance: '100',
    afterBalance: '49',
    beforeNonce: 0,
    afterNonce: 2,
  },
  {
    account: '0x742d...88E4',
    beforeBalance: '5',
    afterBalance: '56',
    beforeNonce: 0,
    afterNonce: 0,
  },
]

export const rootChecks: ProtocolRootCheck[] = [
  {
    name: 'state_root',
    expected: '0x6d93...a11c',
    computed: '0x6d93...a11c',
    status: 'pass',
    note: 'Computed after executing transactions against a working state.',
  },
  {
    name: 'receipt_root',
    expected: '0x42ff...9001',
    computed: '0x7b2a...e812',
    status: 'fail',
    note: 'Receipt root must hash serialized receipt contents, not only transaction hashes.',
  },
  {
    name: 'tx_root',
    expected: '0x8a31...4f0b',
    computed: '0x8a31...4f0b',
    status: 'pass',
    note: 'Transaction ordering is deterministic for the selected block candidate.',
  },
]

export const protocolFindings: ProtocolFinding[] = [
  {
    priority: 'P1',
    title: 'u128 transaction values are truncated to u64 during execution',
    file: 'src/executor.rs:39-50',
    impact: 'Large transaction values can execute with incorrect balance semantics.',
    fix: 'Use u128 balances end-to-end or reject non-fitting values with try_from.',
  },
  {
    priority: 'P1',
    title: 'Failed block execution leaves earlier state mutations applied',
    file: 'src/executor.rs:65-71',
    impact: 'A failed block can partially mutate canonical state.',
    fix: 'Execute against a cloned working state and commit only after all checks pass.',
  },
  {
    priority: 'P1',
    title: 'Receipt root is computed from transaction hashes',
    file: 'src/executor.rs:80-85',
    impact: 'Receipt metadata changes may not be detected by validation.',
    fix: 'Hash serialized receipt fields including status, gas used and logs.',
  },
  {
    priority: 'P2',
    title: 'All-target clippy fails on test code',
    file: 'src/state.rs:109-119',
    impact: 'The Week 2 quality gate is not fully green.',
    fix: 'Remove needless test borrows and keep all-target clippy in CI.',
  },
]

export const forkNodes: ForkNode[] = [
  { number: 120, hash: '0xaa10', parent: '0xgen', canonical: true },
  { number: 121, hash: '0xbb20', parent: '0xaa10', canonical: true },
  { number: 122, hash: '0xcc30', parent: '0xbb20', canonical: true },
  { number: 122, hash: '0xdd31', parent: '0xbb20', canonical: false },
  { number: 123, hash: '0xee40', parent: '0xcc30', canonical: true },
]
