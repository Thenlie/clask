package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultPathHonorsXDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg")
	got, err := DefaultPath()
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join("/tmp/xdg", "clask", "config.yaml"); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestDefaultPathFallsBackToDotConfig(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", "/home/u")
	got, err := DefaultPath()
	if err != nil {
		t.Fatal(err)
	}
	if want := "/home/u/.config/clask/config.yaml"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestLoadMissingReturnsDefault(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "nope.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Defaults.File != "TASKS.md" || cfg.Defaults.Priority != 3 || cfg.Scan.Depth != 3 {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	if len(cfg.Projects) != 0 {
		t.Fatalf("expected no projects, got %d", len(cfg.Projects))
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.yaml")
	cfg := Default()
	cfg.Projects = []Project{{
		Name: "falcon",
		Path: "~/Cvent/Code/falcon",
		Files: []File{
			{Path: "TASKS.md"},
			{Path: "packages/app/TASKS.md", Name: "app"},
		},
	}}
	if err := cfg.Save(path); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	p, ok := got.Project("falcon")
	if !ok {
		t.Fatal("project falcon missing after round trip")
	}
	if len(p.Files) != 2 || p.Files[1].Name != "app" {
		t.Fatalf("files not preserved: %+v", p.Files)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("config not written: %v", err)
	}
}

func TestLoadRejectsUnknownKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	os.WriteFile(path, []byte("projects: []\nbogus: 1\n"), 0o644)
	_, err := Load(path)
	if err == nil || !strings.Contains(err.Error(), "bogus") {
		t.Fatalf("expected unknown-key error, got %v", err)
	}
}

func TestLoadPartialFileKeepsDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	os.WriteFile(path, []byte("projects:\n  - name: a\n    path: /tmp/a\n"), 0o644)
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Defaults.File != "TASKS.md" || cfg.Scan.Depth != 3 || len(cfg.Scan.Ignore) == 0 {
		t.Fatalf("defaults not filled: %+v", cfg)
	}
}

func TestValidate(t *testing.T) {
	cases := []struct {
		name string
		mut  func(*Config)
		want string
	}{
		{"dup project", func(c *Config) {
			c.Projects = []Project{{Name: "a", Path: "/a"}, {Name: "a", Path: "/b"}}
		}, "duplicate"},
		{"bad name", func(c *Config) {
			c.Projects = []Project{{Name: "a:b", Path: "/a"}}
		}, "may not contain"},
		{"abs file", func(c *Config) {
			c.Projects = []Project{{Name: "a", Path: "/a", Files: []File{{Path: "/abs/TASKS.md"}}}}
		}, "relative"},
		{"dup file", func(c *Config) {
			c.Projects = []Project{{Name: "a", Path: "/a", Files: []File{{Path: "TASKS.md"}, {Path: "./TASKS.md"}}}}
		}, "twice"},
		{"priority", func(c *Config) { c.Defaults.Priority = 9 }, "1-5"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := Default()
			tc.mut(cfg)
			err := cfg.Validate()
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("want error containing %q, got %v", tc.want, err)
			}
		})
	}
}

func TestExpandCollapseHome(t *testing.T) {
	t.Setenv("HOME", "/home/u")
	got, err := ExpandHome("~/x/../y")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/home/u/y" {
		t.Fatalf("ExpandHome = %q", got)
	}
	if c := CollapseHome("/home/u/y"); c != "~/y" {
		t.Fatalf("CollapseHome = %q", c)
	}
	if c := CollapseHome("/home/user2/y"); c != "/home/user2/y" {
		t.Fatalf("CollapseHome must not match sibling prefix, got %q", c)
	}
}
