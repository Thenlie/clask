---
title: Task block
description: Every key a task can have, its type and its rules.
---

Every block after the first is a task. Keys are written in this order by
clask. Any other key is an error.

| key | type | required | notes |
|---|---|---|---|
| `id` | string | yes | Lowercase letters, digits and `-`, starting with a letter or digit, at most 40 characters. Unique within the file. |
| `title` | string | yes | One line. |
| `priority` | integer 1–5 | no | 1 is highest. Defaults to the file's `default_priority`, then the config default. |
| `status` | enum | no | `todo` (default), `in-progress`, `blocked`, `done`, `dropped`. |
| `tags` | list of strings | no | `[a, b]` or a block list. |
| `due` | date | no | `YYYY-MM-DD`. Open tasks past their due date are highlighted. |
| `created` | date | no | Set by `clask add`. |
| `completed` | date | no | Set by `done` and `drop`, cleared by `reopen`. |
| `effort` | string | no | Free-form, such as `s`, `l` or `2h`. Displayed only. |
| `blocked_by` | list of refs | no | Tasks this one waits on. See below. |
| `links` | list of strings | no | URLs. |

clask always writes `id`, `title`, `priority` and `status`, and leaves out
optional keys that are empty.

## ids

`clask add` derives the id from the title: lowercase, punctuation runs
become `-`, cut at a word boundary to 40 characters. A clash in the same
file gets `-2`, `-3` and so on. Rename an id with `clask set <ref> id=new-id`.

## Status and completed

`done` and `drop` are **closed** statuses. Closed tasks are hidden from
`clask list` unless you pass `--all` or ask for that status explicitly.

`completed` should be set exactly when the status is closed. The status
commands and the web UI maintain it for you. `clask check` reports a
closed task without a `completed` date, or an open task with one.

## blocked_by

Each entry refers to another task:

| form | meaning |
|---|---|
| `upgrade-cbl` | A task with that id in the same file. |
| `falcon/upgrade-cbl` | A task in project `falcon`, in any of its files. |
| `falcon:app/upgrade-cbl` | A task in one specific file. |

`clask check` reports entries that don't match any task. clask doesn't
change status based on blockers; `blocked` is a status you set yourself.

## Dates

Dates are plain `YYYY-MM-DD` with no quotes and no time. clask compares
them against your local date.
