// Package registry mutates the project list in the config and resolves
// projects, files and scopes by name or by filesystem location.
package registry

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Thenlie/clask/internal/config"
)

// ErrNotFound is returned when a project or scope reference matches nothing.
var ErrNotFound = errors.New("not found")

// GitRoot walks up from dir looking for a .git entry (directory, or file for
// worktrees). ok is false when dir is not inside a git repository.
func GitRoot(dir string) (root string, ok bool) {
	dir = filepath.Clean(dir)
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

// Register adds the project rooted at dir. If dir is inside a git repository
// the repository root is used instead. When name is empty the directory
// basename is used. If the default task file exists at the root it is added
// as the project's first file.
func Register(cfg *config.Config, dir, name string) (*config.Project, error) {
	abs, err := config.ExpandHome(dir)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return nil, fmt.Errorf("register: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("register: %s is not a directory", abs)
	}
	if root, ok := GitRoot(abs); ok {
		abs = root
	}
	if existing, ok := ProjectByPath(cfg, abs); ok {
		return nil, fmt.Errorf("register: %s is already registered as %q", abs, existing.Name)
	}
	if name == "" {
		name = filepath.Base(abs)
	}
	if _, taken := cfg.Project(name); taken {
		return nil, fmt.Errorf("register: project name %q is already in use", name)
	}

	p := config.Project{Name: name, Path: config.CollapseHome(abs)}
	if _, err := os.Stat(filepath.Join(abs, cfg.Defaults.File)); err == nil {
		p.Files = append(p.Files, config.File{Path: cfg.Defaults.File})
	}
	cfg.Projects = append(cfg.Projects, p)
	if err := cfg.Validate(); err != nil {
		cfg.Projects = cfg.Projects[:len(cfg.Projects)-1]
		return nil, err
	}
	return &cfg.Projects[len(cfg.Projects)-1], nil
}

// AddFile records a task file inside an existing project. rel is relative to
// the project root; name is the optional explicit scope label.
func AddFile(cfg *config.Config, p *config.Project, rel, name string) (*config.File, error) {
	rel = filepath.Clean(rel)
	if filepath.IsAbs(rel) || strings.HasPrefix(rel, "..") {
		return nil, fmt.Errorf("add file: %q must be inside the project", rel)
	}
	for _, f := range p.Files {
		if filepath.Clean(f.Path) == rel {
			return nil, fmt.Errorf("add file: %s is already registered in %q", rel, p.Name)
		}
	}
	candidate := config.File{Path: rel, Name: name}
	for _, f := range p.Files {
		if Scope(*p, f) == Scope(*p, candidate) {
			return nil, fmt.Errorf("add file: scope %q already used by %s in %q; pass --name", Scope(*p, candidate), f.Path, p.Name)
		}
	}
	p.Files = append(p.Files, candidate)
	if err := cfg.Validate(); err != nil {
		p.Files = p.Files[:len(p.Files)-1]
		return nil, err
	}
	return &p.Files[len(p.Files)-1], nil
}

// Unregister removes a whole project ("name") or a single file
// ("name:scope"). Disk files are never touched.
func Unregister(cfg *config.Config, ref string) error {
	projName, scope, hasScope := strings.Cut(ref, ":")
	idx := -1
	for i := range cfg.Projects {
		if cfg.Projects[i].Name == projName {
			idx = i
			break
		}
	}
	if idx < 0 {
		return fmt.Errorf("unregister: project %q: %w", projName, ErrNotFound)
	}
	if !hasScope {
		cfg.Projects = append(cfg.Projects[:idx], cfg.Projects[idx+1:]...)
		return nil
	}
	p := &cfg.Projects[idx]
	for i, f := range p.Files {
		if Scope(*p, f) == scope {
			p.Files = append(p.Files[:i], p.Files[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("unregister: scope %q in project %q: %w", scope, projName, ErrNotFound)
}

// Scope returns the label for a file: its explicit name, the project name
// for a file at the project root, otherwise the basename of its directory.
func Scope(p config.Project, f config.File) string {
	if f.Name != "" {
		return f.Name
	}
	dir := filepath.Dir(filepath.Clean(f.Path))
	if dir == "." {
		return p.Name
	}
	return filepath.Base(dir)
}

// ProjectByPath finds the project whose root is exactly abs.
func ProjectByPath(cfg *config.Config, abs string) (*config.Project, bool) {
	abs = filepath.Clean(abs)
	for i := range cfg.Projects {
		root, err := cfg.Projects[i].AbsPath()
		if err == nil && root == abs {
			return &cfg.Projects[i], true
		}
	}
	return nil, false
}

// ResolveScope finds a project and optionally one of its files from a
// "project" or "project:scope" reference. file is nil when no scope given.
func ResolveScope(cfg *config.Config, ref string) (p *config.Project, file *config.File, err error) {
	projName, scope, hasScope := strings.Cut(ref, ":")
	p, ok := cfg.Project(projName)
	if !ok {
		return nil, nil, fmt.Errorf("project %q: %w", projName, ErrNotFound)
	}
	if !hasScope {
		return p, nil, nil
	}
	for i := range p.Files {
		if Scope(*p, p.Files[i]) == scope {
			return p, &p.Files[i], nil
		}
	}
	return nil, nil, fmt.Errorf("scope %q in project %q: %w", scope, projName, ErrNotFound)
}

// Context locates the project and the nearest registered file for a
// working directory. file is nil when cwd is inside a project but under no
// registered file's directory. ok is false when cwd is in no project.
func Context(cfg *config.Config, cwd string) (p *config.Project, file *config.File, ok bool) {
	cwd = filepath.Clean(cwd)
	best := -1
	for i := range cfg.Projects {
		root, err := cfg.Projects[i].AbsPath()
		if err != nil || !isWithin(cwd, root) {
			continue
		}
		if best < 0 || len(root) > len(mustAbs(cfg.Projects[best])) {
			best = i
		}
	}
	if best < 0 {
		return nil, nil, false
	}
	p = &cfg.Projects[best]
	root := mustAbs(*p)
	bestLen := -1
	for i := range p.Files {
		dir := filepath.Join(root, filepath.Dir(p.Files[i].Path))
		if isWithin(cwd, dir) && len(dir) > bestLen {
			bestLen = len(dir)
			file = &p.Files[i]
		}
	}
	return p, file, true
}

// RemoveFile drops the file with the given relative path from a project.
func RemoveFile(p *config.Project, rel string) bool {
	rel = filepath.Clean(rel)
	for i, f := range p.Files {
		if filepath.Clean(f.Path) == rel {
			p.Files = append(p.Files[:i], p.Files[i+1:]...)
			return true
		}
	}
	return false
}

func mustAbs(p config.Project) string {
	abs, _ := p.AbsPath()
	return abs
}

// isWithin reports whether path equals dir or is nested under it.
func isWithin(path, dir string) bool {
	return path == dir || strings.HasPrefix(path, dir+string(filepath.Separator))
}
