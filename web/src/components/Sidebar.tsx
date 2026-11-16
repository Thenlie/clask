import type { ProjectInfo } from '../types'

interface Props {
  projects: ProjectInfo[]
  selected: string | null
  onSelect: (selector: string | null) => void
  version: string
}

export function Sidebar({ projects, selected, onSelect, version }: Props) {
  const totalOpen = projects.reduce((n, p) => n + p.files.reduce((m, f) => m + f.open, 0), 0)
  return (
    <nav className="sidebar">
      <div className="brand">
        clask <span className="version">{version}</span>
      </div>
      <button className={`nav-item ${selected === null ? 'active' : ''}`} onClick={() => onSelect(null)}>
        <span>All tasks</span>
        <span className="count">{totalOpen}</span>
      </button>
      {projects.map((p) => {
        const open = p.files.reduce((m, f) => m + f.open, 0)
        const warn = p.files.some((f) => f.warnings.length > 0)
        // A project with one root file needs no nested scope rows.
        const nested = p.files.filter((f) => f.selector !== p.name)
        return (
          <div key={p.name} className="nav-group">
            <button
              className={`nav-item ${selected === p.name ? 'active' : ''}`}
              onClick={() => onSelect(p.name)}
              title={p.path}
            >
              <span>
                {p.name}
                {warn && (
                  <span className="warn-dot" title="file has warnings">
                    !
                  </span>
                )}
              </span>
              <span className="count">{open}</span>
            </button>
            {nested.map((f) => (
              <button
                key={f.selector || f.path}
                className={`nav-item nested ${selected === f.selector ? 'active' : ''}`}
                onClick={() => onSelect(f.selector)}
                title={f.file}
                disabled={!f.selector}
              >
                <span>{f.scope || f.path}</span>
                <span className="count">{f.open}</span>
              </button>
            ))}
          </div>
        )
      })}
      {projects.length === 0 && <p className="hint">No projects registered. Run `clask register` in a repo.</p>}
    </nav>
  )
}
