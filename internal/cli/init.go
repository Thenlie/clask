package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/Thenlie/clask/internal/config"
	"github.com/Thenlie/clask/internal/registry"
)

// taskFileTemplate renders a fresh task file: a file block carrying the
// scope name, followed by a note that the parser ignores.
func taskFileTemplate(name string) []byte {
	return []byte(fmt.Sprintf(`---
name: %s
---
Tasks for %s. Add one with: clask add "Title"
Format: https://github.com/Thenlie/clask#format
`, name, name))
}

func newInitCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "init [path]",
		Short: "Create a task file here and register it",
		Long: `Create a TASKS.md in the current directory (or PATH) and add it to the
registry. Inside a registered project the file becomes a new scope of that
project; otherwise the directory (or its git root) is registered as a new
project. Use --name to set the scope or project name explicitly.`,
		Example: `  clask init                   # new project (or new scope) for the current directory
  clask init packages/app      # add a scope inside a registered monorepo
  clask init --name career     # set the name explicitly`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) == 1 {
				dir = args[0]
			}
			abs, err := config.ExpandHome(dir)
			if err != nil {
				return err
			}
			if err := os.MkdirAll(abs, 0o755); err != nil {
				return err
			}
			cfg, cfgPath, err := loadConfig()
			if err != nil {
				return err
			}
			target := filepath.Join(abs, cfg.Defaults.File)
			if fileExists(target) {
				return fmt.Errorf("%s already exists; use \"clask register\" or \"clask discover\" to add it", target)
			}

			var label, scope string
			if p, _, ok := registry.Context(cfg, abs); ok {
				root, _ := p.AbsPath()
				rel, _ := filepath.Rel(root, target)
				f, err := registry.AddFile(cfg, p, rel, name)
				if err != nil {
					return err
				}
				scope = registry.Scope(*p, *f)
				label = p.Name + ":" + scope
			} else {
				p, err := registry.Register(cfg, abs, name)
				if err != nil {
					return err
				}
				root, _ := p.AbsPath()
				rel, _ := filepath.Rel(root, target)
				if len(p.Files) == 0 {
					if _, err := registry.AddFile(cfg, p, rel, ""); err != nil {
						return err
					}
				}
				scope = p.Name
				label = p.Name
			}
			if err := os.WriteFile(target, taskFileTemplate(scope), 0o644); err != nil {
				return fmt.Errorf("create %s: %w", target, err)
			}
			if err := cfg.Save(cfgPath); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "created %s (%s)\n", config.CollapseHome(target), label)
			return nil
		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "", "scope name (inside a project) or project name (new project)")
	return cmd
}
