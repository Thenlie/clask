---
name: falcon-app
default_priority: 2
---
Notes about this file. Ignored by clask.

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
Horizontal rules use *** here.

***

Done.

---
id: upgrade-cbl
title: Upgrade Couchbase Lite
status: todo
created: 2026-11-10
---

---
id: shipped
title: Already shipped
priority: 5
status: done
completed: 2026-11-01
---
Body right after.
