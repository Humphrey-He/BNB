import type { PriceTicker as PriceTickerType } from '../types'
import { PriceTicker } from './PriceTicker'

interface MarketOverviewProps {
  tickers: PriceTickerType[]
  selectedSymbol?: string
  onSelectSymbol?: (symbol: string) => void
}

export function MarketOverview({ tickers, selectedSymbol, onSelectSymbol }: MarketOverviewProps) {
  return (
    <div className="market-overview">
      <div className="market-tickers-grid">
        {tickers.map((ticker) => (
          <PriceTicker
            key={ticker.symbol}
            ticker={ticker}
            selected={selectedSymbol === ticker.symbol}
            onClick={() => onSelectSymbol?.(ticker.symbol)}
          />
        ))}
      </div>
    </div>
  )
}
