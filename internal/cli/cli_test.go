package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// run executes clask with args against an isolated config, returning
// stdout and the error. Each test gets its own XDG_CONFIG_HOME.
func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	flags = globalFlags{}
	root := newRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(args)
	err := root.Execute()
	return out.String(), err
}

func must(t *testing.T, args ...string) string {
	t.Helper()
	out, err := run(t, args...)
	if err != nil {
		t.Fatalf("clask %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return out
}

func setup(t *testing.T) (repo string) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("NO_COLOR", "1")
	repo = t.TempDir()
	os.Mkdir(filepath.Join(repo, ".git"), 0o755)
	os.MkdirAll(filepath.Join(repo, "packages", "app"), 0o755)
	return repo
}

func TestEndToEnd(t *testing.T) {
	repo := setup(t)
	must(t, "init", repo, "--name", "demo")
	must(t, "init", filepath.Join(repo, "packages", "app"))

	must(t, "add", "First task", "-p", "demo", "--priority", "1", "-t", "x", "--body", "Body text")
	must(t, "add", "App task", "-p", "demo:app", "--due", "2020-01-01")
	must(t, "add", "First task", "-p", "demo") // duplicate title -> -2

	out := must(t, "list")
	for _, want := range []string{"first-task ", "first-task-2", "app-task", "2020-01-01"} {
		if !strings.Contains(out, want) {
			t.Errorf("list missing %q:\n%s", want, out)
		}
	}
	if strings.Count(out, "\n") != 4 { // header + 3 rows
		t.Errorf("unexpected row count:\n%s", out)
	}

	out = must(t, "list", "--overdue")
	if !strings.Contains(out, "app-task") || strings.Contains(out, "first-task") {
		t.Errorf("overdue filter:\n%s", out)
	}

	must(t, "start", "first-task")
	must(t, "done", "first-task")
	out = must(t, "list")
	if strings.Contains(out, "first-task ") {
		t.Errorf("done task still listed:\n%s", out)
	}
	out = must(t, "list", "--all", "--json")
	var payload struct {
		Tasks []struct {
			ID, Status, Completed string
		}
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("json: %v\n%s", err, out)
	}
	found := false
	for _, tk := range payload.Tasks {
		if tk.ID == "first-task" {
			found = true
			if tk.Status != "done" || tk.Completed == "" {
				t.Errorf("first-task = %+v", tk)
			}
		}
	}
	if !found {
		t.Fatal("first-task missing from --all")
	}

	out = must(t, "set", "demo:app/app", "priority=2", "tags+=ui", "due=")
	if !strings.Contains(out, "priority    2") || !strings.Contains(out, "ui") || strings.Contains(out, "2020-01-01") {
		t.Errorf("set output:\n%s", out)
	}

	out = must(t, "show", "first-task-2")
	if !strings.Contains(out, "demo/first-task-2") || !strings.Contains(out, "file ") {
		t.Errorf("show output:\n%s", out)
	}
	out = must(t, "show", "first-task", "--md")
	if !strings.HasPrefix(out, "---\nid: first-task\n") || !strings.Contains(out, "Body text") {
		t.Errorf("show --md output:\n%s", out)
	}

	out = must(t, "list", "-o", "tsv", "--all")
	if !strings.HasPrefix(out, "project\tid\tpriority\tstatus\tdue\ttitle\n") || !strings.Contains(out, "demo\tfirst-task\t1\tdone\t") {
		t.Errorf("tsv output:\n%s", out)
	}
	out = must(t, "list", "-o", "jsonl")
	if strings.Count(out, "\n") != 2 || !strings.HasPrefix(out, `{"project":"demo"`) {
		t.Errorf("jsonl output:\n%s", out)
	}
	if _, err := run(t, "list", "-o", "xml"); err == nil {
		t.Error("bad format should fail")
	}

	out = must(t, "next")
	if strings.Count(out, "\n") != 2 {
		t.Errorf("next should print one row:\n%s", out)
	}

	out = must(t, "projects", "--files")
	if !strings.Contains(out, "demo") || !strings.Contains(out, "app") {
		t.Errorf("projects --files:\n%s", out)
	}

	out = must(t, "check")
	if !strings.HasPrefix(out, "ok:") {
		t.Errorf("check: %s", out)
	}

	// break a file and make sure check fails and list still works
	path := filepath.Join(repo, "packages", "app", "TASKS.md")
	b, _ := os.ReadFile(path)
	os.WriteFile(path, append(b, "\n---\nid: x\ntitle: X\nbogus: 1\n---\n"...), 0o644)
	out, err := run(t, "check")
	if err == nil || !strings.Contains(out, "field bogus is not a task key") {
		t.Errorf("check should fail on unknown key: %v\n%s", err, out)
	}
	if _, err := run(t, "list"); err != nil {
		t.Errorf("list must survive a bad block: %v", err)
	}

	must(t, "unregister", "demo:app")
	out = must(t, "projects", "--files")
	if strings.Contains(out, "packages/app") {
		t.Errorf("scope not unregistered:\n%s", out)
	}
	must(t, "unregister", "demo")
	out = must(t, "projects")
	if !strings.Contains(out, "no projects registered") {
		t.Errorf("projects after unregister:\n%s", out)
	}
}

func TestErrors(t *testing.T) {
	repo := setup(t)
	must(t, "init", repo, "--name", "demo")
	cases := [][]string{
		{"show", "nothing"},
		{"done", "nothing"},
		{"set", "nothing", "priority=1"},
		{"add", "x", "-p", "ghost"},
		{"add", "x", "-p", "demo", "--due", "yesterday"},
		{"list", "--status", "weird"},
		{"list", "--sort", "weird"},
		{"register", repo},
		{"unregister", "ghost"},
		{"init", repo},
	}
	for _, args := range cases {
		if _, err := run(t, args...); err == nil {
			t.Errorf("clask %s should fail", strings.Join(args, " "))
		}
	}
}

func TestScanAndDiscoverAndPrune(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("NO_COLOR", "1")
	base := t.TempDir()
	for _, r := range []string{"one", "two"} {
		os.MkdirAll(filepath.Join(base, r, ".git"), 0o755)
	}
	os.WriteFile(filepath.Join(base, "one", "TASKS.md"), []byte("---\n---\n"), 0o644)
	os.MkdirAll(filepath.Join(base, "one", "pkg", "a"), 0o755)
	os.WriteFile(filepath.Join(base, "one", "pkg", "a", "TASKS.md"), []byte("---\n---\n"), 0o644)

	out := must(t, "scan", base, "--existing", "-y")
	if !strings.Contains(out, "registered one") || strings.Contains(out, "registered two") {
		t.Errorf("scan --existing:\n%s", out)
	}
	out = must(t, "scan", base, "-y")
	if !strings.Contains(out, "registered two") {
		t.Errorf("scan all:\n%s", out)
	}
	out = must(t, "discover", "one", "-y")
	if !strings.Contains(out, "added pkg/a/TASKS.md as one:a") {
		t.Errorf("discover:\n%s", out)
	}
	os.RemoveAll(filepath.Join(base, "two"))
	os.RemoveAll(filepath.Join(base, "one", "pkg"))
	out = must(t, "prune", "-y")
	if !strings.Contains(out, "pruned 2 entries") {
		t.Errorf("prune:\n%s", out)
	}
	out = must(t, "projects", "--files")
	if strings.Contains(out, "two") || strings.Contains(out, "pkg/a") {
		t.Errorf("after prune:\n%s", out)
	}
}
