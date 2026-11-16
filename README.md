# clask

Cross-repo task lists from the terminal. Each project keeps a `TASKS.md`
next to its code; `clask` aggregates every registered file on your machine
into one prioritized view.

```
$ clask
PROJECT      ID             P  STATUS       DUE         TITLE
falcon:app   fix-sync-race  1  in-progress  2026-12-01  Fix sync race on event switch
clask        docs-site      2  todo                     Astro docs site
falcon       upgrade-cbl    2  todo                     Upgrade Couchbase Lite
```

- Tasks live in the repo, versioned with the code they describe.
- Nothing is scanned ambiently. You register projects; `scan` and
  `discover` only walk directories you name, to a bounded depth.
- A monorepo can have one task file per package, each with its own scope.
- Every read command takes `--json`; `show --md` pipes into `glow`.
- Mutations rewrite only the task's metadata block. The rest of the file is
  byte-for-byte untouched, so diffs stay small.

## Install

Requires Go 1.23+.

```bash
go install github.com/Thenlie/clask/cmd/clask@latest   # into $GOBIN or $(go env GOPATH)/bin
# or from a checkout:
make install        # mise.toml sets GOBIN=~/.local/bin
```

Shell completion: `clask completion zsh --help` (also bash, fish, powershell).

## Quick start

```bash
cd ~/code/my-repo
clask init                      # creates TASKS.md, registers the repo
clask add "Fix login redirect" --priority 1 -t auth --due 2026-12-01
clask add "Write docs" --body "Cover install and format."
clask                           # list open tasks across every project
clask start fix-login           # unique id prefixes work
clask done fix-login
clask show write-docs --md | glow -

# onboard existing repos (walks only ~/code, 3 levels deep)
clask scan ~/code
```

## Format

`TASKS.md` is a YAML multi-document stream. The **first block is file
metadata** (it may be empty). **Every later block is a task**, optionally
followed by a markdown body that runs until the next `---`.

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

Rules:

- A line that is exactly `---` opens a block; the next `---` closes it.
  Text between a close and the next open is the preceding task's body.
- Keys are validated strictly. An unknown key makes that task invalid; it
  is skipped with a warning and `clask check` fails.
- A `---` inside a body would open a new block, so horizontal rules must
  be `***` or `___`.

### File keys

| key | type | default | notes |
|---|---|---|---|
| `name` | string | registry scope | label for this file, e.g. a package name |
| `default_priority` | 1–5 | 3 | priority for tasks that omit one |

### Task keys

| key | type | required | notes |
|---|---|---|---|
| `id` | `^[a-z0-9][a-z0-9-]{0,39}$` | yes | unique within the file; `add` derives it from the title |
| `title` | string | yes | one line |
| `priority` | 1–5 | no | 1 is highest; defaults to the file or config default |
| `status` | enum | no | `todo` (default), `in-progress`, `blocked`, `done`, `dropped` |
| `tags` | list | no | `[a, b]` |
| `due` | `YYYY-MM-DD` | no | overdue open tasks are highlighted |
| `created` | date | no | set by `add` |
| `completed` | date | no | set by `done` / `drop`, cleared by `reopen` |
| `effort` | string | no | free-form: `s`, `l`, `2h` |
| `blocked_by` | list of refs | no | `id` in the same file, `project/id`, or `project:scope/id` |
| `links` | list of URLs | no | |

## Registry

`$XDG_CONFIG_HOME/clask/config.yaml` (default `~/.config/clask/config.yaml`).
Managed by `init`, `register`, `discover`, `scan`, `unregister` and `prune`;
editable by hand.

```yaml
projects:
  - name: falcon
    path: ~/code/falcon
    files:
      - path: TASKS.md                      # scope "falcon"
      - path: packages/app/TASKS.md         # scope "app" (dir basename)
      - path: packages/automation/TASKS.md
        name: e2e                           # explicit scope
defaults:
  file: TASKS.md
  priority: 3
scan:
  depth: 3
  ignore: [node_modules, .git, vendor, dist, build, Pods, .next, target]
```

A file's **scope** is, in order: the registry `name`, the file's own
`name`, the project name for a file at the project root, otherwise the
basename of its directory.

## Referencing tasks

Commands that take a `<ref>` accept `[project[:scope]/]id`. The id may be
any unique prefix.

```
fix-sync                  # searched across every project
falcon/fix-sync           # within project falcon (any scope)
falcon:app/fix            # within one file
```

Selectors (`-p`) are `project` or `project:scope`; `-p .` means the
nearest registered file above the current directory.

## Commands

Task commands

| command | what it does |
|---|---|
| `clask` / `clask list` | open tasks, priority then due then created. `-p`, `-s`, `-t` (all must match), `--priority`, `--due-before`, `--overdue`, `--all`, `--sort priority\|due\|created\|project` |
| `clask next` | the single top task (same flags as `list`) |
| `clask show <ref> [--md]` | every field plus body; `--md` prints frontmatter + body |
| `clask add <title>` | `-p`, `--priority`, `-t`, `--due`, `--effort`, `--link`, `--blocked-by`, `--body`, `--edit` |
| `clask start\|block\|done\|drop\|reopen <ref>...` | set status; `done`/`drop` stamp `completed`, `reopen` clears it |
| `clask set <ref> k=v...` | `title`, `id`, `priority`, `status`, `due`, `created`, `completed`, `effort`; lists take `tags=a,b`, `tags+=c`, `tags-=a` (same for `blocked_by`, `links`) |
| `clask edit [ref]` | open in `$VISUAL`/`$EDITOR` at the task's line (or `-p` for a file) |
| `clask check` | lint; exit 1 on any issue. Git-hook friendly |

Registry commands

| command | what it does |
|---|---|
| `clask init [path] [--name]` | create a task file and register it (new project, or new scope inside one) |
| `clask register [path] [--name]` | register a directory (its git root); picks up an existing root file |
| `clask discover [project]` | add existing task files inside a project (`--depth`, `-y`) |
| `clask scan <dir>...` | register git repos under the named dirs (`--depth`, `--existing`, `-y`) |
| `clask projects [--files]` | registered projects, or one row per file with scope |
| `clask unregister <project[:scope]>` | forget a project or one file; never touches disk |
| `clask prune` | drop entries whose paths vanished |

Global: `--json`, `--no-color` (also honors `NO_COLOR`), `--config`.

### JSON

`list --json` returns `{"tasks": [...], "warnings": [...]}`; `show`,
`add`, `set` return one task object; status verbs return a list; `check`
returns issues. Task objects carry `project`, `scope`, `file`, `line` plus
every task key and `body`.

## Development

```bash
mise install          # Go from mise.toml
make test             # go vet + go test
make build            # bin/clask
```

Layout follows Go conventions: `cmd/clask` is the entrypoint,
`internal/core` holds all behavior (the CLI is a thin layer over it, so a
future `clask serve` can reuse it), `internal/taskfile` is the format.
