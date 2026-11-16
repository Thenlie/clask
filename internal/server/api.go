package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Thenlie/clask/internal/core"
	"github.com/Thenlie/clask/internal/taskfile"
)

func (s *Server) handleMeta(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"version": s.opts.Version,
		"config":  s.opts.ConfigPath,
		"today":   taskfile.Today(),
	})
}

// fileInfo is one registered task file as the UI sees it.
type fileInfo struct {
	Path     string   `json:"path"` // relative to the project
	File     string   `json:"file"` // absolute
	Scope    string   `json:"scope"`
	Selector string   `json:"selector"`
	Open     int      `json:"open"`
	Total    int      `json:"total"`
	Warnings []string `json:"warnings"`
}

type projectInfo struct {
	Name  string     `json:"name"`
	Path  string     `json:"path"`
	Files []fileInfo `json:"files"`
}

func (s *Server) handleProjects(w http.ResponseWriter, r *http.Request) {
	ws, err := s.load()
	if err != nil {
		writeError(w, err)
		return
	}
	byPath := map[string]*core.LoadedFile{}
	for _, lf := range ws.Files {
		byPath[lf.Path] = lf
	}
	out := []projectInfo{}
	for _, p := range ws.Config.Projects {
		pi := projectInfo{Name: p.Name, Path: p.Path, Files: []fileInfo{}}
		for _, f := range p.Files {
			fi := fileInfo{Path: f.Path, Warnings: []string{}}
			abs, err := p.FileAbsPath(f)
			if err != nil {
				fi.Warnings = append(fi.Warnings, err.Error())
				pi.Files = append(pi.Files, fi)
				continue
			}
			fi.File = abs
			lf := byPath[abs]
			if lf == nil {
				pi.Files = append(pi.Files, fi)
				continue
			}
			fi.Scope = lf.Scope
			fi.Selector = core.Entry{Project: p.Name, Scope: lf.Scope}.Selector()
			if lf.File != nil {
				for _, t := range lf.File.Tasks {
					fi.Total++
					if !t.Status.Closed() {
						fi.Open++
					}
				}
			}
			for _, wn := range ws.Warnings {
				if wn.Path == abs {
					fi.Warnings = append(fi.Warnings, wn.Message)
				}
			}
			pi.Files = append(pi.Files, fi)
		}
		out = append(out, pi)
	}
	writeJSON(w, http.StatusOK, out)
}

type tasksResponse struct {
	Tasks    []core.Entry   `json:"tasks"`
	Warnings []core.Warning `json:"warnings"`
}

// handleListTasks mirrors `clask list` flags as query parameters:
// project (repeatable, "project" or "project:scope"), status (repeatable),
// tag (repeatable, all must match), priority, due_before, overdue, all, sort.
func (s *Server) handleListTasks(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	var f core.Filter
	for _, st := range q["status"] {
		parsed, err := taskfile.ParseStatus(st)
		if err != nil {
			writeError(w, err)
			return
		}
		f.Statuses = append(f.Statuses, parsed)
	}
	f.Tags = q["tag"]
	if v := q.Get("priority"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 5 {
			writeError(w, &httpError{Status: http.StatusBadRequest, Message: "priority must be 1-5"})
			return
		}
		f.Priority = n
	}
	if v := q.Get("due_before"); v != "" {
		d, err := taskfile.ParseDate(v)
		if err != nil {
			writeError(w, err)
			return
		}
		f.DueBefore = &d
	}
	f.Overdue = isTrue(q.Get("overdue"))
	f.IncludeClosed = isTrue(q.Get("all"))
	key := core.SortPriority
	if v := q.Get("sort"); v != "" {
		var err error
		if key, err = core.ParseSortKey(v); err != nil {
			writeError(w, err)
			return
		}
	}
	ws, err := s.load(q["project"]...)
	if err != nil {
		writeError(w, err)
		return
	}
	resp := tasksResponse{Tasks: ws.List(f, key), Warnings: ws.Warnings}
	if resp.Tasks == nil {
		resp.Tasks = []core.Entry{}
	}
	if resp.Warnings == nil {
		resp.Warnings = []core.Warning{}
	}
	writeJSON(w, http.StatusOK, resp)
}

func isTrue(v string) bool {
	switch strings.ToLower(v) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}

