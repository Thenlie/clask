package registry

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/Thenlie/clask/internal/config"
)

// FindTaskFiles returns paths (relative to root) of every file named
// fileName within root, descending at most depth directory levels and
// skipping directories listed in ignore. depth 1 looks only at root.
func FindTaskFiles(root, fileName string, depth int, ignore []string) ([]string, error) {
	skip := toSet(ignore)
	var found []string
	err := walkBounded(root, depth, skip, func(dir string, entries []fs.DirEntry) bool {
		for _, e := range entries {
			if !e.IsDir() && e.Name() == fileName {
				rel, _ := filepath.Rel(root, filepath.Join(dir, e.Name()))
				found = append(found, rel)
			}
		}
		return true
	})
	sort.Strings(found)
	return found, err
}

// FindRepos returns absolute paths of git repositories under each of dirs,
// descending at most depth levels and never into a repository itself. A dir
// that is itself a repository is returned on its own.
func FindRepos(dirs []string, depth int, ignore []string) ([]string, error) {
	skip := toSet(ignore)
	seen := map[string]bool{}
	var found []string
	for _, d := range dirs {
		abs, err := config.ExpandHome(d)
		if err != nil {
			return nil, err
		}
		err = walkBounded(abs, depth, skip, func(dir string, entries []fs.DirEntry) bool {
			for _, e := range entries {
				if e.Name() == ".git" {
					if !seen[dir] {
						seen[dir] = true
						found = append(found, dir)
					}
					return false // do not descend into a repository
				}
			}
			return true
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Strings(found)
	return found, nil
}

// walkBounded visits root and its subdirectories up to depth levels deep.
// visit receives each directory's entries and returns false to stop
// descending below that directory.
func walkBounded(root string, depth int, skip map[string]bool, visit func(dir string, entries []fs.DirEntry) bool) error {
	info, err := os.Stat(root)
	if err != nil {
		return fmt.Errorf("walk: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("walk: %s is not a directory", root)
	}
	var rec func(dir string, level int) error
	rec = func(dir string, level int) error {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil // unreadable dir: skip silently, it is the user's machine
		}
		if !visit(dir, entries) || level >= depth {
			return nil
		}
		for _, e := range entries {
			if !e.IsDir() || skip[e.Name()] {
				continue
			}
			if err := rec(filepath.Join(dir, e.Name()), level+1); err != nil {
				return err
			}
		}
		return nil
	}
	return rec(root, 1)
}

func toSet(items []string) map[string]bool {
	m := make(map[string]bool, len(items))
	for _, s := range items {
		m[s] = true
	}
	return m
}

// Pruned describes an entry removed by Prune.
type Pruned struct {
	Project string
	File    string // empty when the whole project was removed
	Path    string
}

// Missing lists projects whose path no longer exists and files whose path
// no longer exists, without modifying the config.
func Missing(cfg *config.Config) []Pruned {
	var out []Pruned
	for _, p := range cfg.Projects {
		root, err := p.AbsPath()
		if err != nil || !exists(root) {
			out = append(out, Pruned{Project: p.Name, Path: p.Path})
			continue
		}
		for _, f := range p.Files {
			if !exists(filepath.Join(root, f.Path)) {
				out = append(out, Pruned{Project: p.Name, File: f.Path, Path: filepath.Join(root, f.Path)})
			}
		}
	}
	return out
}

// Prune removes everything Missing reports and returns what was removed.
func Prune(cfg *config.Config) []Pruned {
	gone := Missing(cfg)
	for _, g := range gone {
		p, ok := cfg.Project(g.Project)
		if !ok {
			continue
		}
		if g.File == "" {
			_ = Unregister(cfg, g.Project)
			continue
		}
		RemoveFile(p, g.File)
	}
	return gone
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
