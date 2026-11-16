package core

import (
	"fmt"
	"sort"
	"strings"
)

// Issue is one finding from Check.
type Issue struct {
	Project string `json:"project"`
	Scope   string `json:"scope"`
	Path    string `json:"file"`
	Line    int    `json:"line,omitempty"`
	Message string `json:"message"`
}

func (i Issue) String() string {
	if i.Line > 0 {
		return fmt.Sprintf("%s:%d: %s", i.Path, i.Line, i.Message)
	}
	return fmt.Sprintf("%s: %s", i.Path, i.Message)
}

// Check lints every loaded file: load warnings (unreadable or structurally
// broken files, skipped task blocks) plus cross-task rules such as
// blocked_by references that point nowhere.
func (ws *Workspace) Check() []Issue {
	var issues []Issue
	for _, w := range ws.Warnings {
		issues = append(issues, Issue(w))
	}
	entries := ws.Entries()
	byRef := map[string]bool{}
	for _, e := range entries {
		byRef[e.Ref()] = true
		byRef[e.Project+"/"+e.ID] = true
		byRef[e.Project+":"+e.Scope+"/"+e.ID] = true
	}
	for _, e := range entries {
		for _, ref := range e.BlockedBy {
			if resolvesBlockedBy(e, ref, byRef) {
				continue
			}
			issues = append(issues, Issue{
				Project: e.Project, Scope: e.Scope, Path: e.Path, Line: e.Line,
				Message: fmt.Sprintf("blocked_by %q does not match any task", ref),
			})
		}
		if e.Status.Closed() && e.Completed == nil {
			issues = append(issues, Issue{
				Project: e.Project, Scope: e.Scope, Path: e.Path, Line: e.Line,
				Message: fmt.Sprintf("task %q is %s but has no completed date", e.ID, e.Status),
			})
		}
		if !e.Status.Closed() && e.Completed != nil {
			issues = append(issues, Issue{
				Project: e.Project, Scope: e.Scope, Path: e.Path, Line: e.Line,
				Message: fmt.Sprintf("task %q is %s but has a completed date", e.ID, e.Status),
			})
		}
	}
	sort.SliceStable(issues, func(i, j int) bool {
		if issues[i].Path != issues[j].Path {
			return issues[i].Path < issues[j].Path
		}
		return issues[i].Line < issues[j].Line
	})
	return issues
}

// resolvesBlockedBy accepts a bare id (same file), "project/id" or
// "project:scope/id".
func resolvesBlockedBy(from Entry, ref string, byRef map[string]bool) bool {
	if strings.Contains(ref, "/") {
		return byRef[ref]
	}
	return byRef[from.Selector()+"/"+ref]
}
