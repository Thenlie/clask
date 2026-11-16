package taskfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func load(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestParseFull(t *testing.T) {
	data := load(t, "full.md")
	f, err := Parse(data, Options{DefaultPriority: 3})
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Problems) != 0 {
		t.Fatalf("unexpected problems: %v", f.Problems)
	}
	if f.Meta.Name != "falcon-app" || f.Meta.DefaultPriority != 2 {
		t.Fatalf("meta = %+v", f.Meta)
	}
	if len(f.Tasks) != 3 {
		t.Fatalf("got %d tasks", len(f.Tasks))
	}

	a := f.Tasks[0]
	if a.ID != "fix-sync-race" || a.Priority != 1 || a.Status != StatusInProgress {
		t.Fatalf("task 0 = %+v", a)
	}
	if a.Due.String() != "2026-12-01" || a.Created.String() != "2026-11-16" {
		t.Fatalf("dates = %v %v", a.Due, a.Created)
	}
	if len(a.Tags) != 2 || a.Tags[1] != "couchbase" || a.BlockedBy[0] != "upgrade-cbl" || a.Links[0] != "https://example.com/HUB-123" {
		t.Fatalf("lists = %v %v %v", a.Tags, a.BlockedBy, a.Links)
	}
	if !strings.HasPrefix(a.Body, "Replicator restarts") || !strings.HasSuffix(a.Body, "Done.") {
		t.Fatalf("body = %q", a.Body)
	}
	if a.Summary() != "Replicator restarts before the old DB closes. Repro: switch events twice fast." {
		t.Fatalf("summary = %q", a.Summary())
	}
	if a.Span.Line != 7 {
		t.Fatalf("span line = %d", a.Span.Line)
	}
	if got := string(data[a.Span.YAMLStart:a.Span.YAMLEnd]); !strings.HasPrefix(got, "id: fix-sync-race\n") || !strings.HasSuffix(got, "HUB-123\n") {
		t.Fatalf("yaml range = %q", got)
	}

	b := f.Tasks[1]
	if b.Priority != 2 {
		t.Fatalf("file default_priority not applied: %d", b.Priority)
	}
	if b.Body != "" {
		t.Fatalf("expected empty body, got %q", b.Body)
	}

	c := f.Tasks[2]
	if c.Status != StatusDone || c.Completed.String() != "2026-11-01" || c.Body != "Body right after." {
		t.Fatalf("task 2 = %+v", c)
	}
}

func TestParseEmptyFileBlock(t *testing.T) {
	f, err := Parse(load(t, "empty.md"), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Tasks) != 0 || f.Meta.Name != "" {
		t.Fatalf("unexpected content: %+v", f)
	}
}

func TestParseCRLF(t *testing.T) {
	f, err := Parse(load(t, "crlf.md"), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if f.Meta.Name != "crlf" || len(f.Tasks) != 1 || f.Tasks[0].Title != "CRLF task" || f.Tasks[0].Body != "body line" {
		t.Fatalf("crlf parse = %+v %+v", f.Meta, f.Tasks)
	}
}

func TestParseProblems(t *testing.T) {
	f, err := Parse(load(t, "problems.md"), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Tasks) != 1 || f.Tasks[0].ID != "good" {
		t.Fatalf("expected only 'good' to survive, got %+v", f.Tasks)
	}
	want := []string{
		"must match",      // Bad_ID
		"id is required",  // missing id
		"field color not", // unknown key
		"duplicate id",    // good again
		"invalid date",    // tomorrow
		"invalid status",  // doing
	}
	if len(f.Problems) != len(want) {
		for _, p := range f.Problems {
			t.Log(p)
		}
		t.Fatalf("got %d problems, want %d", len(f.Problems), len(want))
	}
	for i, w := range want {
		if !strings.Contains(f.Problems[i].Error(), w) {
			t.Errorf("problem %d = %q, want substring %q", i, f.Problems[i].Error(), w)
		}
	}
}

func TestParseStructuralErrors(t *testing.T) {
	cases := map[string]string{
		"":                                "must start",
		"# Tasks\n---\n---\n":             "must start",
		"---\nname: x\n":                  "never closed",
		"---\n---\n---\nid: a\n":          "never closed",
		"---\nbogus: 1\n---\n":            "field bogus",
		"---\ndefault_priority: 9\n---\n": "default_priority",
	}
	for in, want := range cases {
		_, err := Parse([]byte(in), Options{})
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("Parse(%q) error = %v, want substring %q", in, err, want)
		}
	}
}

func TestSlug(t *testing.T) {
	cases := map[string]string{
		"Fix sync race on event switch": "fix-sync-race-on-event-switch",
		"  Hello, World!  ":             "hello-world",
		"Ünïcödé & symbols":             "n-c-d-symbols",
		"":                              "task",
		"!!!":                           "task",
		strings.Repeat("a", 50):         strings.Repeat("a", 40),
		strings.Repeat("ab ", 20):       "ab-ab-ab-ab-ab-ab-ab-ab-ab-ab-ab-ab-ab",
	}
	for in, want := range cases {
		if got := Slug(in); got != want {
			t.Errorf("Slug(%q) = %q, want %q", in, got, want)
		}
		if got := Slug(in); !ValidID(got) {
			t.Errorf("Slug(%q) = %q is not a valid id", in, got)
		}
	}
}

func TestOverdue(t *testing.T) {
	today, _ := ParseDate("2026-11-16")
	past, _ := ParseDate("2026-11-15")
	tk := &Task{Status: StatusTodo, Due: &past}
	if !tk.Overdue(today) {
		t.Fatal("expected overdue")
	}
	tk.Status = StatusDone
	if tk.Overdue(today) {
		t.Fatal("closed tasks are never overdue")
	}
	tk.Status = StatusTodo
	tk.Due = &today
	if tk.Overdue(today) {
		t.Fatal("due today is not overdue")
	}
}
