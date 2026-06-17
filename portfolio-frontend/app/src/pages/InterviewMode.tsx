import type { TimeBudget, ProjectId } from '../types'
import { projects } from '../data/projects'
import { useTranslation } from 'react-i18next'

interface InterviewModeProps {
  timeBudget: TimeBudget
  setTimeBudget: (value: TimeBudget) => void
  onOpenProject: (id: ProjectId) => void
}

export function InterviewMode({ timeBudget, setTimeBudget, onOpenProject }: InterviewModeProps) {
  const { t } = useTranslation()
  const steps = t(`interview.steps.${timeBudget}`, { returnObjects: true }) as string[]

  return (
    <section className="interview-page">
      <div className="interview-header">
        <div>
          <h2>{t('interview.title', { minutes: timeBudget })}</h2>
          <p>{t('interview.seniorBody')}</p>
        </div>
        <div className="segmented">
          {(['3', '8', '20'] as const).map((item) => (
            <button
              key={item}
              className={timeBudget === item ? 'selected' : ''}
              onClick={() => setTimeBudget(item)}
            >
              {item} {t('common.min')}
            </button>
          ))}
        </div>
      </div>
      <div className="talk-track">
        {steps.map((step, index) => (
          <div className="talk-step" key={step}>
            <span>{String(index + 1).padStart(2, '0')}</span>
            <p>{step}</p>
          </div>
        ))}
      </div>
      <div className="interview-projects">
        {projects.map((project) => (
          <InterviewProjectButton
            key={project.id}
            project={project}
            onClick={() => onOpenProject(project.id)}
          />
        ))}
      </div>
    </section>
  )
}

function InterviewProjectButton({
  project,
  onClick,
}: {
  project: (typeof projects)[number]
  onClick: () => void
}) {
  const { t } = useTranslation()
  const text = t(`projects.${project.key}`, { returnObjects: true }) as { shortName: string; role: string }

  return (
    <button onClick={onClick}>
      <strong>{text.shortName}</strong>
      <span>{text.role}</span>
    </button>
  )
}
