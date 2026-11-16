---
title: Task refs
description: How to name a task on the command line and in blocked_by.
---

Commands that act on one task, such as `show`, `done`, `set` and `edit`,
take a **ref**:

```
[project[:scope]/]id
```

The id can be any prefix that matches exactly one task.

| ref | searches |
|---|---|
| `fix-sync` | Every registered file. |
| `falcon/fix-sync` | Every file in project `falcon`. |
| `falcon:app/fix` | Only the `app` file in `falcon`. |

An exact id match beats prefix matches. If a ref matches several tasks,
clask lists the full refs of every candidate so you can pick one:

```
clask: ambiguous reference "fix": falcon:app/fix-sync, blog/fix-typo
```

The same forms are used in [`blocked_by`](/clask/format/task-block/#blocked_by),
except that a bare id there means "in the same file".
