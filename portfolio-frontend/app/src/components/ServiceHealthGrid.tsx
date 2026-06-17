import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { AlertCircle, CheckCircle2, Clock3, ChevronDown } from 'lucide-react'
import type { Service } from '../types'
import { SectionTitle } from './SectionTitle'

interface ServiceHealthGridProps {
  services: Service[]
}

export function ServiceHealthGrid({ services }: ServiceHealthGridProps) {
  const { t } = useTranslation()
  const [selectedService, setSelectedService] = useState<Service | null>(null)

  return (
    <section className="panel">
      <SectionTitle title={t('ops.serviceHealth.title')} subtitle={t('ops.serviceHealth.subtitle')} />
      <div className="service-health-grid">
        {services.map((service) => (
          <ServiceCard
            key={service.id}
            service={service}
            selected={selectedService?.id === service.id}
            onClick={() => setSelectedService(selectedService?.id === service.id ? null : service)}
          />
        ))}
      </div>
      {selectedService && (
        <ServiceDetailPanel service={selectedService} onClose={() => setSelectedService(null)} />
      )}
    </section>
  )
}

function ServiceCard({
  service,
  selected,
  onClick,
}: {
  service: Service
  selected: boolean
  onClick: () => void
}) {
  const { t } = useTranslation()

  return (
    <button
      className={`service-card ${service.status} ${selected ? 'selected' : ''}`}
      onClick={onClick}
    >
      <div className="service-card-header">
        <span className={`service-status-icon ${service.status}`}>
          {service.status === 'healthy' && <CheckCircle2 aria-hidden="true" />}
          {service.status === 'degraded' && <Clock3 aria-hidden="true" />}
          {service.status === 'down' && <AlertCircle aria-hidden="true" />}
        </span>
        <span className="service-name">{service.name}</span>
      </div>
      <div className="service-card-metrics">
        <div className="service-metric">
          <span>{t('ops.serviceHealth.uptime')}</span>
          <strong>{service.uptime}</strong>
        </div>
        <div className="service-metric">
          <span>{t('ops.serviceHealth.latency')}</span>
          <strong>{service.avgLatency}</strong>
        </div>
        <div className="service-metric">
          <span>{t('ops.serviceHealth.errors')}</span>
          <strong className={service.errorCount > 0 ? 'text-danger' : ''}>{service.errorCount}</strong>
        </div>
      </div>
      <div className="service-card-footer">
        <span>{t('ops.serviceHealth.lastHeartbeat')}: {service.lastHeartbeat}</span>
      </div>
    </button>
  )
}

function ServiceDetailPanel({ service, onClose }: { service: Service; onClose: () => void }) {
  const { t } = useTranslation()

  return (
    <div className="service-detail-panel">
      <div className="service-detail-header">
        <div>
          <span className={`service-status-badge ${service.status}`}>
            {t(`ops.serviceHealth.status.${service.status}`)}
          </span>
          <h3>{service.name}</h3>
        </div>
        <button className="detail-close-btn" onClick={onClose}>
          <ChevronDown aria-hidden="true" />
        </button>
      </div>
      <div className="service-detail-grid">
        <div className="service-detail-item">
          <span>{t('ops.serviceHealth.uptime')}</span>
          <strong>{service.uptime}</strong>
        </div>
        <div className="service-detail-item">
          <span>{t('ops.serviceHealth.latency')}</span>
          <strong>{service.avgLatency}</strong>
        </div>
        <div className="service-detail-item">
          <span>{t('ops.serviceHealth.errors')}</span>
          <strong className={service.errorCount > 0 ? 'text-danger' : ''}>{service.errorCount}</strong>
        </div>
        <div className="service-detail-item">
          <span>{t('ops.serviceHealth.lastHeartbeat')}</span>
          <strong>{service.lastHeartbeat}</strong>
        </div>
      </div>
      {service.nextAction && (
        <div className="service-next-action">
          <span>{t('ops.serviceHealth.nextAction')}</span>
          <p>{service.nextAction}</p>
        </div>
      )}
    </div>
  )
}
