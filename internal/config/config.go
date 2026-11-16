// Package config loads and saves the clask registry file.
//
// The registry lists every project and the TASKS.md files inside it. Nothing
// is discovered ambiently: if a file is not in the registry, clask does not
// read it.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config is the on-disk registry. Decoded strictly: unknown keys are errors.
type Config struct {
	Projects []Project `yaml:"projects"`
	Defaults Defaults  `yaml:"defaults"`
	Scan     Scan      `yaml:"scan"`
}

// Project is one registered repository or directory.
type Project struct {
	Name  string `yaml:"name"`
	Path  string `yaml:"path"` // absolute, or ~-prefixed
	Files []File `yaml:"files"`
}

// File is one task file inside a project. Path is relative to Project.Path.
// Name is the optional scope label; when empty the label is derived.
type File struct {
	Path string `yaml:"path"`
	Name string `yaml:"name,omitempty"`
}

// Defaults apply when a file or task omits a value.
type Defaults struct {
	File     string `yaml:"file"`
	Priority int    `yaml:"priority"`
}

// Scan bounds directory walks performed by `scan` and `discover`.
type Scan struct {
	Depth  int      `yaml:"depth"`
	Ignore []string `yaml:"ignore"`
}

// Default returns a config with no projects and built-in defaults.
func Default() *Config {
	return &Config{
		Defaults: Defaults{File: "TASKS.md", Priority: 3},
		Scan: Scan{
			Depth:  3,
			Ignore: []string{"node_modules", ".git", "vendor", "dist", "build", "Pods", ".next", "target"},
		},
	}
}

// DefaultPath returns $XDG_CONFIG_HOME/clask/config.yaml, falling back to
// ~/.config/clask/config.yaml.
func DefaultPath() (string, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home dir: %w", err)
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "clask", "config.yaml"), nil
}

// Load reads the registry at path. A missing file yields Default(), not an
// error, so a fresh install works before anything is registered.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return Default(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	cfg := Default()
	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)
	if err := dec.Decode(cfg); err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	cfg.fillDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config %s: %w", path, err)
	}
	return cfg, nil
}

// fillDefaults restores built-in values for fields a user left zero.
func (c *Config) fillDefaults() {
	d := Default()
	if c.Defaults.File == "" {
		c.Defaults.File = d.Defaults.File
	}
	if c.Defaults.Priority == 0 {
		c.Defaults.Priority = d.Defaults.Priority
	}
	if c.Scan.Depth == 0 {
		c.Scan.Depth = d.Scan.Depth
	}
	if c.Scan.Ignore == nil {
		c.Scan.Ignore = d.Scan.Ignore
	}
}

// Validate checks structural rules the YAML schema cannot express.
func (c *Config) Validate() error {
	seen := map[string]bool{}
	for i, p := range c.Projects {
		if p.Name == "" {
			return fmt.Errorf("projects[%d]: name is required", i)
		}
		if strings.ContainsAny(p.Name, ":/ ") {
			return fmt.Errorf("project %q: name may not contain ':', '/' or spaces", p.Name)
		}
		if seen[p.Name] {
			return fmt.Errorf("project %q: duplicate name", p.Name)
		}
		seen[p.Name] = true
		if p.Path == "" {
			return fmt.Errorf("project %q: path is required", p.Name)
		}
		files := map[string]bool{}
		for j, f := range p.Files {
			if f.Path == "" {
				return fmt.Errorf("project %q: files[%d]: path is required", p.Name, j)
			}
			if filepath.IsAbs(f.Path) {
				return fmt.Errorf("project %q: file %q must be relative to the project path", p.Name, f.Path)
			}
			clean := filepath.Clean(f.Path)
			if files[clean] {
				return fmt.Errorf("project %q: file %q listed twice", p.Name, f.Path)
			}
			files[clean] = true
			if f.Name != "" && strings.ContainsAny(f.Name, ":/ ") {
				return fmt.Errorf("project %q: file name %q may not contain ':', '/' or spaces", p.Name, f.Name)
			}
		}
	}
	if c.Defaults.Priority < 1 || c.Defaults.Priority > 5 {
		return fmt.Errorf("defaults.priority must be 1-5, got %d", c.Defaults.Priority)
	}
	if c.Scan.Depth < 1 {
		return fmt.Errorf("scan.depth must be >= 1, got %d", c.Scan.Depth)
	}
	return nil
}

// Save writes the registry atomically (temp file + rename) and creates the
// parent directory if needed.
func (c *Config) Save(path string) error {
	if err := c.Validate(); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(c); err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	if err := enc.Close(); err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	data := buf.Bytes()
	tmp, err := os.CreateTemp(filepath.Dir(path), ".config-*.yaml")
	if err != nil {
		return fmt.Errorf("create temp config: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("write temp config: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("close temp config: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("replace config: %w", err)
	}
	return nil
}

// Project returns the project with the given name.
func (c *Config) Project(name string) (*Project, bool) {
	for i := range c.Projects {
		if c.Projects[i].Name == name {
			return &c.Projects[i], true
		}
	}
	return nil, false
}

// AbsPath returns the project path with ~ expanded and made absolute.
func (p Project) AbsPath() (string, error) {
	return ExpandHome(p.Path)
}

// FileAbsPath joins the project path with a registered file path.
func (p Project) FileAbsPath(f File) (string, error) {
	root, err := p.AbsPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, f.Path), nil
}

// ExpandHome replaces a leading ~ with the user's home directory and returns
// an absolute, cleaned path.
func ExpandHome(p string) (string, error) {
	if p == "~" || strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home dir: %w", err)
		}
		p = filepath.Join(home, strings.TrimPrefix(p, "~"))
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("resolve %q: %w", p, err)
	}
	return filepath.Clean(abs), nil
}

// CollapseHome replaces the home directory prefix with ~ for display and for
// storing portable paths in the registry.
func CollapseHome(p string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return p
	}
	if p == home {
		return "~"
	}
	if strings.HasPrefix(p, home+string(filepath.Separator)) {
		return "~" + p[len(home):]
	}
	return p
}

// DefaultPriority returns the configured fallback priority.
func (c *Config) DefaultPriority() int {
	if c.Defaults.Priority == 0 {
		return 3
	}
	return c.Defaults.Priority
}
