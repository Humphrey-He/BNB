import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { AlertTriangle, CheckCircle2, Clock3, RadioTower, RefreshCcw } from 'lucide-react'
import { ChainSelector, CopyableHash, SectionTitle } from '../components'
import { balanceAssets, chainFilters, depositRecords } from '../data/assets'
import type { ChainFilter, DepositRecord, DepositStatus } from '../types'
import type { ReactNode } from 'react'

function depositStatusIcon(status: DepositStatus) {
  if (status === 'confirmed') return <CheckCircle2 aria-hidden="true" className="asset-status-icon" />
  if (status === 'reorged') return <RefreshCcw aria-hidden="true" className="asset-status-icon" />
  if (status === 'pending') return <Clock3 aria-hidden="true" className="asset-status-icon" />
  return <RadioTower aria-hidden="true" className="asset-status-icon" />
}

export function AssetDashboard() {
  const { t } = useTranslation()
  const [selectedChain, setSelectedChain] = useState<ChainFilter>('All')
  const [selectedDepositId, setSelectedDepositId] = useState(depositRecords[0].id)

  const visibleBalances = useMemo(
    () => selectedChain === 'All' ? balanceAssets : balanceAssets.filter((asset) => asset.chain === selectedChain),
    [selectedChain],
  )

  const visibleDeposits = useMemo(
    () => selectedChain === 'All' ? depositRecords : depositRecords.filter((deposit) => deposit.chain === selectedChain),
    [selectedChain],
  )

  const selectedDeposit = visibleDeposits.find((deposit) => deposit.id === selectedDepositId) ?? visibleDeposits[0] ?? depositRecords[0]

  const totals = {
    balance: '$215,?'.replace('?', '029.32'),
    pending: depositRecords.filter((deposit) => deposit.status === 'pending' || deposit.status === 'detected').length,
    confirmed: depositRecords.filter((deposit) => deposit.status === 'confirmed').length,
    alerts: depositRecords.filter((deposit) => deposit.status === 'reorged').length,
  }

  return (
    <section className="asset-page">
      <div className="asset-hero">
        <div>
          <span className="terminal-line">{t('assetDashboard.terminal')}</span>
          <h2>{t('assetDashboard.title')}</h2>
          <p>{t('assetDashboard.subtitle')}</p>
        </div>
        <div className="asset-hero-status">
          <span>{t('assetDashboard.pipelineStatus')}</span>
          <strong>scanner → parser → confirm → ledger</strong>
          <p>{t('assetDashboard.pipelineBody')}</p>
        </div>
      </div>

      <div className="asset-summary-grid">
        <AssetSummaryCard label={t('assetDashboard.totalBalance')} value={totals.balance} tone="yellow" />
        <AssetSummaryCard label={t('assetDashboard.pendingDeposits')} value={String(totals.pending)} tone="cyan" />
        <AssetSummaryCard label={t('assetDashboard.confirmedCredits')} value={String(totals.confirmed)} tone="green" />
        <AssetSummaryCard label={t('assetDashboard.riskAlerts')} value={String(totals.alerts)} tone="red" />
      </div>

      <div className="asset-toolbar">
        <SectionTitle title={t('assetDashboard.assetTableTitle')} subtitle={t('assetDashboard.assetTableSubtitle')} />
        <ChainSelector
          chains={chainFilters}
          selectedChain={selectedChain}
          onSelectChain={(chain) => setSelectedChain(chain as ChainFilter)}
        />
      </div>

      <div className="asset-layout">
        <div className="asset-main">
          <TokenBalanceTable assets={visibleBalances} />
          <DepositLifecycle
            deposits={visibleDeposits}
            selectedDeposit={selectedDeposit}
            onSelectDeposit={setSelectedDepositId}
          />
        </div>
        <DepositDetailPanel deposit={selectedDeposit} />
      </div>
    </section>
  )
}

function AssetSummaryCard({ label, value, tone }: { label: string; value: string; tone: 'yellow' | 'cyan' | 'green' | 'red' }) {
  return (
    <article className={`asset-summary-card ${tone}`}>
      <span>{label}</span>
      <strong>{value}</strong>
    </article>
  )
}

