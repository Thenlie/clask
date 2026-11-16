---
title: Validate in a git hook
description: Run clask check before every commit so broken task files never land.
---

`clask check` exits 1 when a task file has a problem, which makes it a
natural pre-commit hook. Limit it to the current project with `-p .` so an
unrelated repo's file can't block your commit.

## Plain git hook

```bash title=".git/hooks/pre-commit"
#!/bin/sh
# Only check when a task file is staged.
if git diff --cached --name-only | grep -q 'TASKS\.md$'; then
  clask check -p . || {
    echo "Fix TASKS.md (see above) or commit with --no-verify." >&2
    exit 1
  }
fi
```

```bash
chmod +x .git/hooks/pre-commit
```

## lefthook

```yaml title="lefthook.yml"
pre-commit:
  commands:
    clask:
      glob: "**/TASKS.md"
      run: clask check -p .
```

## pre-commit (the Python framework)

```yaml title=".pre-commit-config.yaml"
repos:
  - repo: local
    hooks:
      - id: clask-check
        name: clask check
        entry: clask check -p .
        language: system
        files: TASKS\.md$
        pass_filenames: false
```

## What check catches

- Broken block structure, such as a `---` inside a body.
- Unknown keys, bad values, invalid ids and duplicate ids.
- `blocked_by` entries that don't match any task.
- Closed tasks without a `completed` date, and open tasks with one.

`clask check --json` prints the issues as a list for other tools.
