import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { CheckCircle2, AlertTriangle, XCircle, ExternalLink } from 'lucide-react'
import type { RpcProvider } from '../types'
import { SectionTitle } from './SectionTitle'

interface RpcProviderTableProps {
  providers: RpcProvider[]
}

export function RpcProviderTable({ providers }: RpcProviderTableProps) {
  const { t } = useTranslation()
  const [selectedProvider, setSelectedProvider] = useState<RpcProvider | null>(null)

  return (
    <section className="panel">
      <SectionTitle title={t('ops.rpc.title')} subtitle={t('ops.rpc.subtitle')} />
      <div className="rpc-table">
        <div className="rpc-table-head">
          <span>{t('ops.rpc.chain')}</span>
          <span>{t('ops.rpc.endpoint')}</span>
          <span>{t('ops.rpc.status')}</span>
          <span>{t('ops.rpc.latency')}</span>
          <span>{t('ops.rpc.errorRate')}</span>
          <span>{t('ops.rpc.latestBlock')}</span>
          <span>{t('ops.rpc.lastUpdate')}</span>
        </div>
        {providers.map((provider) => (
          <RpcProviderRow
            key={provider.id}
            provider={provider}
            selected={selectedProvider?.id === provider.id}
            onClick={() => setSelectedProvider(selectedProvider?.id === provider.id ? null : provider)}
          />
        ))}
      </div>
    </section>
  )
}

function RpcProviderRow({
  provider,
  selected,
  onClick,
}: {
  provider: RpcProvider
  selected: boolean
  onClick: () => void
}) {
  const { t } = useTranslation()

  return (
    <button
      className={`rpc-row ${selected ? 'selected' : ''} ${provider.status}`}
      onClick={onClick}
    >
      <span className="rpc-chain">{provider.chain}</span>
      <span className="rpc-endpoint">
        {provider.endpoint}
        <ExternalLink aria-hidden="true" className="rpc-link-icon" />
      </span>
      <span className={`rpc-status-badge ${provider.status}`}>
        {provider.status === 'online' && <CheckCircle2 aria-hidden="true" />}
        {provider.status === 'degraded' && <AlertTriangle aria-hidden="true" />}
        {provider.status === 'down' && <XCircle aria-hidden="true" />}
        {t(`ops.rpc.status.${provider.status}`)}
      </span>
      <span className={`rpc-latency ${provider.latency === 'timeout' ? 'text-danger' : ''}`}>
        {provider.latency}
      </span>
      <span className={`rpc-error-rate ${parseFloat(provider.errorRate) > 1 ? 'text-danger' : ''}`}>
        {provider.errorRate}
      </span>
      <span className="rpc-block">{provider.latestBlock}</span>
      <span className="rpc-update">{provider.lastUpdate}</span>
    </button>
  )
}
