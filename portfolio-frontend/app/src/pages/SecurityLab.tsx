import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { AlertTriangle, CheckCircle2, FlaskConical, LockKeyhole, ShieldAlert, XCircle } from 'lucide-react'
import { SectionTitle } from '../components'
import { attackCases, auditChecklist, testCoverage, vaultSimulations } from '../data/securityLab'
import type { AttackCase, SecuritySeverity, TestCoverageStatus, VaultAction } from '../types'

const severityTone: Record<SecuritySeverity, string> = {
  critical: 'danger',
  high: 'warning',
  medium: 'info',
}

const coverageTone: Record<TestCoverageStatus, string> = {
  missing: 'danger',
  planned: 'warning',
  covered: 'ready',
}

export function SecurityLab() {
  const { t } = useTranslation()
  const [action, setAction] = useState<VaultAction>('deposit')
  const [selectedAttackId, setSelectedAttackId] = useState(attackCases[0].id)
  const selectedSimulation = useMemo(
    () => vaultSimulations.find((item) => item.action === action) ?? vaultSimulations[0],
    [action],
  )
  const selectedAttack = attackCases.find((item) => item.id === selectedAttackId) ?? attackCases[0]
  const criticalCount = attackCases.filter((item) => item.severity === 'critical').length

  return (
    <section className="security-page">
      <div className="security-hero">
        <div>
          <span className="terminal-line">{t('securityLab.terminal')}</span>
          <h2>{t('securityLab.title')}</h2>
          <p>{t('securityLab.subtitle')}</p>
        </div>
        <div className="security-verdict">
          <span>{t('securityLab.readiness')}</span>
          <strong>{t('securityLab.notProductionReady')}</strong>
          <p>{t('securityLab.readinessBody', { count: criticalCount })}</p>
        </div>
      </div>

      <div className="security-summary-grid">
        <SecurityMetric label={t('securityLab.attackCases')} value={String(attackCases.length)} tone="danger" />
        <SecurityMetric label={t('securityLab.criticalFindings')} value={String(criticalCount)} tone="danger" />
        <SecurityMetric label={t('securityLab.coverageAreas')} value={String(testCoverage.length)} tone="warning" />
        <SecurityMetric label={t('securityLab.auditItems')} value={String(auditChecklist.length)} tone="cyan" />
      </div>

      <div className="security-layout">
        <div className="security-main">
          <VaultAccountingPanel action={action} setAction={setAction} />
          <VaultInvariantPanel simulation={selectedSimulation} />
          <AttackCaseList selectedAttack={selectedAttack} onSelectAttack={setSelectedAttackId} />
          <TestCoverageMatrix />
        </div>
        <aside className="security-side">
          <AttackDetail attack={selectedAttack} />
          <AuditChecklist />
        </aside>
      </div>
    </section>
  )
}

function SecurityMetric({ label, value, tone }: { label: string; value: string; tone: 'danger' | 'warning' | 'cyan' }) {
  return (
    <article className={`security-metric ${tone}`}>
      <span>{label}</span>
      <strong>{value}</strong>
    </article>
  )
}

function VaultAccountingPanel({ action, setAction }: { action: VaultAction; setAction: (action: VaultAction) => void }) {
  const { t } = useTranslation()

  return (
    <section className="panel security-panel">
      <div className="security-panel-head">
        <SectionTitle title={t('securityLab.vaultTitle')} subtitle={t('securityLab.vaultSubtitle')} />
        <div className="risk-toggle security-toggle">
          {(['deposit', 'withdraw', 'redeem'] as const).map((item) => (
            <button className={action === item ? 'selected' : ''} key={item} onClick={() => setAction(item)}>
              {t(`securityLab.actions.${item}`)}
            </button>
          ))}
        </div>
      </div>
      <div className="vault-flow">
        <div className="vault-node">
          <LockKeyhole aria-hidden="true" />
          <span>{t('securityLab.userAssets')}</span>
        </div>
        <div className="vault-arrow">-&gt;</div>
        <div className="vault-node">
          <FlaskConical aria-hidden="true" />
          <span>{t('securityLab.vaultAccounting')}</span>
        </div>
        <div className="vault-arrow">-&gt;</div>
        <div className="vault-node">
          <ShieldAlert aria-hidden="true" />
          <span>{t('securityLab.securityBoundary')}</span>
        </div>
      </div>
    </section>
  )
}

