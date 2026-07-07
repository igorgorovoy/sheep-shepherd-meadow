import { useState } from 'react'
import { getMeadowApiBase } from '../api/config'

export function PullCommand({ repo, tag }: { repo: string; tag: string }) {
  const [copied, setCopied] = useState(false)
  const host = getMeadowApiBase().replace(/^https?:\/\//, '')
  const cmd = `docker pull ${host}/${repo}:${tag}`

  async function copy() {
    try {
      await navigator.clipboard.writeText(cmd)
      setCopied(true)
      window.setTimeout(() => setCopied(false), 2000)
    } catch {
      /* ignore */
    }
  }

  return (
    <div className="pull-cmd">
      <code className="pull-cmd__text mono">{cmd}</code>
      <button type="button" className="btn" onClick={() => void copy()}>
        {copied ? 'Copied' : 'Copy'}
      </button>
    </div>
  )
}
