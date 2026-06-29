import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Link2 } from 'lucide-react'
import { EmptyState, LoadingState } from '../components'
import { WalletOverview } from '../components/WalletOverview'
import { WalletBalances } from '../components/WalletBalances'
import { WalletActivity } from '../components/WalletActivity'
import {
  getAddressBalances,
  getAddressTransactions,
  getChains,
  getDemoWalletAddress,
  getDeposits,
  getTokens,
  hasApiBase,
  type BalanceDTO,
  type ChainDTO,
  type DepositDTO,
  type TokenDTO,
  type TransactionDTO,
} from '../lib/api'
import { walletAccount, walletTransactions } from '../data/wallet'

export function Wallet() {
  const { t } = useTranslation()
  const configuredDemoAddress = getDemoWalletAddress()
  const [resolvedAddress, setResolvedAddress] = useState(configuredDemoAddress || walletAccount.address)
  const [isConnected, setIsConnected] = useState(Boolean(configuredDemoAddress) || walletAccount.isConnected)
  const [liveChains, setLiveChains] = useState<ChainDTO[]>([])
  const [liveTokens, setLiveTokens] = useState<TokenDTO[]>([])
  const [liveBalances, setLiveBalances] = useState<BalanceDTO[]>([])
  const [liveTransactions, setLiveTransactions] = useState<TransactionDTO[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    let cancelled = false

    async function load() {
      if (!hasApiBase()) {
        setLoading(false)
        return
      }

      try {
        setLoading(true)
        setError('')

        const [chainsResult, tokensResult, depositsResult] = await Promise.all([
          getChains(),
          getTokens(),
          getDeposits(20),
        ])

        if (cancelled) {
          return
        }

        const deposits = (depositsResult as DepositDTO[]) ?? []
        const detectedAddress = configuredDemoAddress || deposits[0]?.to_address || ''

        setLiveChains(chainsResult)
        setLiveTokens(tokensResult)
        setResolvedAddress(detectedAddress || walletAccount.address)

        if (!detectedAddress) {
          setLiveBalances([])
          setLiveTransactions([])
          return
        }

        const [balancesResult, transactionsResult] = await Promise.all([
          getAddressBalances(detectedAddress),
          getAddressTransactions(detectedAddress, 20),
        ])
        if (cancelled) {
          return
        }

        setLiveBalances(balancesResult)
        setLiveTransactions(transactionsResult)
      } catch (loadError) {
        if (cancelled) {
          return
        }
        setError(loadError instanceof Error ? loadError.message : 'Failed to load wallet data')
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }

    void load()
    return () => {
      cancelled = true
    }
  }, [configuredDemoAddress])

  const chainById = useMemo(
    () => new Map(liveChains.map((chain) => [chain.chain_id, chain])),
    [liveChains],
  )

  const tokenByKey = useMemo(
    () => new Map(liveTokens.map((token) => [`${token.chain_id}:${token.id}`, token])),
    [liveTokens],
  )

  const liveBalanceAssets = useMemo(
    () =>
      liveBalances.map((balance) => {
        const chain = chainById.get(balance.chain_id)
        const token = tokenByKey.get(`${balance.chain_id}:${balance.token_id}`)
        const symbol = token?.symbol || `Token #${balance.token_id}`
        const decimals = token?.decimals ?? 18

        return {
          chain: chain?.name || `Chain ${balance.chain_id}`,
          token: token?.is_native ? chain?.native_symbol || symbol : symbol,
          symbol,
          available: formatTokenAmount(balance.available_balance, decimals),
          locked: formatTokenAmount(balance.frozen_balance, decimals),
          value: `${formatTokenAmount(addIntegerStrings(balance.available_balance, balance.frozen_balance), decimals)} ${symbol}`,
          change: 'live',
        }
      }),
    [chainById, liveBalances, tokenByKey],
  )

  const mappedTransactions = useMemo(
    () =>
      liveTransactions.map((transaction, index) => {
        const token = tokenByKey.get(`${transaction.chain_id}:${transaction.token_id}`)
        const symbol = token?.symbol || `Token #${transaction.token_id}`
        const amount = formatTokenAmount(transaction.amount, token?.decimals ?? 18)

        return {
          id: `${transaction.tx_hash || transaction.created_at}-${index}`,
          type: transaction.type === 'withdrawal' ? 'send' : 'receive',
          status: mapWalletStatus(transaction.status),
          amount: transaction.type === 'withdrawal' ? `-${amount}` : `+${amount}`,
          token: symbol,
          timestamp: transaction.created_at,
          txHash: transaction.tx_hash || `pending-${index}`,
          from: transaction.from_address,
          to: transaction.to_address,
        } as const
      }),
    [liveTransactions, tokenByKey],
  )

  const networkLabel = useMemo(() => {
    const names = [...new Set(liveBalanceAssets.map((asset) => asset.chain))]
    if (names.length === 0) {
      return walletAccount.network
    }
    if (names.length === 1) {
      return names[0]
    }
    return `${names.length} chains connected`
  }, [liveBalanceAssets])

  const useFallbackData = !hasApiBase() || Boolean(error)
  const usingLiveData = !useFallbackData
  const displayAccount = {
    ...walletAccount,
    address: resolvedAddress,
    network: usingLiveData ? networkLabel : walletAccount.network,
    label: usingLiveData ? 'BNB Demo Wallet' : walletAccount.label,
    isConnected,
  }
  const displayBalances = useFallbackData ? balanceAssetsFallback() : liveBalanceAssets
  const displayTransactions = useFallbackData ? walletTransactions : mappedTransactions

  return (
    <section className="wallet-page">
      <div className="wallet-hero">
        <div className="wallet-hero-content">
          <span className="terminal-line">{t('app.wallet.terminal')}</span>
          <h2>{t('app.wallet.title')}</h2>
          <p>{t('app.wallet.subtitle')}</p>
        </div>
        <button
          className={`wallet-connect-btn ${isConnected ? 'disconnect' : 'connect'}`}
          onClick={() => setIsConnected(!isConnected)}
        >
          <Link2 aria-hidden="true" className="wallet-connect-icon" />
          {isConnected ? t('app.wallet.disconnect') : t('app.wallet.connect')}
        </button>
      </div>

      {loading ? <LoadingState message="Loading wallet state..." /> : null}
      {!loading && error ? <EmptyState title="Live wallet data unavailable" description={error} /> : null}
      {!loading && hasApiBase() && !usingLiveData ? (
        <EmptyState
          title="No live wallet activity detected"
          description="The page can auto-bind from recent deposits, but there are no public wallet records to infer yet."
        />
      ) : null}

      <div className="wallet-layout">
        <div className="wallet-main">
          <WalletOverview account={displayAccount} />
          <WalletBalances assets={displayBalances} />
        </div>
        <div className="wallet-side">
          <WalletActivity transactions={displayTransactions} />
        </div>
      </div>
    </section>
  )
}

