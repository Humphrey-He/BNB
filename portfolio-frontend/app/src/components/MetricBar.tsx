interface MetricBarProps {
  label?: string
  value: number
}

export function MetricBar({ label, value }: MetricBarProps) {
  return (
    <div className="metric-bar">
      {label && (
        <div className="metric-bar-label">
          <span>{label}</span>
          <strong>{value}</strong>
        </div>
      )}
      <div className="bar-track">
        <div className="bar-fill" style={{ width: `${value}%` }} />
      </div>
    </div>
  )
}
