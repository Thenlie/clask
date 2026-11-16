---
title: Open in editor
description: Ask the server to open a task in your editor.
---

```
POST /api/open/{ref}
```

The server builds the same command as [`clask edit`](/clask/guides/editor/)
from its own `$VISUAL` or `$EDITOR`.

| editor | response |
|---|---|
| A GUI editor: VS Code, Cursor, VSCodium, Windsurf, Sublime Text or Zed | `204`. The editor opens at the task's line. |
| Anything else, such as vim or nano | `409` with the command to run in a terminal instead. |

```bash
curl -s -X POST http://127.0.0.1:7373/api/open/fix-sync -H 'Content-Type: application/json'
```

```json
{ "error": "vi is a terminal editor and cannot be opened from the browser", "command": "clask edit falcon/fix-sync" }
```

The editor variables are read from the environment `clask serve` started
in. Restart the server after changing them.
