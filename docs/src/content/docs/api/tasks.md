---
title: Tasks
description: List, read, create and update tasks over HTTP.
---

## List tasks

```
GET /api/tasks
```

Query parameters match `clask list` flags. Repeatable ones may appear more
than once.

| parameter | repeatable | meaning |
|---|---|---|
| `project` | yes | Selector: `falcon` or `falcon:app`. |
| `status` | yes | Only these statuses. Including a closed status shows closed tasks. |
| `tag` | yes | Tasks must have every tag listed. |
| `priority` | no | Exactly this priority, 1–5. |
| `due_before` | no | Only tasks due before this `YYYY-MM-DD` date. |
| `overdue` | no | `1` or `true`: open tasks past their due date. |
| `all` | no | `1` or `true`: include done and dropped tasks. |
| `sort` | no | `priority` (default), `due`, `created` or `project`. |

```bash
curl -s 'http://127.0.0.1:7373/api/tasks?project=falcon&tag=sync&sort=due'
```

```json
{ "tasks": [ { "project": "falcon", "scope": "app", "id": "fix-sync-race", "…": "…" } ], "warnings": [] }
```

Both arrays are always present, even when empty.

## Get a task

```
GET /api/tasks/{ref}
```

```bash
curl -s http://127.0.0.1:7373/api/tasks/falcon:app/fix-sync
```

Returns one task object. Closed tasks can be fetched too. A ref that
matches nothing is 404, and an ambiguous prefix is 409.

## Create a task

```
POST /api/tasks
```

| field | required | meaning |
|---|---|---|
| `selector` | yes | Target file: `falcon` or `falcon:app`. A project with several files needs a scope unless it has a root file. |
| `title` | yes | One line. The id is derived from it. |
| `priority` | no | 1–5. Defaults to the file, then the config default. |
| `tags`, `links`, `blocked_by` | no | Lists of strings. |
| `due` | no | `YYYY-MM-DD`. |
| `effort` | no | Free-form string. |
| `body` | no | Markdown body. |

```bash
curl -s -X POST http://127.0.0.1:7373/api/tasks \
  -H 'Content-Type: application/json' \
  -d '{"selector":"falcon:app","title":"Retry on 503","priority":2,"tags":["sync"]}'
```

Returns `201` and the new task, including its generated `id` and `created`
date. Unknown fields are rejected with 400.

## Update a task

```
PATCH /api/tasks/{ref}
```

Send only the fields to change.

| field | type | notes |
|---|---|---|
| `title` | string | Must not be empty. |
| `priority` | number | 1–5. |
| `status` | string | `completed` is set or cleared to match, like the status commands. |
| `due` | string or `null` | `null` or `""` clears it. |
| `effort` | string or `null` | `null` or `""` clears it. |
| `tags`, `blocked_by`, `links` | list or `null` | Replaces the whole list. `null` clears it. |

```bash
curl -s -X PATCH http://127.0.0.1:7373/api/tasks/fix-sync \
  -H 'Content-Type: application/json' \
  -d '{"status":"done"}'
```

Returns the updated task. Only that task's metadata block is rewritten.
The body and every other task stay byte for byte the same.

`id` and the body can't be changed over the API. Use `clask set <ref> id=…`
and your editor for those.
