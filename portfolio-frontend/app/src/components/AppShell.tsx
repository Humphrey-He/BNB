import type { ReactNode } from 'react'
import type { View } from '../types'
import { SideRail } from './SideRail'
import { Topbar } from './Topbar'

interface AppShellProps {
  children: ReactNode
  currentView: View
  onNavigate: (view: View) => void
  wallet: boolean
  onWalletToggle: () => void
}

export function AppShell({
  children,
  currentView,
  onNavigate,
  wallet,
  onWalletToggle,
}: AppShellProps) {
  return (
    <div className="app-shell">
      <SideRail currentView={currentView} onNavigate={onNavigate} />
      <main className="workspace">
        <Topbar
          wallet={wallet}
          onWalletToggle={onWalletToggle}
        />
        {children}
      </main>
    </div>
  )
}
