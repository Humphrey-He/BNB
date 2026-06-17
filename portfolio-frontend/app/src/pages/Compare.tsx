import { useTranslation } from 'react-i18next'
import { projects } from '../data/projects'
import { MetricBar, SectionTitle } from '../components'

function useProjectText(projectKey: string) {
  const { t } = useTranslation()
  return t(`projects.${projectKey}`, { returnObjects: true }) as {
    shortName: string
    role: string
    hrPitch: string
  }
}

export function Compare() {
  const { t } = useTranslation()

  return (
    <section className="compare-page">
      <SectionTitle title={t('compare.title')} subtitle={t('compare.subtitle')} />
      <div className="compare-table">
        <div className="compare-head">
          <span>{t('compare.project')}</span>
          <span>{t('compare.role')}</span>
          <span>{t('compare.marketFit')}</span>
          <span>{t('compare.testReadiness')}</span>
          <span>{t('compare.positioning')}</span>
        </div>
        {projects.map((project) => (
          <CompareRow key={project.id} project={project} />
        ))}
      </div>
      <div className="decision-panel">
        <h2>{t('compare.packageTitle')}</h2>
        <p>{t('compare.packageBody')}</p>
      </div>
    </section>
  )
}

function CompareRow({ project }: { project: (typeof projects)[number] }) {
  const text = useProjectText(project.key)
  return (
    <div className="compare-row">
      <strong>{text.shortName}</strong>
      <span>{text.role}</span>
      <MetricBar value={project.metrics.marketFit} />
      <MetricBar value={project.metrics.testReadiness} />
      <p>{text.hrPitch}</p>
    </div>
  )
}
