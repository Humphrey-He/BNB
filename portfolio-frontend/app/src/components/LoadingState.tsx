import { Loader2 } from 'lucide-react'

interface LoadingStateProps {
  message?: string
}

export function LoadingState({ message = 'Loading...' }: LoadingStateProps) {
  return (
    <div className="loading-state">
      <Loader2 className="spinner" aria-hidden="true" />
      <span>{message}</span>
    </div>
  )
}
