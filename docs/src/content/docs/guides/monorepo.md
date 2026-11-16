---
title: Monorepos
description: One task file per package, each with its own scope.
---

A monorepo can keep a task file at the root for cross-cutting work and one
in each package that has its own backlog.

## Set it up

```bash
cd ~/code/falcon
clask init                          # TASKS.md at the root, project "falcon"
cd packages/app && clask init       # scope "app"
cd ../automation && clask init --name e2e
```

Running `init` inside a registered project adds a new file to that project
rather than creating a new project. The scope comes from `--name` or the
directory name, and is written into the new file's `name:`.

Already have task files in packages? Let clask find them:

```bash
clask discover falcon
```

## Work with scopes

```bash
clask list -p falcon              # every file in the project
clask list -p falcon:app          # one package
clask list -p .                   # whichever file you're standing in
clask add "Flaky login test" -p falcon:e2e
```

Inside `packages/app`, a plain `clask add` writes to the app file, because
it's the nearest registered file above you.

The project column shows `falcon` for the root file and `falcon:app` for a
package. Refs work the same way: `falcon:app/fix-sync`.

## Cross-package blockers

```yaml
blocked_by: [falcon/upgrade-cbl]     # any file in falcon
blocked_by: [falcon:e2e/stable-ci]   # one specific file
```

`clask check -p falcon` reports blockers that don't resolve.
