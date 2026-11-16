---
title: Quick start
description: Create a task file, add tasks and see them across projects in two minutes.
---

## Start a task list in a repo

```bash
cd ~/code/my-repo
clask init
```

`init` creates `TASKS.md` and registers the repo's git root as a project.
The project name defaults to the directory name; pass `--name` to choose
another.

## Add and work tasks

```bash
clask add "Fix login redirect" --priority 1 -t auth --due 2026-12-01
clask add "Write docs" --body "Cover install and format."
clask                       # same as clask list
clask start fix-login       # any unique id prefix works
clask done fix-login
```

`add` writes a new block at the end of `TASKS.md` with an id derived from
the title and today's `created` date. Status changes rewrite only that
task's metadata block.

## Read a task

```bash
clask show write-docs
clask show write-docs --md | glow -
```

## Add more projects

```bash
clask scan ~/code             # register git repos under ~/code, 3 levels deep
clask scan ~/code --existing  # only repos that already have a TASKS.md
```

Then run `clask` anywhere to see every open task, highest priority first.
Use `-p` to narrow it:

```bash
clask list -p my-repo
clask list -p .             # the project you are standing in
```

## Next

- [The task file format](/clask/format/overview/) explains every key.
- [Guides](/clask/guides/monorepo/) cover monorepos, git hooks and piping.
- [`clask serve`](/clask/web-ui/overview/) opens the same tasks in a browser.
