---
title: Install
description: Install clask from source or with go install, and set up shell completion.
---

clask is a single Go binary. The web UI is compiled into it at build time.

## From a checkout

This is the recommended way. You get the CLI and the embedded web UI.

```bash
git clone https://github.com/Thenlie/clask
cd clask
mise install        # Go, node and pnpm versions from mise.toml
make web install    # build the UI, then go install
```

`mise.toml` sets `GOBIN=~/.local/bin`, so `clask` lands there. Make sure
that directory is on your `PATH`. Without mise you need Go 1.23+, Node 22
and pnpm.

## With go install

```bash
go install github.com/Thenlie/clask/cmd/clask@latest
```

This installs into `$GOBIN`, or `$(go env GOPATH)/bin`. Builds from the
module proxy don't include the web UI, so `clask serve` shows a page
explaining how to build it. Every other command works the same.

## Shell completion

```bash
clask completion zsh --help     # also bash, fish, powershell
```

The help text shows where to put the script for your shell.

## Check it works

```bash
clask version
clask projects        # empty until you register something
```
