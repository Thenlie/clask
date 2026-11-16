---
title: Projects
description: List registered projects and their task files.
---

```
GET /api/projects
```

```bash
curl -s http://127.0.0.1:7373/api/projects
```

```json
[
  {
    "name": "falcon",
    "path": "~/code/falcon",
    "files": [
      {
        "path": "packages/app/TASKS.md",
        "file": "/Users/you/code/falcon/packages/app/TASKS.md",
        "scope": "app",
        "selector": "falcon:app",
        "open": 4,
        "total": 9,
        "warnings": []
      }
    ]
  }
]
```

| field | meaning |
|---|---|
| `name`, `path` | As written in the [registry](/clask/registry/config/). |
| `files[].path` | Path relative to the project. |
| `files[].file` | Absolute path. Empty if the path couldn't be resolved. |
| `files[].scope` | Effective [scope](/clask/registry/scopes/). Empty if the file couldn't be loaded. |
| `files[].selector` | Value to pass as `project` or `selector` elsewhere in the API. |
| `files[].open` | Tasks that aren't done or dropped. |
| `files[].total` | All valid tasks in the file. |
| `files[].warnings` | Load problems for this file, as messages. |
