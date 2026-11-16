package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/Thenlie/clask/internal/config"
	"github.com/Thenlie/clask/internal/registry"
)

func newDiscoverCmd() *cobra.Command {
	var depth int
	var yes bool
	cmd := &cobra.Command{
		Use:   "discover [project]",
		Short: "Find unregistered task files inside a project",
		Long: `Walk a registered project's directory (bounded by --depth, skipping the
configured ignore list) for task files not yet in the registry, then add
them. Defaults to the project containing the current directory.`,
		Example: `  clask discover               # project containing the current directory
  clask discover falcon --depth 4
  clask discover falcon -y     # add without asking`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, cfgPath, err := loadConfig()
			if err != nil {
				return err
			}
			var p *config.Project
			if len(args) == 1 {
				var ok bool
				if p, ok = cfg.Project(args[0]); !ok {
					return fmt.Errorf("project %q: %w", args[0], registry.ErrNotFound)
				}
			} else {
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}
				var ok bool
				if p, _, ok = registry.Context(cfg, cwd); !ok {
					return fmt.Errorf("current directory is not inside a registered project; pass a project name")
				}
			}
			root, err := p.AbsPath()
			if err != nil {
				return err
			}
			if depth == 0 {
				depth = cfg.Scan.Depth
			}
			found, err := registry.FindTaskFiles(root, cfg.Defaults.File, depth, cfg.Scan.Ignore)
			if err != nil {
				return err
			}
			registered := map[string]bool{}
			for _, f := range p.Files {
				registered[filepath.Clean(f.Path)] = true
			}
			var missing []string
			for _, rel := range found {
				if !registered[filepath.Clean(rel)] {
					missing = append(missing, rel)
				}
			}
			out := cmd.OutOrStdout()
			if len(missing) == 0 {
				fmt.Fprintf(out, "no new %s files under %s (depth %d)\n", cfg.Defaults.File, p.Path, depth)
				return nil
			}
			fmt.Fprintf(out, "found %d unregistered file(s) in %s:\n", len(missing), p.Name)
			for _, rel := range missing {
				fmt.Fprintf(out, "  %s\n", rel)
			}
			if !confirm(cmd, yes, "add them?") {
				return nil
			}
			added := 0
			for _, rel := range missing {
				f, err := registry.AddFile(cfg, p, rel, "")
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "skipped %s: %v\n", rel, err)
					continue
				}
				fmt.Fprintf(out, "added %s as %s:%s\n", rel, p.Name, registry.Scope(*p, *f))
				added++
			}
			if added == 0 {
				return nil
			}
			return cfg.Save(cfgPath)
		},
	}
	cmd.Flags().IntVar(&depth, "depth", 0, "directory levels to descend (default: scan.depth from config)")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "add without asking")
	return cmd
}
