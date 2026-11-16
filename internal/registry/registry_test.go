package registry

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Thenlie/clask/internal/config"
)

func mkrepo(t *testing.T, withTasks bool) string {
	t.Helper()
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, ".git"), 0o755)
	if withTasks {
		os.WriteFile(filepath.Join(dir, "TASKS.md"), []byte("---\n---\n"), 0o644)
	}
	return dir
}

func TestRegisterUsesGitRootAndPicksUpRootFile(t *testing.T) {
	repo := mkrepo(t, true)
	sub := filepath.Join(repo, "packages", "a")
	os.MkdirAll(sub, 0o755)

	cfg := config.Default()
	p, err := Register(cfg, sub, "")
	if err != nil {
		t.Fatal(err)
	}
	abs, _ := p.AbsPath()
	if abs != repo {
		t.Fatalf("project path = %q, want git root %q", abs, repo)
	}
	if p.Name != filepath.Base(repo) {
		t.Fatalf("name = %q, want basename", p.Name)
	}
	if len(p.Files) != 1 || p.Files[0].Path != "TASKS.md" {
		t.Fatalf("root TASKS.md not added: %+v", p.Files)
	}
}

func TestRegisterRejectsDuplicates(t *testing.T) {
	repo := mkrepo(t, false)
	cfg := config.Default()
	if _, err := Register(cfg, repo, "one"); err != nil {
		t.Fatal(err)
	}
	if _, err := Register(cfg, repo, "two"); err == nil || !strings.Contains(err.Error(), "already registered") {
		t.Fatalf("expected duplicate-path error, got %v", err)
	}
	other := mkrepo(t, false)
	if _, err := Register(cfg, other, "one"); err == nil || !strings.Contains(err.Error(), "already in use") {
		t.Fatalf("expected duplicate-name error, got %v", err)
	}
}

func TestRegisterNonGitDirUsesDirItself(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Default()
	p, err := Register(cfg, dir, "plain")
	if err != nil {
		t.Fatal(err)
	}
	abs, _ := p.AbsPath()
	if abs != dir {
		t.Fatalf("path = %q, want %q", abs, dir)
	}
}

func TestScopeAndAddFile(t *testing.T) {
	p := config.Project{Name: "mono", Path: "/mono"}
	if s := Scope(p, config.File{Path: "TASKS.md"}); s != "mono" {
		t.Fatalf("root scope = %q", s)
	}
	if s := Scope(p, config.File{Path: "packages/app/TASKS.md"}); s != "app" {
		t.Fatalf("dir scope = %q", s)
	}
	if s := Scope(p, config.File{Path: "packages/app/TASKS.md", Name: "ui"}); s != "ui" {
		t.Fatalf("explicit scope = %q", s)
	}

	cfg := config.Default()
	cfg.Projects = []config.Project{p}
	proj := &cfg.Projects[0]
	if _, err := AddFile(cfg, proj, "TASKS.md", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := AddFile(cfg, proj, "./TASKS.md", ""); err == nil {
		t.Fatal("expected duplicate file error")
	}
	if _, err := AddFile(cfg, proj, "a/app/TASKS.md", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := AddFile(cfg, proj, "b/app/TASKS.md", ""); err == nil || !strings.Contains(err.Error(), "scope") {
		t.Fatalf("expected scope clash, got %v", err)
	}
	if _, err := AddFile(cfg, proj, "b/app/TASKS.md", "app2"); err != nil {
		t.Fatal(err)
	}
	if _, err := AddFile(cfg, proj, "../x/TASKS.md", ""); err == nil {
		t.Fatal("expected outside-project error")
	}
}

func TestUnregisterProjectAndScope(t *testing.T) {
	cfg := config.Default()
	cfg.Projects = []config.Project{
		{Name: "a", Path: "/a", Files: []config.File{{Path: "TASKS.md"}, {Path: "pkg/x/TASKS.md"}}},
		{Name: "b", Path: "/b"},
	}
	if err := Unregister(cfg, "a:x"); err != nil {
		t.Fatal(err)
	}
	if len(cfg.Projects[0].Files) != 1 {
		t.Fatalf("scope not removed: %+v", cfg.Projects[0].Files)
	}
	if err := Unregister(cfg, "a:nope"); err == nil {
		t.Fatal("expected missing scope error")
	}
	if err := Unregister(cfg, "b"); err != nil {
		t.Fatal(err)
	}
	if len(cfg.Projects) != 1 || cfg.Projects[0].Name != "a" {
		t.Fatalf("project b not removed: %+v", cfg.Projects)
	}
	if err := Unregister(cfg, "zzz"); err == nil {
		t.Fatal("expected missing project error")
	}
}

func TestResolveScope(t *testing.T) {
	cfg := config.Default()
	cfg.Projects = []config.Project{
		{Name: "a", Path: "/a", Files: []config.File{{Path: "TASKS.md"}, {Path: "pkg/x/TASKS.md"}}},
	}
	p, f, err := ResolveScope(cfg, "a")
	if err != nil || p.Name != "a" || f != nil {
		t.Fatalf("project-only resolve: %v %v %v", p, f, err)
	}
	_, f, err = ResolveScope(cfg, "a:x")
	if err != nil || f == nil || f.Path != "pkg/x/TASKS.md" {
		t.Fatalf("scope resolve: %v %v", f, err)
	}
	if _, _, err := ResolveScope(cfg, "a:y"); err == nil {
		t.Fatal("expected error for unknown scope")
	}
}
