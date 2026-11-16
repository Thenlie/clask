---
title: Task file overview
description: How a TASKS.md file is structured as a YAML multi-document stream.
---

A task file is usually called `TASKS.md`. It is a YAML multi-document
stream with markdown in between:

- The **first block is the file block**. It holds settings for the whole
  file and may be empty.
- **Every later block is one task**. Its markdown body runs until the next
  block opens.

```markdown
---
name: app
default_priority: 3
---
Notes about this file. clask ignores this text.

---
id: fix-sync-race
title: Fix sync race on event switch
priority: 1
status: in-progress
tags: [sync, couchbase]
due: 2026-12-01
created: 2026-11-16
effort: m
blocked_by: [upgrade-cbl]
links:
  - https://example.com/HUB-123
---
Replicator restarts before the old DB closes.
Repro: switch events twice fast.

### Notes
Markdown is fine here. Use `***` for horizontal rules, never `---`.

---
id: upgrade-cbl
title: Upgrade Couchbase Lite
priority: 2
status: todo
created: 2026-11-10
---
```

## Blocks

A line that is exactly `---` opens a block, and the next `---` line closes
it. Text between a block's close and the next block's open is that block's
body. Leading and trailing blank lines in a body are trimmed. Everything
else is kept byte for byte.

The file block's body is free text for humans. A task's body is shown by
`clask show`, the web UI, and `--json` as `body`. Its first paragraph is the
task's summary.

## Strict keys

Each block decodes against a fixed list of keys. An unknown or misspelled
key makes that task invalid. clask skips it with a warning that names the
file and line, and `clask check` fails. This is deliberate. A typo like
`priorty: 1` would otherwise be silently ignored.

- [File block keys](/clask/format/file-block/)
- [Task block keys](/clask/format/task-block/)
- [Writing rules](/clask/format/writing-rules/) for editing by hand

## Why this format

- **Plain text in the repo.** Tasks are reviewed, branched and merged with
  the code.
- **Standard YAML.** Each block is ordinary YAML, so editors highlight it
  and other tools can read it.
- **Small diffs.** clask edits a task by rewriting only the lines between
  its two `---` markers. Bodies and other tasks are never re-serialised.
