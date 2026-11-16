package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/Thenlie/clask/internal/cli"
)

func TestGenerate(t *testing.T) {
	dir := t.TempDir()
	stale := filepath.Join(dir, "clask-removed.md")
	os.WriteFile(stale, []byte("old"), 0o644)

	n, err := generate(cli.NewRootCmd(), dir, "/base/")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Error("stale page not removed")
	}
	pages, _ := filepath.Glob(filepath.Join(dir, "*.md"))
	if len(pages) != n || n < 20 {
		t.Fatalf("wrote %d pages, glob found %d", n, len(pages))
	}
	have := map[string]bool{}
	for _, p := range pages {
		have[strings.TrimSuffix(filepath.Base(p), ".md")] = true
	}
	for _, want := range []string{"clask", "clask-list", "clask-serve", "clask-completion", "clask-completion-zsh"} {
		if !have[want] {
			t.Errorf("missing page %s", want)
		}
	}
	linkRe := regexp.MustCompile(`\]\(/base/([a-z-]+)/\)`)
	for _, p := range pages {
		b, _ := os.ReadFile(p)
		s := string(b)
		if !strings.HasPrefix(s, "---\ntitle: \"clask") || !strings.Contains(s, "\ndescription: ") {
			t.Errorf("%s: bad frontmatter:\n%.200s", p, s)
		}
		if strings.Contains(s, "\n### ") || strings.Contains(s, "Auto generated") {
			t.Errorf("%s: not tidied", p)
		}
		for _, m := range linkRe.FindAllStringSubmatch(s, -1) {
			if !have[m[1]] {
				t.Errorf("%s links to missing page %s", p, m[1])
			}
		}
	}
}
