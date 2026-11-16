// Package cli wires cobra commands to the core workspace.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Thenlie/clask/internal/render"
)

// version is set at build time via -ldflags.
var version = "dev"

type globalFlags struct {
	json       bool
	noColor    bool
	configPath string
}

var flags globalFlags

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "clask",
		Short: "Cross-repo task lists from TASKS.md files",
		Long: `clask aggregates TASKS.md files from every registered project on this
machine into one view. Tasks live in the repo next to the code they describe.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	// Bare "clask" behaves like "clask list" with default flags.
	defaultList := &listFlags{sortKey: "priority"}
	root.RunE = func(cmd *cobra.Command, args []string) error {
		return runList(cmd, defaultList, 0)
	}

	pf := root.PersistentFlags()
	pf.BoolVar(&flags.json, "json", false, "emit JSON instead of a table")
	pf.BoolVar(&flags.noColor, "no-color", false, "disable colored output")
	pf.StringVar(&flags.configPath, "config", "", "path to config file (default $XDG_CONFIG_HOME/clask/config.yaml)")

	root.AddCommand(
		newListCmd(),
		newShowCmd(),
		newNextCmd(),
		newInitCmd(),
		newRegisterCmd(),
		newUnregisterCmd(),
		newDiscoverCmd(),
		newScanCmd(),
		newProjectsCmd(),
		newPruneCmd(),
		newVersionCmd(),
	)
	return root
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the clask version",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), "clask", version)
		},
	}
}

// Execute runs the root command and returns a process exit code.
func Execute() int {
	root := newRootCmd()
	cobra.OnInitialize(func() {
		if flags.noColor {
			render.DisableColor()
		}
	})
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "clask:", err)
		return 1
	}
	return 0
}
