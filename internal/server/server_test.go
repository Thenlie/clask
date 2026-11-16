package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/Thenlie/clask/internal/config"
	"github.com/Thenlie/clask/internal/core"
)

const fixture = `---
name: demo
---

---
id: alpha
title: Alpha task
priority: 1
status: todo
tags: [x]
due: 2020-01-01
created: 2026-01-01
---
Alpha body.

---
id: beta
title: Beta task
priority: 3
status: in-progress
created: 2026-01-02
---

---
id: shipped
title: Shipped task
priority: 2
status: done
created: 2026-01-01
completed: 2026-01-03
---
`

type env struct {
	t        *testing.T
	srv      *httptest.Server
	taskFile string
	launched [][]string
}

func newEnv(t *testing.T) *env {
	t.Helper()
	t.Setenv("EDITOR", "vi")
	t.Setenv("VISUAL", "")
	repo := t.TempDir()
	taskFile := filepath.Join(repo, "TASKS.md")
	if err := os.WriteFile(taskFile, []byte(fixture), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.Projects = []config.Project{{Name: "demo", Path: repo, Files: []config.File{{Path: "TASKS.md"}}}}
	cfgPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := cfg.Save(cfgPath); err != nil {
		t.Fatal(err)
	}
	e := &env{t: t, taskFile: taskFile}
	s := New(Options{ConfigPath: cfgPath, Version: "test", Launch: func(argv []string) error {
		e.launched = append(e.launched, argv)
		return nil
	}})
	e.srv = httptest.NewServer(s.Handler())
	t.Cleanup(e.srv.Close)
	return e
}

func (e *env) do(method, path string, body any, hdr ...string) (*http.Response, []byte) {
	e.t.Helper()
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, e.srv.URL+path, rd)
	if method != http.MethodGet {
		req.Header.Set("Content-Type", "application/json")
	}
	for i := 0; i+1 < len(hdr); i += 2 {
		req.Header.Set(hdr[i], hdr[i+1])
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		e.t.Fatal(err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return resp, data
}

func (e *env) get(path string, into any) *http.Response {
	e.t.Helper()
	resp, data := e.do(http.MethodGet, path, nil)
	if resp.StatusCode != http.StatusOK {
		e.t.Fatalf("GET %s: %d %s", path, resp.StatusCode, data)
	}
	if into != nil {
		if err := json.Unmarshal(data, into); err != nil {
			e.t.Fatalf("GET %s: bad JSON %v: %s", path, err, data)
		}
	}
	return resp
}

func (e *env) file() string {
	b, err := os.ReadFile(e.taskFile)
	if err != nil {
		e.t.Fatal(err)
	}
	return string(b)
}

func wantStatus(t *testing.T, resp *http.Response, data []byte, want int) {
	t.Helper()
	if resp.StatusCode != want {
		t.Fatalf("status %d, want %d: %s", resp.StatusCode, want, data)
	}
}

func ids(es []core.Entry) []string {
	out := make([]string, len(es))
	for i, e := range es {
		out[i] = e.ID
	}
	return out
}

func TestMetaAndProjects(t *testing.T) {
	e := newEnv(t)
	var meta map[string]any
	e.get("/api/meta", &meta)
	if meta["version"] != "test" {
		t.Errorf("meta = %v", meta)
	}
	var projects []projectInfo
	e.get("/api/projects", &projects)
	if len(projects) != 1 || len(projects[0].Files) != 1 {
		t.Fatalf("projects = %+v", projects)
	}
	f := projects[0].Files[0]
	if f.Scope != "demo" || f.Selector != "demo" || f.Open != 2 || f.Total != 3 || f.File != e.taskFile {
		t.Errorf("file = %+v", f)
	}
}

func TestListTasks(t *testing.T) {
	e := newEnv(t)
	var resp tasksResponse
	e.get("/api/tasks", &resp)
	if got := ids(resp.Tasks); strings.Join(got, ",") != "alpha,beta" {
		t.Errorf("default list = %v", got)
	}
	e.get("/api/tasks?all=1&sort=created", &resp)
	if len(resp.Tasks) != 3 {
		t.Errorf("all = %v", ids(resp.Tasks))
	}
	e.get("/api/tasks?status=done", &resp)
	if got := ids(resp.Tasks); strings.Join(got, ",") != "shipped" {
		t.Errorf("status filter = %v", got)
	}
	e.get("/api/tasks?overdue=true&project=demo", &resp)
	if got := ids(resp.Tasks); strings.Join(got, ",") != "alpha" {
		t.Errorf("overdue = %v", got)
	}
	if resp.Warnings == nil {
		t.Error("warnings should be [] not null")
	}
	for path, want := range map[string]int{
		"/api/tasks?status=bogus":     400,
		"/api/tasks?priority=9":       400,
		"/api/tasks?sort=title":       400,
		"/api/tasks?project=nope":     404,
		"/api/tasks?project=demo:zzz": 404,
		"/api/nothing":                404,
	} {
		r, data := e.do(http.MethodGet, path, nil)
		wantStatus(t, r, data, want)
		if !strings.Contains(string(data), `"error"`) {
			t.Errorf("%s: no error body: %s", path, data)
		}
	}
}

func TestGetTask(t *testing.T) {
	e := newEnv(t)
	var entry core.Entry
	e.get("/api/tasks/demo/alpha", &entry)
	if entry.Title != "Alpha task" || entry.Body != "Alpha body." || entry.Line != 5 {
		t.Errorf("entry = %+v", entry)
	}
	e.get("/api/tasks/be", &entry) // unique prefix
	if entry.ID != "beta" {
		t.Errorf("prefix lookup = %s", entry.ID)
	}
	r, data := e.do(http.MethodGet, "/api/tasks/zzz", nil)
	wantStatus(t, r, data, 404)
	// "alpha" and "shipped" don't share a prefix; make one ambiguous.
	e.do(http.MethodPost, "/api/tasks", map[string]any{"selector": "demo", "title": "Alpha two"})
	r, data = e.do(http.MethodGet, "/api/tasks/alpha", nil) // exact wins
	wantStatus(t, r, data, 200)
	r, data = e.do(http.MethodGet, "/api/tasks/alp", nil)
	wantStatus(t, r, data, 409)
}

func TestCreateTask(t *testing.T) {
	e := newEnv(t)
	r, data := e.do(http.MethodPost, "/api/tasks", map[string]any{
		"selector": "demo", "title": "New thing", "priority": 2, "tags": []string{"web"},
		"due": "2026-12-01", "body": "Some body",
	})
	wantStatus(t, r, data, 201)
	var entry core.Entry
	json.Unmarshal(data, &entry)
	if entry.ID != "new-thing" || entry.Priority != 2 || entry.Due.String() != "2026-12-01" || entry.Body != "Some body" {
		t.Errorf("created = %+v", entry)
	}
	if !strings.HasSuffix(e.file(), "---\nid: new-thing\ntitle: New thing\npriority: 2\nstatus: todo\ntags: [web]\ndue: 2026-12-01\ncreated: "+entry.Created.String()+"\n---\nSome body\n") {
		t.Errorf("file tail:\n%s", e.file())
	}
	for _, body := range []map[string]any{
		{"title": "x"},
		{"selector": "demo"},
		{"selector": "demo", "title": "x", "bogus": 1},
		{"selector": "demo", "title": "x", "due": "tomorrow"},
		{"selector": "demo", "title": "x", "priority": 9},
	} {
		r, data := e.do(http.MethodPost, "/api/tasks", body)
		wantStatus(t, r, data, 400)
	}
	r, data = e.do(http.MethodPost, "/api/tasks", map[string]any{"selector": "nope", "title": "x"})
	wantStatus(t, r, data, 404)
}

func TestPatchTask(t *testing.T) {
	e := newEnv(t)
	before := e.file()
	r, data := e.do(http.MethodPatch, "/api/tasks/beta", map[string]any{"status": "done"})
	wantStatus(t, r, data, 200)
	var entry core.Entry
	json.Unmarshal(data, &entry)
	if entry.Status != "done" || entry.Completed == nil {
		t.Errorf("patched = %+v", entry)
	}
	diff := changedLines(before, e.file())
	sort.Strings(diff)
	if len(diff) != 2 || !strings.HasPrefix(diff[0], "completed: ") || !strings.HasPrefix(diff[1], "status: ") {
		t.Errorf("changed lines = %q", diff)
	}

	r, data = e.do(http.MethodPatch, "/api/tasks/beta", map[string]any{"status": "todo"})
	wantStatus(t, r, data, 200)
	entry = core.Entry{} // Unmarshal leaves omitted fields as they were
	json.Unmarshal(data, &entry)
	if entry.Completed != nil {
		t.Error("reopen should clear completed")
	}

	r, data = e.do(http.MethodPatch, "/api/tasks/alpha", map[string]any{
		"title": "Alpha renamed", "priority": 4, "due": nil, "effort": "m",
		"tags": []string{"a", "b"}, "links": []string{"https://x"},
	})
	wantStatus(t, r, data, 200)
	entry = core.Entry{}
	json.Unmarshal(data, &entry)
	if entry.Title != "Alpha renamed" || entry.Priority != 4 || entry.Due != nil || entry.Effort != "m" ||
		strings.Join(entry.Tags, ",") != "a,b" || len(entry.Links) != 1 {
		t.Errorf("patched = %+v", entry)
	}
	if !strings.Contains(e.file(), "Alpha body.") {
		t.Error("body lost")
	}

	for _, body := range []map[string]any{
		{},
		{"priority": 0},
		{"status": "maybe"},
		{"due": "soon"},
		{"id": "new"},
		{"tags": "notalist"},
		{"title": ""},
	} {
		r, data := e.do(http.MethodPatch, "/api/tasks/alpha", body)
		wantStatus(t, r, data, 400)
	}
	r, data = e.do(http.MethodPatch, "/api/tasks/zzz", map[string]any{"priority": 1})
	wantStatus(t, r, data, 404)
}

// changedLines returns lines present in b but not in a (as a multiset).
func changedLines(a, b string) []string {
	have := map[string]int{}
	for _, l := range strings.Split(a, "\n") {
		have[l]++
	}
	var out []string
	for _, l := range strings.Split(b, "\n") {
		if have[l] > 0 {
			have[l]--
			continue
		}
		out = append(out, l)
	}
	return out
}

func TestOpen(t *testing.T) {
	e := newEnv(t)
	r, data := e.do(http.MethodPost, "/api/open/alpha", nil)
	wantStatus(t, r, data, 409)
	var body errorBody
	json.Unmarshal(data, &body)
	if body.Command != "clask edit demo/alpha" {
		t.Errorf("command = %q", body.Command)
	}
	t.Setenv("EDITOR", "code")
	r, data = e.do(http.MethodPost, "/api/open/alpha", nil)
	wantStatus(t, r, data, 204)
	if len(e.launched) != 1 || strings.Join(e.launched[0], " ") != "code --goto "+e.taskFile+":5" {
		t.Errorf("launched = %q", e.launched)
	}
	r, data = e.do(http.MethodPost, "/api/open/zzz", nil)
	wantStatus(t, r, data, 404)
}

func TestCheck(t *testing.T) {
	e := newEnv(t)
	var issues []core.Issue
	e.get("/api/check", &issues)
	if len(issues) != 0 {
		t.Errorf("issues = %+v", issues)
	}
	e.do(http.MethodPatch, "/api/tasks/alpha", map[string]any{"blocked_by": []string{"ghost"}})
	e.get("/api/check", &issues)
	if len(issues) != 1 || !strings.Contains(issues[0].Message, "ghost") {
		t.Errorf("issues = %+v", issues)
	}
}

func TestGuard(t *testing.T) {
	e := newEnv(t)
	body := map[string]any{"priority": 1}
	r, data := e.do(http.MethodPatch, "/api/tasks/alpha", body, "Sec-Fetch-Site", "cross-site")
	wantStatus(t, r, data, 403)
	r, data = e.do(http.MethodPatch, "/api/tasks/alpha", body, "Origin", "http://evil.example")
	wantStatus(t, r, data, 403)
	r, data = e.do(http.MethodPatch, "/api/tasks/alpha", body, "Origin", e.srv.URL, "Sec-Fetch-Site", "same-origin")
	wantStatus(t, r, data, 200)
	r, data = e.do(http.MethodPatch, "/api/tasks/alpha", body, "Content-Type", "text/plain")
	wantStatus(t, r, data, 415)
	r, data = e.do(http.MethodDelete, "/api/tasks/alpha", nil)
	wantStatus(t, r, data, 404) // unknown method falls to the JSON catch-all
}

func TestStaticFallback(t *testing.T) {
	e := newEnv(t)
	for _, path := range []string{"/", "/some/route"} {
		r, data := e.do(http.MethodGet, path, nil)
		wantStatus(t, r, data, 200)
		if !strings.Contains(string(data), "UI not built") {
			t.Errorf("%s: %s", path, data)
		}
	}
}
