import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { ChevronDown, RefreshCcw, AlertTriangle, CheckCircle2 } from 'lucide-react'
import type { ReorgEvent } from '../types'
import { SectionTitle } from './SectionTitle'

interface ReorgMonitorProps {
  events: ReorgEvent[]
}

export function ReorgMonitor({ events }: ReorgMonitorProps) {
  const { t } = useTranslation()

  return (
    <section className="panel">
      <SectionTitle title={t('ops.reorg.title')} subtitle={t('ops.reorg.subtitle')} />
      <div className="reorg-timeline">
        {events.map((event) => (
          <ReorgEventCard key={event.id} event={event} />
        ))}
      </div>
    </section>
  )
}

function ReorgEventCard({ event }: { event: ReorgEvent }) {
  const { t } = useTranslation()
  const [expanded, setExpanded] = useState(false)

  return (
    <div className={`reorg-card ${event.severity}`}>
      <button className="reorg-card-header" onClick={() => setExpanded(!expanded)}>
        <div className="reorg-severity-indicator">
          {event.severity === 'high' && <AlertTriangle aria-hidden="true" />}
          {event.severity === 'medium' && <RefreshCcw aria-hidden="true" />}
          {event.severity === 'low' && <CheckCircle2 aria-hidden="true" />}
        </div>
        <div className="reorg-info">
          <span className="reorg-chain">{event.chain}</span>
          <span className="reorg-time">{event.detectedAt}</span>
        </div>
        <div className="reorg-metrics">
          <span className={`reorg-severity-badge ${event.severity}`}>
            {t(`ops.reorg.severity.${event.severity}`)}
          </span>
          <span className="reorg-depth">{event.depth} blocks</span>
        </div>
        <ChevronDown
          aria-hidden="true"
          className={`reorg-expand-icon ${expanded ? 'expanded' : ''}`}
        />
      </button>
      {expanded && (
        <div className="reorg-detail">
          <div className="reorg-blocks">
            <span>{t('ops.reorg.affectedBlocks')}:</span>
            <div className="reorg-block-list">
              {event.affectedBlocks.map((block) => (
                <span key={block} className="reorg-block">{block}</span>
              ))}
            </div>
          </div>
          {event.resolvedAt && (
            <div className="reorg-resolved">
              <span>{t('ops.reorg.resolvedAt')}: {event.resolvedAt}</span>
            </div>
          )}
          <div className="reorg-compensation">
            <span>{t('ops.reorg.compensation')}:</span>
            <p>{event.compensation}</p>
          </div>
        </div>
      )}
    </div>
  )
}
