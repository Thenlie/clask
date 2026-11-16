---
title: Editing in your editor
description: How clask edit and the web UI open task files at the right line.
---

clask writes metadata for you, but bodies are plain markdown. Write them in
your editor.

## clask edit

```bash
clask edit fix-sync          # opens the file at the task's block
clask edit -p falcon:app     # opens one task file
clask edit                   # the nearest task file
clask add "New thing" --edit # add, then jump to it
```

clask uses `$VISUAL`, then `$EDITOR`, then `vi`. It knows how to jump to a
line in these editors:

| editor | command clask runs |
|---|---|
| vi, vim, nvim, nano, micro, emacs, helix, kakoune | `editor +LINE file` |
| VS Code, Cursor, VSCodium, Windsurf | `code --goto file:LINE` |
| Sublime Text, Zed | `subl file:LINE` |

Other editors open the file at the top. Arguments in the variable are kept,
so `EDITOR="code --wait"` works.

## From the web UI

"Open in editor" in the [web UI](/clask/web-ui/overview/) runs the same
command, but from the server process. That works for GUI editors like VS Code
or Zed. A terminal editor such as vim needs your terminal, so the page shows
the `clask edit` command to copy instead.

## Editing by hand

Follow the [writing rules](/clask/format/writing-rules/) and run
`clask check -p .` when you're done.
