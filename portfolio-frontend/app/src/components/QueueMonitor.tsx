import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { AlertTriangle, CheckCircle2, XCircle, Filter } from 'lucide-react'
import type { QueueEntry, QueueStatus } from '../types'
import { SectionTitle } from './SectionTitle'

interface QueueMonitorProps {
  queues: QueueEntry[]
}

export function QueueMonitor({ queues }: QueueMonitorProps) {
  const { t } = useTranslation()
  const [filter, setFilter] = useState<'all' | QueueStatus>('all')

  const topics = [...new Set(queues.map((q) => q.topic))]
  const [topicFilter, setTopicFilter] = useState<string>('all')

  const filteredQueues = queues.filter((q) => {
    if (filter !== 'all' && q.status !== filter) return false
    if (topicFilter !== 'all' && q.topic !== topicFilter) return false
    return true
  })

  return (
    <section className="panel">
      <SectionTitle title={t('ops.queue.title')} subtitle={t('ops.queue.subtitle')} />
      <div className="queue-toolbar">
        <div className="queue-filter-group">
          <Filter aria-hidden="true" />
          <span>{t('ops.queue.statusFilter')}:</span>
          {(['all', 'active', 'backlogged', 'dead'] as const).map((status) => (
            <button
              key={status}
              className={filter === status ? 'filter-btn active' : 'filter-btn'}
              onClick={() => setFilter(status)}
            >
              {status === 'all' ? t('ops.queue.all') : t(`ops.queue.status.${status}`)}
            </button>
          ))}
        </div>
        <div className="queue-filter-group">
          <span>{t('ops.queue.topicFilter')}:</span>
          <button
            className={topicFilter === 'all' ? 'filter-btn active' : 'filter-btn'}
            onClick={() => setTopicFilter('all')}
          >
            {t('ops.queue.all')}
          </button>
          {topics.map((topic) => (
            <button
              key={topic}
              className={topicFilter === topic ? 'filter-btn active' : 'filter-btn'}
              onClick={() => setTopicFilter(topic)}
            >
              {topic}
            </button>
          ))}
        </div>
      </div>
      <div className="queue-table">
        <div className="queue-table-head">
          <span>{t('ops.queue.topic')}</span>
          <span>{t('ops.queue.status')}</span>
          <span>{t('ops.queue.backlog')}</span>
          <span>{t('ops.queue.retry')}</span>
          <span>{t('ops.queue.deadLetter')}</span>
          <span>{t('ops.queue.avgProcessTime')}</span>
        </div>
        {filteredQueues.map((queue) => (
          <QueueRow key={queue.id} queue={queue} />
        ))}
      </div>
    </section>
  )
}

function QueueRow({ queue }: { queue: QueueEntry }) {
  const { t } = useTranslation()

  return (
    <div className={`queue-row ${queue.status}`}>
      <span className="queue-topic">{queue.topic}</span>
      <span className={`queue-status-badge ${queue.status}`}>
        {queue.status === 'active' && <CheckCircle2 aria-hidden="true" />}
        {queue.status === 'backlogged' && <AlertTriangle aria-hidden="true" />}
        {queue.status === 'dead' && <XCircle aria-hidden="true" />}
        {t(`ops.queue.status.${queue.status}`)}
      </span>
      <span className={`queue-backlog ${queue.backlogCount > 1000 ? 'text-danger' : queue.backlogCount > 100 ? 'text-warning' : ''}`}>
        {queue.backlogCount.toLocaleString()}
      </span>
      <span className={queue.retryCount > 0 ? 'text-warning' : ''}>{queue.retryCount}</span>
      <span className={queue.deadLetterCount > 0 ? 'text-danger' : ''}>{queue.deadLetterCount}</span>
      <span className="queue-process-time">{queue.avgProcessTime}</span>
    </div>
  )
}