func (s *Server) handleGetTask(w http.ResponseWriter, r *http.Request) {
	ws, err := s.load()
	if err != nil {
		writeError(w, err)
		return
	}
	e, err := ws.Get(r.PathValue("ref"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, e)
}

// createRequest is the body of POST /api/tasks.
type createRequest struct {
	Selector  string         `json:"selector"`
	Title     string         `json:"title"`
	Priority  int            `json:"priority,omitempty"`
	Tags      []string       `json:"tags,omitempty"`
	Due       *taskfile.Date `json:"due,omitempty"`
	Effort    string         `json:"effort,omitempty"`
	Links     []string       `json:"links,omitempty"`
	BlockedBy []string       `json:"blocked_by,omitempty"`
	Body      string         `json:"body,omitempty"`
}

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if req.Selector == "" {
		writeError(w, &httpError{Status: http.StatusBadRequest, Message: "selector is required (\"project\" or \"project:scope\")"})
		return
	}
	if strings.TrimSpace(req.Title) == "" {
		writeError(w, &httpError{Status: http.StatusBadRequest, Message: "title is required"})
		return
	}
	ws, err := s.load(req.Selector)
	if err != nil {
		writeError(w, err)
		return
	}
	e, err := ws.Add(req.Selector, core.NewTask{
		Title: req.Title, Priority: req.Priority, Tags: req.Tags, Due: req.Due,
		Effort: req.Effort, Links: req.Links, BlockedBy: req.BlockedBy, Body: req.Body,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, e)
}

// handlePatchTask applies a partial update. Scalar keys (title, priority,
// status, due, effort) go through core.Assign so validation and the
// completed-date rules match `clask set`; list keys replace the list.
// null clears a nullable field.
func (s *Server) handlePatchTask(w http.ResponseWriter, r *http.Request) {
	var patch map[string]json.RawMessage
	if err := decodeJSON(r, &patch); err != nil {
		writeError(w, err)
		return
	}
	if len(patch) == 0 {
		writeError(w, &httpError{Status: http.StatusBadRequest, Message: "empty patch"})
		return
	}
	ws, err := s.load()
	if err != nil {
		writeError(w, err)
		return
	}
	e, err := ws.Update(r.PathValue("ref"), func(t *taskfile.Task) error {
		return applyPatch(t, patch)
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func applyPatch(t *taskfile.Task, patch map[string]json.RawMessage) error {
	for key, raw := range patch {
		switch key {
		case "title", "priority", "status", "due", "effort":
			val, err := scalarString(raw)
			if err != nil {
				return fmt.Errorf("%s: %w", key, err)
			}
			if err := core.Assign(t, []string{key + "=" + val}); err != nil {
				return err
			}
		case "tags", "blocked_by", "links":
			var list []string
			if string(raw) != "null" {
				if err := json.Unmarshal(raw, &list); err != nil {
					return fmt.Errorf("%s must be a list of strings", key)
				}
			}
			switch key {
			case "tags":
				t.Tags = list
			case "blocked_by":
				t.BlockedBy = list
			case "links":
				t.Links = list
			}
		default:
			return fmt.Errorf("unknown key %q (want title, priority, status, due, effort, tags, blocked_by or links)", key)
		}
	}
	return nil
}

// scalarString renders a JSON string, number or null as the text
// core.Assign expects; null becomes "" which clears nullable fields.
func scalarString(raw json.RawMessage) (string, error) {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return "", err
	}
	switch x := v.(type) {
	case nil:
		return "", nil
	case string:
		return x, nil
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64), nil
	}
	return "", fmt.Errorf("must be a string, number or null")
}

// guiEditors can be launched from a server process; terminal editors
// need the user's own terminal, so the UI shows a `clask edit` command.
var guiEditors = map[string]bool{
	"code": true, "code-insiders": true, "cursor": true, "codium": true,
	"windsurf": true, "subl": true, "zed": true,
}

func (s *Server) handleOpen(w http.ResponseWriter, r *http.Request) {
	ws, err := s.load()
	if err != nil {
		writeError(w, err)
		return
	}
	e, err := ws.Get(r.PathValue("ref"))
	if err != nil {
		writeError(w, err)
		return
	}
	argv := core.EditorCommand(e.Path, e.Line)
	if !guiEditors[filepath.Base(argv[0])] {
		writeError(w, &httpError{
			Status:  http.StatusConflict,
			Message: fmt.Sprintf("%s is a terminal editor and cannot be opened from the browser", argv[0]),
			Command: "clask edit " + e.Ref(),
		})
		return
	}
	if err := s.opts.Launch(argv); err != nil {
		writeError(w, &httpError{Status: http.StatusInternalServerError, Message: "launch editor: " + err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleCheck(w http.ResponseWriter, r *http.Request) {
	ws, err := s.load(r.URL.Query()["project"]...)
	if err != nil {
		writeError(w, err)
		return
	}
	issues := ws.Check()
	if issues == nil {
		issues = []core.Issue{}
	}
	writeJSON(w, http.StatusOK, issues)
}
