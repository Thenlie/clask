---
title: Writing rules
description: What to keep in mind when editing task files by hand.
---

clask is happy for you to edit task files in your editor. These rules keep
the file valid.

## Never use `---` in a body

A line that is exactly `---` always starts a new block. In a body, use
`***` or `___` for a horizontal rule instead.

```markdown
---
id: release
title: Cut the release
---
Checklist below.

***

- [ ] Tag
```

If you forget, clask reports the line and suggests `***`.

## One task file per folder

The registry lists task files by path. A project can have many, such as one
at the repo root and one per package, but at most one per directory. See
the [monorepo guide](/clask/guides/monorepo/).

## Keep ids unique within a file

Two tasks with the same id in one file make the second invalid. The same id
in two different files is fine. Use a [ref](/clask/registry/refs/) with a
scope to tell them apart.

## Quote titles with special characters

Titles are YAML strings. Quote a title that starts with a special character
or contains `: `:

```yaml
title: 'Local web UI: clask serve'
```

`clask add` and `clask set` quote for you.

## Check your work

```bash
clask check -p .
```

`check` validates structure, keys, values, duplicate ids, `blocked_by`
references and `completed` dates. It exits 1 on any problem, so it works as
a [git hook](/clask/guides/git-hook/).

## Line endings

clask reads files with CRLF line endings. Files it writes use LF.
