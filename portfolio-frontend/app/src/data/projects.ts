import type { Project, IconMap } from '../types'
import {
  LayoutDashboard,
  Server,
  Blocks,
  ShieldCheck,
  GitBranch,
  Terminal,
  Wallet,
  CheckCircle2,
  AlertTriangle,
  ChevronRight,
  Network,
  CircleDot,
  BarChart3,
  ArrowLeftRight,
} from 'lucide-react'

export const iconMap: IconMap = {
  dashboard: LayoutDashboard,
  assets: Wallet,
  ops: Server,
  server: Server,
  blocks: Blocks,
  shield: ShieldCheck,
  compare: GitBranch,
  terminal: Terminal,
  wallet: Wallet,
  check: CheckCircle2,
  alert: AlertTriangle,
  arrow: ChevronRight,
  network: Network,
  dot: CircleDot,
  chart: BarChart3,
  trade: ArrowLeftRight,
}

export const projects: Project[] = [
  {
    id: 'web3-backend',
    key: 'web3Backend',
    status: 'Flagship',
    verification: 'Blocked',
    metrics: { marketFit: 92, depth: 84, testReadiness: 42, riskControl: 68 },
  },
  {
    id: 'protocol-rust',
    key: 'protocolRust',
    status: 'Depth',
    verification: 'Partial',
    metrics: { marketFit: 78, depth: 86, testReadiness: 76, riskControl: 70 },
  },
  {
    id: 'smart-contract',
    key: 'smartContract',
    status: 'Security',
    verification: 'No tests',
    metrics: { marketFit: 58, depth: 62, testReadiness: 12, riskControl: 48 },
  },
]

export function getProjectById(id: string): Project | undefined {
  return projects.find((project) => project.id === id)
}
