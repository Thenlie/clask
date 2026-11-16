---
title: Adding projects
description: init, register, discover, scan and prune, and when to use each.
---

clask never scans your disk on its own. These commands change the registry,
and the two that walk directories only walk the ones you name, to a bounded
depth.

| command | use it when |
|---|---|
| `clask init [path]` | You want a new task file. It creates `TASKS.md` and registers it as a new project, or as a new scope if you're inside a registered project. |
| `clask register [path]` | A repo should be known to clask. It registers the git root and picks up an existing root `TASKS.md`. |
| `clask discover [project]` | A registered project has task files that aren't listed yet, such as packages in a monorepo. |
| `clask scan <dir>...` | You want to register many repos at once. |
| `clask unregister <project[:scope]>` | You want clask to forget a project or one file. Files on disk are not touched. |
| `clask prune` | Directories or files were deleted or moved. It drops registry entries whose paths are gone. |

## How walking works

`scan` and `discover` descend at most `--depth` levels, which defaults to
`scan.depth` in the config. They skip any directory named in `scan.ignore`.
`scan` stops descending once it finds a `.git` directory, so it doesn't
wander into a repo's own subfolders.

Both list what they found and ask before changing anything. Pass `-y` to
skip the question in scripts.

See the [onboarding guide](/clask/guides/onboard-many-repos/) for a
walkthrough.
