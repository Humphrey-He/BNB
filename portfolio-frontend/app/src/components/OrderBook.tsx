import type { OrderBookEntry } from '../types'

interface OrderBookProps {
  bids: OrderBookEntry[]
  asks: OrderBookEntry[]
  symbol?: string
}

export function OrderBook({ bids, asks, symbol = 'BTC/USDT' }: OrderBookProps) {
  const maxTotal = Math.max(
    ...bids.map((b) => parseFloat(b.total.replace(/,/g, ''))),
    ...asks.map((a) => parseFloat(a.total.replace(/,/g, ''))),
  )

  return (
    <div className="order-book">
      <div className="order-book-header">
        <h3>Order Book</h3>
        <span className="order-book-symbol">{symbol}</span>
      </div>

      <div className="order-book-columns">
        <span>Price</span>
        <span>Amount</span>
        <span>Total</span>
      </div>

      <div className="order-book-asks">
        {[...asks].reverse().map((ask, i) => {
          const totalNum = parseFloat(ask.total.replace(/,/g, ''))
          const widthPercent = (totalNum / maxTotal) * 100
          return (
            <div key={`ask-${i}`} className="order-book-row ask">
              <div className="depth-bar ask" style={{ width: `${widthPercent}%` }} />
              <span className="price">{ask.price}</span>
              <span className="amount">{ask.amount}</span>
              <span className="total">{ask.total}</span>
            </div>
          )
        })}
      </div>

      <div className="order-book-spread">
        <span>Spread</span>
        <span>0.50 (0.0007%)</span>
      </div>

      <div className="order-book-bids">
        {bids.map((bid, i) => {
          const totalNum = parseFloat(bid.total.replace(/,/g, ''))
          const widthPercent = (totalNum / maxTotal) * 100
          return (
            <div key={`bid-${i}`} className="order-book-row bid">
              <div className="depth-bar bid" style={{ width: `${widthPercent}%` }} />
              <span className="price">{bid.price}</span>
              <span className="amount">{bid.amount}</span>
              <span className="total">{bid.total}</span>
            </div>
          )
        })}
      </div>
    </div>
  )
}
