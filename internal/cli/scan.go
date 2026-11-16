package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/Thenlie/clask/internal/config"
	"github.com/Thenlie/clask/internal/registry"
)

func newScanCmd() *cobra.Command {
	var depth int
	var yes, existing bool
	cmd := &cobra.Command{
		Use:   "scan <dir>...",
		Short: "Find git repositories under directories and register them",
		Long: `Walk each DIR (bounded by --depth, skipping the configured ignore list)
for git repositories and register the ones not already known. Only the
directories you name are walked; clask never scans your whole disk.

With --existing only repositories that already contain a task file at
their root are offered.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, cfgPath, err := loadConfig()
			if err != nil {
				return err
			}
			if depth == 0 {
				depth = cfg.Scan.Depth
			}
			repos, err := registry.FindRepos(args, depth, cfg.Scan.Ignore)
			if err != nil {
				return err
			}
			var candidates []string
			for _, r := range repos {
				if _, ok := registry.ProjectByPath(cfg, r); ok {
					continue
				}
				if existing && !fileExists(filepath.Join(r, cfg.Defaults.File)) {
					continue
				}
				candidates = append(candidates, r)
			}
			out := cmd.OutOrStdout()
			if len(candidates) == 0 {
				fmt.Fprintf(out, "no new repositories found (depth %d)\n", depth)
				return nil
			}
			fmt.Fprintf(out, "found %d unregistered repositor%s:\n", len(candidates), plural(len(candidates), "y", "ies"))
			for _, r := range candidates {
				mark := ""
				if fileExists(filepath.Join(r, cfg.Defaults.File)) {
					mark = "  (has " + cfg.Defaults.File + ")"
				}
				fmt.Fprintf(out, "  %s%s\n", config.CollapseHome(r), mark)
			}
			if !confirm(cmd, yes, "register them?") {
				return nil
			}
			added := 0
			for _, r := range candidates {
				p, err := registry.Register(cfg, r, "")
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "skipped %s: %v (try: clask register %s --name <other>)\n", config.CollapseHome(r), err, config.CollapseHome(r))
					continue
				}
				fmt.Fprintf(out, "registered %s\n", p.Name)
				added++
			}
			if added == 0 {
				return nil
			}
			return cfg.Save(cfgPath)
		},
	}
	cmd.Flags().IntVar(&depth, "depth", 0, "directory levels to descend (default: scan.depth from config)")
	cmd.Flags().BoolVar(&existing, "existing", false, "only repositories that already have a task file")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "register without asking")
	return cmd
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

func newPruneCmd() *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Remove registry entries whose paths no longer exist",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, cfgPath, err := loadConfig()
			if err != nil {
				return err
			}
			gone := registry.Missing(cfg)
			out := cmd.OutOrStdout()
			if len(gone) == 0 {
				fmt.Fprintln(out, "nothing to prune")
				return nil
			}
			for _, g := range gone {
				if g.File == "" {
					fmt.Fprintf(out, "  project %s: %s missing\n", g.Project, g.Path)
				} else {
					fmt.Fprintf(out, "  %s: file %s missing\n", g.Project, g.File)
				}
			}
			if !confirm(cmd, yes, "remove these entries?") {
				return nil
			}
			registry.Prune(cfg)
			if err := cfg.Save(cfgPath); err != nil {
				return err
			}
			fmt.Fprintf(out, "pruned %d entr%s\n", len(gone), plural(len(gone), "y", "ies"))
			return nil
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "remove without asking")
	return cmd
}
