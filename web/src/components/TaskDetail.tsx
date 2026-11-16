import { useState, type KeyboardEvent } from 'react'
import Markdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { api, ApiError } from '../api'
import { isOverdue, PRIORITIES, refOf, selectorOf, STATUSES, type Entry, type Patch } from '../types'

interface Props {
  entry: Entry
  today: string
  busy: boolean
  onPatch: (p: Patch) => void
  onSelectRef: (ref: string) => void
  onClose: () => void
  onError: (e: unknown) => void
}

export function TaskDetail({ entry, today, busy, onPatch, onSelectRef, onClose, onError }: Props) {
  const [editingTitle, setEditingTitle] = useState(false)
  const [title, setTitle] = useState(entry.title)
  const [newTag, setNewTag] = useState('')
  const [editCommand, setEditCommand] = useState<string | null>(null)
  const ref = refOf(entry)

  const saveTitle = () => {
    setEditingTitle(false)
    if (title.trim() && title.trim() !== entry.title) onPatch({ title: title.trim() })
    else setTitle(entry.title)
  }
  const titleKey = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') saveTitle()
    if (e.key === 'Escape') {
      setTitle(entry.title)
      setEditingTitle(false)
    }
  }

  const addTag = () => {
    const t = newTag.trim()
    if (!t) return
    setNewTag('')
    if (!(entry.tags ?? []).includes(t)) onPatch({ tags: [...(entry.tags ?? []), t] })
  }
  const removeTag = (t: string) => onPatch({ tags: (entry.tags ?? []).filter((x) => x !== t) })

  // blocked_by entries are bare ids (same file) or full refs.
  const blockerRef = (b: string) => (b.includes('/') ? b : `${selectorOf(entry)}/${b}`)

  const openInEditor = async () => {
    setEditCommand(null)
    try {
      await api.open(ref)
    } catch (e) {
      if (e instanceof ApiError && e.command) setEditCommand(e.command)
      else onError(e)
    }
  }

  return (
    <aside className="detail">
      <div className="detail-head">
        <span className="project">{selectorOf(entry)}</span>
        <code className="id">{entry.id}</code>
        <span className="spacer" />
        <button className="icon" onClick={onClose} title="Close">
          ×
        </button>
      </div>

      {editingTitle ? (
        <input
          className="title-input"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          onBlur={saveTitle}
          onKeyDown={titleKey}
          autoFocus
        />
      ) : (
        <h2 className="title" onClick={() => setEditingTitle(true)} title="Click to edit">
          {entry.title}
        </h2>
      )}

      <div className="status-row">
        {STATUSES.map((s) => (
          <button
            key={s}
            className={`chip status-${s} ${entry.status === s ? 'on' : ''}`}
            disabled={busy || entry.status === s}
            onClick={() => onPatch({ status: s })}
          >
            {s}
          </button>
        ))}
      </div>

      <dl className="fields">
        <dt>priority</dt>
        <dd>
          <select value={entry.priority} disabled={busy} onChange={(e) => onPatch({ priority: Number(e.target.value) })}>
            {PRIORITIES.map((p) => (
              <option key={p} value={p}>
                P{p}
              </option>
            ))}
          </select>
        </dd>
        <dt>due</dt>
        <dd className={isOverdue(entry, today) ? 'overdue' : ''}>
          <input
            type="date"
            value={entry.due ?? ''}
            disabled={busy}
            onChange={(e) => onPatch({ due: e.target.value || null })}
          />
        </dd>
        <dt>effort</dt>
        <dd>
          <input
            key={entry.effort ?? ''}
            defaultValue={entry.effort ?? ''}
            placeholder="—"
            disabled={busy}
            onBlur={(e) => {
              const v = e.target.value.trim()
              if (v !== (entry.effort ?? '')) onPatch({ effort: v || null })
            }}
          />
        </dd>
        <dt>tags</dt>
        <dd className="taglist">
          {(entry.tags ?? []).map((t) => (
            <span key={t} className="tag">
              #{t}
              <button className="x" disabled={busy} onClick={() => removeTag(t)} title="Remove tag">
                ×
              </button>
            </span>
          ))}
          <input
            className="tag-input"
            value={newTag}
            placeholder="+ tag"
            disabled={busy}
            onChange={(e) => setNewTag(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && addTag()}
            onBlur={addTag}
          />
        </dd>
        {entry.blocked_by && entry.blocked_by.length > 0 && (
          <>
            <dt>blocked by</dt>
            <dd>
              {entry.blocked_by.map((b) => (
                <button key={b} className="link" onClick={() => onSelectRef(blockerRef(b))}>
                  {b}
                </button>
              ))}
            </dd>
          </>
        )}
        {entry.links && entry.links.length > 0 && (
          <>
            <dt>links</dt>
            <dd>
              {entry.links.map((l) => (
                <a key={l} href={l} target="_blank" rel="noreferrer">
                  {l}
                </a>
              ))}
            </dd>
          </>
        )}
        <dt>created</dt>
        <dd>{entry.created ?? '—'}</dd>
        {entry.completed && (
          <>
            <dt>completed</dt>
            <dd>{entry.completed}</dd>
          </>
        )}
      </dl>

      <div className="body markdown">
        {entry.body ? (
          <Markdown remarkPlugins={[remarkGfm]}>{entry.body}</Markdown>
        ) : (
          <p className="muted">No notes. Use “Open in editor” to add some.</p>
        )}
      </div>

      <div className="detail-foot">
        <code className="path" title={entry.file}>
          {entry.file.replace(/^.*\/([^/]+\/[^/]+)$/, '$1')}:{entry.line}
        </code>
        <span className="spacer" />
        <button onClick={() => void openInEditor()}>Open in editor</button>
      </div>
      {editCommand && (
        <div className="banner info">
          Your editor runs in a terminal. Run:
          <code>{editCommand}</code>
          <button className="link" onClick={() => void navigator.clipboard.writeText(editCommand)}>
            copy
          </button>
        </div>
      )}
    </aside>
  )
}
