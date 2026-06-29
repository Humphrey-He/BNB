import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { ProjectId, View } from '../types'
import { projects } from '../data/projects'
import { AppIcon, EmptyState, LoadingState, SectionTitle, StatusPill, Metric } from '../components'
import {
  getAddressBalances,
  getChains,
  getDemoWalletAddress,
  getDeposits,
  hasApiBase,
  type BalanceDTO,
  type ChainDTO,
  type DepositDTO,
} from '../lib/api'

function useProjectText(projectKey: string) {
  const { t } = useTranslation()
  return t(`projects.${projectKey}`, { returnObjects: true })
}

interface DashboardProps {
  onOpenProject: (id: ProjectId) => void
  setView: (view: View) => void
}

const releaseCards = [
  {
    title: 'v0.1.0-beta',
    body: 'Core backend, withdrawal pipeline fixes, contract dependency recovery and release docs are in place.',
  },
  {
    title: 'Next target',
    body: 'Push toward a testnet-verified withdrawal loop with receipt settlement and operational retry confidence.',
  },
]

export function Dashboard({ onOpenProject, setView }: DashboardProps) {
  const { t } = useTranslation()
  const [chains, setChains] = useState<ChainDTO[]>([])
  const [balances, setBalances] = useState<BalanceDTO[]>([])
  const [deposits, setDeposits] = useState<DepositDTO[]>([])
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

        const [chainsResult, depositResult] = await Promise.all([getChains(), getDeposits(20)])
        if (cancelled) {
          return
        }

        const liveChains = (chainsResult as ChainDTO[]) ?? []
        const liveDeposits = (depositResult as DepositDTO[]) ?? []
        const resolvedAddress = getDemoWalletAddress() || liveDeposits[0]?.to_address || ''

        setChains(liveChains)
        setDeposits(liveDeposits)

        if (resolvedAddress) {
          const balanceResult = await getAddressBalances(resolvedAddress)
          if (cancelled) {
            return
          }
          setBalances(balanceResult)
        } else {
          setBalances([])
        }
      } catch (loadError) {
        if (cancelled) {
          return
        }
        setError(loadError instanceof Error ? loadError.message : 'Failed to load dashboard data')
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
  }, [])

  const dashboardStats = useMemo(() => {
    const totalAvailable = balances.reduce((sum, item) => sum + Number(item.available_balance || 0), 0)
    const activeChains = chains.filter((item) => item.is_active).length
    const pendingDeposits = deposits.filter((item) => item.status === 'detected' || item.status === 'pending_confirmation').length

    return [
      { label: 'Available units', value: formatCompact(totalAvailable), tone: 'yellow' },
      { label: 'Chains online', value: String(activeChains || chains.length || 0), tone: 'cyan' },
      { label: 'Pending deposits', value: String(pendingDeposits), tone: 'red' },
      { label: 'Tracked deposits', value: String(deposits.length), tone: 'green' },
    ] as const
  }, [balances, chains, deposits])

  const chainRows = useMemo(
    () =>
      chains.slice(0, 6).map((item) => ({
        name: item.name,
        active: item.is_active,
        finality: item.finality_confirmations,
        nativeSymbol: item.native_symbol,
        chainId: item.chain_id,
      })),
    [chains],
  )

  const depositRows = useMemo(
    () =>
      deposits.slice(0, 4).map((item) => ({
        id: item.id,
        status: item.status,
        amount: item.amount,
        hash: shortHash(item.tx_hash),
        chain: item.chain_id,
        confirmations: item.confirmations,
      })),
    [deposits],
  )

  return (
    <section className="dashboard-grid dashboard-premium">
      <div className="hero-panel dashboard-hero-premium">
        <div className="hero-copy">
          <div className="terminal-line">{t('dashboard.terminal')}</div>
          <div className="hero-eyebrow">
            <span>BNB Asset Control</span>
            <span className="hero-live-dot" />
            <span>{hasApiBase() ? 'API-linked dashboard' : 'Demo mode dashboard'}</span>
          </div>
          <h2>{t('dashboard.title')}</h2>
          <p>{t('dashboard.body')}</p>
          <div className="hero-actions">
            <button className="primary-btn" onClick={() => onOpenProject('web3-backend')}>
              {t('dashboard.openBackend')}
            </button>
            <button className="secondary-btn" onClick={() => setView('assets')}>
              Open Asset Surface
            </button>
            <button className="secondary-btn" onClick={() => setView('ops')}>
              Open Ops Console
            </button>
          </div>
        </div>

        <div className="hero-spotlight">
          {loading ? (
            <LoadingState message="Loading backend telemetry..." />
          ) : error ? (
            <EmptyState
              title="Backend data unavailable"
              description={error}
            />
          ) : (
            <>
              <div className="hero-spotlight-grid">
                {dashboardStats.map((item) => (
                  <div className={`hero-stat-card ${item.tone}`} key={item.label}>
                    <span>{item.label}</span>
                    <strong>{item.value}</strong>
                  </div>
                ))}
              </div>

              <div className="hero-flow-card">
                <div className="flow-card-head">
                  <span>Live control plane</span>
                  <strong>Chain health and recent deposits</strong>
                </div>
                <div className="dashboard-live-grid">
                  <div className="live-list">
                    {chainRows.length > 0 ? (
                      chainRows.map((item) => (
                        <div className="live-row" key={item.name}>
                          <div>
                            <strong>{item.name}</strong>
                            <p>{`${item.nativeSymbol} · chain ${item.chainId}`}</p>
                          </div>
                          <div className={item.active ? 'live-status healthy' : 'live-status danger'}>
                            {item.active ? `finality ${item.finality}` : 'inactive'}
                          </div>
                        </div>
                      ))
                    ) : (
                      <EmptyState title="No chain data yet" description="Public chain configuration will appear here once the API is reachable." />
                    )}
                  </div>

                  <div className="live-list">
                    {depositRows.length > 0 ? (
                      depositRows.map((item) => (
                        <div className="live-row" key={item.id}>
                          <div>
                            <strong>{item.amount}</strong>
                            <p>{`chain ${item.chain} · ${item.hash}`}</p>
                          </div>
                          <div className="live-status neutral">{`${item.status} · ${item.confirmations}`}</div>
                        </div>
                      ))
                    ) : (
                      <EmptyState title="No recent deposits" description="Deposit events will appear here once the backend is ingesting chain data." />
                    )}
                  </div>
                </div>
              </div>
            </>
          )}
        </div>
      </div>

      <section className="dashboard-surface-grid">
        <div className="surface-primary panel">
          <SectionTitle
            title="Control Surfaces"
            subtitle="The homepage behaves like a product cockpit instead of a portfolio shelf."
          />
          <div className="surface-cards">
            <button className="surface-card primary" onClick={() => setView('assets')}>
              <div className="surface-card-head">
                <AppIcon name="wallet" />
                <span>Assets</span>
              </div>
              <strong>Balances, confirmations, ledger truth</strong>
              <p>Track available and frozen balances with deposit lifecycle evidence.</p>
            </button>

            <button className="surface-card cyan" onClick={() => setView('ops')}>
              <div className="surface-card-head">
                <AppIcon name="server" />
                <span>Infrastructure</span>
              </div>
              <strong>RPC health, queue depth, reorg risk</strong>
              <p>Operate the platform like a live service, not a static demo page.</p>
            </button>

            <button className="surface-card purple" onClick={() => setView('risk')}>
              <div className="surface-card-head">
                <AppIcon name="shield" />
                <span>Risk</span>
              </div>
              <strong>Manual review and delivery blockers</strong>
              <p>Surface the real issues that still separate beta from production.</p>
            </button>
          </div>
        </div>

        <div className="surface-side panel">
          <SectionTitle
            title="Release Narrative"
            subtitle="What this version says about the project today."
          />
          <div className="release-card-list">
            {releaseCards.map((card) => (
              <div className="release-card" key={card.title}>
                <span>{card.title}</span>
                <p>{card.body}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      <div className="project-cards dashboard-project-cards">
        {projects.map((project) => (
          <ProjectCard key={project.id} project={project} onOpen={() => onOpenProject(project.id)} />
        ))}
      </div>

      <div className="matrix-panel">
        <SectionTitle title={t('dashboard.matrixTitle')} subtitle={t('dashboard.matrixSubtitle')} />
        <div className="matrix">
          {projects.map((project) => (
            <MatrixDot key={project.id} project={project} />
          ))}
        </div>
      </div>

      <div className="blockers-panel">
        <SectionTitle title={t('dashboard.blockersTitle')} subtitle={t('dashboard.blockersSubtitle')} />
        {projects.map((project) => (
          <BlockerRow key={project.id} project={project} />
        ))}
      </div>

      <RoadmapPanel />
    </section>
  )
}

function MatrixDot({ project }: { project: (typeof projects)[number] }) {
  const text = useProjectText(project.key)
  return (
    <div
      className={`matrix-dot ${project.id}`}
      style={{ left: `${project.metrics.marketFit}%`, bottom: `${project.metrics.depth}%` }}
    >
      <span>{(text as { shortName: string }).shortName}</span>
    </div>
  )
}

function BlockerRow({ project }: { project: (typeof projects)[number] }) {
  const text = useProjectText(project.key)
  return (
    <div className="blocker-row">
      <StatusPill value={project.verification} />
      <span>{(text as { shortName: string }).shortName}</span>
      <p>{(text as { findings: { title: string }[] }).findings[0].title}</p>
    </div>
  )
}

function ProjectCard({ project, onOpen }: { project: (typeof projects)[number]; onOpen: () => void }) {
  const { t } = useTranslation()
  const text = useProjectText(project.key) as {
    shortName: string
    role: string
    tagline: string
    stack: string[]
    findings: { title: string }[]
  }

  return (
    <article className={`project-card ${project.id}`}>
      <div className="project-card-head">
        <div className="project-icon">
          <AppIcon
            name={
              project.id === 'web3-backend'
                ? 'server'
                : project.id === 'protocol-rust'
                  ? 'blocks'
                  : 'shield'
            }
          />
        </div>
        <div>
          <h3>{text.shortName}</h3>
          <p>{text.role}</p>
        </div>
      </div>
      <p className="project-tagline">{text.tagline}</p>
      <div className="stack-row">
        {text.stack.slice(0, 4).map((item) => (
          <span key={item}>{item}</span>
        ))}
      </div>
      <div className="metric-pair">
        <Metric label={t('compare.marketFit')} value={project.metrics.marketFit} />
        <Metric label={t('detail.technicalDepth')} value={project.metrics.depth} />
      </div>
      <div className="main-risk">
        <AppIcon name="alert" />
        <span>{t('common.mainRisk')}</span>
        <p>{text.findings[0].title}</p>
      </div>
      <div className="card-footer">
        <StatusPill value={project.verification} />
        <button onClick={onOpen}>
          {t('common.openDeepDive')} <AppIcon name="arrow" />
        </button>
      </div>
    </article>
  )
}

function RoadmapPanel() {
  const { t } = useTranslation()
  const items = t('roadmap.items', { returnObjects: true }) as [string, string, string][]

  return (
    <section className="roadmap-panel">
      <SectionTitle title={t('roadmap.title')} subtitle={t('roadmap.subtitle')} />
      <div className="roadmap">
        {items.map(([week, area, text]) => (
          <div className="roadmap-item" key={week}>
            <span>{week}</span>
            <strong>{area}</strong>
            <p>{text}</p>
          </div>
        ))}
      </div>
    </section>
  )
}

function formatCompact(value: number) {
  if (!Number.isFinite(value) || value <= 0) {
    return '0'
  }
  if (value >= 1_000_000_000) {
    return `${(value / 1_000_000_000).toFixed(2)}B`
  }
  if (value >= 1_000_000) {
    return `${(value / 1_000_000).toFixed(2)}M`
  }
  if (value >= 1_000) {
    return `${(value / 1_000).toFixed(2)}K`
  }
  return value.toFixed(2)
}

function shortHash(value: string) {
  if (value.length <= 14) {
    return value
  }
  return `${value.slice(0, 6)}...${value.slice(-4)}`
}
