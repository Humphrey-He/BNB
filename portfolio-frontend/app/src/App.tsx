import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import type { ProjectId, TimeBudget, View } from './types'
import { projects } from './data/projects'
import { AppShell } from './components'
import {
  AssetDashboard,
  Compare,
  Dashboard,
  InterviewMode,
  Market,
  OpsConsole,
  ProjectDetail,
  ProtocolVisualizer,
  RiskCenter,
  SecurityLab,
  Trading,
  Wallet,
} from './pages'

type RouteState = {
  view: View
  projectId: ProjectId
  routeKey: string
}

type Direction = 'forward' | 'backward' | 'replace'

const defaultProjectId: ProjectId = 'web3-backend'

const routeOrder: View[] = [
  'dashboard',
  'assets',
  'ops',
  'protocol',
  'security',
  'risk',
  'project',
  'compare',
  'interview',
  'wallet',
  'market',
  'trading',
]

const viewPaths: Record<Exclude<View, 'project'>, string> = {
  dashboard: '/',
  assets: '/assets',
  ops: '/ops',
  protocol: '/protocol',
  security: '/security',
  compare: '/compare',
  interview: '/interview',
  risk: '/risk',
  wallet: '/wallet',
  market: '/market',
  trading: '/trading',
}

function parsePath(pathname: string): RouteState {
  const [, section, detail] = pathname.split('/')

  if (section === 'projects') {
    const projectId = projects.some((project) => project.id === detail)
      ? (detail as ProjectId)
      : defaultProjectId

    return {
      view: 'project',
      projectId,
      routeKey: `/projects/${projectId}`,
    }
  }

  const view = (Object.entries(viewPaths).find(([, path]) => path === pathname)?.[0] ??
    'dashboard') as View

  return {
    view,
    projectId: defaultProjectId,
    routeKey: viewPaths[view as Exclude<View, 'project'>],
  }
}

function pathFor(view: View, projectId: ProjectId = defaultProjectId) {
  if (view === 'project') {
    return `/projects/${projectId}`
  }

  return viewPaths[view]
}

function getDirection(from: View, to: View): Direction {
  if (from === to) {
    return 'replace'
  }

  return routeOrder.indexOf(to) >= routeOrder.indexOf(from) ? 'forward' : 'backward'
}

function App() {
  const [route, setRoute] = useState<RouteState>(() => parsePath(window.location.pathname))
  const [timeBudget, setTimeBudget] = useState<TimeBudget>('8')
  const [wallet, setWallet] = useState(false)
  const [direction, setDirection] = useState<Direction>('replace')
  const previousView = useRef(route.view)

  const current = useMemo(
    () => projects.find((project) => project.id === route.projectId) ?? projects[0],
    [route.projectId],
  )

  const pushRoute = useCallback((nextRoute: RouteState) => {
    const nextPath = pathFor(nextRoute.view, nextRoute.projectId)
    setDirection(getDirection(previousView.current, nextRoute.view))
    previousView.current = nextRoute.view
    window.history.pushState(null, '', nextPath)
    setRoute(nextRoute)
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }, [])

  const navigate = useCallback(
    (view: View) => {
      pushRoute({
        view,
        projectId: view === 'project' ? route.projectId : defaultProjectId,
        routeKey: pathFor(view, view === 'project' ? route.projectId : defaultProjectId),
      })
    },
    [pushRoute, route.projectId],
  )

  const navigateProject = useCallback(
    (id: ProjectId) => {
      pushRoute({
        view: 'project',
        projectId: id,
        routeKey: pathFor('project', id),
      })
    },
    [pushRoute],
  )

  useEffect(() => {
    const handlePopState = () => {
      const nextRoute = parsePath(window.location.pathname)
      setDirection(getDirection(previousView.current, nextRoute.view))
      previousView.current = nextRoute.view
      setRoute(nextRoute)
    }

    window.addEventListener('popstate', handlePopState)
    return () => window.removeEventListener('popstate', handlePopState)
  }, [])

  return (
    <AppShell
      currentView={route.view}
      onNavigate={navigate}
      wallet={wallet}
      onWalletToggle={() => setWallet((value) => !value)}
    >
      <div className={`route-frame ${direction}`} key={route.routeKey}>
        {route.view === 'dashboard' && (
          <Dashboard onOpenProject={navigateProject} setView={navigate} />
        )}
        {route.view === 'assets' && <AssetDashboard />}
        {route.view === 'ops' && <OpsConsole />}
        {route.view === 'protocol' && <ProtocolVisualizer />}
        {route.view === 'security' && <SecurityLab />}
        {route.view === 'project' && (
          <ProjectDetail project={current} onSelectProject={navigateProject} />
        )}
        {route.view === 'compare' && <Compare />}
        {route.view === 'interview' && (
          <InterviewMode
            timeBudget={timeBudget}
            setTimeBudget={setTimeBudget}
            onOpenProject={navigateProject}
          />
        )}
        {route.view === 'risk' && <RiskCenter />}
        {route.view === 'wallet' && <Wallet />}
        {route.view === 'market' && <Market />}
        {route.view === 'trading' && <Trading />}
      </div>
    </AppShell>
  )
}

export default App
