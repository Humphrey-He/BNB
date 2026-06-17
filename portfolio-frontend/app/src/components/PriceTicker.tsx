import type { PriceTicker as PriceTickerType } from '../types'

interface PriceTickerProps {
  ticker: PriceTickerType
  onClick?: () => void
  selected?: boolean
}

export function PriceTicker({ ticker, onClick, selected }: PriceTickerProps) {
  const isPositive = !ticker.change24h.startsWith('-')

  return (
    <button
      className={`price-ticker-card ${selected ? 'selected' : ''}`}
      onClick={onClick}
      type="button"
    >
      <div className="ticker-header">
        <span className="ticker-symbol">{ticker.symbol}</span>
        <span className="ticker-name">{ticker.name}</span>
      </div>
      <div className="ticker-price">${ticker.price}</div>
      <div className={`ticker-change ${isPositive ? 'positive' : 'negative'}`}>
        {ticker.change24h}
      </div>
      <div className="ticker-meta">
        <span>Vol: {ticker.volume24h}</span>
      </div>
      <div className="ticker-range">
        <div className="range-item">
          <span>H</span>
          <strong>{ticker.high24h}</strong>
        </div>
        <div className="range-item">
          <span>L</span>
          <strong>{ticker.low24h}</strong>
        </div>
      </div>
    </button>
  )
}
