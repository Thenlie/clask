---
title: JSON output
description: The stable JSON shapes shared by --json and the HTTP API.
---

Every command that prints tasks takes `--json`. The [HTTP API](/clask/api/overview/)
returns the same shapes, so a script can switch between them.

## Task object

```json
{
  "project": "falcon",
  "scope": "app",
  "file": "/Users/you/code/falcon/packages/app/TASKS.md",
  "line": 7,
  "id": "fix-sync-race",
  "title": "Fix sync race on event switch",
  "priority": 1,
  "status": "in-progress",
  "tags": ["sync", "couchbase"],
  "due": "2026-12-01",
  "created": "2026-11-16",
  "effort": "m",
  "blocked_by": ["upgrade-cbl"],
  "links": ["https://example.com/HUB-123"],
  "body": "Replicator restarts before the old DB closes."
}
```

| field | always present | notes |
|---|---|---|
| `project` | yes | Project name. |
| `scope` | yes | The file's [scope](/clask/registry/scopes/). Equals `project` for a root file. |
| `file` | yes | Absolute path of the task file. |
| `line` | yes | Line of the task's opening `---`. |
| `id`, `title`, `priority`, `status` | yes | See [task keys](/clask/format/task-block/). |
| `tags`, `due`, `created`, `completed`, `effort`, `blocked_by`, `links` | no | Omitted when empty. Dates are `YYYY-MM-DD` strings. |
| `body` | no | Markdown body, trimmed. Omitted when empty. |

A task's full ref is `project/id` when `scope` equals `project`, and
`project:scope/id` otherwise.

## Warning object

Problems found while loading files, such as an unreadable file or a skipped
task block:

```json
{ "project": "falcon", "scope": "app", "file": "/…/TASKS.md", "line": 12, "message": "\"priorty\" is not a task key" }
```

`line` is omitted when the problem isn't tied to one line. Issues from
`clask check` use the same shape.

## Per command

| command | output |
|---|---|
| `list`, `next` | `{"tasks": [Task…], "warnings": [Warning…]}` |
| `list -o jsonl` | One Task per line, no warnings. |
| `show`, `add`, `set` | One Task. |
| `start`, `block`, `done`, `drop`, `reopen` | `[Task…]`, one per ref. |
| `check` | `[Issue…]`, empty when clean. |
| `projects` | `[{"project", "path", "files", "warnings"}…]` |

Without `--json`, warnings go to stderr so stdout stays clean.