function VaultInvariantPanel({ simulation }: { simulation: typeof vaultSimulations[number] }) {
  const { t } = useTranslation()

  return (
    <section className="panel security-panel">
      <SectionTitle title={t('securityLab.invariantTitle')} subtitle={t('securityLab.invariantSubtitle')} />
      <div className="vault-sim-grid">
        <SimCell label={t('securityLab.assets')} value={String(simulation.assets)} />
        <SimCell label={t('securityLab.totalAssets')} value={`${simulation.totalAssetsBefore} -> ${simulation.totalAssetsAfter}`} />
        <SimCell label={t('securityLab.totalShares')} value={`${simulation.totalSharesBefore} -> ${simulation.totalSharesAfter}`} />
        <SimCell label={t('securityLab.sharesDelta')} value={String(simulation.sharesDelta)} />
      </div>
      <div className="invariant-callout">
        <CheckCircle2 aria-hidden="true" />
        <p>{simulation.invariant}</p>
      </div>
    </section>
  )
}

function SimCell({ label, value }: { label: string; value: string }) {
  return (
    <div className="sim-cell">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  )
}

function AttackCaseList({ selectedAttack, onSelectAttack }: { selectedAttack: AttackCase; onSelectAttack: (id: string) => void }) {
  const { t } = useTranslation()

  return (
    <section className="panel security-panel">
      <SectionTitle title={t('securityLab.attackTitle')} subtitle={t('securityLab.attackSubtitle')} />
      <div className="attack-list">
        {attackCases.map((attack) => (
          <button
            className={selectedAttack.id === attack.id ? `attack-row active ${attack.severity}` : `attack-row ${attack.severity}`}
            key={attack.id}
            onClick={() => onSelectAttack(attack.id)}
          >
            <span className={`security-chip ${severityTone[attack.severity]}`}>{t(`securityLab.severity.${attack.severity}`)}</span>
            <strong>{attack.title}</strong>
            <small>{attack.contract}</small>
          </button>
        ))}
      </div>
    </section>
  )
}

function TestCoverageMatrix() {
  const { t } = useTranslation()

  return (
    <section className="panel security-panel">
      <SectionTitle title={t('securityLab.coverageTitle')} subtitle={t('securityLab.coverageSubtitle')} />
      <div className="coverage-table">
        <div className="coverage-head">
          <span>{t('securityLab.area')}</span>
          <span>Unit</span>
          <span>Fuzz</span>
          <span>Invariant</span>
          <span>PoC</span>
        </div>
        {testCoverage.map((row) => (
          <div className="coverage-row" key={row.area}>
            <strong>{row.area}</strong>
            <CoverageBadge status={row.unit} />
            <CoverageBadge status={row.fuzz} />
            <CoverageBadge status={row.invariant} />
            <CoverageBadge status={row.attackPoC} />
          </div>
        ))}
      </div>
    </section>
  )
}

function CoverageBadge({ status }: { status: TestCoverageStatus }) {
  const { t } = useTranslation()
  return <span className={`security-chip ${coverageTone[status]}`}>{t(`securityLab.coverage.${status}`)}</span>
}

function AttackDetail({ attack }: { attack: AttackCase }) {
  const { t } = useTranslation()

  return (
    <section className="security-inspector">
      <span>{t('securityLab.selectedAttack')}</span>
      <strong>{attack.title}</strong>
      <div className={`security-chip ${severityTone[attack.severity]}`}>{t(`securityLab.severity.${attack.severity}`)}</div>
      <DetailBlock label={t('securityLab.contract')} value={attack.contract} />
      <DetailBlock label={t('securityLab.vector')} value={attack.vector} />
      <DetailBlock label={t('securityLab.impact')} value={attack.impact} />
      <DetailBlock label={t('securityLab.fix')} value={attack.fix} />
    </section>
  )
}

function DetailBlock({ label, value }: { label: string; value: string }) {
  return (
    <div className="security-detail-block">
      <span>{label}</span>
      <p>{value}</p>
    </div>
  )
}

function AuditChecklist() {
  const { t } = useTranslation()

  return (
    <section className="security-inspector">
      <SectionTitle title={t('securityLab.auditTitle')} subtitle={t('securityLab.auditSubtitle')} />
      <div className="audit-list">
        {auditChecklist.map((item) => (
          <article className={`audit-item ${item.status}`} key={item.item}>
            {item.status === 'done' ? <CheckCircle2 aria-hidden="true" /> : item.status === 'in_progress' ? <AlertTriangle aria-hidden="true" /> : <XCircle aria-hidden="true" />}
            <div>
              <strong>{item.item}</strong>
              <p>{item.note}</p>
              <span>{t(`securityLab.auditStatus.${item.status}`)}</span>
            </div>
          </article>
        ))}
      </div>
    </section>
  )
}
