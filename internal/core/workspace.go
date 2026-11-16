// Package core loads every registered task file and answers queries over
// them. The CLI and any future server are thin layers over this package.
package core

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/Thenlie/clask/internal/config"
	"github.com/Thenlie/clask/internal/registry"
	"github.com/Thenlie/clask/internal/taskfile"
)

// Workspace is the set of loaded task files for one invocation.
type Workspace struct {
	Config   *config.Config
	Files    []*LoadedFile
	Warnings []Warning
}

// LoadedFile is one registered task file, parsed or failed.
type LoadedFile struct {
	Project *config.Project
	Cfg     config.File
	Scope   string
	Path    string // absolute
	File    *taskfile.File
}

// Warning is a non-fatal issue found while loading: an unreadable file,
// a structurally broken file, or a skipped task block.
type Warning struct {
	Project string `json:"project"`
	Scope   string `json:"scope"`
	Path    string `json:"file"`
	Line    int    `json:"line,omitempty"`
	Message string `json:"message"`
}

func (w Warning) String() string {
	if w.Line > 0 {
		return fmt.Sprintf("%s:%d: %s", w.Path, w.Line, w.Message)
	}
	return fmt.Sprintf("%s: %s", w.Path, w.Message)
}

// Entry is a task together with where it lives. It is the JSON contract
// shared by `--json` output and any future API.
type Entry struct {
	Project string `json:"project"`
	Scope   string `json:"scope"`
	Path    string `json:"file"`
	Line    int    `json:"line"`
	*taskfile.Task
}

// Selector returns "project" for a root-scoped file or "project:scope".
func (e Entry) Selector() string {
	if e.Scope == e.Project {
		return e.Project
	}
	return e.Project + ":" + e.Scope
}

// Ref returns the fully qualified task reference, e.g. "falcon:app/fix-sync".
func (e Entry) Ref() string { return e.Selector() + "/" + e.ID }

// Load parses the files selected by selectors ("project" or
// "project:scope"). With no selectors every registered file is loaded.
// A scope matches either the registry-derived label or the file's own
// "name". Unknown selectors are errors; unreadable files become Warnings.
func Load(cfg *config.Config, selectors ...string) (*Workspace, error) {
	ws := &Workspace{Config: cfg}
	loaded := map[string]*LoadedFile{}
	loadProject := func(p *config.Project) []*LoadedFile {
		var out []*LoadedFile
		for _, f := range p.Files {
			abs, err := p.FileAbsPath(f)
			if err != nil {
				ws.warn(p.Name, registry.Scope(*p, f), f.Path, 0, err.Error())
				continue
			}
			lf, ok := loaded[abs]
			if !ok {
				lf = ws.loadFile(p, f, abs)
				loaded[abs] = lf
			}
			out = append(out, lf)
		}
		return out
	}
	add := func(lf *LoadedFile) {
		for _, have := range ws.Files {
			if have == lf {
				return
			}
		}
		ws.Files = append(ws.Files, lf)
	}

	if len(selectors) == 0 {
		for i := range cfg.Projects {
			for _, lf := range loadProject(&cfg.Projects[i]) {
				add(lf)
			}
		}
		return ws, nil
	}
	for _, sel := range selectors {
		projName, scope, hasScope := strings.Cut(sel, ":")
		p, ok := cfg.Project(projName)
		if !ok {
			return nil, fmt.Errorf("project %q: %w", projName, registry.ErrNotFound)
		}
		files := loadProject(p)
		if !hasScope {
			for _, lf := range files {
				add(lf)
			}
			continue
		}
		matched := false
		for _, lf := range files {
			if lf.Scope == scope || registry.Scope(*p, lf.Cfg) == scope {
				add(lf)
				matched = true
			}
		}
		if !matched {
			return nil, fmt.Errorf("scope %q in project %q: %w", scope, projName, registry.ErrNotFound)
		}
	}
	return ws, nil
}

func (ws *Workspace) loadFile(p *config.Project, f config.File, abs string) *LoadedFile {
	lf := &LoadedFile{Project: p, Cfg: f, Path: abs, Scope: registry.Scope(*p, f)}
	data, err := os.ReadFile(abs)
	if err != nil {
		ws.warn(p.Name, lf.Scope, abs, 0, err.Error())
		return lf
	}
	parsed, err := taskfile.Parse(data, taskfile.Options{DefaultPriority: ws.Config.DefaultPriority()})
	if err != nil {
		ws.warn(p.Name, lf.Scope, abs, 0, err.Error())
		return lf
	}
	lf.File = parsed
	if f.Name == "" && parsed.Meta.Name != "" {
		lf.Scope = parsed.Meta.Name
	}
	for _, pr := range parsed.Problems {
		ws.warn(p.Name, lf.Scope, abs, pr.Line, pr.Err.Error())
	}
	return lf
}

func (ws *Workspace) warn(project, scope, path string, line int, msg string) {
	ws.Warnings = append(ws.Warnings, Warning{Project: project, Scope: scope, Path: path, Line: line, Message: msg})
}

