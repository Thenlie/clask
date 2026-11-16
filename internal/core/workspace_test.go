package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Thenlie/clask/internal/config"
	"github.com/Thenlie/clask/internal/taskfile"
)

func fixture(t *testing.T) *config.Config {
	t.Helper()
	root := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(root, rel)
		os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("alpha/TASKS.md", `---
---
---
id: a-one
title: Alpha one
priority: 2
due: 2026-11-10
created: 2026-11-01
tags: [x, y]
---
body of a-one

---
id: a-two
title: Alpha two
priority: 1
status: done
completed: 2026-11-02
---
`)
	write("mono/TASKS.md", `---
---
---
id: root-task
title: Root task
priority: 3
---
`)
	write("mono/packages/app/TASKS.md", `---
name: ui
---
---
id: app-task
title: App task
priority: 1
tags: [x]
---

---
id: broken
title: Broken
color: red
---
`)
	write("mono/packages/missing/TASKS.md", "no file block here\n")

	cfg := config.Default()
	cfg.Projects = []config.Project{
		{Name: "alpha", Path: filepath.Join(root, "alpha"), Files: []config.File{{Path: "TASKS.md"}}},
		{Name: "mono", Path: filepath.Join(root, "mono"), Files: []config.File{
			{Path: "TASKS.md"},
			{Path: "packages/app/TASKS.md"},
			{Path: "packages/missing/TASKS.md"},
			{Path: "packages/gone/TASKS.md"},
		}},
	}
	return cfg
}

func TestLoadAllAndWarnings(t *testing.T) {
	ws, err := Load(fixture(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(ws.Files) != 5 {
		t.Fatalf("loaded %d files", len(ws.Files))
	}
	es := ws.Entries()
	if len(es) != 4 {
		t.Fatalf("got %d entries: %+v", len(es), es)
	}
	// scope from file meta name
	var ui *Entry
	for i := range es {
		if es[i].ID == "app-task" {
			ui = &es[i]
		}
	}
	if ui == nil || ui.Scope != "ui" || ui.Ref() != "mono:ui/app-task" {
		t.Fatalf("meta name scope not applied: %+v", ui)
	}
	// warnings: broken task, structural file, missing file
	if len(ws.Warnings) != 3 {
		t.Fatalf("warnings = %v", ws.Warnings)
	}
	joined := ""
	for _, w := range ws.Warnings {
		joined += w.String() + "\n"
	}
	for _, want := range []string{"field color", "must start", "no such file"} {
		if !strings.Contains(joined, want) {
			t.Errorf("warnings missing %q:\n%s", want, joined)
		}
	}
}

func TestLoadSelectors(t *testing.T) {
	cfg := fixture(t)
	// "ui" comes from the file's own name; "app" is the registry-derived label.
	ws, err := Load(cfg, "mono:ui")
	if err != nil || len(ws.Files) != 1 || len(ws.Entries()) != 1 {
		t.Fatalf("meta-name selector: %v %d", err, len(ws.Files))
	}
	ws, err = Load(cfg, "mono:app")
	if err != nil || len(ws.Files) != 1 || len(ws.Entries()) != 1 {
		t.Fatalf("scope selector: %v %d", err, len(ws.Files))
	}
	ws, err = Load(cfg, "alpha", "mono")
	if err != nil || len(ws.Files) != 5 {
		t.Fatalf("multi selector: %v %d", err, len(ws.Files))
	}
	if _, err := Load(cfg, "nope"); err == nil {
		t.Fatal("expected unknown project error")
	}
	if _, err := Load(cfg, "mono:nope"); err == nil {
		t.Fatal("expected unknown scope error")
	}
}

func TestListFilterAndSort(t *testing.T) {
	ws, _ := Load(fixture(t))
	today, _ := taskfile.ParseDate("2026-11-16")

	open := ws.List(Filter{Today: today}, SortPriority)
	ids := idsOf(open)
	if strings.Join(ids, ",") != "app-task,a-one,root-task" {
		t.Fatalf("default order = %v", ids)
	}
	all := ws.List(Filter{IncludeClosed: true, Today: today}, SortPriority)
	if len(all) != 4 || all[0].ID != "a-two" {
		t.Fatalf("all = %v", idsOf(all))
	}
	done := ws.List(Filter{Statuses: []taskfile.Status{taskfile.StatusDone}}, SortPriority)
	if len(done) != 1 || done[0].ID != "a-two" {
		t.Fatalf("status filter = %v", idsOf(done))
	}
	tagged := ws.List(Filter{Tags: []string{"x", "y"}}, SortPriority)
	if len(tagged) != 1 || tagged[0].ID != "a-one" {
		t.Fatalf("tag AND filter = %v", idsOf(tagged))
	}
	over := ws.List(Filter{Overdue: true, Today: today}, SortPriority)
	if len(over) != 1 || over[0].ID != "a-one" {
		t.Fatalf("overdue = %v", idsOf(over))
	}
	byProj := ws.List(Filter{}, SortProject)
	if ids := idsOf(byProj); strings.Join(ids, ",") != "a-one,root-task,app-task" {
		t.Fatalf("project sort = %v", ids)
	}
	byDue := ws.List(Filter{}, SortDue)
	if byDue[0].ID != "a-one" {
		t.Fatalf("due sort = %v", idsOf(byDue))
	}
}

func TestGet(t *testing.T) {
	ws, _ := Load(fixture(t))
	e, err := ws.Get("a-one")
	if err != nil || e.ID != "a-one" {
		t.Fatalf("exact: %v %v", e, err)
	}
	e, err = ws.Get("root")
	if err != nil || e.ID != "root-task" {
		t.Fatalf("prefix: %v %v", e, err)
	}
	_, err = ws.Get("a-")
	if err == nil || !strings.Contains(err.Error(), "ambiguous") || !strings.Contains(err.Error(), "alpha/a-two") {
		t.Fatalf("ambiguous: %v", err)
	}
	e, err = ws.Get("mono:ui/app")
	if err != nil || e.ID != "app-task" {
		t.Fatalf("scoped: %v %v", e, err)
	}
	e, err = ws.Get("mono/app")
	if err != nil || e.ID != "app-task" {
		t.Fatalf("project-scoped: %v %v", e, err)
	}
	e, err = ws.Get("mono:mono/root")
	if err != nil || e.ID != "root-task" {
		t.Fatalf("explicit root scope: %v %v", e, err)
	}
	if _, err := ws.Get("mono:mono/app-task"); err == nil {
		t.Fatal("root scope must not match a scoped file's task")
	}
	if _, err := ws.Get("alpha/app-task"); err == nil {
		t.Fatal("wrong project must not match")
	}
	if _, err := ws.Get("zzz"); err == nil {
		t.Fatal("expected not found")
	}
}

func idsOf(es []Entry) []string {
	out := make([]string, len(es))
	for i, e := range es {
		out[i] = e.ID
	}
	return out
}
