const runtimeOrigin = typeof window !== 'undefined' ? window.location.origin : ''
const apiBase = import.meta.env.VITE_API_BASE_URL?.replace(/\/$/, '') ?? runtimeOrigin
const apiToken = import.meta.env.VITE_API_AUTH_TOKEN?.trim() ?? ''
const demoAddress = import.meta.env.VITE_DEMO_WALLET_ADDRESS?.trim() ?? ''

function buildHeaders(authenticated = false): HeadersInit {
  if (authenticated && apiToken) {
    return {
      Authorization: `Bearer ${apiToken}`,
    }
  }
  return {}
}

async function request<T>(path: string, authenticated = false): Promise<T> {
  const response = await fetch(`${apiBase}${path}`, {
    headers: buildHeaders(authenticated),
  })

  if (!response.ok) {
    const text = await response.text()
    throw new Error(text || `Request failed: ${response.status}`)
  }

  return response.json() as Promise<T>
}

function buildQuery(params: Record<string, string | number | undefined>) {
  const search = new URLSearchParams()

  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== '') {
      search.set(key, String(value))
    }
  })

  const query = search.toString()
  return query ? `?${query}` : ''
}

export type ChainDTO = {
  id: number
  chain_id: number
  name: string
  native_symbol: string
  finality_confirmations: number
  is_active: boolean
}

export type TokenDTO = {
  id: number
  chain_id: number
  contract_address: string
  symbol: string
  decimals: number
  is_native: boolean
  is_active: boolean
}

export type ChainStatusDTO = {
  chain_id: number
  name: string
  last_scanned_block: number
  latest_block: number
  scan_lag: number
  is_active: boolean
  rpc_healthy: boolean
  rpc_provider?: string
  rpc_error?: string
  provider_count: number
}

export type BalanceDTO = {
  account_address: string
  chain_id: number
  token_id: number
  available_balance: string
  frozen_balance: string
}

export type DepositDTO = {
  id: number
  chain_id: number
  token_id: number
  tx_hash: string
  from_address: string
  to_address: string
  amount: string
  block_number: number
  status: string
  confirmations: number
  created_at: string
}

export type TransactionDTO = {
  type: string
  chain_id: number
  token_id: number
  tx_hash: string
  from_address: string
  to_address: string
  amount: string
  status: string
  block_number?: number
  created_at: string
}

export async function getChains() {
  return request<ChainDTO[]>('/api/v1/chains')
}

export async function getTokens(chainId?: number) {
  return request<TokenDTO[]>(`/api/v1/tokens${buildQuery({ chain_id: chainId })}`)
}

export async function getChainStatus() {
  return request<ChainStatusDTO[]>('/api/v1/chain-status', true)
}

export async function getAddressBalances(address: string) {
  return request<BalanceDTO[]>(`/api/v1/addresses/${address}/balances`)
}

export async function getDeposits(
  limit = 20,
  filters?: { address?: string; chainId?: number; status?: string },
) {
  return request<DepositDTO[]>(
    `/api/v1/deposits${buildQuery({
      limit,
      address: filters?.address,
      chain_id: filters?.chainId,
      status: filters?.status,
    })}`,
  )
}

export async function getAddressTransactions(address: string, limit = 50) {
  return request<TransactionDTO[]>(
    `/api/v1/addresses/${address}/transactions${buildQuery({ limit })}`,
  )
}

export function getDemoWalletAddress() {
  return demoAddress
}

export function hasApiBase() {
  return Boolean(apiBase)
}