// Entries returns every loaded task in file order.
func (ws *Workspace) Entries() []Entry {
	var out []Entry
	for _, lf := range ws.Files {
		if lf.File == nil {
			continue
		}
		for _, t := range lf.File.Tasks {
			out = append(out, Entry{Project: lf.Project.Name, Scope: lf.Scope, Path: lf.Path, Line: t.Span.Line, Task: t})
		}
	}
	return out
}

// ErrAmbiguous is wrapped when a reference matches several tasks.
var ErrAmbiguous = errors.New("ambiguous reference")

// Get resolves "[project[:scope]/]id-or-prefix" to exactly one task.
func (ws *Workspace) Get(ref string) (Entry, error) {
	sel, idPart, scoped := strings.Cut(ref, "/")
	if !scoped {
		idPart, sel = sel, ""
	}
	if idPart == "" {
		return Entry{}, fmt.Errorf("reference %q has no task id", ref)
	}
	selProj, selScope, hasScope := strings.Cut(sel, ":")
	var exact, prefix []Entry
	for _, e := range ws.Entries() {
		if scoped && (e.Project != selProj || (hasScope && e.Scope != selScope)) {
			continue
		}
		switch {
		case e.ID == idPart:
			exact = append(exact, e)
		case strings.HasPrefix(e.ID, idPart):
			prefix = append(prefix, e)
		}
	}
	if len(exact) == 1 {
		return exact[0], nil
	}
	if len(exact) == 0 && len(prefix) == 1 {
		return prefix[0], nil
	}
	cands := append(exact, prefix...)
	if len(cands) == 0 {
		return Entry{}, fmt.Errorf("no task matches %q", ref)
	}
	refs := make([]string, len(cands))
	for i, c := range cands {
		refs[i] = c.Ref()
	}
	return Entry{}, fmt.Errorf("%w %q: %s", ErrAmbiguous, ref, strings.Join(refs, ", "))
}

// Filter narrows List results. Zero values mean "no constraint".
type Filter struct {
	Statuses      []taskfile.Status
	Tags          []string // all must be present
	Priority      int
	DueBefore     *taskfile.Date
	Overdue       bool
	IncludeClosed bool
	Today         taskfile.Date
}

// SortKey names a List ordering.
type SortKey string

// Sort keys accepted by List.
const (
	SortPriority SortKey = "priority"
	SortDue      SortKey = "due"
	SortCreated  SortKey = "created"
	SortProject  SortKey = "project"
)

// ParseSortKey validates a sort flag value.
func ParseSortKey(s string) (SortKey, error) {
	switch SortKey(s) {
	case SortPriority, SortDue, SortCreated, SortProject:
		return SortKey(s), nil
	}
	return "", fmt.Errorf("invalid sort %q (want priority, due, created or project)", s)
}

// List returns matching entries in the requested order.
func (ws *Workspace) List(f Filter, key SortKey) []Entry {
	today := f.Today
	if today.IsZero() {
		today = taskfile.Today()
	}
	var out []Entry
	for _, e := range ws.Entries() {
		if !f.IncludeClosed && len(f.Statuses) == 0 && e.Status.Closed() {
			continue
		}
		if len(f.Statuses) > 0 && !containsStatus(f.Statuses, e.Status) {
			continue
		}
		if f.Priority != 0 && e.Priority != f.Priority {
			continue
		}
		if !hasAllTags(e.Tags, f.Tags) {
			continue
		}
		if f.DueBefore != nil && (e.Due == nil || !e.Due.Before(*f.DueBefore)) {
			continue
		}
		if f.Overdue && !e.Overdue(today) {
			continue
		}
		out = append(out, e)
	}
	sortEntries(out, key)
	return out
}

func containsStatus(list []taskfile.Status, s taskfile.Status) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func hasAllTags(have, want []string) bool {
	for _, w := range want {
		found := false
		for _, h := range have {
			if h == w {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func sortEntries(es []Entry, key SortKey) {
	sort.SliceStable(es, func(i, j int) bool {
		a, b := es[i], es[j]
		switch key {
		case SortDue:
			if c := cmpDate(a.Due, b.Due); c != 0 {
				return c < 0
			}
		case SortCreated:
			if c := cmpDate(a.Created, b.Created); c != 0 {
				return c < 0
			}
		case SortProject:
			if a.Project != b.Project {
				return a.Project < b.Project
			}
			if a.Scope != b.Scope {
				return a.Scope < b.Scope
			}
		}
		if a.Priority != b.Priority {
			return a.Priority < b.Priority
		}
		if c := cmpDate(a.Due, b.Due); c != 0 {
			return c < 0
		}
		if c := cmpDate(a.Created, b.Created); c != 0 {
			return c < 0
		}
		if a.Project != b.Project {
			return a.Project < b.Project
		}
		if a.Scope != b.Scope {
			return a.Scope < b.Scope
		}
		return a.ID < b.ID
	})
}

// cmpDate orders nil after any date.
func cmpDate(a, b *taskfile.Date) int {
	switch {
	case a == nil && b == nil:
		return 0
	case a == nil:
		return 1
	case b == nil:
		return -1
	case a.Before(*b):
		return -1
	case b.Before(*a):
		return 1
	}
	return 0
}
