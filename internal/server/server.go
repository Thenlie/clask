// Package server exposes a core.Workspace over HTTP for the local web UI.
// Every request reloads config and task files, so the page always reflects
// what the CLI or an editor wrote last; nothing is cached in memory.
package server

import (
	"encoding/json"
	"mime"
	"net/http"
	"net/url"
	"os/exec"

	"github.com/Thenlie/clask/internal/config"
	"github.com/Thenlie/clask/internal/core"
)

// Options configures a Server.
type Options struct {
	ConfigPath string // registry file; "" means the XDG default
	Version    string
	// Launch starts an editor detached from the server process. nil uses
	// exec with no attached stdio.
	Launch func(argv []string) error
}

// Server is the HTTP API plus the embedded frontend.
type Server struct {
	opts Options
	mux  *http.ServeMux
}

// New builds a Server. Call Handler to mount it.
func New(opts Options) *Server {
	if opts.Launch == nil {
		opts.Launch = launchDetached
	}
	s := &Server{opts: opts, mux: http.NewServeMux()}
	api := http.NewServeMux()
	api.HandleFunc("GET /api/meta", s.handleMeta)
	api.HandleFunc("GET /api/projects", s.handleProjects)
	api.HandleFunc("GET /api/tasks", s.handleListTasks)
	api.HandleFunc("POST /api/tasks", s.handleCreateTask)
	api.HandleFunc("GET /api/tasks/{ref...}", s.handleGetTask)
	api.HandleFunc("PATCH /api/tasks/{ref...}", s.handlePatchTask)
	api.HandleFunc("POST /api/open/{ref...}", s.handleOpen)
	api.HandleFunc("GET /api/check", s.handleCheck)
	api.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, &httpError{Status: http.StatusNotFound, Message: "no such endpoint"})
	})
	s.mux.Handle("/api/", guard(api))
	s.mux.Handle("/", Static())
	return s
}

// Handler returns the root handler.
func (s *Server) Handler() http.Handler { return s.mux }

// load reads the registry and the selected task files.
func (s *Server) load(selectors ...string) (*core.Workspace, error) {
	path := s.opts.ConfigPath
	if path == "" {
		var err error
		if path, err = config.DefaultPath(); err != nil {
			return nil, &httpError{Status: http.StatusInternalServerError, Message: err.Error()}
		}
	}
	cfg, err := config.Load(path)
	if err != nil {
		return nil, &httpError{Status: http.StatusInternalServerError, Message: err.Error()}
	}
	return core.Load(cfg, selectors...)
}

// guard rejects mutating requests that a browser could send cross-site.
// The server only listens on loopback, but any page the user has open
// could still POST to it; same-origin pages send a JSON content type and
// never carry a cross-site fetch marker.
func guard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			next.ServeHTTP(w, r)
			return
		}
		if r.Header.Get("Sec-Fetch-Site") == "cross-site" {
			writeError(w, &httpError{Status: http.StatusForbidden, Message: "cross-site requests are not allowed"})
			return
		}
		if origin := r.Header.Get("Origin"); origin != "" {
			u, err := url.Parse(origin)
			if err != nil || u.Host != r.Host {
				writeError(w, &httpError{Status: http.StatusForbidden, Message: "origin does not match"})
				return
			}
		}
		mt, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if mt != "application/json" {
			writeError(w, &httpError{Status: http.StatusUnsupportedMediaType, Message: "Content-Type must be application/json"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// decodeJSON reads a request body strictly: unknown fields are errors.
func decodeJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return &httpError{Status: http.StatusBadRequest, Message: "invalid JSON body: " + err.Error()}
	}
	return nil
}

func launchDetached(argv []string) error {
	cmd := exec.Command(argv[0], argv[1:]...)
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}
