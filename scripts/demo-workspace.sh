#!/bin/sh
# Build a throwaway clask workspace with realistic tasks, for screenshots
# and demos. Nothing outside the target directory is touched.
#
#   scripts/demo-workspace.sh [dir]     # default: $TMPDIR/clask-demo
#   clask --config <dir>/config.yaml serve
set -eu

root=${1:-${TMPDIR:-/tmp}/clask-demo}
root=${root%/}
rm -rf "$root"
mkdir -p "$root/acme-app/packages/mobile" "$root/acme-app/packages/api" "$root/blog" "$root/dotfiles"
for d in acme-app blog dotfiles; do mkdir -p "$root/$d/.git"; done

today=$(date +%Y-%m-%d)
past=$(date -v-5d +%Y-%m-%d 2>/dev/null || date -d '5 days ago' +%Y-%m-%d)
soon=$(date -v+3d +%Y-%m-%d 2>/dev/null || date -d '3 days' +%Y-%m-%d)
later=$(date -v+21d +%Y-%m-%d 2>/dev/null || date -d '21 days' +%Y-%m-%d)

cat > "$root/acme-app/TASKS.md" <<MD
---
name: acme-app
---
Cross-cutting work for the Acme monorepo.

---
id: upgrade-node-22
title: Upgrade the monorepo to Node 22
priority: 2
status: in-progress
tags: [infra]
due: $soon
created: $past
effort: m
links:
  - https://nodejs.org/en/blog/release
---
Bump \`.nvmrc\`, CI images and the Dockerfile together.

- [x] CI images
- [ ] Dockerfile
- [ ] Local dev docs

---
id: release-notes-2-4
title: Write release notes for 2.4
priority: 3
status: todo
tags: [docs]
due: $later
created: $today
---

---
id: drop-ie11-polyfills
title: Remove IE11 polyfills
priority: 4
status: done
tags: [cleanup]
created: $past
completed: $today
---
MD

cat > "$root/acme-app/packages/mobile/TASKS.md" <<MD
---
name: mobile
default_priority: 2
---

---
id: fix-offline-sync
title: Fix sync race when switching accounts offline
priority: 1
status: in-progress
tags: [sync, bug]
due: $past
created: $past
effort: l
blocked_by: [acme-app:api/retry-on-503]
---
The replicator restarts before the old database closes, so the second
account sometimes sees the first account's drafts.

### Repro
1. Go offline.
2. Switch accounts twice quickly.
3. Open Drafts.

***

Likely fix: await \`replicator.stop()\` before opening the next DB.

---
id: dark-mode-icons
title: Dark mode app icons
status: todo
tags: [design]
created: $today
effort: s
---

---
id: push-permission-copy
title: Rewrite push permission prompt copy
priority: 3
status: blocked
tags: [design, copy]
created: $past
---
Waiting on legal review of the new wording.
MD

cat > "$root/acme-app/packages/api/TASKS.md" <<MD
---
name: api
---

---
id: retry-on-503
title: Retry on 503 from the sync gateway
priority: 1
status: todo
tags: [sync]
due: $soon
created: $today
effort: m
---
Exponential backoff, max 5 attempts, honor Retry-After.

---
id: rate-limit-dashboard
title: Rate limit dashboard
priority: 3
status: todo
tags: [observability]
created: $past
---
MD

cat > "$root/blog/TASKS.md" <<MD
---
name: blog
---

---
id: post-tasks-in-repo
title: 'Post: why my tasks live in the repo'
priority: 2
status: in-progress
tags: [writing]
due: $later
created: $past
---
Outline: problem, prior art, the TASKS.md format, one week of use.

---
id: rss-feed-fix
title: Fix RSS feed dates
priority: 3
status: todo
tags: [bug]
created: $today
---
MD

cat > "$root/dotfiles/TASKS.md" <<MD
---
name: dotfiles
---

---
id: switch-to-ghostty
title: Switch terminal config to Ghostty
priority: 4
status: todo
tags: [terminal]
created: $past
---

---
id: old-zsh-plugins
title: Remove unused zsh plugins
priority: 5
status: dropped
created: $past
completed: $today
---
MD

cat > "$root/config.yaml" <<YAML
projects:
  - name: acme-app
    path: $root/acme-app
    files:
      - path: TASKS.md
      - path: packages/mobile/TASKS.md
      - path: packages/api/TASKS.md
  - name: blog
    path: $root/blog
    files:
      - path: TASKS.md
  - name: dotfiles
    path: $root/dotfiles
    files:
      - path: TASKS.md
YAML

echo "Demo workspace: $root"
echo "Run:  clask --config $root/config.yaml serve"
