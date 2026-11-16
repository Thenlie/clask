---
title: API overview
description: The JSON HTTP API served by clask serve.
---

`clask serve` exposes a small JSON API under `/api`. The web UI uses it, and
so can your scripts. It runs on `http://127.0.0.1:7373` by default.

| method | path | does |
|---|---|---|
| `GET` | `/api/meta` | Version, config path and today's date. |
| `GET` | [`/api/projects`](/clask/api/projects/) | Registered projects and files with counts. |
| `GET` | [`/api/tasks`](/clask/api/tasks/#list-tasks) | List tasks, like `clask list`. |
| `GET` | [`/api/tasks/{ref}`](/clask/api/tasks/#get-a-task) | One task, like `clask show`. |
| `POST` | [`/api/tasks`](/clask/api/tasks/#create-a-task) | Add a task, like `clask add`. |
| `PATCH` | [`/api/tasks/{ref}`](/clask/api/tasks/#update-a-task) | Change fields, like `clask set`. |
| `POST` | [`/api/open/{ref}`](/clask/api/open/) | Open the task in your editor. |
| `GET` | [`/api/check`](/clask/api/check/) | Lint, like `clask check`. |

## Conventions

- Bodies and responses are JSON. Task objects use the
  [same shape as `--json`](/clask/reference/json/).
- `{ref}` is a [task ref](/clask/registry/refs/). Its `/` and `:` don't need
  escaping: `/api/tasks/falcon:app/fix-sync`.
- Every request rereads the registry and task files. There's no cache, so
  edits from the CLI or your editor show up on the next request.
- Errors are `{"error": "message"}` with a meaningful status. See
  [errors](/clask/api/errors/).

## Security

The server binds to `127.0.0.1` and has no authentication. Anyone who can
reach the port can read and change your task files, which is why
`--host` prints a warning when you change it.

Browsers can send requests to localhost from any page you visit, so
requests that change data are rejected with 403 or 415 unless they:

- send `Content-Type: application/json`, and
- aren't marked `Sec-Fetch-Site: cross-site`, and
- have an `Origin` header, if any, that matches the server's host.

`curl` and same-origin pages meet these rules without extra work.
