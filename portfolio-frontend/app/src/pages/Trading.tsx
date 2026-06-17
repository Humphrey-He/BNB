import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { SectionTitle } from '../components'
import { OrderBook } from '../components/OrderBook'
import { RecentTrades } from '../components/RecentTrades'
import { TradingPairSelector } from '../components/TradingPairSelector'
import {
  tradingPairs,
  orderBookData,
  askBookData,
  recentTrades,
} from '../data/market'
import type { TradingPair } from '../types'

export function Trading() {
  const { t } = useTranslation()
  const [selectedPair, setSelectedPair] = useState<TradingPair>(tradingPairs[0])

  return (
    <div className="trading-page">
      <div className="trading-hero">
        <div className="terminal-line">trading: orderbook_live trades_streaming</div>
        <h2>{t('trading.title')}</h2>
        <p>{t('trading.subtitle')}</p>
      </div>

      <div className="trading-layout">
        <div className="trading-main">
          <div className="trading-pair-section">
            <TradingPairSelector
              pairs={tradingPairs}
              selectedPair={selectedPair}
              onSelectPair={setSelectedPair}
            />
            <div className="current-price">
              <span className="price-label">{t('trading.lastPrice')}</span>
              <span className="price-value">67,841.50</span>
              <span className="price-change positive">+0.12%</span>
            </div>
          </div>

          <div className="trading-panels">
            <OrderBook
              bids={orderBookData}
              asks={askBookData}
              symbol={selectedPair?.symbol}
            />
            <RecentTrades
              trades={recentTrades}
              symbol={selectedPair?.symbol}
            />
          </div>
        </div>

        <div className="trading-sidebar">
          <div className="trading-form-section">
            <SectionTitle
              title={t('trading.placeOrder')}
              subtitle={t('trading.placeOrderSubtitle')}
            />

            <div className="order-type-tabs">
              <button className="order-tab active" type="button">
                {t('trading.limit')}
              </button>
              <button className="order-tab" type="button">
                {t('trading.market')}
              </button>
              <button className="order-tab" type="button">
                {t('trading.stop')}
              </button>
            </div>

            <div className="order-form">
              <div className="form-group">
                <label>{t('trading.side')}</label>
                <div className="side-toggle">
                  <button className="side-btn buy active" type="button">
                    {t('trading.buy')}
                  </button>
                  <button className="side-btn sell" type="button">
                    {t('trading.sell')}
                  </button>
                </div>
              </div>

              <div className="form-group">
                <label>{t('trading.price')}</label>
                <input
                  type="text"
                  className="form-input"
                  placeholder="0.00"
                  defaultValue="67,841.50"
                />
              </div>

              <div className="form-group">
                <label>{t('trading.amount')}</label>
                <input
                  type="text"
                  className="form-input"
                  placeholder="0.00"
                />
              </div>

              <div className="form-group">
                <label>{t('trading.total')}</label>
                <input
                  type="text"
                  className="form-input"
                  placeholder="0.00"
                />
              </div>

              <div className="form-slider">
                <div className="slider-labels">
                  <span>0%</span>
                  <span>25%</span>
                  <span>50%</span>
                  <span>75%</span>
                  <span>100%</span>
                </div>
                <input
                  type="range"
                  min="0"
                  max="100"
                  defaultValue="0"
                  className="slider"
                />
              </div>

              <button className="submit-order-btn buy" type="button">
                {t('trading.buy')} {selectedPair?.base || 'BTC'}
              </button>
              <button className="submit-order-btn sell" type="button">
                {t('trading.sell')} {selectedPair?.base || 'BTC'}
              </button>
            </div>
          </div>

          <div className="positions-section">
            <SectionTitle
              title={t('trading.positions')}
              subtitle={t('trading.positionsSubtitle')}
            />
            <div className="empty-positions">
              <p>{t('trading.noPositions')}</p>
            </div>
          </div>

          <div className="orders-section">
            <SectionTitle
              title={t('trading.openOrders')}
              subtitle={t('trading.openOrdersSubtitle')}
            />
            <div className="empty-orders">
              <p>{t('trading.noOrders')}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