function TokenBalanceTable({ assets }: { assets: typeof balanceAssets }) {
  const { t } = useTranslation()

  return (
    <section className="panel asset-panel">
      <SectionTitle title={t('assetDashboard.balances')} subtitle={t('assetDashboard.balancesSubtitle')} />
      <div className="asset-table">
        <div className="asset-table-head">
          <span>{t('assetDashboard.token')}</span>
          <span>{t('assetDashboard.chain')}</span>
          <span>{t('assetDashboard.available')}</span>
          <span>{t('assetDashboard.locked')}</span>
          <span>{t('assetDashboard.value')}</span>
          <span>{t('assetDashboard.contract')}</span>
        </div>
        {assets.map((asset) => (
          <div className="asset-table-row" key={`${asset.chain}-${asset.symbol}`}>
            <strong>{asset.symbol}<small>{asset.token}</small></strong>
            <span>{asset.chain}</span>
            <span>{asset.available}</span>
            <span>{asset.locked}</span>
            <span className={asset.change.startsWith('+') ? 'positive' : 'negative'}>{asset.value} <small>{asset.change}</small></span>
            {asset.contract === 'native' ? <span>{asset.contract}</span> : <CopyableHash hash={asset.contract} />}
          </div>
        ))}
      </div>
    </section>
  )
}

function DepositLifecycle({
  deposits,
  selectedDeposit,
  onSelectDeposit,
}: {
  deposits: DepositRecord[]
  selectedDeposit: DepositRecord
  onSelectDeposit: (id: string) => void
}) {
  const { t } = useTranslation()

  return (
    <section className="panel asset-panel">
      <SectionTitle title={t('assetDashboard.depositLifecycle')} subtitle={t('assetDashboard.depositLifecycleSubtitle')} />
      <div className="deposit-list">
        {deposits.map((deposit) => {
          const progress = Math.min(100, Math.round((deposit.confirmations / deposit.requiredConfirmations) * 100))
          return (
            <button
              className={selectedDeposit.id === deposit.id ? `deposit-row active ${deposit.status}` : `deposit-row ${deposit.status}`}
              key={deposit.id}
              onClick={() => onSelectDeposit(deposit.id)}
            >
              <div className="deposit-row-main">
                <span className={`deposit-status ${deposit.status}`}>{depositStatusIcon(deposit.status)}{t(`assetDashboard.status.${deposit.status}`)}</span>
                <strong>{deposit.amount} {deposit.token}</strong>
                <small>{deposit.chain} / {deposit.block}</small>
              </div>
              <div className="confirmation-track">
                <span>{deposit.confirmations}/{deposit.requiredConfirmations}</span>
                <div><i style={{ width: `${progress}%` }} /></div>
              </div>
            </button>
          )
        })}
      </div>
    </section>
  )
}

function DepositDetailPanel({ deposit }: { deposit: DepositRecord }) {
  const { t } = useTranslation()
  const steps = [
    ['scanner', deposit.status === 'reorged' ? 'warning' : 'done'],
    ['parser', deposit.status === 'reorged' ? 'warning' : 'done'],
    ['confirmWorker', deposit.status === 'confirmed' ? 'done' : deposit.status === 'pending' ? 'active' : 'waiting'],
    ['ledger', deposit.status === 'confirmed' ? 'done' : deposit.status === 'reorged' ? 'warning' : 'waiting'],
  ] as const

  return (
    <aside className="asset-detail-panel">
      <div className="asset-detail-header">
        <span>{t('assetDashboard.selectedDeposit')}</span>
        <strong>{deposit.id}</strong>
        <p>{deposit.amount} {deposit.token} / {deposit.chain}</p>
      </div>
      <div className="asset-stepper">
        {steps.map(([step, state], index) => (
          <div className={`asset-step ${state}`} key={step}>
            <span>{index + 1}</span>
            <p>{t(`assetDashboard.steps.${step}`)}</p>
          </div>
        ))}
      </div>
      <div className="asset-detail-list">
        <DetailItem label={t('assetDashboard.txHash')} value={<CopyableHash hash={deposit.txHash} />} />
        <DetailItem label={t('assetDashboard.account')} value={<CopyableHash hash={deposit.account} />} />
        <DetailItem label={t('assetDashboard.timestamp')} value={deposit.timestamp} />
        <DetailItem label={t('assetDashboard.balanceDelta')} value={deposit.balanceDelta} className={deposit.balanceDelta.startsWith('+') ? 'positive' : deposit.balanceDelta.includes('reverted') ? 'negative' : undefined} />
        <DetailItem label={t('assetDashboard.ledgerStatus')} value={deposit.ledgerStatus} />
      </div>
      <div className="raw-event-box">
        <div>
          <AlertTriangle aria-hidden="true" />
          <span>{t('assetDashboard.rawAndParsed')}</span>
        </div>
        <p>{deposit.rawLog}</p>
        <p>{deposit.parsedEvent}</p>
      </div>
    </aside>
  )
}

function DetailItem({ label, value, className }: { label: string; value: ReactNode; className?: string }) {
  return (
    <div className="detail-item">
      <span>{label}</span>
      <strong className={className}>{value}</strong>
    </div>
  )
}
