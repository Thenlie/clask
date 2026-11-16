package cli

import (
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Thenlie/clask/internal/core"
	"github.com/Thenlie/clask/internal/registry"
	"github.com/Thenlie/clask/internal/render"
)

type projectRow struct {
	Project  string   `json:"project"`
	Scope    string   `json:"scope,omitempty"`
	Path     string   `json:"path"`
	Files    int      `json:"files,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

func newProjectsCmd() *cobra.Command {
	var showFiles bool
	cmd := &cobra.Command{
		Use:   "projects",
		Short: "List registered projects",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := loadConfig()
			if err != nil {
				return err
			}
			// Effective scopes may come from each file's own name, so load them.
			scopes := map[string]string{}
			if showFiles {
				ws, err := core.Load(cfg)
				if err != nil {
					return err
				}
				for _, lf := range ws.Files {
					scopes[lf.Path] = lf.Scope
				}
			}
			var rows []projectRow
			for _, p := range cfg.Projects {
				root, err := p.AbsPath()
				var warn []string
				if err != nil {
					warn = append(warn, err.Error())
				} else if !fileExists(root) {
					warn = append(warn, "path missing")
				}
				if !showFiles {
					for _, f := range p.Files {
						if fp, err := p.FileAbsPath(f); err == nil && !fileExists(fp) {
							warn = append(warn, "missing "+f.Path)
						}
					}
					rows = append(rows, projectRow{Project: p.Name, Path: p.Path, Files: len(p.Files), Warnings: warn})
					continue
				}
				if len(p.Files) == 0 {
					rows = append(rows, projectRow{Project: p.Name, Path: p.Path, Warnings: append(warn, "no files")})
				}
				for _, f := range p.Files {
					fw := append([]string(nil), warn...)
					scope := registry.Scope(p, f)
					if fp, err := p.FileAbsPath(f); err == nil {
						if !fileExists(fp) {
							fw = append(fw, "file missing")
						} else if s, ok := scopes[fp]; ok {
							scope = s
						}
					}
					rows = append(rows, projectRow{Project: p.Name, Scope: scope, Path: f.Path, Warnings: fw})
				}
			}
			out := cmd.OutOrStdout()
			if flags.json {
				if rows == nil {
					rows = []projectRow{}
				}
				return render.JSON(out, rows)
			}
			if len(rows) == 0 {
				_, err := out.Write([]byte("no projects registered; run \"clask register\" or \"clask scan <dir>\"\n"))
				return err
			}
			var headers []render.Column
			var table [][]string
			if showFiles {
				headers = render.Headers("PROJECT", "SCOPE", "FILE", "WARNINGS")
				for _, r := range rows {
					table = append(table, []string{r.Project, r.Scope, r.Path, strings.Join(r.Warnings, "; ")})
				}
			} else {
				headers = render.Headers("PROJECT", "PATH", "FILES", "WARNINGS")
				for _, r := range rows {
					table = append(table, []string{r.Project, r.Path, strconv.Itoa(r.Files), strings.Join(r.Warnings, "; ")})
				}
			}
			return render.Table(out, headers, table)
		},
	}
	cmd.Flags().BoolVar(&showFiles, "files", false, "one row per task file with its scope")
	return cmd
}
