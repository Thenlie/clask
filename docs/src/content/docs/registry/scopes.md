---
title: Scopes
description: How each task file gets its label, and how to select files with -p.
---

Every task file has a **scope**, a short label shown in the project column
as `project:scope`. A file at the project root shows just the project name.

## Where the label comes from

The first of these that applies wins:

1. `name` on the file in the [registry config](/clask/registry/config/).
2. `name` in the file's own [file block](/clask/format/file-block/).
3. The project name, if the file sits at the project root.
4. The name of the directory that contains the file.

So `packages/app/TASKS.md` in project `falcon` is `falcon:app` unless you
rename it. Scopes must be unique within a project.

## Selecting with -p

Commands that read or write task files take `-p` / `--project` with a
**selector**:

| selector | selects |
|---|---|
| `falcon` | Every file in project `falcon`. |
| `falcon:app` | Only the file with scope `app`. |
| `.` | The nearest registered task file above the current directory, or the whole project if you're in a registered project but above all its files. |

`-p` can be repeated on `list`, `next` and `check`. Without `-p`, reading
commands cover every registered file, and `add` uses `.`.

When `add` targets a project with several files, it uses the root file if
there is one. Otherwise it asks you to pick a scope.
