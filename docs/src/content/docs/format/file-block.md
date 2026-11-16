---
title: File block
description: Keys allowed in the first block of a task file.
---

The first block of every task file describes the file itself. It is
required, but it can be empty:

```markdown
---
---
```

| key | type | default | meaning |
|---|---|---|---|
| `name` | string | see below | Label for this file, shown as its scope, e.g. `falcon:app`. |
| `default_priority` | integer 1–5 | `defaults.priority` in the config, which is 3 | Priority for tasks that don't set one. |

`clask init` writes `name:` for you, using `--name` or the directory name.

## How `name` is used

The `name` here is one of several sources for the file's
[scope](/clask/registry/scopes/). A `name` set on the file in the registry
config wins over this one. Use the file block when you want the label to
travel with the repo, and the registry when it's a personal preference.

## Notes after the block

Anything between the file block and the first task is ignored by clask.
Use it for a description of the file or a reminder of conventions.
