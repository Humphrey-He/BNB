import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts'
import { SectionTitle } from '../components'
import { MarketOverview } from '../components/MarketOverview'
import { priceTickers, marketChartData } from '../data/market'

export function Market() {
  const { t } = useTranslation()
  const [selectedSymbol, setSelectedSymbol] = useState<string | undefined>('BTC')

  return (
    <div className="market-page">
      <div className="market-hero">
        <div className="market-hero-content">
          <div className="terminal-line">market: live_tickers sentiment_neutral</div>
          <h2>{t('market.title')}</h2>
          <p>{t('market.subtitle')}</p>
        </div>
        <div className="market-sentiment">
          <div className="sentiment-indicator">
            <span className="sentiment-label">{t('market.sentiment')}</span>
            <div className="sentiment-bar">
              <div className="sentiment-fill fear" style={{ width: '32%' }} />
              <div className="sentiment-fill neutral" style={{ width: '45%' }} />
              <div className="sentiment-fill greed" style={{ width: '23%' }} />
            </div>
            <div className="sentiment-labels">
              <span>{t('market.fear')}</span>
              <span>{t('market.neutral')}</span>
              <span>{t('market.greed')}</span>
            </div>
          </div>
        </div>
      </div>

      <div className="market-chart-section">
        <div className="market-chart-container">
          <div className="chart-header">
            <h3>{selectedSymbol}/USDT {t('market.priceChart')}</h3>
            <span className="chart-timeframe">24h</span>
          </div>
          <div className="price-chart">
            <ResponsiveContainer width="100%" height={280}>
              <AreaChart data={marketChartData} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
                <defs>
                  <linearGradient id="priceGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#f0b90b" stopOpacity={0.3} />
                    <stop offset="95%" stopColor="#f0b90b" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <XAxis
                  dataKey="time"
                  axisLine={false}
                  tickLine={false}
                  tick={{ fill: '#6f7a8c', fontSize: 11 }}
                />
                <YAxis
                  domain={['auto', 'auto']}
                  axisLine={false}
                  tickLine={false}
                  tick={{ fill: '#6f7a8c', fontSize: 11 }}
                  width={70}
                  tickFormatter={(value) => `$${value.toLocaleString()}`}
                />
                <Tooltip
                  contentStyle={{
                    background: '#111722',
                    border: '1px solid #202838',
                    borderRadius: '6px',
                    color: '#f4f7fb',
                  }}
                  labelStyle={{ color: '#6f7a8c' }}
                />
                <Area
                  type="monotone"
                  dataKey="price"
                  stroke="#f0b90b"
                  strokeWidth={2}
                  fill="url(#priceGradient)"
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>

      <div className="market-overview-section">
        <SectionTitle
          title={t('market.overview')}
          subtitle={t('market.overviewSubtitle')}
        />
        <MarketOverview
          tickers={priceTickers}
          selectedSymbol={selectedSymbol}
          onSelectSymbol={setSelectedSymbol}
        />
      </div>

      <div className="market-tickers-detail">
        <SectionTitle
          title={t('market.marketStats')}
          subtitle={t('market.marketStatsSubtitle')}
        />
        {selectedSymbol && (
          <div className="selected-ticker-stats">
            {priceTickers
              .filter((t) => t.symbol === selectedSymbol)
              .map((ticker) => (
                <div key={ticker.symbol} className="ticker-detail-card">
                  <div className="ticker-detail-header">
                    <div>
                      <h4>{ticker.name}</h4>
                      <span className="ticker-detail-symbol">{ticker.symbol}</span>
                    </div>
                    <div className={`ticker-detail-change ${ticker.change24h.startsWith('-') ? 'negative' : 'positive'}`}>
                      {ticker.change24h}
                    </div>
                  </div>
                  <div className="ticker-detail-price">${ticker.price}</div>
                  <div className="ticker-detail-grid">
                    <div className="ticker-stat">
                      <span>{t('market.high24h')}</span>
                      <strong>{ticker.high24h}</strong>
                    </div>
                    <div className="ticker-stat">
                      <span>{t('market.low24h')}</span>
                      <strong>{ticker.low24h}</strong>
                    </div>
                    <div className="ticker-stat">
                      <span>{t('market.volume24h')}</span>
                      <strong>{ticker.volume24h}</strong>
                    </div>
                  </div>
                </div>
              ))}
          </div>
        )}
      </div>
    </div>
  )
}
