---
title: Errors
description: Status codes and error bodies returned by the API.
---

Errors have a JSON body:

```json
{ "error": "no task matches \"zzz\"" }
```

Some errors add a `command` field with a CLI command that does what the
request couldn't, as with [open](/clask/api/open/).

| status | when |
|---|---|
| `400` | Bad input: invalid JSON, an unknown field, a bad status, date or priority, an empty title, or a value that fails validation. |
| `403` | A data-changing request came from another site. See [security](/clask/api/overview/#security). |
| `404` | Unknown project, scope or task ref, or an unknown `/api` path or method. |
| `409` | Ambiguous ref (the message lists every candidate), or an editor that can't be opened from the server. |
| `415` | A data-changing request without `Content-Type: application/json`. |
| `500` | The registry or a task file couldn't be read or written. |

The messages are the same ones the CLI prints, so they're safe to show to a
user as they are.
