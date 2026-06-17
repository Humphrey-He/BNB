import { useState, useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { Shield, AlertTriangle, CheckCircle, XCircle } from 'lucide-react'
import { SectionTitle } from '../components'
import { RiskBoard } from '../components/RiskBoard'
import { RiskDetailDrawer } from '../components/RiskDetailDrawer'
import { risks } from '../data/riskCenter'
import type { RiskCenterRisk } from '../types'

export function RiskCenter() {
  const { t } = useTranslation()
  const [selectedRisk, setSelectedRisk] = useState<RiskCenterRisk | null>(null)

  const summaryStats = useMemo(() => {
    const stats = { total: 0, p0: 0, p1: 0, p2: 0, open: 0, inProgress: 0, fixed: 0, blocked: 0 }
    risks.forEach((r) => {
      stats.total++
      if (r.priority === 'P0') stats.p0++
      if (r.priority === 'P1') stats.p1++
      if (r.priority === 'P2') stats.p2++
      if (r.status === 'open') stats.open++
      if (r.status === 'in_progress') stats.inProgress++
      if (r.status === 'fixed') stats.fixed++
      if (r.status === 'blocked') stats.blocked++
    })
    return stats
  }, [])

  return (
    <div className="risk-center-page">
      <div className="topbar">
        <div>
          <h1>{t('riskCenter.title')}</h1>
          <p>{t('riskCenter.subtitle')}</p>
        </div>
      </div>

      <div className="risk-center-hero">
        <div className="risk-hero-content">
          <div className="terminal-line">risk_center: aggregated_view active</div>
          <h2>{t('riskCenter.heading')}</h2>
          <p>{t('riskCenter.description')}</p>
        </div>
        <div className="risk-summary-cards">
          <div className="risk-summary-card total">
            <Shield style={{ width: 24, height: 24, color: 'var(--accent-cyan)' }} />
            <div className="summary-content">
              <span>{t('riskCenter.totalRisks')}</span>
              <strong>{summaryStats.total}</strong>
            </div>
          </div>
          <div className="risk-summary-card p0">
            <AlertTriangle style={{ width: 24, height: 24, color: 'var(--danger)' }} />
            <div className="summary-content">
              <span>P0</span>
              <strong>{summaryStats.p0}</strong>
            </div>
          </div>
          <div className="risk-summary-card p1">
            <AlertTriangle style={{ width: 24, height: 24, color: 'var(--accent-yellow)' }} />
            <div className="summary-content">
              <span>P1</span>
              <strong>{summaryStats.p1}</strong>
            </div>
          </div>
          <div className="risk-summary-card p2">
            <AlertTriangle style={{ width: 24, height: 24, color: 'var(--accent-blue)' }} />
            <div className="summary-content">
              <span>P2</span>
              <strong>{summaryStats.p2}</strong>
            </div>
          </div>
          <div className="risk-summary-card open">
            <XCircle style={{ width: 24, height: 24, color: 'var(--danger)' }} />
            <div className="summary-content">
              <span>{t('riskCenter.status_open')}</span>
              <strong>{summaryStats.open}</strong>
            </div>
          </div>
          <div className="risk-summary-card fixed">
            <CheckCircle style={{ width: 24, height: 24, color: 'var(--success)' }} />
            <div className="summary-content">
              <span>{t('riskCenter.status_fixed')}</span>
              <strong>{summaryStats.fixed}</strong>
            </div>
          </div>
        </div>
      </div>

      <div className="risk-center-content">
        <div className="risk-board-container">
          <SectionTitle title={t('riskCenter.riskBoard')} subtitle={t('riskCenter.riskBoardSubtitle')} />
          <RiskBoard
            risks={risks}
            onSelectRisk={setSelectedRisk}
            selectedRiskId={selectedRisk?.id}
          />
        </div>
        <div className="risk-drawer-container">
          <RiskDetailDrawer risk={selectedRisk} onClose={() => setSelectedRisk(null)} />
        </div>
      </div>
    </div>
  )
}
