import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { AlertTriangle, CheckCircle2, Clock3, RadioTower, RefreshCcw } from 'lucide-react'
import { ChainSelector, CopyableHash, EmptyState, LoadingState, SectionTitle } from '../components'
import { balanceAssets, chainFilters, depositRecords } from '../data/assets'
import {
  getAddressBalances,
  getChains,
  getDemoWalletAddress,
  getDeposits,
  getTokens,
  hasApiBase,
  type BalanceDTO,
  type ChainDTO,
  type DepositDTO,
  type TokenDTO,
} from '../lib/api'
import type { DepositStatus } from '../types'
import type { ReactNode } from 'react'

type AssetRow = {
  chain: string
  token: string
  symbol: string
  available: string
  locked: string
  value: string
  change: string
  contract: string
}

type AssetDepositRecord = {
  id: string
  chain: string
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

function depositStatusIcon(status: DepositStatus) {
  if (status === 'confirmed') return <CheckCircle2 aria-hidden="true" className="asset-status-icon" />
  if (status === 'reorged') return <RefreshCcw aria-hidden="true" className="asset-status-icon" />
  if (status === 'pending') return <Clock3 aria-hidden="true" className="asset-status-icon" />
  return <RadioTower aria-hidden="true" className="asset-status-icon" />
}

export function AssetDashboard() {
  const { t } = useTranslation()
  const configuredDemoAddress = getDemoWalletAddress()
  const [selectedChain, setSelectedChain] = useState('All')
  const [selectedDepositId, setSelectedDepositId] = useState(String(depositRecords[0].id))
  const [liveChains, setLiveChains] = useState<ChainDTO[]>([])
  const [liveTokens, setLiveTokens] = useState<TokenDTO[]>([])
  const [liveBalances, setLiveBalances] = useState<BalanceDTO[]>([])
  const [liveDeposits, setLiveDeposits] = useState<DepositDTO[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    let cancelled = false

    async function load() {
      if (!hasApiBase()) {
        setLoading(false)
        return
      }

      try {
        setLoading(true)
        setError('')

        const [chainsResult, tokensResult, depositsResult] = await Promise.all([
          getChains(),
          getTokens(),
          getDeposits(30),
        ])
        if (cancelled) {
          return
        }

        const chains = (chainsResult as ChainDTO[]) ?? []
        const tokens = (tokensResult as TokenDTO[]) ?? []
        const deposits = (depositsResult as DepositDTO[]) ?? []
        const resolvedAddress = configuredDemoAddress || deposits[0]?.to_address || ''

        setLiveChains(chains)
        setLiveTokens(tokens)
        setLiveDeposits(
          resolvedAddress
            ? deposits.filter((deposit) => deposit.to_address.toLowerCase() === resolvedAddress.toLowerCase())
            : deposits,
        )

        if (resolvedAddress) {
          const balancesResult = await getAddressBalances(resolvedAddress)
          if (cancelled) {
            return
          }
          setLiveBalances(balancesResult)
        } else {
          setLiveBalances([])
        }
      } catch (loadError) {
        if (cancelled) {
          return
        }
        setError(loadError instanceof Error ? loadError.message : 'Failed to load asset dashboard data')
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }

    void load()
    return () => {
      cancelled = true
    }
  }, [configuredDemoAddress])

  const chainById = useMemo(
    () => new Map(liveChains.map((chain) => [chain.chain_id, chain])),
    [liveChains],
  )

  const tokenByKey = useMemo(
    () => new Map(liveTokens.map((token) => [`${token.chain_id}:${token.id}`, token])),
    [liveTokens],
  )

  const liveBalanceAssets = useMemo(
    () =>
      liveBalances.map((balance) => {
        const chain = chainById.get(balance.chain_id)
        const token = tokenByKey.get(`${balance.chain_id}:${balance.token_id}`)
        const symbol = token?.symbol || `Token #${balance.token_id}`
        const decimals = token?.decimals ?? 18
        const available = formatTokenAmount(balance.available_balance, decimals)
        const locked = formatTokenAmount(balance.frozen_balance, decimals)
        const total = formatTokenAmount(addIntegerStrings(balance.available_balance, balance.frozen_balance), decimals)

        return {
          chain: chain?.name || `Chain ${balance.chain_id}`,
          token: token?.is_native ? chain?.native_symbol || symbol : symbol,
          symbol,
          available,
          locked,
          value: `${total} ${symbol}`,
          change: 'live',
          contract: token?.is_native ? 'native' : token?.contract_address || `token:${balance.token_id}`,
        }
      }),
    [chainById, liveBalances, tokenByKey],
  )

  const liveDepositRecords = useMemo(
    () =>
      liveDeposits.map((deposit) => {
        const chain = chainById.get(deposit.chain_id)
        const token = tokenByKey.get(`${deposit.chain_id}:${deposit.token_id}`)
        const status = mapDepositStatus(deposit.status)
        const requiredConfirmations = chain?.finality_confirmations || Math.max(deposit.confirmations, 1)
        const amount = formatTokenAmount(deposit.amount, token?.decimals ?? 18)
        const symbol = token?.symbol || `Token #${deposit.token_id}`

        return {
          id: String(deposit.id),
          chain: chain?.name || `Chain ${deposit.chain_id}`,
          token: symbol,
          amount,
          balanceDelta: status === 'reorged' ? '0.00 (reverted)' : `+${amount}`,
          status,
          confirmations: deposit.confirmations,
          requiredConfirmations,
          txHash: deposit.tx_hash,
          account: deposit.to_address,
          block: deposit.block_number.toLocaleString('en-US'),
          timestamp: deposit.created_at,
          rawLog: `Transfer(from=${shortHash(deposit.from_address)}, to=${shortHash(deposit.to_address)}, value=${deposit.amount})`,
          parsedEvent: `token=${symbol}, account=${shortHash(deposit.to_address)}, amount=${amount}`,
          ledgerStatus: deposit.status,
        } satisfies AssetDepositRecord
      }),
    [chainById, liveDeposits, tokenByKey],
  )

  const liveChainFilters = useMemo(() => {
    const names = new Set<string>()
    liveChains.forEach((chain) => names.add(chain.name))
    liveBalanceAssets.forEach((asset) => names.add(asset.chain))
    liveDepositRecords.forEach((deposit) => names.add(deposit.chain))
    return ['All', ...names]
  }, [liveBalanceAssets, liveChains, liveDepositRecords])

  const useFallbackData = !hasApiBase() || Boolean(error)
  const currentBalances = useFallbackData ? balanceAssets : liveBalanceAssets
  const currentDeposits = useFallbackData ? depositRecords : liveDepositRecords
  const currentChainFilters = useFallbackData ? chainFilters : liveChainFilters.length > 0 ? liveChainFilters : ['All']
  const usingLiveData = !useFallbackData

  const visibleBalances = useMemo(
    () => selectedChain === 'All' ? currentBalances : currentBalances.filter((asset) => asset.chain === selectedChain),
    [currentBalances, selectedChain],
  )

  const visibleDeposits = useMemo(
    () => selectedChain === 'All' ? currentDeposits : currentDeposits.filter((deposit) => deposit.chain === selectedChain),
    [currentDeposits, selectedChain],
  )

  useEffect(() => {
    if (visibleDeposits.length === 0) {
      setSelectedDepositId('')
      return
    }

    if (!visibleDeposits.some((deposit) => deposit.id === selectedDepositId)) {
      setSelectedDepositId(visibleDeposits[0].id)
    }
  }, [selectedDepositId, visibleDeposits])

  const selectedDeposit = visibleDeposits.find((deposit) => deposit.id === selectedDepositId) ?? visibleDeposits[0] ?? currentDeposits[0]

  const totals = {
    balance: usingLiveData ? `${currentBalances.length} live assets` : '$215,029.32',
    pending: currentDeposits.filter((deposit) => deposit.status === 'pending' || deposit.status === 'detected').length,
    confirmed: currentDeposits.filter((deposit) => deposit.status === 'confirmed').length,
    alerts: currentDeposits.filter((deposit) => deposit.status === 'reorged').length,
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
          <p>{usingLiveData ? 'Live API mode is reading balances and deposit history from the backend.' : t('assetDashboard.pipelineBody')}</p>
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
          chains={currentChainFilters}
          selectedChain={selectedChain}
          onSelectChain={setSelectedChain}
        />
      </div>

      {loading ? <LoadingState message="Loading live asset data..." /> : null}
      {!loading && error ? <EmptyState title="Live asset data unavailable" description={error} /> : null}

      <div className="asset-layout">
        <div className="asset-main">
          <TokenBalanceTable assets={visibleBalances} />
          {visibleDeposits.length > 0 ? (
            <DepositLifecycle
              deposits={visibleDeposits}
              selectedDeposit={selectedDeposit}
              onSelectDeposit={setSelectedDepositId}
            />
          ) : (
            <section className="panel asset-panel">
              <EmptyState
                title="No deposits for this view"
                description="Deposit events will appear here once the scanner and parser have persisted matching records."
              />
            </section>
          )}
        </div>
        {selectedDeposit ? (
          <DepositDetailPanel deposit={selectedDeposit} />
        ) : (
          <aside className="asset-detail-panel">
            <EmptyState
              title="No deposit selected"
              description="Select a deposit row to inspect its chain-to-ledger lifecycle."
            />
          </aside>
        )}
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

function TokenBalanceTable({ assets }: { assets: AssetRow[] }) {
  const { t } = useTranslation()

  return (
    <section className="panel asset-panel">
      <SectionTitle title={t('assetDashboard.balances')} subtitle={t('assetDashboard.balancesSubtitle')} />
      {assets.length > 0 ? (
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
              <span className={asset.change.startsWith('+') ? 'positive' : asset.change.startsWith('-') ? 'negative' : ''}>{asset.value} <small>{asset.change}</small></span>
              {asset.contract === 'native' ? <span>{asset.contract}</span> : <CopyableHash hash={asset.contract} />}
            </div>
          ))}
        </div>
      ) : (
        <EmptyState
          title="No live balances yet"
          description="Once the tracked address has credited balances, they will appear here from the backend ledger."
        />
      )}
    </section>
  )
}

