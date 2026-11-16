package registry

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Thenlie/clask/internal/config"
)

func touch(t *testing.T, path string) {
	t.Helper()
	os.MkdirAll(filepath.Dir(path), 0o755)
	if err := os.WriteFile(path, []byte("---\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestFindTaskFiles(t *testing.T) {
	root := t.TempDir()
	touch(t, filepath.Join(root, "TASKS.md"))
	touch(t, filepath.Join(root, "packages", "a", "TASKS.md"))
	touch(t, filepath.Join(root, "packages", "a", "deep", "er", "TASKS.md")) // depth 5
	touch(t, filepath.Join(root, "node_modules", "x", "TASKS.md"))
	touch(t, filepath.Join(root, "packages", "b", "tasks.md")) // wrong case

	got, err := FindTaskFiles(root, "TASKS.md", 3, []string{"node_modules"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(got, ",") != "TASKS.md,packages/a/TASKS.md" {
		t.Fatalf("got %v", got)
	}
	got, _ = FindTaskFiles(root, "TASKS.md", 1, nil)
	if len(got) != 1 || got[0] != "TASKS.md" {
		t.Fatalf("depth 1 = %v", got)
	}
	got, _ = FindTaskFiles(root, "TASKS.md", 10, nil)
	if len(got) != 4 {
		t.Fatalf("unbounded no-ignore = %v", got)
	}
}

func TestFindRepos(t *testing.T) {
	root := t.TempDir()
	mk := func(parts ...string) {
		os.MkdirAll(filepath.Join(append([]string{root}, parts...)...), 0o755)
	}
	mk("work", "repo1", ".git")
	mk("work", "repo1", "sub", ".git") // nested: must not be found
	mk("work", "group", "repo2", ".git")
	mk("work", "node_modules", "pkg", ".git")
	mk("work", "plain")
	mk("deep", "a", "b", "c", "repo3", ".git") // beyond depth 3

	got, err := FindRepos([]string{filepath.Join(root, "work"), filepath.Join(root, "deep")}, 3, []string{"node_modules"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		filepath.Join(root, "work", "group", "repo2"),
		filepath.Join(root, "work", "repo1"),
	}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("got %v, want %v", got, want)
	}
	// scanning a repo itself returns just that repo
	got, _ = FindRepos([]string{filepath.Join(root, "work", "repo1")}, 3, nil)
	if len(got) != 1 {
		t.Fatalf("self scan = %v", got)
	}
	if _, err := FindRepos([]string{filepath.Join(root, "nope")}, 1, nil); err == nil {
		t.Fatal("expected error for missing dir")
	}
}

func TestPrune(t *testing.T) {
	root := t.TempDir()
	touch(t, filepath.Join(root, "keep", "TASKS.md"))
	os.MkdirAll(filepath.Join(root, "partial"), 0o755)
	touch(t, filepath.Join(root, "partial", "TASKS.md"))

	cfg := config.Default()
	cfg.Projects = []config.Project{
		{Name: "keep", Path: filepath.Join(root, "keep"), Files: []config.File{{Path: "TASKS.md"}}},
		{Name: "partial", Path: filepath.Join(root, "partial"), Files: []config.File{{Path: "TASKS.md"}, {Path: "gone/TASKS.md"}}},
		{Name: "vanished", Path: filepath.Join(root, "vanished")},
	}
	gone := Prune(cfg)
	if len(gone) != 2 {
		t.Fatalf("pruned = %+v", gone)
	}
	if len(cfg.Projects) != 2 {
		t.Fatalf("projects = %+v", cfg.Projects)
	}
	p, _ := cfg.Project("partial")
	if len(p.Files) != 1 {
		t.Fatalf("partial files = %+v", p.Files)
	}
	if again := Prune(cfg); len(again) != 0 {
		t.Fatalf("second prune removed %+v", again)
	}
}
