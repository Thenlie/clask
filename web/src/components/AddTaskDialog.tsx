import { useEffect, useRef, useState, type FormEvent } from 'react'
import { PRIORITIES, type CreateRequest, type ProjectInfo } from '../types'

interface Props {
  projects: ProjectInfo[]
  defaultSelector: string
  busy: boolean
  onSubmit: (req: CreateRequest) => Promise<boolean>
  onClose: () => void
}

function splitList(s: string): string[] | undefined {
  const items = s
    .split(/[,\s]+/)
    .map((x) => x.trim())
    .filter(Boolean)
  return items.length ? items : undefined
}

export function AddTaskDialog({ projects, defaultSelector, busy, onSubmit, onClose }: Props) {
  const ref = useRef<HTMLDialogElement>(null)
  const [selector, setSelector] = useState(defaultSelector)
  const [title, setTitle] = useState('')
  const [priority, setPriority] = useState(0)
  const [tags, setTags] = useState('')
  const [due, setDue] = useState('')
  const [effort, setEffort] = useState('')
  const [body, setBody] = useState('')

  useEffect(() => {
    const d = ref.current
    if (d && !d.open) d.showModal()
  }, [])

  const submit = async (e: FormEvent) => {
    e.preventDefault()
    await onSubmit({
      selector,
      title: title.trim(),
      priority: priority || undefined,
      tags: splitList(tags),
      due: due || undefined,
      effort: effort.trim() || undefined,
      body: body.trim() || undefined,
    })
  }

  const files = projects.flatMap((p) => p.files.filter((f) => f.selector))
  return (
    <dialog ref={ref} className="dialog" onClose={onClose} onCancel={onClose}>
      <form onSubmit={submit}>
        <h2>New task</h2>
        <label>
          File
          <select value={selector} onChange={(e) => setSelector(e.target.value)} required>
            <option value="" disabled>
              choose a task file
            </option>
            {files.map((f) => (
              <option key={f.selector} value={f.selector}>
                {f.selector} — {f.path}
              </option>
            ))}
          </select>
        </label>
        <label>
          Title
          <input value={title} onChange={(e) => setTitle(e.target.value)} required autoFocus />
        </label>
        <div className="grid3">
          <label>
            Priority
            <select value={priority} onChange={(e) => setPriority(Number(e.target.value))}>
              <option value={0}>file default</option>
              {PRIORITIES.map((p) => (
                <option key={p} value={p}>
                  P{p}
                </option>
              ))}
            </select>
          </label>
          <label>
            Due
            <input type="date" value={due} onChange={(e) => setDue(e.target.value)} />
          </label>
          <label>
            Effort
            <input value={effort} onChange={(e) => setEffort(e.target.value)} placeholder="s, m, l, 2h" />
          </label>
        </div>
        <label>
          Tags
          <input value={tags} onChange={(e) => setTags(e.target.value)} placeholder="comma or space separated" />
        </label>
        <label>
          Body (markdown)
          <textarea value={body} onChange={(e) => setBody(e.target.value)} rows={5} />
        </label>
        <div className="actions">
          <button type="button" onClick={onClose}>
            Cancel
          </button>
          <button type="submit" className="primary" disabled={busy || !selector || !title.trim()}>
            Add task
          </button>
        </div>
      </form>
    </dialog>
  )
}
