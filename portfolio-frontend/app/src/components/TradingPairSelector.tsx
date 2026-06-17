import type { TradingPair } from '../types'
import { useState } from 'react'

interface TradingPairSelectorProps {
  pairs: TradingPair[]
  selectedPair: TradingPair | null
  onSelectPair: (pair: TradingPair) => void
}

export function TradingPairSelector({ pairs, selectedPair, onSelectPair }: TradingPairSelectorProps) {
  const [isOpen, setIsOpen] = useState(false)

  return (
    <div className="trading-pair-selector">
      <button
        className="pair-selector-trigger"
        onClick={() => setIsOpen(!isOpen)}
        type="button"
      >
        <span className="pair-symbol">
          {selectedPair ? selectedPair.symbol : 'Select Pair'}
        </span>
        <span className={`pair-arrow ${isOpen ? 'open' : ''}`}>&#9662;</span>
      </button>

      {isOpen && (
        <div className="pair-dropdown">
          {pairs.map((pair) => (
            <button
              key={pair.id}
              className={`pair-option ${selectedPair?.id === pair.id ? 'selected' : ''}`}
              onClick={() => {
                onSelectPair(pair)
                setIsOpen(false)
              }}
              type="button"
            >
              <span className="pair-base">{pair.base}</span>
              <span className="pair-sep">/</span>
              <span className="pair-quote">{pair.quote}</span>
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