function mapWalletStatus(status: string): 'pending' | 'confirmed' | 'failed' {
  if (status === 'confirmed' || status === 'broadcasted') {
    return 'confirmed'
  }
  if (status === 'failed' || status === 'canceled' || status === 'orphaned') {
    return 'failed'
  }
  return 'pending'
}

function addIntegerStrings(left: string, right: string) {
  try {
    return (BigInt(left || '0') + BigInt(right || '0')).toString()
  } catch {
    return left || right || '0'
  }
}

function formatTokenAmount(value: string, decimals: number) {
  try {
    const normalizedDecimals = Math.max(decimals, 0)
    const negative = value.startsWith('-')
    const digits = negative ? value.slice(1) : value
    const padded = digits.padStart(normalizedDecimals + 1, '0')
    const whole = padded.slice(0, padded.length - normalizedDecimals) || '0'
    const fraction = normalizedDecimals === 0 ? '' : padded.slice(-normalizedDecimals).replace(/0+$/, '')
    const composed = fraction ? `${whole}.${fraction}` : whole
    return `${negative ? '-' : ''}${Number(composed).toLocaleString('en-US', {
      maximumFractionDigits: 6,
    })}`
  } catch {
    return value
  }
}

function balanceAssetsFallback() {
  return [
    {
      chain: 'Ethereum',
      token: 'USD Coin',
      symbol: 'USDC',
      available: '42,180.42',
      locked: '1,200.00',
      value: '$43,380.42',
      change: '+2.8%',
    },
    {
      chain: 'BNB Chain',
      token: 'BNB',
      symbol: 'BNB',
      available: '128.45',
      locked: '8.00',
      value: '$91,248.20',
      change: '+4.1%',
    },
    {
      chain: 'Polygon',
      token: 'Tether USD',
      symbol: 'USDT',
      available: '18,902.10',
      locked: '0.00',
      value: '$18,902.10',
      change: '-0.2%',
    },
    {
      chain: 'Solana',
      token: 'Solana',
      symbol: 'SOL',
      available: '384.22',
      locked: '12.00',
      value: '$62,498.60',
      change: '+6.4%',
    },
  ]
}
