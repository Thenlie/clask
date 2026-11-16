# clask

Cross-repo task lists from the terminal. Each project keeps a `TASKS.md`
next to its code; `clask` aggregates every registered file on your machine
into one prioritized view.

**Docs: https://thenlie.github.io/clask/**

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
- Mutations rewrite only the task's metadata block, so diffs stay small.
- `clask serve` opens a local web UI over the same files.

A task file is a YAML multi-document stream: a file block, then one block
per task with an optional markdown body.

```markdown
---
name: app
---

---
id: fix-sync-race
title: Fix sync race on event switch
priority: 1
status: in-progress
tags: [sync]
due: 2026-12-01
---
Replicator restarts before the old DB closes.
```

## Install

```bash
git clone https://github.com/Thenlie/clask && cd clask
mise install          # Go, node, pnpm
make web install      # CLI + embedded web UI into ~/.local/bin

# or CLI only (clask serve shows a hint page):
go install github.com/Thenlie/clask/cmd/clask@latest
```

## Quick start

```bash
cd ~/code/my-repo
clask init                      # creates TASKS.md, registers the repo
clask add "Fix login redirect" --priority 1 -t auth --due 2026-12-01
clask                           # open tasks across every project
clask done fix-login            # unique id prefixes work
clask scan ~/code               # register more repos (bounded walk)
clask serve                     # browser UI
```

## Documentation

- [Task file format](https://thenlie.github.io/clask/format/overview/)
- [Registry, scopes and refs](https://thenlie.github.io/clask/registry/config/)
- [Guides](https://thenlie.github.io/clask/guides/monorepo/): monorepos, onboarding, git hooks, piping, editors
- [Web UI](https://thenlie.github.io/clask/web-ui/overview/) and [HTTP API](https://thenlie.github.io/clask/api/overview/)
- [CLI reference](https://thenlie.github.io/clask/reference/cli/clask/), generated from `--help`
- [JSON output](https://thenlie.github.io/clask/reference/json/)

## Development

```bash
mise install          # Go, node and pnpm from mise.toml
make test             # go vet + go test
make build            # web UI + bin/clask (single binary)
make build-go         # Go only, skips the frontend build
make docs-dev         # docs site with hot reload on :4321
make docs             # build docs/dist
```

Frontend dev loop with hot reload:

```bash
clask serve --no-open   # API on :7373
make web-dev            # Vite on :5173, proxies /api to the Go server
```

Layout follows Go conventions: `cmd/clask` is the entrypoint,
`internal/core` holds all behavior, `internal/taskfile` is the format,
`internal/server` is the HTTP API with the built frontend embedded from
`internal/server/dist`, and `web/` is the React source. Both the CLI and
the server are thin layers over `internal/core`. `docs/` is the Astro
Starlight site; `cmd/gendocs` generates its CLI reference from the cobra
commands, so edit help text in `internal/cli`, not the generated pages.

`scripts/demo-workspace.sh` builds a throwaway workspace with realistic
tasks for screenshots and demos.
