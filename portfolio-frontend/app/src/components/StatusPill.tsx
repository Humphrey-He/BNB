import { useTranslation } from 'react-i18next'
import { XCircle } from 'lucide-react'
import type { Verification, AppIconName } from '../types'
import { AppIcon } from './AppIcon'

interface StatusPillProps {
  value: Verification
}

export function StatusPill({ value }: StatusPillProps) {
  const { t } = useTranslation()
  const iconName: AppIconName = value === 'Partial' ? 'dot' : 'alert'

  return (
    <span className={`status-pill ${value.toLowerCase().replace(' ', '-')}`}>
      {value === 'Blocked' || value === 'No tests' ? (
        <XCircle aria-hidden="true" className="status-icon" />
      ) : (
        <AppIcon name={iconName} className="status-icon" />
      )}
      {t(`status.${value}`)}
    </span>
  )
}
