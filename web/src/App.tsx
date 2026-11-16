import { useCallback, useEffect, useMemo, useState } from 'react'
import { api, ApiError } from './api'
import { AddTaskDialog } from './components/AddTaskDialog'
import { Sidebar } from './components/Sidebar'
import { TaskDetail } from './components/TaskDetail'
import { TaskTable } from './components/TaskTable'
import { Toolbar, type Filters } from './components/Toolbar'
import { Warnings } from './components/Warnings'
import {
  refOf,
  today as localToday,
  type CreateRequest,
  type Entry,
  type Patch,
  type ProjectInfo,
  type Warning,
} from './types'

const defaultFilters: Filters = {
  statuses: [],
  tags: [],
  priority: 0,
  overdue: false,
  all: false,
  sort: 'priority',
  search: '',
}

// Filters that round-trip through the URL so reload and deep links work.
function readUrl(): { project: string | null; task: string | null } {
  const p = new URLSearchParams(window.location.search)
  return { project: p.get('project'), task: p.get('task') }
}

function writeUrl(project: string | null, task: string | null) {
  const p = new URLSearchParams()
  if (project) p.set('project', project)
  if (task) p.set('task', task)
  const qs = p.toString()
  window.history.replaceState(null, '', qs ? `?${qs}` : window.location.pathname)
}

export default function App() {
  const initial = useMemo(readUrl, [])
  const [project, setProject] = useState<string | null>(initial.project)
  const [selectedRef, setSelectedRef] = useState<string | null>(initial.task)
  const [filters, setFilters] = useState<Filters>(defaultFilters)
  const [projects, setProjects] = useState<ProjectInfo[]>([])
  const [tasks, setTasks] = useState<Entry[]>([])
  const [warnings, setWarnings] = useState<Warning[]>([])
  const [selected, setSelected] = useState<Entry | null>(null)
  const [version, setVersion] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)
  const [adding, setAdding] = useState(false)
  const today = localToday()

  useEffect(() => writeUrl(project, selectedRef), [project, selectedRef])

  const fail = (e: unknown) => setError(e instanceof Error ? e.message : String(e))

  const refresh = useCallback(async () => {
    setError(null)
    try {
      const [p, t] = await Promise.all([
        api.projects(),
        api.tasks({
          project: project ?? undefined,
          statuses: filters.statuses,
          tags: filters.tags,
          priority: filters.priority || undefined,
          overdue: filters.overdue,
          all: filters.all,
          sort: filters.sort,
        }),
      ])
      setProjects(p)
      setTasks(t.tasks)
      setWarnings(t.warnings)
    } catch (e) {
      fail(e)
    }
  }, [project, filters])

  useEffect(() => {
    void refresh()
  }, [refresh])

  useEffect(() => {
    api.meta().then((m) => setVersion(m.version)).catch(() => {})
  }, [])

  // Load the selected task from the list when possible, else fetch it
  // (closed tasks are hidden from the default list but still linkable).
  useEffect(() => {
    if (!selectedRef) {
      setSelected(null)
      return
    }
    const inList = tasks.find((t) => refOf(t) === selectedRef)
    if (inList) {
      setSelected(inList)
      return
    }
    let cancelled = false
    api
      .task(selectedRef)
      .then((e) => !cancelled && setSelected(e))
      .catch((e: unknown) => {
        if (cancelled) return
        if (e instanceof ApiError && e.status === 404) setSelectedRef(null)
        else fail(e)
      })
    return () => {
      cancelled = true
    }
  }, [selectedRef, tasks])

  const visible = useMemo(() => {
    const q = filters.search.trim().toLowerCase()
    if (!q) return tasks
    return tasks.filter(
      (t) =>
        t.id.includes(q) ||
        t.title.toLowerCase().includes(q) ||
        (t.tags ?? []).some((tag) => tag.toLowerCase().includes(q)),
    )
  }, [tasks, filters.search])

  const allTags = useMemo(() => {
    const s = new Set<string>()
    for (const t of tasks) for (const tag of t.tags ?? []) s.add(tag)
    return [...s].sort()
  }, [tasks])

  const mutate = async (fn: () => Promise<Entry | void>) => {
    setBusy(true)
    setError(null)
    try {
      const e = await fn()
      await refresh()
      if (e) {
        setSelected(e)
        setSelectedRef(refOf(e))
      }
      return true
    } catch (e) {
      fail(e)
      return false
    } finally {
      setBusy(false)
    }
  }

  const patch = (p: Patch) => selected && mutate(() => api.patch(refOf(selected), p))
  const create = async (req: CreateRequest) => {
    const ok = await mutate(() => api.create(req))
    if (ok) setAdding(false)
    return ok
  }

  const selectProject = (sel: string | null) => {
    setProject(sel)
    setSelectedRef(null)
  }

  return (
    <div className="app">
      <Sidebar projects={projects} selected={project} onSelect={selectProject} version={version} />
      <main className="main">
        <Toolbar
          filters={filters}
          onChange={setFilters}
          tags={allTags}
          count={visible.length}
          onRefresh={() => void refresh()}
          onAdd={() => setAdding(true)}
        />
        {error && (
          <div className="banner error" role="alert">
            {error}
            <button className="link" onClick={() => setError(null)}>
              dismiss
            </button>
          </div>
        )}
        <Warnings warnings={warnings} />
        <TaskTable tasks={visible} today={today} selectedRef={selectedRef} onSelect={setSelectedRef} />
      </main>
      {selected && (
        <TaskDetail
          key={refOf(selected)}
          entry={selected}
          today={today}
          busy={busy}
          onPatch={(p) => void patch(p)}
          onSelectRef={setSelectedRef}
          onClose={() => setSelectedRef(null)}
          onError={fail}
        />
      )}
      {adding && (
        <AddTaskDialog
          projects={projects}
          defaultSelector={project ?? selected?.project ?? projects[0]?.files[0]?.selector ?? ''}
          busy={busy}
          onSubmit={create}
          onClose={() => setAdding(false)}
        />
      )}
    </div>
  )
}
