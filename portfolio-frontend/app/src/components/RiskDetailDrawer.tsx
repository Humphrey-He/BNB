import { useTranslation } from 'react-i18next'
import { X, AlertTriangle, Clock, CheckCircle, XCircle, ExternalLink, Calendar, User, Link2 } from 'lucide-react'
import type { RiskCenterRisk, ProjectId, RiskStatus } from '../types'

interface RiskDetailDrawerProps {
  risk: RiskCenterRisk | null
  onClose: () => void
}

const projectLabels: Record<ProjectId, string> = {
  'web3-backend': 'Web3 Backend',
  'protocol-rust': 'Rust Protocol',
  'smart-contract': 'Smart Contract',
}

const projectColorVars: Record<ProjectId, string> = {
  'web3-backend': 'var(--accent-yellow)',
  'protocol-rust': 'var(--accent-cyan)',
  'smart-contract': 'var(--purple)',
}

const statusConfig: Record<RiskStatus, { icon: React.ReactNode; labelKey: string; color: string }> = {
  open: { icon: <AlertTriangle style={{ width: 16, height: 16 }} />, labelKey: 'riskCenter.status_open', color: 'var(--danger)' },
  in_progress: { icon: <Clock style={{ width: 16, height: 16 }} />, labelKey: 'riskCenter.status_in_progress', color: 'var(--accent-yellow)' },
  fixed: { icon: <CheckCircle style={{ width: 16, height: 16 }} />, labelKey: 'riskCenter.status_fixed', color: 'var(--success)' },
  blocked: { icon: <XCircle style={{ width: 16, height: 16 }} />, labelKey: 'riskCenter.status_blocked', color: 'var(--danger)' },
}

export function RiskDetailDrawer({ risk, onClose }: RiskDetailDrawerProps) {
  const { t } = useTranslation()

  if (!risk) return null

  const status = statusConfig[risk.status]

  return (
    <div className="risk-detail-drawer">
      <div className="drawer-header">
        <div className="drawer-title-row">
          <span
            className="priority-badge"
            style={{
              color: risk.priority === 'P0' ? 'var(--danger)' : risk.priority === 'P1' ? 'var(--accent-yellow)' : 'var(--accent-blue)',
              background: risk.priority === 'P0' ? 'rgba(246, 70, 93, 0.15)' : risk.priority === 'P1' ? 'rgba(240, 185, 11, 0.15)' : 'rgba(55, 125, 255, 0.15)',
              borderColor: risk.priority === 'P0' ? 'rgba(246, 70, 93, 0.4)' : risk.priority === 'P1' ? 'rgba(240, 185, 11, 0.4)' : 'rgba(55, 125, 255, 0.4)',
            }}
          >
            {risk.priority}
          </span>
          <span
            className="project-badge"
            style={{ color: projectColorVars[risk.projectId], borderColor: projectColorVars[risk.projectId] }}
          >
            {projectLabels[risk.projectId]}
          </span>
        </div>
        <button className="drawer-close-btn" onClick={onClose} aria-label="Close">
          <X style={{ width: 20, height: 20 }} />
        </button>
      </div>

      <h2 className="drawer-risk-title">{risk.title}</h2>

      <div className="drawer-status-row">
        <span className="drawer-status-badge" style={{ color: status.color }}>
          {status.icon}
          {t(status.labelKey)}
        </span>
      </div>

      <div className="drawer-section">
        <h3>
          <AlertTriangle style={{ width: 16, height: 16 }} />
          {t('riskCenter.impact')}
        </h3>
        <p>{risk.impact}</p>
      </div>

      <div className="drawer-section">
        <h3>
          <CheckCircle style={{ width: 16, height: 16 }} />
          {t('riskCenter.fixPlan')}
        </h3>
        <p>{risk.fixPlan}</p>
      </div>

      <div className="drawer-section">
        <h3>
          <Link2 style={{ width: 16, height: 16 }} />
          {t('riskCenter.evidenceLinks')} ({risk.evidenceLinks.length})
        </h3>
        <div className="evidence-links-list">
          {risk.evidenceLinks.length === 0 ? (
            <p className="no-evidence">{t('riskCenter.noEvidenceLinks')}</p>
          ) : (
            risk.evidenceLinks.map((link, idx) => (
              <a
                key={idx}
                href={link}
                target="_blank"
                rel="noopener noreferrer"
                className="evidence-link-item"
              >
                <ExternalLink style={{ width: 14, height: 14 }} />
                <span className="evidence-link-text">{link}</span>
              </a>
            ))
          )}
        </div>
      </div>

      <div className="drawer-meta">
        <div className="drawer-meta-item">
          <Calendar style={{ width: 14, height: 14 }} />
          <span className="meta-label">{t('riskCenter.created')}:</span>
          <span className="meta-value">{new Date(risk.createdAt).toLocaleDateString()}</span>
        </div>
        <div className="drawer-meta-item">
          <Calendar style={{ width: 14, height: 14 }} />
          <span className="meta-label">{t('riskCenter.updated')}:</span>
          <span className="meta-value">{new Date(risk.updatedAt).toLocaleDateString()}</span>
        </div>
        {risk.assignee && (
          <div className="drawer-meta-item">
            <User style={{ width: 14, height: 14 }} />
            <span className="meta-label">{t('riskCenter.assignee')}:</span>
            <span className="meta-value">{risk.assignee}</span>
          </div>
        )}
      </div>

      <div className="drawer-timeline">
        <h3>{t('riskCenter.fixProgress')}</h3>
        <div className="fix-timeline">
          <div className={`timeline-item ${risk.status === 'fixed' ? 'completed' : ''}`}>
            <div className="timeline-dot" />
            <div className="timeline-content">
              <span className="timeline-label">{t('riskCenter.created')}</span>
              <span className="timeline-date">{new Date(risk.createdAt).toLocaleDateString()}</span>
            </div>
          </div>
          {risk.status !== 'open' && (
            <div className={`timeline-item ${risk.status === 'fixed' || risk.status === 'blocked' ? 'completed' : 'active'}`}>
              <div className="timeline-dot" />
              <div className="timeline-content">
                <span className="timeline-label">
                  {risk.status === 'in_progress' ? t('riskCenter.inProgress') : risk.status === 'blocked' ? t('riskCenter.blocked') : t('riskCenter.started')}
                </span>
                <span className="timeline-date">{new Date(risk.updatedAt).toLocaleDateString()}</span>
              </div>
            </div>
          )}
          {risk.status === 'fixed' && (
            <div className="timeline-item completed">
              <div className="timeline-dot" />
              <div className="timeline-content">
                <span className="timeline-label">{t('riskCenter.fixed')}</span>
                <span className="timeline-date">{new Date(risk.updatedAt).toLocaleDateString()}</span>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
