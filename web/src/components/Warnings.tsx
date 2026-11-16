import { useState } from 'react'
import type { Warning } from '../types'

export function Warnings({ warnings }: { warnings: Warning[] }) {
  const [open, setOpen] = useState(false)
  if (warnings.length === 0) return null
  return (
    <div className="banner warn">
      <button className="link" onClick={() => setOpen(!open)}>
        {warnings.length} warning{warnings.length === 1 ? '' : 's'} while loading task files {open ? '▾' : '▸'}
      </button>
      {open && (
        <ul>
          {warnings.map((w, i) => (
            <li key={i}>
              <code>
                {w.file}
                {w.line ? `:${w.line}` : ''}
              </code>{' '}
              {w.message}
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
