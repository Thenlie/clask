package cli

import (
	"fmt"
	"os"

	"github.com/Thenlie/clask/internal/config"
	"github.com/Thenlie/clask/internal/core"
	"github.com/Thenlie/clask/internal/registry"
	"github.com/Thenlie/clask/internal/render"
)

// resolveSelectors expands "." into the current project (and nearest file)
// and passes other selectors through unchanged.
func resolveSelectors(cfg *config.Config, sels []string) ([]string, error) {
	var out []string
	for _, s := range sels {
		if s != "." {
			out = append(out, s)
			continue
		}
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		p, f, ok := registry.Context(cfg, cwd)
		if !ok {
			return nil, fmt.Errorf("current directory is not inside a registered project")
		}
		if f == nil {
			out = append(out, p.Name)
		} else {
			out = append(out, p.Name+":"+registry.Scope(*p, *f))
		}
	}
	return out, nil
}

// loadWorkspace loads config and the selected files, printing warnings to
// stderr unless JSON output was requested (callers embed them there).
func loadWorkspace(sels []string) (*core.Workspace, error) {
	cfg, _, err := loadConfig()
	if err != nil {
		return nil, err
	}
	sels, err = resolveSelectors(cfg, sels)
	if err != nil {
		return nil, err
	}
	ws, err := core.Load(cfg, sels...)
	if err != nil {
		return nil, err
	}
	if !flags.json {
		render.Warnings(os.Stderr, ws.Warnings)
	}
	return ws, nil
}
