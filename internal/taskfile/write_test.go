package taskfile

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReplaceUnchangedIsByteIdentical(t *testing.T) {
	data := load(t, "canonical.md")
	f, err := Parse(data, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Problems) != 0 {
		t.Fatal(f.Problems)
	}
	for _, task := range f.Tasks {
		out, err := f.Replace(task)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(out, data) {
			t.Errorf("Replace(%s) changed bytes:\n%s", task.ID, diffHint(data, out))
		}
	}
}

func TestReplaceOnlyTouchesYAMLRange(t *testing.T) {
	data := load(t, "full.md")
	f, _ := Parse(data, Options{})
	task := f.Tasks[1] // upgrade-cbl: omitted priority, will be made explicit
	task.Status = StatusDone
	today, _ := ParseDate("2026-11-16")
	task.Completed = &today

	out, err := f.Replace(task)
	if err != nil {
		t.Fatal(err)
	}
	sp := task.Span
	if !bytes.Equal(out[:sp.YAMLStart], data[:sp.YAMLStart]) {
		t.Fatal("prefix changed")
	}
	tail := len(data) - sp.YAMLEnd
	if !bytes.Equal(out[len(out)-tail:], data[sp.YAMLEnd:]) {
		t.Fatal("suffix changed")
	}
	re, err := Parse(out, Options{})
	if err != nil {
		t.Fatal(err)
	}
	got, _ := re.Task("upgrade-cbl")
	if got.Status != StatusDone || got.Completed.String() != "2026-11-16" || got.Priority != 2 || got.Created.String() != "2026-11-10" {
		t.Fatalf("reparsed = %+v", got)
	}
	if len(re.Tasks) != 3 || re.Tasks[0].Body != f.Tasks[0].Body || re.Tasks[2].Body != f.Tasks[2].Body {
		t.Fatal("neighbouring tasks disturbed")
	}
}

func TestAppend(t *testing.T) {
	cases := map[string]string{
		"empty":            "---\n---\n",
		"trailing blanks":  "---\n---\n\n\n\n",
		"no final newline": "---\n---\n---\nid: a\ntitle: A\n---\nbody",
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			f, err := Parse([]byte(in), Options{})
			if err != nil {
				t.Fatal(err)
			}
			nt := &Task{ID: "new", Title: "New one", Priority: 3, Status: StatusTodo, Tags: []string{"x"}}
			out, err := f.Append(nt, "\n\nSome body\n\n")
			if err != nil {
				t.Fatal(err)
			}
			s := string(out)
			if !strings.HasSuffix(s, "\n\n---\nid: new\ntitle: New one\npriority: 3\nstatus: todo\ntags: [x]\n---\nSome body\n") {
				t.Fatalf("append output:\n%s", s)
			}
			if strings.Contains(s, "\n\n\n") {
				t.Fatalf("more than one blank line:\n%s", s)
			}
			re, err := Parse(out, Options{})
			if err != nil || len(re.Problems) != 0 {
				t.Fatalf("reparse: %v %v", err, re.Problems)
			}
			got, ok := re.Task("new")
			if !ok || got.Body != "Some body" {
				t.Fatalf("appended task = %+v", got)
			}
		})
	}
}

func TestWriteFileAtomicKeepsMode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "TASKS.md")
	os.WriteFile(path, []byte("old"), 0o600)
	if err := WriteFile(path, []byte("new")); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(path)
	info, _ := os.Stat(path)
	if string(b) != "new" || info.Mode().Perm() != 0o600 {
		t.Fatalf("content=%q mode=%v", b, info.Mode().Perm())
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatalf("temp file left behind: %v", entries)
	}
}

func diffHint(a, b []byte) string {
	i := 0
	for i < len(a) && i < len(b) && a[i] == b[i] {
		i++
	}
	lo := i - 40
	if lo < 0 {
		lo = 0
	}
	end := func(s []byte) int {
		if i+40 < len(s) {
			return i + 40
		}
		return len(s)
	}
	return "want ..." + string(a[lo:end(a)]) + "...\n got ..." + string(b[lo:end(b)]) + "..."
}
