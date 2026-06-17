import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Link2 } from 'lucide-react'
import { WalletOverview } from '../components/WalletOverview'
import { WalletBalances } from '../components/WalletBalances'
import { WalletActivity } from '../components/WalletActivity'
import { walletAccount, walletTransactions } from '../data/wallet'

export function Wallet() {
  const { t } = useTranslation()
  const [isConnected, setIsConnected] = useState(walletAccount.isConnected)

  const displayAccount = { ...walletAccount, isConnected }

  return (
    <section className="wallet-page">
      <div className="wallet-hero">
        <div className="wallet-hero-content">
          <span className="terminal-line">{t('wallet.terminal')}</span>
          <h2>{t('wallet.title')}</h2>
          <p>{t('wallet.subtitle')}</p>
        </div>
        <button
          className={`wallet-connect-btn ${isConnected ? 'disconnect' : 'connect'}`}
          onClick={() => setIsConnected(!isConnected)}
        >
          <Link2 aria-hidden="true" className="wallet-connect-icon" />
          {isConnected ? t('wallet.disconnect') : t('wallet.connect')}
        </button>
      </div>

      <div className="wallet-layout">
        <div className="wallet-main">
          <WalletOverview account={displayAccount} />
          <WalletBalances />
        </div>
        <div className="wallet-side">
          <WalletActivity transactions={walletTransactions} />
        </div>
      </div>
    </section>
  )
}
