package core

import (
	"os"
	"strings"
	"testing"

	"github.com/Thenlie/clask/internal/taskfile"
)

func TestSetStatusWritesOnlyMetadata(t *testing.T) {
	cfg := fixture(t)
	ws, _ := Load(cfg)
	e, _ := ws.Get("a-one")
	before, _ := os.ReadFile(e.Path)

	got, err := ws.SetStatus("a-one", taskfile.StatusDone)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != taskfile.StatusDone || got.Completed == nil {
		t.Fatalf("entry after = %+v", got.Task)
	}
	after, _ := os.ReadFile(e.Path)
	bl, al := strings.Split(string(before), "\n"), strings.Split(string(after), "\n")
	if len(al) != len(bl)+2 { // status line changed in place, completed added
		t.Fatalf("line count before=%d after=%d\n%s", len(bl), len(al), after)
	}
	if !strings.Contains(string(after), "body of a-one") || !strings.Contains(string(after), "id: a-two") {
		t.Fatal("other content lost")
	}
	// reopen clears completed
	got, err = ws.SetStatus("a-one", taskfile.StatusTodo)
	if err != nil || got.Completed != nil || got.Status != taskfile.StatusTodo {
		t.Fatalf("reopen = %+v %v", got.Task, err)
	}
}

func TestUpdateRollsBackOnInvalid(t *testing.T) {
	ws, _ := Load(fixture(t))
	_, err := ws.Update("a-one", func(tk *taskfile.Task) error { tk.Priority = 9; return nil })
	if err == nil {
		t.Fatal("expected validation error")
	}
	e, _ := ws.Get("a-one")
	if e.Priority != 2 {
		t.Fatalf("in-memory task not restored: %d", e.Priority)
	}
	_, err = ws.Update("a-one", func(tk *taskfile.Task) error { tk.ID = "a-two"; return nil })
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected duplicate id error, got %v", err)
	}
}

func TestAdd(t *testing.T) {
	ws, _ := Load(fixture(t))
	due, _ := taskfile.ParseDate("2026-12-31")
	e, err := ws.Add("alpha", NewTask{Title: "Alpha one", Tags: []string{"new"}, Due: &due, Body: "details here"})
	if err != nil {
		t.Fatal(err)
	}
	if e.ID != "alpha-one" {
		t.Fatalf("id = %q", e.ID)
	}
	if e.Priority != 3 || e.Status != taskfile.StatusTodo || e.Created == nil || e.Body != "details here" || e.Ref() != "alpha/"+e.ID {
		t.Fatalf("added = %+v", e.Task)
	}
	// same title again gets -2
	e2, err := ws.Add("alpha", NewTask{Title: "Alpha one"})
	if err != nil || e2.ID != e.ID+"-2" {
		t.Fatalf("dedupe = %v %v", e2.ID, err)
	}
	// mono has root + scoped files: bare project goes to root file
	e3, err := ws.Add("mono", NewTask{Title: "Root add"})
	if err != nil || e3.Scope != "mono" {
		t.Fatalf("root add = %v %v", e3, err)
	}
	e4, err := ws.Add("mono:ui", NewTask{Title: "UI add", Priority: 1})
	if err != nil || e4.Scope != "ui" || e4.Priority != 1 {
		t.Fatalf("scoped add = %v %v", e4, err)
	}
	if _, err := ws.Add("nope", NewTask{Title: "x"}); err == nil {
		t.Fatal("expected unknown selector error")
	}
	if _, err := ws.Add("alpha", NewTask{Title: "   "}); err == nil {
		t.Fatal("expected empty title error")
	}
	// everything still parses
	ws2, _ := Load(ws.Config)
	if len(ws2.Entries()) != 8 {
		t.Fatalf("expected 8 entries after adds, got %d", len(ws2.Entries()))
	}
}

func TestAssign(t *testing.T) {
	tk := &taskfile.Task{ID: "x", Title: "X", Priority: 3, Status: taskfile.StatusTodo, Tags: []string{"a", "b"}}
	err := Assign(tk, []string{
		"priority=1", "title= New title ", "tags+=c,a", "tags-=b", "due=2026-01-02",
		"effort=l", "links+=https://x", "status=done", "blocked_by=p,q",
	})
	if err != nil {
		t.Fatal(err)
	}
	if tk.Priority != 1 || tk.Title != "New title" || strings.Join(tk.Tags, ",") != "a,c" || tk.Due.String() != "2026-01-02" ||
		tk.Effort != "l" || tk.Links[0] != "https://x" || tk.Status != taskfile.StatusDone || tk.Completed == nil || len(tk.BlockedBy) != 2 {
		t.Fatalf("after assign = %+v", tk)
	}
	if err := Assign(tk, []string{"due="}); err != nil || tk.Due != nil {
		t.Fatalf("clear due: %v %v", err, tk.Due)
	}
	if err := Assign(tk, []string{"status=todo"}); err != nil || tk.Completed != nil {
		t.Fatal("reopen via set should clear completed")
	}
	for _, bad := range []string{"nokey", "color=red", "priority=high", "title+=x", "due=soon", "status=meh"} {
		if err := Assign(tk, []string{bad}); err == nil {
			t.Errorf("expected error for %q", bad)
		}
	}
}

func TestEditorCommand(t *testing.T) {
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "code --wait")
	if got := strings.Join(EditorCommand("/f.md", 7), " "); got != "code --wait --goto /f.md:7" {
		t.Fatal(got)
	}
	t.Setenv("EDITOR", "/usr/bin/nvim")
	if got := strings.Join(EditorCommand("/f.md", 7), " "); got != "/usr/bin/nvim +7 /f.md" {
		t.Fatal(got)
	}
	t.Setenv("EDITOR", "someeditor")
	if got := strings.Join(EditorCommand("/f.md", 7), " "); got != "someeditor /f.md" {
		t.Fatal(got)
	}
	t.Setenv("EDITOR", "")
	if got := strings.Join(EditorCommand("/f.md", 0), " "); got != "vi /f.md" {
		t.Fatal(got)
	}
}
