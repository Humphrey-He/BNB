import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { CheckCircle2, GitFork, Play, ShieldAlert, XCircle } from 'lucide-react'
import { SectionTitle } from '../components'
import {
  accountDiffs,
  blockCandidate,
  forkNodes,
  protocolFindings,
  protocolTransactions,
  rootChecks,
} from '../data/protocol'
import type { ProtocolTransaction } from '../types'

const statusTone = {
  ready: 'ready',
  nonce_gap: 'warning',
  stale: 'danger',
  overflow_risk: 'danger',
} as const

export function ProtocolVisualizer() {
  const { t } = useTranslation()
  const [selectedTxId, setSelectedTxId] = useState(protocolTransactions[0].id)
  const [scenario, setScenario] = useState<'valid' | 'invalid'>('invalid')

  const selectedTx = useMemo(
    () => protocolTransactions.find((tx) => tx.id === selectedTxId) ?? protocolTransactions[0],
    [selectedTxId],
  )

  const candidateTransactions = protocolTransactions.filter((tx) => blockCandidate.includes(tx.id))
  const invalidChecks = rootChecks.filter((check) => check.status !== 'pass').length
  const blockStatus = scenario === 'valid' ? 'pass' : 'fail'

  return (
    <section className="protocol-page">
      <div className="protocol-hero">
        <div>
          <span className="terminal-line">{t('protocolVisualizer.terminal')}</span>
          <h2>{t('protocolVisualizer.title')}</h2>
          <p>{t('protocolVisualizer.subtitle')}</p>
        </div>
        <div className="protocol-verdict">
          <span>{t('protocolVisualizer.validationVerdict')}</span>
          <strong className={blockStatus}>{blockStatus === 'pass' ? t('protocolVisualizer.validBlock') : t('protocolVisualizer.invalidBlock')}</strong>
          <p>{t('protocolVisualizer.verdictBody', { count: invalidChecks })}</p>
        </div>
      </div>

      <div className="protocol-summary-grid">
        <ProtocolMetric label={t('protocolVisualizer.mempoolSize')} value={String(protocolTransactions.length)} />
        <ProtocolMetric label={t('protocolVisualizer.blockCandidate')} value={String(candidateTransactions.length)} />
        <ProtocolMetric label={t('protocolVisualizer.rootChecks')} value={`${rootChecks.length - invalidChecks}/${rootChecks.length}`} />
        <ProtocolMetric label={t('protocolVisualizer.reviewFindings')} value={String(protocolFindings.length)} />
      </div>

      <div className="protocol-layout">
        <div className="protocol-main">
          <MempoolPanel selectedTx={selectedTx} onSelectTx={setSelectedTxId} />
          <BlockBuilderPanel candidateTransactions={candidateTransactions} scenario={scenario} setScenario={setScenario} />
          <StateTransitionPanel />
          <RootValidationPanel scenario={scenario} />
        </div>
        <aside className="protocol-side">
          <TransactionInspector tx={selectedTx} />
          <ReviewFindingsPanel />
          <ForkStoragePanel />
        </aside>
      </div>
    </section>
  )
}

function ProtocolMetric({ label, value }: { label: string; value: string }) {
  return (
    <article className="protocol-metric">
      <span>{label}</span>
      <strong>{value}</strong>
    </article>
  )
}

function MempoolPanel({ selectedTx, onSelectTx }: { selectedTx: ProtocolTransaction; onSelectTx: (id: string) => void }) {
  const { t } = useTranslation()

  return (
    <section className="panel protocol-panel">
      <SectionTitle title={t('protocolVisualizer.mempoolTitle')} subtitle={t('protocolVisualizer.mempoolSubtitle')} />
      <div className="protocol-table">
        <div className="protocol-table-head">
          <span>{t('protocolVisualizer.tx')}</span>
          <span>{t('protocolVisualizer.sender')}</span>
          <span>{t('protocolVisualizer.nonce')}</span>
          <span>{t('protocolVisualizer.value')}</span>
          <span>{t('protocolVisualizer.fee')}</span>
          <span>{t('protocolVisualizer.status')}</span>
        </div>
        {protocolTransactions.map((tx) => (
          <button
            className={selectedTx.id === tx.id ? 'protocol-table-row active' : 'protocol-table-row'}
            key={tx.id}
            onClick={() => onSelectTx(tx.id)}
          >
            <strong>{tx.id}</strong>
            <span>{tx.sender}</span>
            <span>{tx.nonce}</span>
            <span>{tx.value}</span>
            <span>{tx.fee}</span>
            <span className={`protocol-chip ${statusTone[tx.status]}`}>{t(`protocolVisualizer.txStatus.${tx.status}`)}</span>
          </button>
        ))}
      </div>
    </section>
  )
}

