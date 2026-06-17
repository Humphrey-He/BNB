import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Clock3, CheckCircle2, XCircle, RotateCcw } from 'lucide-react'
import type { OutboxEntry, OutboxStatus } from '../types'
import { SectionTitle } from './SectionTitle'

interface OutboxStatusPanelProps {
  entries: OutboxEntry[]
}

export function OutboxStatusPanel({ entries }: OutboxStatusPanelProps) {
  const { t } = useTranslation()
  const [statusFilter, setStatusFilter] = useState<OutboxStatus | 'all'>('all')

  const filteredEntries = statusFilter === 'all'
    ? entries
    : entries.filter((e) => e.status === statusFilter)

  return (
    <section className="panel">
      <SectionTitle title={t('ops.outbox.title')} subtitle={t('ops.outbox.subtitle')} />
      <div className="outbox-toolbar">
        {(['all', 'pending', 'published', 'failed'] as const).map((status) => (
          <button
            key={status}
            className={statusFilter === status ? 'filter-btn active' : 'filter-btn'}
            onClick={() => setStatusFilter(status)}
          >
            {status === 'all' ? t('ops.outbox.all') : t(`ops.outbox.status.${status}`)}
          </button>
        ))}
      </div>
      <div className="outbox-list">
        {filteredEntries.map((entry) => (
          <OutboxEntryRow key={entry.id} entry={entry} />
        ))}
      </div>
    </section>
  )
}

function OutboxEntryRow({ entry }: { entry: OutboxEntry }) {
  const { t } = useTranslation()

  return (
    <div className={`outbox-entry ${entry.status}`}>
      <div className="outbox-entry-main">
        <span className={`outbox-status-icon ${entry.status}`}>
          {entry.status === 'pending' && <Clock3 aria-hidden="true" />}
          {entry.status === 'published' && <CheckCircle2 aria-hidden="true" />}
          {entry.status === 'failed' && <XCircle aria-hidden="true" />}
        </span>
        <div className="outbox-entry-info">
          <strong>{entry.event}</strong>
          <span>ID: {entry.id}</span>
        </div>
      </div>
      <div className="outbox-entry-meta">
        <span className={`outbox-status-badge ${entry.status}`}>
          {t(`ops.outbox.status.${entry.status}`)}
        </span>
        {entry.retryCount > 0 && (
          <span className="outbox-retry">
            <RotateCcw aria-hidden="true" />
            {entry.retryCount}
          </span>
        )}
      </div>
      <div className="outbox-entry-time">
        <span>{t('ops.outbox.created')}: {entry.createdAt}</span>
        {entry.publishedAt && (
          <span>{t('ops.outbox.published')}: {entry.publishedAt}</span>
        )}
      </div>
      {entry.lastError && (
        <div className="outbox-error">
          <span>{t('ops.outbox.lastError')}:</span>
          <p>{entry.lastError}</p>
        </div>
      )}
    </div>
  )
}
