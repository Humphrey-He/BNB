import type { AppIconName } from '../types'
import { iconMap } from '../data/projects'

interface AppIconProps {
  name: AppIconName
  className?: string
}

export function AppIcon({ name, className = 'icon' }: AppIconProps) {
  const Component = iconMap[name]
  return <Component aria-hidden="true" className={className} strokeWidth={1.8} />
}
