import type {
  CreateRequest,
  Entry,
  Issue,
  Patch,
  ProjectInfo,
  SortKey,
  Status,
  TasksResponse,
} from './types'

export class ApiError extends Error {
  status: number
  // A CLI command the server suggests instead (e.g. `clask edit ref`).
  command?: string

  constructor(status: number, message: string, command?: string) {
    super(message)
    this.status = status
    this.command = command
  }
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const init: RequestInit = { method, headers: {} }
  if (method !== 'GET') {
    ;(init.headers as Record<string, string>)['Content-Type'] = 'application/json'
    init.body = body === undefined ? '{}' : JSON.stringify(body)
  }
  const res = await fetch(path, init)
  if (res.status === 204) return undefined as T
  const text = await res.text()
  let data: unknown = null
  try {
    data = text ? JSON.parse(text) : null
  } catch {
    throw new ApiError(res.status, `${res.status}: ${text.slice(0, 200)}`)
  }
  if (!res.ok) {
    const err = data as { error?: string; command?: string } | null
    throw new ApiError(res.status, err?.error ?? res.statusText, err?.command)
  }
  return data as T
}

export interface TaskQuery {
  project?: string
  statuses?: Status[]
  tags?: string[]
  priority?: number
  overdue?: boolean
  all?: boolean
  sort?: SortKey
}

function query(q: TaskQuery): string {
  const p = new URLSearchParams()
  if (q.project) p.append('project', q.project)
  for (const s of q.statuses ?? []) p.append('status', s)
  for (const t of q.tags ?? []) p.append('tag', t)
  if (q.priority) p.append('priority', String(q.priority))
  if (q.overdue) p.append('overdue', '1')
  if (q.all) p.append('all', '1')
  if (q.sort) p.append('sort', q.sort)
  const s = p.toString()
  return s ? `?${s}` : ''
}

export const api = {
  meta: () => request<{ version: string; config: string; today: string }>('GET', '/api/meta'),
  projects: () => request<ProjectInfo[]>('GET', '/api/projects'),
  tasks: (q: TaskQuery) => request<TasksResponse>('GET', `/api/tasks${query(q)}`),
  task: (ref: string) => request<Entry>('GET', `/api/tasks/${ref}`),
  create: (req: CreateRequest) => request<Entry>('POST', '/api/tasks', req),
  patch: (ref: string, patch: Patch) => request<Entry>('PATCH', `/api/tasks/${ref}`, patch),
  open: (ref: string) => request<void>('POST', `/api/open/${ref}`),
  check: () => request<Issue[]>('GET', '/api/check'),
}
