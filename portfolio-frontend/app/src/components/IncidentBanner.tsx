import { useTranslation } from 'react-i18next'
import { AlertTriangle, XCircle, Info, CheckCircle2 } from 'lucide-react'
import type { Incident } from '../types'

interface IncidentBannerProps {
  incidents: Incident[]
}

export function IncidentBanner({ incidents }: IncidentBannerProps) {
  const { t } = useTranslation()

  if (incidents.length === 0) return null

  return (
    <div className="incident-banner">
      {incidents.map((incident) => (
        <div
          key={incident.id}
          className={`incident-item ${incident.severity} ${incident.acknowledged ? 'acknowledged' : ''}`}
        >
          <span className="incident-icon">
            {incident.severity === 'critical' && <XCircle aria-hidden="true" />}
            {incident.severity === 'warning' && <AlertTriangle aria-hidden="true" />}
            {incident.severity === 'info' && <Info aria-hidden="true" />}
          </span>
          <div className="incident-content">
            <strong>{incident.service}</strong>
            <p>{incident.message}</p>
          </div>
          <span className="incident-time">{incident.timestamp}</span>
          {incident.acknowledged && (
            <span className="incident-ack">
              <CheckCircle2 aria-hidden="true" />
              {t('ops.incident.acknowledged')}
            </span>
          )}
        </div>
      ))}
    </div>
  )
}
