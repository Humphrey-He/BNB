import { useTranslation } from 'react-i18next'
import { ArrowDownLeft, Lock } from 'lucide-react'
import { SectionTitle } from './SectionTitle'

type WalletBalanceRow = {
  chain: string
  token: string
  symbol: string
  available: string
  locked: string
  value: string
  change: string
}

interface WalletBalancesProps {
  assets: WalletBalanceRow[]
}

export function WalletBalances({ assets }: WalletBalancesProps) {
  const { t } = useTranslation()

  const totalLocked = assets.reduce((sum, asset) => {
    const lockedStr = asset.locked.replace(/,/g, '')
    return sum + parseFloat(lockedStr)
  }, 0)

  return (
    <section className="wallet-balances panel">
      <SectionTitle title={t('app.wallet.balances.title')} subtitle={t('app.wallet.balances.subtitle')} />

      <div className="wallet-balance-summary">
        <div className="wallet-balance-total">
          <span className="wallet-balance-label">{t('app.wallet.balances.totalValue')}</span>
          <strong className="wallet-balance-value">{`${assets.length} live positions`}</strong>
        </div>
        <div className="wallet-balance-locked">
          <Lock aria-hidden="true" className="wallet-locked-icon" />
          <span className="wallet-balance-label">{t('app.wallet.balances.locked')}</span>
          <strong className="wallet-balance-locked-value">{totalLocked.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</strong>
        </div>
      </div>

      <div className="wallet-chain-list">
        {assets.map((asset) => (
          <div key={`${asset.chain}-${asset.symbol}`} className="wallet-chain-row">
            <div className="wallet-chain-info">
              <span className="wallet-chain-name">{asset.chain}</span>
              <span className="wallet-chain-token">{asset.token}</span>
            </div>
            <div className="wallet-chain-balances">
              <div className="wallet-balance-item">
                <ArrowDownLeft aria-hidden="true" className="wallet-balance-icon available" />
                <span className="wallet-balance-amount">{asset.available}</span>
                <span className="wallet-balance-symbol">{asset.symbol}</span>
              </div>
              {parseFloat(asset.locked.replace(/,/g, '')) > 0 && (
                <div className="wallet-balance-item locked">
                  <Lock aria-hidden="true" className="wallet-balance-icon locked" />
                  <span className="wallet-balance-amount">{asset.locked}</span>
                  <span className="wallet-balance-symbol">{asset.symbol}</span>
                </div>
              )}
            </div>
            <div className="wallet-chain-value">
              <strong>{asset.value}</strong>
              <span className={`wallet-balance-change ${asset.change.startsWith('+') ? 'positive' : asset.change.startsWith('-') ? 'negative' : ''}`}>
                {asset.change}
              </span>
            </div>
          </div>
        ))}
      </div>
    </section>
  )
}
