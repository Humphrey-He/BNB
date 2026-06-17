interface ChainSelectorProps {
  chains: string[]
  selectedChain: string
  onSelectChain: (chain: string) => void
}

export function ChainSelector({ chains, selectedChain, onSelectChain }: ChainSelectorProps) {
  return (
    <div className="chain-selector" role="tablist">
      {chains.map((chain) => (
        <button
          key={chain}
          role="tab"
          aria-selected={selectedChain === chain}
          className={selectedChain === chain ? 'chain-btn active' : 'chain-btn'}
          onClick={() => onSelectChain(chain)}
        >
          {chain}
        </button>
      ))}
    </div>
  )
}
