import { useTranslation } from 'react-i18next'
import type { ProjectId, View } from '../types'
import { projects } from '../data/projects'
import { AppIcon, SectionTitle, StatusPill, Metric } from '../components'

function useProjectText(projectKey: string) {
  const { t } = useTranslation()
  return t(`projects.${projectKey}`, { returnObjects: true })
}

interface DashboardProps {
  onOpenProject: (id: ProjectId) => void
  setView: (view: View) => void
}

export function Dashboard({ onOpenProject, setView }: DashboardProps) {
  const { t } = useTranslation()

  return (
    <section className="dashboard-grid">
      <div className="hero-panel">
        <div className="hero-copy">
          <div className="terminal-line">{t('dashboard.terminal')}</div>
          <h2>{t('dashboard.title')}</h2>
          <p>{t('dashboard.body')}</p>
          <div className="hero-actions">
            <button className="primary-btn" onClick={() => onOpenProject('web3-backend')}>
              {t('dashboard.openBackend')}
            </button>
            <button className="secondary-btn" onClick={() => setView('interview')}>
              {t('dashboard.interviewMode')}
            </button>
          </div>
        </div>
        <div className="strategy-panel">
          <span>{t('dashboard.strategyLabel')}</span>
          <strong>{t('dashboard.strategyTitle')}</strong>
          <p>{t('dashboard.strategyBody')}</p>
        </div>
      </div>

      <div className="project-cards">
        {projects.map((project) => (
          <ProjectCard key={project.id} project={project} onOpen={() => onOpenProject(project.id)} />
        ))}
      </div>

      <div className="matrix-panel">
        <SectionTitle title={t('dashboard.matrixTitle')} subtitle={t('dashboard.matrixSubtitle')} />
        <div className="matrix">
          {projects.map((project) => (
            <MatrixDot key={project.id} project={project} />
          ))}
        </div>
      </div>

      <div className="blockers-panel">
        <SectionTitle title={t('dashboard.blockersTitle')} subtitle={t('dashboard.blockersSubtitle')} />
        {projects.map((project) => (
          <BlockerRow key={project.id} project={project} />
        ))}
      </div>

      <RoadmapPanel />
    </section>
  )
}

function MatrixDot({ project }: { project: (typeof projects)[number] }) {
  const text = useProjectText(project.key)
  return (
    <div
      className={`matrix-dot ${project.id}`}
      style={{ left: `${project.metrics.marketFit}%`, bottom: `${project.metrics.depth}%` }}
    >
      <span>{(text as { shortName: string }).shortName}</span>
    </div>
  )
}

function BlockerRow({ project }: { project: (typeof projects)[number] }) {
  const text = useProjectText(project.key)
  return (
    <div className="blocker-row">
      <StatusPill value={project.verification} />
      <span>{(text as { shortName: string }).shortName}</span>
      <p>{(text as { findings: { title: string }[] }).findings[0].title}</p>
    </div>
  )
}

function ProjectCard({ project, onOpen }: { project: (typeof projects)[number]; onOpen: () => void }) {
  const { t } = useTranslation()
  const text = useProjectText(project.key) as {
    shortName: string
    role: string
    tagline: string
    stack: string[]
    findings: { title: string }[]
  }

  return (
    <article className={`project-card ${project.id}`}>
      <div className="project-card-head">
        <div className="project-icon">
          <AppIcon
            name={
              project.id === 'web3-backend'
                ? 'server'
                : project.id === 'protocol-rust'
                  ? 'blocks'
                  : 'shield'
            }
          />
        </div>
        <div>
          <h3>{text.shortName}</h3>
          <p>{text.role}</p>
        </div>
      </div>
      <p className="project-tagline">{text.tagline}</p>
      <div className="stack-row">
        {text.stack.slice(0, 4).map((item) => (
          <span key={item}>{item}</span>
        ))}
      </div>
      <div className="metric-pair">
        <Metric label={t('compare.marketFit')} value={project.metrics.marketFit} />
        <Metric label={t('detail.technicalDepth')} value={project.metrics.depth} />
      </div>
      <div className="main-risk">
        <AppIcon name="alert" />
        <span>{t('common.mainRisk')}</span>
        <p>{text.findings[0].title}</p>
      </div>
      <div className="card-footer">
        <StatusPill value={project.verification} />
        <button onClick={onOpen}>
          {t('common.openDeepDive')} <AppIcon name="arrow" />
        </button>
      </div>
    </article>
  )
}

function RoadmapPanel() {
  const { t } = useTranslation()
  const items = t('roadmap.items', { returnObjects: true }) as [string, string, string][]

  return (
    <section className="roadmap-panel">
      <SectionTitle title={t('roadmap.title')} subtitle={t('roadmap.subtitle')} />
      <div className="roadmap">
        {items.map(([week, area, text]) => (
          <div className="roadmap-item" key={week}>
            <span>{week}</span>
            <strong>{area}</strong>
            <p>{text}</p>
          </div>
        ))}
      </div>
    </section>
  )
}
