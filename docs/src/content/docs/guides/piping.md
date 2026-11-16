---
title: Piping and scripting
description: Send clask output to glow, tabiew, jq, spreadsheets and scripts.
---

The table is for reading. For everything else, pick a machine format.

| flag | output |
|---|---|
| `--json` | One JSON document. See the [JSON reference](/clask/reference/json/). |
| `-o jsonl` | One task per line, for streaming tools. |
| `-o tsv` / `-o csv` | Header row plus one row per task, no colour. |
| `show --md` | The task as a frontmatter block plus markdown body. |

Colour turns off automatically when output isn't a terminal. Set
`NO_COLOR=1` or pass `--no-color` to force it off.

## Render a task with glow

```bash
clask show fix-sync --md | glow -
```

## Browse in tabiew

```bash
clask list --all -o tsv | tw
```

Or with any CSV viewer: `clask list -o csv > tasks.csv`.

## jq recipes

```bash
# Titles of overdue tasks
clask list --overdue --json | jq -r '.tasks[].title'

# Open task count per project
clask list --json | jq -r '.tasks | group_by(.project)[] | "\(.[0].project)\t\(length)"'

# Everything tagged infra, as refs
clask list -t infra --json | jq -r '.tasks[] | "\(.project):\(.scope)/\(.id)"'
```

## Shell prompt or status bar

```bash
clask next -p . --json | jq -r '.tasks[0].title // empty'
```

## Bulk changes

Status commands accept several refs:

```bash
clask list -t sprint-12 --json | jq -r '.tasks[] | "\(.project):\(.scope)/\(.id)"' \
  | xargs clask done
```
