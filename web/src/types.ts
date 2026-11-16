// Mirrors the JSON contract of internal/core (Entry, Warning, Issue) and
// internal/server (projects, tasks responses). Keep field names in sync.

export type Status = 'todo' | 'in-progress' | 'blocked' | 'done' | 'dropped'

export const STATUSES: Status[] = ['todo', 'in-progress', 'blocked', 'done', 'dropped']

export const PRIORITIES = [1, 2, 3, 4, 5] as const

export type SortKey = 'priority' | 'due' | 'created' | 'project'

export const SORT_KEYS: SortKey[] = ['priority', 'due', 'created', 'project']

export interface Task {
  id: string
  title: string
  priority: number
  status: Status
  tags?: string[]
  due?: string
  created?: string
  completed?: string
  effort?: string
  blocked_by?: string[]
  links?: string[]
  body?: string
}

export interface Entry extends Task {
  project: string
  scope: string
  file: string
  line: number
}

export interface Warning {
  project: string
  scope: string
  file: string
  line?: number
  message: string
}

export type Issue = Warning

export interface FileInfo {
  path: string
  file: string
  scope: string
  selector: string
  open: number
  total: number
  warnings: string[]
}

export interface ProjectInfo {
  name: string
  path: string
  files: FileInfo[]
}

export interface TasksResponse {
  tasks: Entry[]
  warnings: Warning[]
}

export interface CreateRequest {
  selector: string
  title: string
  priority?: number
  tags?: string[]
  due?: string
  effort?: string
  links?: string[]
  blocked_by?: string[]
  body?: string
}

// A partial update; null clears a nullable field.
export interface Patch {
  title?: string
  priority?: number
  status?: Status
  due?: string | null
  effort?: string | null
  tags?: string[]
  blocked_by?: string[]
  links?: string[]
}

export function isClosed(s: Status): boolean {
  return s === 'done' || s === 'dropped'
}

export function selectorOf(e: Pick<Entry, 'project' | 'scope'>): string {
  return e.scope === e.project ? e.project : `${e.project}:${e.scope}`
}

export function refOf(e: Entry): string {
  return `${selectorOf(e)}/${e.id}`
}

export function isOverdue(e: Task, today: string): boolean {
  return !!e.due && !isClosed(e.status) && e.due < today
}

// Local calendar date as YYYY-MM-DD, matching taskfile.Today on the server
// that runs on the same machine.
export function today(): string {
  const d = new Date()
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}
