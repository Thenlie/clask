package core

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Thenlie/clask/internal/registry"
	"github.com/Thenlie/clask/internal/taskfile"
)

// fileOf finds the loaded file an entry came from.
func (ws *Workspace) fileOf(e Entry) (*LoadedFile, error) {
	for _, lf := range ws.Files {
		if lf.Path == e.Path && lf.File != nil {
			return lf, nil
		}
	}
	return nil, fmt.Errorf("file %s is not loaded", e.Path)
}

// commit writes data to lf's path and reparses it so later calls on this
// workspace see fresh spans.
func (ws *Workspace) commit(lf *LoadedFile, data []byte) error {
	if err := taskfile.WriteFile(lf.Path, data); err != nil {
		return err
	}
	parsed, err := taskfile.Parse(data, taskfile.Options{DefaultPriority: ws.Config.DefaultPriority()})
	if err != nil {
		return fmt.Errorf("reparse after write: %w", err)
	}
	lf.File = parsed
	return nil
}

// Update applies fn to the task ref names, validates, and writes only that
// task's metadata block back to disk.
func (ws *Workspace) Update(ref string, fn func(*taskfile.Task) error) (Entry, error) {
	e, err := ws.Get(ref)
	if err != nil {
		return Entry{}, err
	}
	lf, err := ws.fileOf(e)
	if err != nil {
		return Entry{}, err
	}
	before := *e.Task
	if err := fn(e.Task); err != nil {
		*e.Task = before
		return Entry{}, err
	}
	if err := e.Task.Validate(); err != nil {
		*e.Task = before
		return Entry{}, err
	}
	if e.ID != before.ID {
		if _, dup := lf.File.Task(e.ID); dup {
			*e.Task = before
			return Entry{}, fmt.Errorf("id %q already exists in %s", e.ID, lf.Path)
		}
	}
	data, err := lf.File.Replace(e.Task)
	if err != nil {
		*e.Task = before
		return Entry{}, err
	}
	if err := ws.commit(lf, data); err != nil {
		return Entry{}, err
	}
	return ws.Get(e.Ref())
}

// SetStatus changes a task's status, maintaining the completed date.
func (ws *Workspace) SetStatus(ref string, s taskfile.Status) (Entry, error) {
	return ws.Update(ref, func(t *taskfile.Task) error {
		t.Status = s
		if s.Closed() {
			today := taskfile.Today()
			t.Completed = &today
		} else {
			t.Completed = nil
		}
		return nil
	})
}

// NewTask holds the user-supplied fields for Add.
type NewTask struct {
	Title     string
	Priority  int // 0 = file/config default
	Tags      []string
	Due       *taskfile.Date
	Effort    string
	Links     []string
	BlockedBy []string
	Body      string
}

// Add appends a task to the file selected by selector ("project" or
// "project:scope"). A project with several files needs a scope unless one
// of them is the root file. The id is derived from the title and made
// unique within the file.
func (ws *Workspace) Add(selector string, nt NewTask) (Entry, error) {
	lf, err := ws.targetFile(selector)
	if err != nil {
		return Entry{}, err
	}
	if lf.File == nil {
		return Entry{}, fmt.Errorf("%s could not be parsed; fix it first (see warnings)", lf.Path)
	}
	today := taskfile.Today()
	t := &taskfile.Task{
		ID:        uniqueID(lf.File, taskfile.Slug(nt.Title)),
		Title:     strings.TrimSpace(nt.Title),
		Priority:  nt.Priority,
		Status:    taskfile.StatusTodo,
		Tags:      nt.Tags,
		Due:       nt.Due,
		Created:   &today,
		Effort:    nt.Effort,
		BlockedBy: nt.BlockedBy,
		Links:     nt.Links,
	}
	if t.Priority == 0 {
		t.Priority = lf.File.Meta.DefaultPriority
	}
	if t.Priority == 0 {
		t.Priority = ws.Config.DefaultPriority()
	}
	if err := t.Validate(); err != nil {
		return Entry{}, err
	}
	data, err := lf.File.Append(t, nt.Body)
	if err != nil {
		return Entry{}, err
	}
	if err := ws.commit(lf, data); err != nil {
		return Entry{}, err
	}
	return ws.Get(lf.Project.Name + ":" + lf.Scope + "/" + t.ID)
}

// targetFile picks exactly one loaded file for a selector.
func (ws *Workspace) targetFile(selector string) (*LoadedFile, error) {
	projName, scope, hasScope := strings.Cut(selector, ":")
	var matches []*LoadedFile
	for _, lf := range ws.Files {
		if lf.Project.Name != projName {
			continue
		}
		if hasScope && lf.Scope != scope && registry.Scope(*lf.Project, lf.Cfg) != scope {
			continue
		}
		matches = append(matches, lf)
	}
	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("no task file matches %q", selector)
	case 1:
		return matches[0], nil
	}
	for _, lf := range matches {
		if lf.Scope == lf.Project.Name {
			return lf, nil
		}
	}
	scopes := make([]string, len(matches))
	for i, lf := range matches {
		scopes[i] = projName + ":" + lf.Scope
	}
	return nil, fmt.Errorf("project %q has several task files; pick one: %s", projName, strings.Join(scopes, ", "))
}

