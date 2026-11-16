---
title: Onboard many repos
description: Register every repo under a directory at once, safely.
---

If your repos live under a few directories, `scan` registers them in one
go.

## Preview first

```bash
clask scan ~/code
```

clask walks `~/code` up to `scan.depth` levels (3 by default), stops at each
`.git`, skips directories in `scan.ignore`, and lists the repos it would
register. Nothing changes until you confirm.

## Only repos that already have tasks

```bash
clask scan ~/code --existing
```

This offers only repos with a task file at their root. Useful after
cloning a set of repos that already use clask.

## Tune the walk

```bash
clask scan ~/code ~/work --depth 2
clask scan ~/code -y          # no prompt, for scripts
```

For a permanent change, edit `scan.depth` and `scan.ignore` in the
[registry config](/clask/registry/config/).

## Add task files later

`scan` registers repos, not files deep inside them. To start a list in a
registered repo, run `clask init` there. To pick up existing package files,
run `clask discover <project>`.

## Clean up

When repos are deleted or moved:

```bash
clask projects        # shows "path missing" warnings
clask prune           # removes those entries after asking
```
