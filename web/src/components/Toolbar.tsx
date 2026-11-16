import { PRIORITIES, SORT_KEYS, STATUSES, type SortKey, type Status } from '../types'

export interface Filters {
  statuses: Status[]
  tags: string[]
  priority: number
  overdue: boolean
  all: boolean
  sort: SortKey
  search: string
}

interface Props {
  filters: Filters
  onChange: (f: Filters) => void
  tags: string[]
  count: number
  onRefresh: () => void
  onAdd: () => void
}

export function Toolbar({ filters, onChange, tags, count, onRefresh, onAdd }: Props) {
  const set = (patch: Partial<Filters>) => onChange({ ...filters, ...patch })
  const toggleStatus = (s: Status) =>
    set({
      statuses: filters.statuses.includes(s)
        ? filters.statuses.filter((x) => x !== s)
        : [...filters.statuses, s],
    })
  const toggleTag = (t: string) =>
    set({ tags: filters.tags.includes(t) ? filters.tags.filter((x) => x !== t) : [...filters.tags, t] })

  return (
    <div className="toolbar">
      <div className="row">
        <input
          className="search"
          type="search"
          placeholder="Search id, title, tags"
          value={filters.search}
          onChange={(e) => set({ search: e.target.value })}
        />
        <select value={filters.priority} onChange={(e) => set({ priority: Number(e.target.value) })}>
          <option value={0}>any priority</option>
          {PRIORITIES.map((p) => (
            <option key={p} value={p}>
              P{p}
            </option>
          ))}
        </select>
        <select value={filters.sort} onChange={(e) => set({ sort: e.target.value as SortKey })}>
          {SORT_KEYS.map((k) => (
            <option key={k} value={k}>
              sort: {k}
            </option>
          ))}
        </select>
        <label className="check">
          <input type="checkbox" checked={filters.overdue} onChange={(e) => set({ overdue: e.target.checked })} />
          overdue
        </label>
        <label className="check">
          <input type="checkbox" checked={filters.all} onChange={(e) => set({ all: e.target.checked })} />
          show closed
        </label>
        <span className="spacer" />
        <span className="muted">{count} tasks</span>
        <button onClick={onRefresh} title="Reload from disk">
          Refresh
        </button>
        <button className="primary" onClick={onAdd}>
          + Add
        </button>
      </div>
      <div className="row chips">
        {STATUSES.map((s) => (
          <button
            key={s}
            className={`chip status-${s} ${filters.statuses.includes(s) ? 'on' : ''}`}
            onClick={() => toggleStatus(s)}
          >
            {s}
          </button>
        ))}
        {tags.length > 0 && <span className="sep" />}
        {tags.map((t) => (
          <button key={t} className={`chip tag ${filters.tags.includes(t) ? 'on' : ''}`} onClick={() => toggleTag(t)}>
            #{t}
          </button>
        ))}
      </div>
    </div>
  )
}