func uniqueID(f *taskfile.File, base string) string {
	if _, taken := f.Task(base); !taken {
		return base
	}
	for i := 2; ; i++ {
		suffix := "-" + strconv.Itoa(i)
		cand := base
		if len(cand)+len(suffix) > 40 {
			cand = strings.TrimRight(cand[:40-len(suffix)], "-")
		}
		cand += suffix
		if _, taken := f.Task(cand); !taken {
			return cand
		}
	}
}

// Assign applies "key=value" style edits used by `clask set`:
//
//	title=..., priority=N, status=..., due=YYYY-MM-DD (empty clears),
//	effort=... (empty clears), id=new-id,
//	tags=a,b (replace), tags+=a (add), tags-=a (remove); same for
//	blocked_by and links.
func Assign(t *taskfile.Task, assignments []string) error {
	for _, a := range assignments {
		key, val, ok := strings.Cut(a, "=")
		if !ok {
			return fmt.Errorf("%q is not key=value", a)
		}
		op := "="
		if strings.HasSuffix(key, "+") || strings.HasSuffix(key, "-") {
			op = key[len(key)-1:] + "="
			key = key[:len(key)-1]
		}
		if err := assignOne(t, key, op, val); err != nil {
			return err
		}
	}
	return nil
}

func assignOne(t *taskfile.Task, key, op, val string) error {
	if op != "=" {
		switch key {
		case "tags", "blocked_by", "links":
		default:
			return fmt.Errorf("%s does not support %s (only tags, blocked_by, links do)", key, op)
		}
	}
	switch key {
	case "title":
		t.Title = strings.TrimSpace(val)
	case "id":
		t.ID = val
	case "priority":
		n, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("priority must be a number, got %q", val)
		}
		t.Priority = n
	case "status":
		s, err := taskfile.ParseStatus(val)
		if err != nil {
			return err
		}
		t.Status = s
		if s.Closed() && t.Completed == nil {
			today := taskfile.Today()
			t.Completed = &today
		}
		if !s.Closed() {
			t.Completed = nil
		}
	case "due":
		return setDate(&t.Due, val)
	case "created":
		return setDate(&t.Created, val)
	case "completed":
		return setDate(&t.Completed, val)
	case "effort":
		t.Effort = val
	case "tags":
		t.Tags = editList(t.Tags, op, val)
	case "blocked_by":
		t.BlockedBy = editList(t.BlockedBy, op, val)
	case "links":
		t.Links = editList(t.Links, op, val)
	default:
		return fmt.Errorf("unknown key %q (want title, id, priority, status, due, created, completed, effort, tags, blocked_by or links)", key)
	}
	return nil
}

func setDate(dst **taskfile.Date, val string) error {
	if val == "" {
		*dst = nil
		return nil
	}
	d, err := taskfile.ParseDate(val)
	if err != nil {
		return err
	}
	*dst = &d
	return nil
}

func editList(cur []string, op, val string) []string {
	items := splitList(val)
	switch op {
	case "=":
		return items
	case "+=":
		out := append([]string(nil), cur...)
		for _, it := range items {
			if !contains(out, it) {
				out = append(out, it)
			}
		}
		return out
	case "-=":
		var out []string
		for _, c := range cur {
			if !contains(items, c) {
				out = append(out, c)
			}
		}
		return out
	}
	return cur
}

func splitList(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// EditorCommand builds the argv to open path at line in the user's editor
// ($VISUAL, then $EDITOR, then vi), using the editor's go-to-line syntax
// when known.
func EditorCommand(path string, line int) []string {
	editor := os.Getenv("VISUAL")
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		editor = "vi"
	}
	argv := strings.Fields(editor)
	if line <= 0 {
		return append(argv, path)
	}
	switch base := baseName(argv[0]); base {
	case "code", "code-insiders", "cursor", "codium", "windsurf":
		return append(argv, "--goto", fmt.Sprintf("%s:%d", path, line))
	case "subl", "zed":
		return append(argv, fmt.Sprintf("%s:%d", path, line))
	case "vi", "vim", "nvim", "nano", "micro", "emacs", "emacsclient", "hx", "helix", "kak":
		return append(argv, fmt.Sprintf("+%d", line), path)
	}
	return append(argv, path)
}

func baseName(p string) string {
	if i := strings.LastIndexAny(p, "/\\"); i >= 0 {
		return p[i+1:]
	}
	return p
}
