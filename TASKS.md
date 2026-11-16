---
---
Tasks for this directory. Add one with: clask add "Title"
Format: https://github.com/Thenlie/clask#format

---
id: local-web-ui-clask-serve-with-embedded
title: 'Local web UI: clask serve with embedded frontend'
priority: 3
status: todo
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