function BlockBuilderPanel({
  candidateTransactions,
  scenario,
  setScenario,
}: {
  candidateTransactions: ProtocolTransaction[]
  scenario: 'valid' | 'invalid'
  setScenario: (scenario: 'valid' | 'invalid') => void
}) {
  const { t } = useTranslation()

  return (
    <section className="panel protocol-panel">
      <div className="protocol-panel-head">
        <SectionTitle title={t('protocolVisualizer.blockBuilderTitle')} subtitle={t('protocolVisualizer.blockBuilderSubtitle')} />
        <div className="risk-toggle protocol-toggle">
          <button className={scenario === 'valid' ? 'selected' : ''} onClick={() => setScenario('valid')}>{t('protocolVisualizer.validScenario')}</button>
          <button className={scenario === 'invalid' ? 'selected' : ''} onClick={() => setScenario('invalid')}>{t('protocolVisualizer.invalidScenario')}</button>
        </div>
      </div>
      <div className="block-builder-flow">
        {candidateTransactions.map((tx, index) => (
          <div className="block-tx-card" key={tx.id}>
            <span>{String(index + 1).padStart(2, '0')}</span>
            <strong>{tx.id}</strong>
            <p>{tx.sender} / nonce {tx.nonce}</p>
          </div>
        ))}
        <div className={scenario === 'valid' ? 'execution-card pass' : 'execution-card fail'}>
          <Play aria-hidden="true" />
          <strong>{scenario === 'valid' ? t('protocolVisualizer.atomicCommit') : t('protocolVisualizer.rollbackRequired')}</strong>
          <p>{scenario === 'valid' ? t('protocolVisualizer.atomicCommitBody') : t('protocolVisualizer.rollbackBody')}</p>
        </div>
      </div>
    </section>
  )
}

function StateTransitionPanel() {
  const { t } = useTranslation()

  return (
    <section className="panel protocol-panel">
      <SectionTitle title={t('protocolVisualizer.stateTitle')} subtitle={t('protocolVisualizer.stateSubtitle')} />
      <div className="state-diff-grid">
        {accountDiffs.map((diff) => (
          <article className="state-diff-card" key={diff.account}>
            <strong>{diff.account}</strong>
            <div>
              <span>{t('protocolVisualizer.balance')}</span>
              <p>{diff.beforeBalance} {'->'} {diff.afterBalance}</p>
            </div>
            <div>
              <span>{t('protocolVisualizer.nonce')}</span>
              <p>{diff.beforeNonce} {'->'} {diff.afterNonce}</p>
            </div>
          </article>
        ))}
      </div>
    </section>
  )
}

function RootValidationPanel({ scenario }: { scenario: 'valid' | 'invalid' }) {
  const { t } = useTranslation()
  const checks = scenario === 'valid'
    ? rootChecks.map((check) => ({ ...check, computed: check.expected, status: 'pass' as const }))
    : rootChecks

  return (
    <section className="panel protocol-panel">
      <SectionTitle title={t('protocolVisualizer.rootTitle')} subtitle={t('protocolVisualizer.rootSubtitle')} />
      <div className="root-check-list">
        {checks.map((check) => (
          <article className={`root-check ${check.status}`} key={check.name}>
            {check.status === 'pass' ? <CheckCircle2 aria-hidden="true" /> : <XCircle aria-hidden="true" />}
            <div>
              <strong>{check.name}</strong>
              <p>{check.note}</p>
              <span>{check.expected} / {check.computed}</span>
            </div>
          </article>
        ))}
      </div>
    </section>
  )
}

function TransactionInspector({ tx }: { tx: ProtocolTransaction }) {
  const { t } = useTranslation()

  return (
    <section className="protocol-inspector">
      <span>{t('protocolVisualizer.selectedTx')}</span>
      <strong>{tx.id}</strong>
      <div className={`protocol-chip ${statusTone[tx.status]}`}>{t(`protocolVisualizer.txStatus.${tx.status}`)}</div>
      <p>{tx.reason}</p>
      <div className="protocol-kv">
        <span>{t('protocolVisualizer.sender')}</span>
        <strong>{tx.sender}</strong>
      </div>
      <div className="protocol-kv">
        <span>{t('protocolVisualizer.value')}</span>
        <strong>{tx.value}</strong>
      </div>
    </section>
  )
}

function ReviewFindingsPanel() {
  const { t } = useTranslation()

  return (
    <section className="protocol-inspector">
      <SectionTitle title={t('protocolVisualizer.findingsTitle')} subtitle={t('protocolVisualizer.findingsSubtitle')} />
      <div className="protocol-finding-list">
        {protocolFindings.map((finding) => (
          <article className="protocol-finding" key={finding.title}>
            <div>
              <span className={`priority ${finding.priority.toLowerCase()}`}>{finding.priority}</span>
              <strong>{finding.title}</strong>
            </div>
            <code>{finding.file}</code>
            <p>{finding.impact}</p>
            <small>{finding.fix}</small>
          </article>
        ))}
      </div>
    </section>
  )
}

function ForkStoragePanel() {
  const { t } = useTranslation()

  return (
    <section className="protocol-inspector">
      <SectionTitle title={t('protocolVisualizer.forkTitle')} subtitle={t('protocolVisualizer.forkSubtitle')} />
      <div className="fork-graph">
        {forkNodes.map((node) => (
          <div className={node.canonical ? 'fork-node canonical' : 'fork-node'} key={node.hash}>
            <GitFork aria-hidden="true" />
            <strong>#{node.number}</strong>
            <span>{node.hash}</span>
          </div>
        ))}
      </div>
      <div className="fork-warning">
        <ShieldAlert aria-hidden="true" />
        <p>{t('protocolVisualizer.forkWarning')}</p>
      </div>
    </section>
  )
}
