---
title: Check
description: Lint task files over HTTP.
---

```
GET /api/check
GET /api/check?project=falcon
```

Runs the same validation as [`clask check`](/clask/reference/cli/clask-check/)
and returns a list of issues. `project` is optional and repeatable.

```json
[
  {
    "project": "falcon",
    "scope": "app",
    "file": "/Users/you/code/falcon/packages/app/TASKS.md",
    "line": 23,
    "message": "blocked_by \"ghost\" does not match any task"
  }
]
```

The status is 200 even when there are issues. An empty list means every
file is clean.
