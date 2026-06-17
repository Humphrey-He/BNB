import { useState } from 'react'
import { Copy, Check } from 'lucide-react'

interface CopyableHashProps {
  hash: string
  truncate?: boolean
  className?: string
}

export function CopyableHash({ hash, truncate = true, className = '' }: CopyableHashProps) {
  const [copied, setCopied] = useState(false)

  const displayHash = truncate && hash.length > 16
    ? `${hash.slice(0, 8)}...${hash.slice(-6)}`
    : hash

  const handleCopy = async () => {
    await navigator.clipboard.writeText(hash)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <button
      className={`copyable-hash ${className}`}
      onClick={handleCopy}
      title={hash}
    >
      <span className="hash-text">{displayHash}</span>
      {copied
        ? <Check aria-hidden="true" className="copy-icon success" />
        : <Copy aria-hidden="true" className="copy-icon" />
      }
    </button>
  )
}
