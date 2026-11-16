---
title: Registry config
description: The config.yaml file that lists your projects and task files.
---

clask only reads task files listed in its registry:

```
$XDG_CONFIG_HOME/clask/config.yaml      # usually ~/.config/clask/config.yaml
```

Use `--config <path>` on any command to point at a different file. The
registry commands (`init`, `register`, `discover`, `scan`, `unregister`,
`prune`) keep it up to date, and it's fine to edit by hand. Like task files,
it is strictly validated, so a misspelled key is an error.

```yaml
projects:
  - name: falcon
    path: ~/code/falcon
    files:
      - path: TASKS.md                      # scope "falcon"
      - path: packages/app/TASKS.md         # scope "app" (directory name)
      - path: packages/automation/TASKS.md
        name: e2e                           # explicit scope
  - name: blog
    path: ~/code/blog
    files:
      - path: TASKS.md
defaults:
  file: TASKS.md
  priority: 3
scan:
  depth: 3
  ignore: [node_modules, .git, vendor, dist, build, Pods, .next, target]
```

## projects

| key | meaning |
|---|---|
| `name` | Unique project name, used in refs and `-p`. |
| `path` | Project root. `~` is expanded. |
| `files` | Task files in this project. |
| `files[].path` | Path relative to the project root. |
| `files[].name` | Optional scope label. See [scopes](/clask/registry/scopes/). |

## defaults

| key | default | meaning |
|---|---|---|
| `file` | `TASKS.md` | File name used by `init`, `register` and `discover`. |
| `priority` | `3` | Priority for tasks without one, after the file's `default_priority`. |

## scan

| key | default | meaning |
|---|---|---|
| `depth` | `3` | How many directory levels `scan` and `discover` descend. |
| `ignore` | see above | Directory names never descended into. |

A missing config file is fine. clask behaves as if it were empty and
creates it on the first registry change.
