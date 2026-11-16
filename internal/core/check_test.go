package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Thenlie/clask/internal/config"
)

func TestCheck(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "TASKS.md"), []byte(`---
---
---
id: a
title: A
blocked_by: [b, missing, other/x, other:other/x, other/nope]
---

---
id: b
title: B
status: done
---

---
id: c
title: C
completed: 2026-01-01
---

---
id: Broken
title: bad id
---
`), 0o644)
	other := filepath.Join(root, "other")
	os.Mkdir(other, 0o755)
	os.WriteFile(filepath.Join(other, "TASKS.md"), []byte("---\n---\n---\nid: x\ntitle: X\n---\n"), 0o644)

	cfg := config.Default()
	cfg.Projects = []config.Project{
		{Name: "main", Path: root, Files: []config.File{{Path: "TASKS.md"}}},
		{Name: "other", Path: other, Files: []config.File{{Path: "TASKS.md"}}},
	}
	ws, _ := Load(cfg)
	issues := ws.Check()
	var msgs []string
	for _, i := range issues {
		msgs = append(msgs, i.Message)
	}
	joined := strings.Join(msgs, "\n")
	for _, want := range []string{
		`blocked_by "missing"`,
		`blocked_by "other/nope"`,
		`"b" is done but has no completed`,
		`"c" is todo but has a completed`,
		`id "Broken" must match`,
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing issue %q in:\n%s", want, joined)
		}
	}
	for _, bad := range []string{`blocked_by "b"`, `blocked_by "other/x"`, `blocked_by "other:other/x"`} {
		if strings.Contains(joined, bad) {
			t.Errorf("false positive %q", bad)
		}
	}
	if len(issues) != 5 {
		t.Fatalf("got %d issues:\n%s", len(issues), joined)
	}
}
