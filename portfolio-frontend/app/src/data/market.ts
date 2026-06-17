import type { PriceTicker, OrderBookEntry, Trade, TradingPair } from '../types'

export const tradingPairs: TradingPair[] = [
  { id: 'btc-usdt', base: 'BTC', quote: 'USDT', symbol: 'BTC/USDT' },
  { id: 'eth-usdt', base: 'ETH', quote: 'USDT', symbol: 'ETH/USDT' },
  { id: 'bnb-usdt', base: 'BNB', quote: 'USDT', symbol: 'BNB/USDT' },
  { id: 'sol-usdt', base: 'SOL', quote: 'USDT', symbol: 'SOL/USDT' },
  { id: 'matic-usdt', base: 'MATIC', quote: 'USDT', symbol: 'MATIC/USDT' },
  { id: 'eth-btc', base: 'ETH', quote: 'BTC', symbol: 'ETH/BTC' },
]

export const priceTickers: PriceTicker[] = [
  {
    symbol: 'BTC',
    name: 'Bitcoin',
    price: '67,842.50',
    change24h: '+2.34%',
    volume24h: '28.4B',
    high24h: '68,215.00',
    low24h: '65,890.00',
  },
  {
    symbol: 'ETH',
    name: 'Ethereum',
    price: '3,542.80',
    change24h: '+1.87%',
    volume24h: '14.2B',
    high24h: '3,598.00',
    low24h: '3,480.00',
  },
  {
    symbol: 'BNB',
    name: 'BNB',
    price: '612.40',
    change24h: '-0.54%',
    volume24h: '1.8B',
    high24h: '625.00',
    low24h: '608.50',
  },
  {
    symbol: 'SOL',
    name: 'Solana',
    price: '178.25',
    change24h: '+5.62%',
    volume24h: '3.2B',
    high24h: '182.50',
    low24h: '168.30',
  },
  {
    symbol: 'MATIC',
    name: 'Polygon',
    price: '0.7234',
    change24h: '-1.23%',
    volume24h: '892M',
    high24h: '0.7450',
    low24h: '0.7150',
  },
  {
    symbol: 'USDT',
    name: 'Tether',
    price: '1.0002',
    change24h: '+0.01%',
    volume24h: '52.1B',
    high24h: '1.0005',
    low24h: '0.9998',
  },
  {
    symbol: 'USDC',
    name: 'USD Coin',
    price: '0.9998',
    change24h: '-0.02%',
    volume24h: '6.4B',
    high24h: '1.0002',
    low24h: '0.9995',
  },
  {
    symbol: 'DOGE',
    name: 'Dogecoin',
    price: '0.1634',
    change24h: '+8.42%',
    volume24h: '1.1B',
    high24h: '0.1680',
    low24h: '0.1510',
  },
]

export const orderBookData: OrderBookEntry[] = [
  { price: '67,841.00', amount: '0.8234', total: '55,872.34' },
  { price: '67,840.50', amount: '1.2340', total: '83,716.97' },
  { price: '67,840.00', amount: '2.1560', total: '146,278.24' },
  { price: '67,839.50', amount: '0.5120', total: '34,735.44' },
  { price: '67,839.00', amount: '3.4210', total: '232,001.88' },
  { price: '67,838.50', amount: '1.8900', total: '128,270.06' },
  { price: '67,838.00', amount: '0.6540', total: '44,366.23' },
  { price: '67,837.50', amount: '2.3050', total: '156,348.39' },
]

export const askBookData: OrderBookEntry[] = [
  { price: '67,843.00', amount: '1.4520', total: '98,568.64' },
  { price: '67,843.50', amount: '0.7890', total: '53,512.10' },
  { price: '67,844.00', amount: '2.1100', total: '143,343.84' },
  { price: '67,844.50', amount: '0.9340', total: '63,394.37' },
  { price: '67,845.00', amount: '1.6780', total: '113,916.11' },
  { price: '67,845.50', amount: '0.4320', total: '29,341.26' },
  { price: '67,846.00', amount: '1.0230', total: '69,436.36' },
  { price: '67,846.50', amount: '2.8900', total: '196,156.44' },
]

export const recentTrades: Trade[] = [
  { id: 't-001', type: 'buy', price: '67,841.00', amount: '0.5234', timestamp: '14:32:18' },
  { id: 't-002', type: 'sell', price: '67,842.50', amount: '1.2340', timestamp: '14:32:15' },
  { id: 't-003', type: 'buy', price: '67,840.50', amount: '0.8120', timestamp: '14:32:12' },
  { id: 't-004', type: 'sell', price: '67,843.00', amount: '2.1050', timestamp: '14:32:08' },
  { id: 't-005', type: 'buy', price: '67,839.00', amount: '0.2340', timestamp: '14:32:05' },
  { id: 't-006', type: 'buy', price: '67,838.50', amount: '1.8900', timestamp: '14:32:02' },
  { id: 't-007', type: 'sell', price: '67,844.00', amount: '0.6540', timestamp: '14:31:58' },
  { id: 't-008', type: 'buy', price: '67,837.50', amount: '0.4230', timestamp: '14:31:55' },
  { id: 't-009', type: 'sell', price: '67,845.00', amount: '1.6780', timestamp: '14:31:52' },
  { id: 't-010', type: 'buy', price: '67,836.00', amount: '0.9320', timestamp: '14:31:48' },
  { id: 't-011', type: 'sell', price: '67,846.00', amount: '1.0230', timestamp: '14:31:45' },
  { id: 't-012', type: 'buy', price: '67,835.50', amount: '2.3050', timestamp: '14:31:42' },
  { id: 't-013', type: 'sell', price: '67,847.00', amount: '0.5210', timestamp: '14:31:38' },
  { id: 't-014', type: 'buy', price: '67,834.00', amount: '0.8120', timestamp: '14:31:35' },
  { id: 't-015', type: 'sell', price: '67,848.00', amount: '1.4560', timestamp: '14:31:32' },
]

export const marketChartData = [
  { time: '00:00', price: 66500 },
  { time: '04:00', price: 66780 },
  { time: '08:00', price: 66420 },
  { time: '12:00', price: 67100 },
  { time: '16:00', price: 67500 },
  { time: '20:00', price: 67380 },
  { time: '24:00', price: 67842 },
]
