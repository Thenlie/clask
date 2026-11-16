---
---
Tasks for this directory. Add one with: clask add "Title"
Format: https://github.com/Thenlie/clask#format

---
id: local-web-ui-clask-serve-with-embedded
title: 'Local web UI: clask serve with embedded frontend'
priority: 3
status: in-progress
tags: [web]
created: 2026-11-16
---
HTTP API in internal/server over core.Workspace; frontend in web/ built into internal/server/dist and shipped via go:embed.

---
id: astro-starlight-docs-site
title: Astro Starlight docs site
priority: 3
status: todo
tags: [docs]
created: 2026-11-16
---
docs/ with the format spec; generate the CLI reference with cobra doc.GenMarkdownTree via make docs.

---
id: preserve-crlf-line-endings-on-write
title: Preserve CRLF line endings on write
priority: 4
status: todo
created: 2026-11-16
---
Parser accepts CRLF; writer always emits LF. Detect and keep the file's existing ending.

---
id: homebrew-tap-and-goreleaser
title: Homebrew tap and goreleaser
priority: 4
status: todo
tags: [release]
created: 2026-11-16
---

---
id: put-clask-on-path-outside-the-repo
title: Put clask on PATH outside the repo
priority: 2
status: done
created: 2026-11-16
completed: 2026-11-16
---
go install lands in mise's Go bin. Options: mise use -g go, or GOBIN=~/.local/bin in mise.toml [env].

---
id: live-refresh-for-the-web-ui
title: Live refresh for the web UI
priority: 3
status: todo
tags: [web]
created: 2026-11-16
---
Page currently needs a manual reload to see CLI/editor changes. Options:
1. fsnotify watcher on every registered file + Server-Sent Events push; page refetches. One Go dep, feels instant.
2. Refetch on window focus and after the page's own mutations. No dep; stale if left visible.
3. Poll /api/tasks every few seconds. No dep; workspace reload is cheap so cost is low.
