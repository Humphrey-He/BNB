import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Filter } from 'lucide-react'
import type { Project, Priority, ProjectId } from '../types'
import { projects } from '../data/projects'
import { AppIcon, StatusPill, MetricBar, SectionTitle } from '../components'

function useProjectText(projectKey: string) {
  const { t } = useTranslation()
  return t(`projects.${projectKey}`, { returnObjects: true }) as {
    name: string
    shortName: string
    role: string
    tagline: string
    repo: string
    stack: string[]
    pipeline: string[]
    wins: string[]
    findings: { priority: Priority; title: string; impact: string; fix: string }[]
    hrPitch: string
    seniorPitch: string
    nextMoves: string[]
  }
}

interface ProjectDetailProps {
  project: Project
  onSelectProject: (id: ProjectId) => void
}

export function ProjectDetail({ project, onSelectProject }: ProjectDetailProps) {
  const { t } = useTranslation()
  const text = useProjectText(project.key)
  const pitch = text.seniorPitch

  return (
    <section className="detail-page">
      <div className="project-tabs">
        {projects.map((item) => (
          <ProjectTab
            key={item.id}
            project={item}
            selected={item.id === project.id}
            onClick={() => onSelectProject(item.id)}
          />
        ))}
      </div>

      <div className="detail-header">
        <div>
          <span className="repo-path">{text.repo}</span>
          <h2>{text.name}</h2>
          <p>{pitch}</p>
        </div>
        <div className="verification-card">
          <span>{t('common.verification')}</span>
          <StatusPill value={project.verification} />
          <strong>{t(`status.${project.status}`)}</strong>
        </div>
      </div>

      <div className="detail-layout">
        <div className="main-column">
          <ArchitectureMap steps={text.pipeline} />
          <section className="panel">
            <SectionTitle
              title={t('detail.evidenceTitle')}
              subtitle={t('detail.evidenceSenior')}
            />
            <div className="evidence-grid">
              {text.wins.map((win) => (
                <div className="evidence-item" key={win}>
                  {win}
                </div>
              ))}
            </div>
          </section>
          <section className="panel">
            <SectionTitle title={t('detail.nextFixesTitle')} subtitle={t('detail.nextFixesSubtitle')} />
            <div className="todo-list">
              {text.nextMoves.map((move, index) => (
                <div key={move} className="todo-row">
                  <span>{index + 1}</span>
                  <p>{move}</p>
                </div>
              ))}
            </div>
          </section>
        </div>
        <aside className="side-column">
          <ScorePanel project={project} />
          <RiskPanel findings={text.findings} />
        </aside>
      </div>
    </section>
  )
}

function ProjectTab({
  project,
  selected,
  onClick,
}: {
  project: Project
  selected: boolean
  onClick: () => void
}) {
  const text = useProjectText(project.key)
  return (
    <button className={selected ? 'selected' : ''} onClick={onClick}>
      {text.shortName}
    </button>
  )
}

function ArchitectureMap({ steps }: { steps: string[] }) {
  const { t } = useTranslation()
  const [activeIndex, setActiveIndex] = useState(0)
  const activeStep = steps[activeIndex]

  return (
    <section className="panel">
      <SectionTitle title={t('detail.architectureTitle')} subtitle={t('detail.architectureSubtitle')} />
      <div className="pipeline">
        {steps.map((step, index) => (
          <button
            aria-pressed={activeIndex === index}
            className={activeIndex === index ? 'pipeline-step active' : 'pipeline-step'}
            key={step}
            onClick={() => setActiveIndex(index)}
            onFocus={() => setActiveIndex(index)}
            onMouseEnter={() => setActiveIndex(index)}
          >
            <span>{index + 1}</span>
            <p>{step}</p>
          </button>
        ))}
      </div>
      <div className="pipeline-inspector">
        <div>
          <AppIcon name="network" />
          <span>
            {t('common.step')} {String(activeIndex + 1).padStart(2, '0')}
          </span>
        </div>
        <p>
          <strong>{activeStep}</strong> {t('detail.architectureBody')}
        </p>
      </div>
    </section>
  )
}

function ScorePanel({ project }: { project: Project }) {
  const { t } = useTranslation()
  return (
    <section className="panel compact">
      <SectionTitle title={t('detail.scoreTitle')} subtitle={t('detail.scoreSubtitle')} />
      <MetricBar label={t('detail.marketFit')} value={project.metrics.marketFit} />
      <MetricBar label={t('detail.technicalDepth')} value={project.metrics.depth} />
      <MetricBar label={t('detail.testReadiness')} value={project.metrics.testReadiness} />
      <MetricBar label={t('detail.riskControl')} value={project.metrics.riskControl} />
    </section>
  )
}

function RiskPanel({
  findings,
}: {
  findings: { priority: Priority; title: string; impact: string; fix: string }[]
}) {
  const { t } = useTranslation()
  const [severity, setSeverity] = useState<'all' | Priority>('all')
  const [focus, setFocus] = useState<'impact' | 'fix'>('fix')
  const visibleFindings = severity === 'all' ? findings : findings.filter((finding) => finding.priority === severity)

  return (
    <section className="panel compact">
      <SectionTitle
        title={t('detail.riskTitle')}
        subtitle={t('detail.riskSenior')}
      />
      <div className="risk-controls">
        <div className="risk-control-label">
          <Filter aria-hidden="true" />
          <span>{t('common.severity')}</span>
        </div>
        <div className="risk-filter-row">
          {(['all', 'P0', 'P1', 'P2'] as const).map((item) => (
            <button
              key={item}
              className={severity === item ? 'filter-btn active' : 'filter-btn'}
              onClick={() => setSeverity(item)}
            >
              {item === 'all' ? t('common.all') : item}
            </button>
          ))}
        </div>
        <div className="risk-toggle">
          <button className={focus === 'impact' ? 'selected' : ''} onClick={() => setFocus('impact')}>
            {t('common.impact')}
          </button>
          <button className={focus === 'fix' ? 'selected' : ''} onClick={() => setFocus('fix')}>
            {t('common.fixPlan')}
          </button>
        </div>
      </div>
      <div className="risk-list">
        {visibleFindings.map((finding) => (
          <article className="risk-item" key={finding.title}>
            <div>
              <span className={`priority ${finding.priority.toLowerCase()}`}>{finding.priority}</span>
              <strong>{finding.title}</strong>
            </div>
            <p>{focus === 'impact' ? finding.impact : finding.fix}</p>
          </article>
        ))}
      </div>
    </section>
  )
}
