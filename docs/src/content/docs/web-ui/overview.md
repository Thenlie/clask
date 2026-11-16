---
title: Web UI
description: Browse and update every task in a local browser page with clask serve.
---

```bash
clask serve
```

This starts a server on `http://127.0.0.1:7373` and opens your browser.
Press Ctrl-C to stop it. Pass `--port` to change the port, or `--no-open`
to skip opening the browser.

## What's on the page

- **Sidebar.** Every project with its open task count. Projects with several
  task files list each scope underneath. Click one to filter.
- **Toolbar.** Search over id, title and tags, a priority filter, sort order,
  overdue and "show closed" toggles, status chips, and a chip for every tag
  in view.
- **Table.** The same columns as `clask list`. Overdue dates are red and
  priorities 1 and 2 stand out.
- **Detail pane.** Click a row to see every field and the rendered markdown
  body.

The selected project and task are kept in the URL, so you can reload or
bookmark a view.

See [screenshots](/clask/web-ui/screenshots/).

## What you can change

| in the browser | how |
|---|---|
| Status | Status buttons. `completed` is set and cleared for you. |
| Title | Click the title to edit. Enter saves, Escape cancels. |
| Priority, due date, effort | Fields in the detail pane. |
| Tags | Remove with ×, add by typing and pressing Enter. |
| New task | "+ Add" opens a form with file, title, priority, due, effort, tags and body. |

Writes go through the same code as the CLI, so only the changed task's
metadata block is rewritten.

**The body is edited in your editor.** "Open in editor" opens the task at
its line in a GUI editor, or shows the `clask edit` command for terminal
editors. See [editing in your editor](/clask/guides/editor/).

## Seeing outside changes

The server rereads files on every request. After you edit a file by hand or
run a CLI command, press **Refresh** or reload the page.

## Developing the UI

The frontend is React and TypeScript in `web/`, built into
`internal/server/dist` and embedded in the binary.

```bash
clask serve --no-open   # API on :7373
make web-dev            # Vite on http://localhost:5173 with hot reload, proxies /api
make build              # rebuild the embedded UI and bin/clask
```

The page talks to the [HTTP API](/clask/api/overview/), which you can use
directly too.
