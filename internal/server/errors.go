package server

import (
	"errors"
	"io/fs"
	"net/http"

	"github.com/Thenlie/clask/internal/core"
	"github.com/Thenlie/clask/internal/registry"
)

// httpError carries an explicit status. Command, when set, is a CLI
// invocation the UI can offer instead (e.g. when the editor cannot be
// launched from a server).
type httpError struct {
	Status  int
	Message string
	Command string
}

func (e *httpError) Error() string { return e.Message }

type errorBody struct {
	Error   string `json:"error"`
	Command string `json:"command,omitempty"`
}

// statusOf maps core and registry errors onto HTTP statuses. Anything
// unrecognised is treated as a bad request because core only returns
// errors for invalid input or missing targets; filesystem failures become
// 500s.
func statusOf(err error) int {
	var he *httpError
	if errors.As(err, &he) {
		return he.Status
	}
	var pe *fs.PathError
	switch {
	case errors.Is(err, registry.ErrNotFound), errors.Is(err, core.ErrNoTask):
		return http.StatusNotFound
	case errors.Is(err, core.ErrAmbiguous):
		return http.StatusConflict
	case errors.As(err, &pe):
		return http.StatusInternalServerError
	}
	return http.StatusBadRequest
}

func writeError(w http.ResponseWriter, err error) {
	body := errorBody{Error: err.Error()}
	var he *httpError
	if errors.As(err, &he) {
		body.Command = he.Command
	}
	writeJSON(w, statusOf(err), body)
}
