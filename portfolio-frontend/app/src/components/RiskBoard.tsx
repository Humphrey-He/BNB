import { useState, useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { Shield, AlertTriangle, Clock, CheckCircle, XCircle, ExternalLink } from 'lucide-react'
import type { RiskCenterRisk, ProjectId, RiskStatus, Priority } from '../types'

interface RiskBoardProps {
  risks: RiskCenterRisk[]
  onSelectRisk: (risk: RiskCenterRisk) => void
  selectedRiskId?: string
}

const statusIcons: Record<RiskStatus, React.ReactNode> = {
  open: <AlertTriangle className="icon" style={{ color: 'var(--danger)' }} />,
  in_progress: <Clock className="icon" style={{ color: 'var(--accent-yellow)' }} />,
  fixed: <CheckCircle className="icon" style={{ color: 'var(--success)' }} />,
  blocked: <XCircle className="icon" style={{ color: 'var(--danger)' }} />,
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

export function RiskBoard({ risks, onSelectRisk, selectedRiskId }: RiskBoardProps) {
  const { t } = useTranslation()
  const [projectFilter, setProjectFilter] = useState<ProjectId | 'all'>('all')
  const [priorityFilter, setPriorityFilter] = useState<Priority | 'all'>('all')
  const [statusFilter, setStatusFilter] = useState<RiskStatus | 'all'>('all')

  const filteredRisks = useMemo(() => {
    return risks.filter((risk) => {
      if (projectFilter !== 'all' && risk.projectId !== projectFilter) return false
      if (priorityFilter !== 'all' && risk.priority !== priorityFilter) return false
      if (statusFilter !== 'all' && risk.status !== statusFilter) return false
      return true
    })
  }, [risks, projectFilter, priorityFilter, statusFilter])

  const projectCounts = useMemo(() => {
    const counts: Record<string, number> = { all: risks.length }
    risks.forEach((r) => {
      counts[r.projectId] = (counts[r.projectId] || 0) + 1
    })
    return counts
  }, [risks])

  const priorityCounts = useMemo(() => {
    const counts: Record<string, number> = { all: risks.length }
    risks.forEach((r) => {
      counts[r.priority] = (counts[r.priority] || 0) + 1
    })
    return counts
  }, [risks])

  const statusCounts = useMemo(() => {
    const counts: Record<string, number> = { all: risks.length }
    risks.forEach((r) => {
      counts[r.status] = (counts[r.status] || 0) + 1
    })
    return counts
  }, [risks])

  return (
    <div className="risk-board">
      <div className="risk-toolbar">
        <div className="risk-filter-group">
          <span className="risk-filter-label">
            <Shield style={{ width: 14, height: 14 }} />
            {t('riskCenter.project')}:
          </span>
          <div className="risk-toggle">
            {(['all', 'web3-backend', 'protocol-rust', 'smart-contract'] as const).map((p) => (
              <button
                key={p}
                className={projectFilter === p ? 'selected' : ''}
                onClick={() => setProjectFilter(p)}
              >
                {p === 'all' ? t('common.all') : projectLabels[p]}
                <span className="filter-count">{projectCounts[p]}</span>
              </button>
            ))}
          </div>
        </div>

        <div className="risk-filter-group">
          <span className="risk-filter-label">{t('riskCenter.priority')}:</span>
          <div className="risk-toggle">
            {(['all', 'P0', 'P1', 'P2'] as const).map((p) => (
              <button
                key={p}
                className={priorityFilter === p ? 'selected' : ''}
                onClick={() => setPriorityFilter(p)}
              >
                {p === 'all' ? t('common.all') : p}
                <span className="filter-count">{priorityCounts[p]}</span>
              </button>
            ))}
          </div>
        </div>

        <div className="risk-filter-group">
          <span className="risk-filter-label">{t('riskCenter.status')}:</span>
          <div className="risk-toggle">
            {(['all', 'open', 'in_progress', 'fixed', 'blocked'] as const).map((s) => (
              <button
                key={s}
                className={statusFilter === s ? 'selected' : ''}
                onClick={() => setStatusFilter(s)}
              >
                {s === 'all' ? t('common.all') : t(`riskCenter.status_${s}`)}
                <span className="filter-count">{statusCounts[s]}</span>
              </button>
            ))}
          </div>
        </div>
      </div>

      <div className="risk-list">
        {filteredRisks.length === 0 ? (
          <div className="empty-state">
            <Shield style={{ width: 32, height: 32, color: 'var(--text-muted)' }} />
            <p>{t('riskCenter.noRisks')}</p>
          </div>
        ) : (
          filteredRisks.map((risk) => (
            <button
              key={risk.id}
              className={`risk-item ${selectedRiskId === risk.id ? 'selected' : ''}`}
              onClick={() => onSelectRisk(risk)}
            >
              <div className="risk-item-header">
                <span
                  className="priority-indicator"
                  style={{ color: risk.priority === 'P0' ? 'var(--danger)' : risk.priority === 'P1' ? 'var(--accent-yellow)' : 'var(--accent-blue)' }}
                >
                  {risk.priority}
                </span>
                <span
                  className="project-tag"
                  style={{ color: projectColorVars[risk.projectId], borderColor: projectColorVars[risk.projectId] }}
                >
                  {projectLabels[risk.projectId]}
                </span>
                <span className="risk-status-badge" data-status={risk.status}>
                  {statusIcons[risk.status]}
                  {t(`riskCenter.status_${risk.status}`)}
                </span>
              </div>
              <h3 className="risk-title">{risk.title}</h3>
              <p className="risk-impact">{risk.impact}</p>
              {risk.assignee && (
                <div className="risk-assignee">
                  <span>{t('riskCenter.assignee')}:</span>
                  <strong>{risk.assignee}</strong>
                </div>
              )}
              <div className="risk-footer">
                <span className="risk-updated">
                  {t('riskCenter.updated')}: {new Date(risk.updatedAt).toLocaleDateString()}
                </span>
                {risk.evidenceLinks.length > 0 && (
                  <span className="risk-evidence-count">
                    <ExternalLink style={{ width: 12, height: 12 }} />
                    {risk.evidenceLinks.length} {t('riskCenter.evidenceLinks')}
                  </span>
                )}
              </div>
            </button>
          ))
        )}
      </div>
    </div>
  )
}
