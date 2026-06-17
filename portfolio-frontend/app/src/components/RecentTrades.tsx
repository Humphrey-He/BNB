import type { Trade } from '../types'

interface RecentTradesProps {
  trades: Trade[]
  symbol?: string
}

export function RecentTrades({ trades, symbol = 'BTC/USDT' }: RecentTradesProps) {
  return (
    <div className="recent-trades">
      <div className="recent-trades-header">
        <h3>Recent Trades</h3>
        <span className="recent-trades-symbol">{symbol}</span>
      </div>

      <div className="recent-trades-columns">
        <span>Type</span>
        <span>Price</span>
        <span>Amount</span>
        <span>Time</span>
      </div>

      <div className="recent-trades-list">
        {trades.map((trade) => (
          <div key={trade.id} className={`trade-row ${trade.type}`}>
            <span className={`trade-type ${trade.type}`}>
              {trade.type === 'buy' ? 'B' : 'S'}
            </span>
            <span className={`trade-price ${trade.type}`}>{trade.price}</span>
            <span className="trade-amount">{trade.amount}</span>
            <span className="trade-time">{trade.timestamp}</span>
          </div>
        ))}
      </div>
    </div>
  )
}
