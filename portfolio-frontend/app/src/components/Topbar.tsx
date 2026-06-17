import { useTranslation } from 'react-i18next'
import { languages } from '../i18n'
import { AppIcon } from '../components'

interface TopbarProps {
  wallet: boolean
  onWalletToggle: () => void
}

export function Topbar({ wallet, onWalletToggle }: TopbarProps) {
  const { i18n, t } = useTranslation()

  return (
    <header className="topbar">
      <div>
        <h1>{t('app.title')}</h1>
        <p>{t('app.subtitle')}</p>
      </div>
      <div className="topbar-actions">
        <div className="language-switch" aria-label="Language selector">
          {languages.map((language) => (
            <button
              aria-label={language.label}
              className={i18n.language === language.code ? 'selected' : ''}
              key={language.code}
              onClick={() => void i18n.changeLanguage(language.code)}
              title={language.label}
            >
              {language.short}
            </button>
          ))}
        </div>
        <button className={wallet ? 'wallet connected' : 'wallet'} onClick={onWalletToggle}>
          <AppIcon name={wallet ? 'check' : 'wallet'} />
          {wallet ? t('app.wallet.connected') : t('app.wallet.readonly')}
        </button>
      </div>
    </header>
  )
}
