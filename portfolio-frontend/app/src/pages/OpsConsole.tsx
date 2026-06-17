import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { OpsView } from '../types'
import {
  services,
  rpcProviders,
  queueEntries,
  reorgEvents,
  outboxEntries,
  incidents,
} from '../data/ops'
import { ServiceHealthGrid } from '../components/ServiceHealthGrid'
import { RpcProviderTable } from '../components/RpcProviderTable'
import { QueueMonitor } from '../components/QueueMonitor'
import { ReorgMonitor } from '../components/ReorgMonitor'
import { OutboxStatusPanel } from '../components/OutboxStatusPanel'
import { IncidentBanner } from '../components/IncidentBanner'

export function OpsConsole() {
  const { t } = useTranslation()
  const [currentView, setCurrentView] = useState<OpsView>('health')

  const views: { key: OpsView; label: string }[] = [
    { key: 'health', label: t('ops.nav.serviceHealth') },
    { key: 'rpc', label: t('ops.nav.rpc') },
    { key: 'queue', label: t('ops.nav.queue') },
    { key: 'reorg', label: t('ops.nav.reorg') },
    { key: 'outbox', label: t('ops.nav.outbox') },
  ]

  return (
    <section className="ops-page">
      <div className="ops-hero">
        <div>
          <span className="terminal-line">{t('ops.terminal')}</span>
          <h2>{t('ops.title')}</h2>
          <p>{t('ops.subtitle')}</p>
        </div>
        <IncidentBanner incidents={incidents.filter((i) => !i.acknowledged)} />
      </div>

      <div className="ops-nav">
        {views.map((view) => (
          <button
            key={view.key}
            className={currentView === view.key ? 'ops-nav-btn active' : 'ops-nav-btn'}
            onClick={() => setCurrentView(view.key)}
          >
            {view.label}
          </button>
        ))}
      </div>

      <div className="ops-content">
        {currentView === 'health' && <ServiceHealthGrid services={services} />}
        {currentView === 'rpc' && <RpcProviderTable providers={rpcProviders} />}
        {currentView === 'queue' && <QueueMonitor queues={queueEntries} />}
        {currentView === 'reorg' && <ReorgMonitor events={reorgEvents} />}
        {currentView === 'outbox' && <OutboxStatusPanel entries={outboxEntries} />}
      </div>
    </section>
  )
}