function DepositLifecycle({
  deposits,
  selectedDeposit,
  onSelectDeposit,
}: {
  deposits: AssetDepositRecord[]
  selectedDeposit: AssetDepositRecord
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

function DepositDetailPanel({ deposit }: { deposit: AssetDepositRecord }) {
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

function mapDepositStatus(status: string): DepositStatus {
  if (status === 'confirmed') return 'confirmed'
  if (status === 'orphaned' || status === 'reorged') return 'reorged'
  if (status === 'pending_confirmation' || status === 'pending') return 'pending'
  return 'detected'
}

function addIntegerStrings(left: string, right: string) {
  try {
    return (BigInt(left || '0') + BigInt(right || '0')).toString()
  } catch {
    return left || right || '0'
  }
}

function formatTokenAmount(value: string, decimals: number) {
  try {
    const normalizedDecimals = Math.max(decimals, 0)
    const negative = value.startsWith('-')
    const digits = negative ? value.slice(1) : value
    const padded = digits.padStart(normalizedDecimals + 1, '0')
    const whole = padded.slice(0, padded.length - normalizedDecimals) || '0'
    const fraction = normalizedDecimals === 0 ? '' : padded.slice(-normalizedDecimals).replace(/0+$/, '')
    const composed = fraction ? `${whole}.${fraction}` : whole
    return `${negative ? '-' : ''}${Number(composed).toLocaleString('en-US', {
      maximumFractionDigits: 6,
    })}`
  } catch {
    return value
  }
}

function shortHash(value: string) {
  if (value.length <= 14) {
    return value
  }
  return `${value.slice(0, 6)}...${value.slice(-4)}`
}
