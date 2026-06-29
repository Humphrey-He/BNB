import { useTranslation } from 'react-i18next'
import { ArrowUpRight, ArrowDownLeft, RefreshCw, ShieldCheck, Clock3, XCircle, CheckCircle2 } from 'lucide-react'
import { CopyableHash } from './CopyableHash'
import type { WalletTransaction, WalletTransaction as WalletTransactionType } from '../types'

interface WalletActivityProps {
  transactions: WalletTransaction[]
}

export function WalletActivity({ transactions }: WalletActivityProps) {
  const { t } = useTranslation()

  const txTypeIcon = (type: WalletTransactionType['type']) => {
    switch (type) {
      case 'send':
        return <ArrowUpRight aria-hidden="true" className="tx-type-icon send" />
      case 'receive':
        return <ArrowDownLeft aria-hidden="true" className="tx-type-icon receive" />
      case 'swap':
        return <RefreshCw aria-hidden="true" className="tx-type-icon swap" />
      case 'approve':
        return <ShieldCheck aria-hidden="true" className="tx-type-icon approve" />
    }
  }

  const txTypeLabel = (type: WalletTransactionType['type']) => {
    return t(`app.wallet.activity.type.${type}`)
  }

  const txStatusIcon = (status: WalletTransactionType['status']) => {
    switch (status) {
      case 'pending':
        return <Clock3 aria-hidden="true" className="tx-status-icon pending" />
      case 'confirmed':
        return <CheckCircle2 aria-hidden="true" className="tx-status-icon confirmed" />
      case 'failed':
        return <XCircle aria-hidden="true" className="tx-status-icon failed" />
    }
  }

  return (
    <section className="wallet-activity panel">
      <div className="wallet-activity-header">
        <h3 className="wallet-activity-title">{t('app.wallet.activity.title')}</h3>
        <span className="wallet-activity-subtitle">{t('app.wallet.activity.subtitle')}</span>
      </div>

      <div className="wallet-tx-list">
        {transactions.map((tx) => (
          <div key={tx.id} className={`wallet-tx-row status-${tx.status}`}>
            <div className="wallet-tx-type">
              {txTypeIcon(tx.type)}
              <div className="wallet-tx-type-info">
                <span className="wallet-tx-type-label">{txTypeLabel(tx.type)}</span>
                <span className="wallet-tx-amount">
                  {tx.amount} {tx.token}
                </span>
              </div>
            </div>

            <div className="wallet-tx-addresses">
              <div className="wallet-tx-address">
                <span className="wallet-tx-address-label">{t('app.wallet.activity.from')}</span>
                <CopyableHash hash={tx.from} truncate />
              </div>
              <div className="wallet-tx-address">
                <span className="wallet-tx-address-label">{t('app.wallet.activity.to')}</span>
                <CopyableHash hash={tx.to} truncate />
              </div>
            </div>

            <div className="wallet-tx-meta">
              <div className="wallet-tx-status">
                {txStatusIcon(tx.status)}
                <span className={`wallet-tx-status-text ${tx.status}`}>{t(`app.wallet.activity.status.${tx.status}`)}</span>
              </div>
              <span className="wallet-tx-time">{tx.timestamp}</span>
            </div>

            <div className="wallet-tx-hash">
              <span className="wallet-tx-hash-label">{t('app.wallet.activity.txHash')}</span>
              <CopyableHash hash={tx.txHash} />
            </div>
          </div>
        ))}
      </div>
    </section>
  )
}
