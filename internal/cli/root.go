// Package cli wires cobra commands to the core workspace.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Thenlie/clask/internal/render"
	"github.com/Thenlie/clask/internal/taskfile"
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
	defaultList := &listFlags{sortKey: "priority", format: "table"}
	root.RunE = func(cmd *cobra.Command, args []string) error {
		return runList(cmd, defaultList, 0)
	}

	pf := root.PersistentFlags()
	pf.BoolVar(&flags.json, "json", false, "emit JSON instead of a table")
	pf.BoolVar(&flags.noColor, "no-color", false, "disable colored output")
	pf.StringVar(&flags.configPath, "config", "", "path to config file (default $XDG_CONFIG_HOME/clask/config.yaml)")

	root.AddGroup(
		&cobra.Group{ID: "tasks", Title: "Task commands:"},
		&cobra.Group{ID: "registry", Title: "Registry commands:"},
	)
	root.AddCommand(
		newListCmd(),
		newShowCmd(),
		newNextCmd(),
		newAddCmd(),
		newStatusCmd("start", "Mark tasks in-progress", taskfile.StatusInProgress),
		newStatusCmd("block", "Mark tasks blocked", taskfile.StatusBlocked),
		newStatusCmd("done", "Mark tasks done", taskfile.StatusDone),
		newStatusCmd("drop", "Mark tasks dropped", taskfile.StatusDropped),
		newStatusCmd("reopen", "Reopen tasks (status todo, clears completed)", taskfile.StatusTodo),
		newSetCmd(),
		newEditCmd(),
		newCheckCmd(),
		newInitCmd(),
		newRegisterCmd(),
		newUnregisterCmd(),
		newDiscoverCmd(),
		newScanCmd(),
		newProjectsCmd(),
		newPruneCmd(),
		newVersionCmd(),
	)
	for _, c := range root.Commands() {
		switch c.Name() {
		case "list", "show", "next", "add", "start", "block", "done", "drop", "reopen", "set", "edit", "check":
			c.GroupID = "tasks"
		case "init", "register", "unregister", "discover", "scan", "projects", "prune":
			c.GroupID = "registry"
		}
	}
	root.SetHelpCommandGroupID("")
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
		if _, ok := err.(errCheckFailed); !ok {
			fmt.Fprintln(os.Stderr, "clask:", err)
		}
		return 1
	}
	return 0
}
