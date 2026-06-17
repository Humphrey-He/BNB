import type { View } from '../types'
import { AppIcon } from '../components'
import { useTranslation } from 'react-i18next'

interface SideRailProps {
  currentView: View
  onNavigate: (view: View) => void
}

export function SideRail({ currentView, onNavigate }: SideRailProps) {
  const { t } = useTranslation()

  const navItems: { view: View; icon: 'dashboard' | 'assets' | 'ops' | 'server' | 'compare' | 'terminal' | 'shield' | 'wallet' | 'chart' | 'trade' | 'blocks' | 'alert'; labelKey: string }[] = [
    { view: 'dashboard', icon: 'dashboard', labelKey: 'app.nav.dashboard' },
    { view: 'assets', icon: 'assets', labelKey: 'app.nav.assets' },
    { view: 'ops', icon: 'ops', labelKey: 'app.nav.ops' },
    { view: 'protocol', icon: 'blocks', labelKey: 'app.nav.protocol' },
    { view: 'security', icon: 'shield', labelKey: 'app.nav.securityLab' },
    { view: 'risk', icon: 'alert', labelKey: 'app.nav.risk' },
    { view: 'project', icon: 'server', labelKey: 'app.nav.projects' },
    { view: 'compare', icon: 'compare', labelKey: 'app.nav.compare' },
    { view: 'interview', icon: 'terminal', labelKey: 'app.nav.interview' },
    { view: 'wallet', icon: 'wallet', labelKey: 'app.nav.wallet' },
    { view: 'market', icon: 'chart', labelKey: 'app.nav.market' },
    { view: 'trading', icon: 'trade', labelKey: 'app.nav.trading' },
  ]

  return (
    <aside className="side-rail">
      <div className="brand-mark">BNB</div>
      {navItems.map((item) => (
        <button
          type="button"
          key={item.view}
          className={currentView === item.view ? 'rail-btn active' : 'rail-btn'}
          onClick={() => onNavigate(item.view)}
          title={t(item.labelKey)}
        >
          <AppIcon name={item.icon} />
        </button>
      ))}
    </aside>
  )
}
