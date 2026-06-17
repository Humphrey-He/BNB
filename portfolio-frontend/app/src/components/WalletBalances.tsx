import { useTranslation } from 'react-i18next'
import { ArrowDownLeft, Lock } from 'lucide-react'
import { balanceAssets } from '../data/assets'
import { SectionTitle } from './SectionTitle'

export function WalletBalances() {
  const { t } = useTranslation()

  const totalValue = balanceAssets.reduce((sum, asset) => {
    const valueStr = asset.value.replace(/[$,]/g, '')
    return sum + parseFloat(valueStr)
  }, 0)

  const totalLocked = balanceAssets.reduce((sum, asset) => {
    const lockedStr = asset.locked.replace(/,/g, '')
    return sum + parseFloat(lockedStr)
  }, 0)

  return (
    <section className="wallet-balances panel">
      <SectionTitle title={t('wallet.balances.title')} subtitle={t('wallet.balances.subtitle')} />

      <div className="wallet-balance-summary">
        <div className="wallet-balance-total">
          <span className="wallet-balance-label">{t('wallet.balances.totalValue')}</span>
          <strong className="wallet-balance-value">${totalValue.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</strong>
        </div>
        <div className="wallet-balance-locked">
          <Lock aria-hidden="true" className="wallet-locked-icon" />
          <span className="wallet-balance-label">{t('wallet.balances.locked')}</span>
          <strong className="wallet-balance-locked-value">${totalLocked.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</strong>
        </div>
      </div>

      <div className="wallet-chain-list">
        {balanceAssets.map((asset) => (
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
              <span className={`wallet-balance-change ${asset.change.startsWith('+') ? 'positive' : 'negative'}`}>
                {asset.change}
              </span>
            </div>
          </div>
        ))}
      </div>
    </section>
  )
}
