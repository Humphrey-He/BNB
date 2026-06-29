import { useTranslation } from 'react-i18next'
import { Wifi, WifiOff, Shield, Key, Usb } from 'lucide-react'
import { CopyableHash } from './CopyableHash'
import type { WalletAccount } from '../types'

interface WalletOverviewProps {
  account: WalletAccount
}

export function WalletOverview({ account }: WalletOverviewProps) {
  const { t } = useTranslation()

  const walletTypeIcon = () => {
    if (account.type === 'Hardware') return <Usb aria-hidden="true" className="wallet-type-icon" />
    if (account.type === 'Contract') return <Shield aria-hidden="true" className="wallet-type-icon" />
    return <Key aria-hidden="true" className="wallet-type-icon" />
  }

  const walletTypeLabel = () => {
    if (account.type === 'EOA') return t('app.wallet.type.eoa')
    if (account.type === 'Contract') return t('app.wallet.type.contract')
    return t('app.wallet.type.hardware')
  }

  return (
    <article className="wallet-overview panel">
      <div className="wallet-overview-header">
        <div className="wallet-status-badge">
          {account.isConnected ? (
            <>
              <Wifi aria-hidden="true" className="wallet-status-icon connected" />
              <span className="wallet-status-text connected">{t('app.wallet.connected')}</span>
            </>
          ) : (
            <>
              <WifiOff aria-hidden="true" className="wallet-status-icon disconnected" />
              <span className="wallet-status-text disconnected">{t('app.wallet.disconnected')}</span>
            </>
          )}
        </div>
      </div>

      <div className="wallet-address-section">
        <span className="wallet-address-label">{t('app.wallet.address')}</span>
        <div className="wallet-address-row">
          <CopyableHash hash={account.address} />
        </div>
      </div>

      <div className="wallet-meta-row">
        <div className="wallet-meta-item">
          <span className="wallet-meta-label">{t('app.wallet.network')}</span>
          <strong className="wallet-meta-value">{account.network}</strong>
        </div>
        <div className="wallet-meta-item">
          <span className="wallet-meta-label">{t('app.wallet.type.label')}</span>
          <div className="wallet-type-row">
            {walletTypeIcon()}
            <strong className="wallet-meta-value">{walletTypeLabel()}</strong>
          </div>
        </div>
        {account.label && (
          <div className="wallet-meta-item">
            <span className="wallet-meta-label">{t('app.wallet.label')}</span>
            <strong className="wallet-meta-value">{account.label}</strong>
          </div>
        )}
      </div>

      <p className="wallet-disclaimer">{t('app.wallet.readonlyDisclaimer')}</p>
    </article>
  )
}
