package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Thenlie/clask/internal/registry"
)

func newRegisterCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "register [path]",
		Short: "Register a project directory",
		Long: `Register the current directory (or PATH) as a project. If the directory is
inside a git repository the repository root is registered. A TASKS.md at the
root is added automatically; use "clask init" to create one.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) == 1 {
				dir = args[0]
			}
			cfg, path, err := loadConfig()
			if err != nil {
				return err
			}
			p, err := registry.Register(cfg, dir, name)
			if err != nil {
				return err
			}
			if err := cfg.Save(path); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "registered %s at %s", p.Name, p.Path)
			if len(p.Files) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), " (no %s yet; run \"clask init\")", cfg.Defaults.File)
			}
			fmt.Fprintln(cmd.OutOrStdout())
			return nil
		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "", "project name (default: directory basename)")
	return cmd
}

func newUnregisterCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unregister <project[:scope]>",
		Short: "Remove a project, or one of its task files, from the registry",
		Long: `Remove a project ("falcon") or a single task file ("falcon:app") from the
registry. Files on disk are never modified.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, path, err := loadConfig()
			if err != nil {
				return err
			}
			if err := registry.Unregister(cfg, args[0]); err != nil {
				return err
			}
			if err := cfg.Save(path); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "unregistered %s\n", args[0])
			return nil
		},
	}
}

// fileExists reports whether path exists on disk.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
