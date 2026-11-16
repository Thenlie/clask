package cli

import (
	"github.com/spf13/cobra"

	"github.com/Thenlie/clask/internal/render"
	"github.com/Thenlie/clask/internal/taskfile"
)

func newShowCmd() *cobra.Command {
	var md bool
	cmd := &cobra.Command{
		Use:   "show <ref>",
		Short: "Show one task in full",
		Long: `Show every field and the markdown body of one task.

A ref is "[project[:scope]/]id". The id may be a unique prefix:
  clask show fix-sync
  clask show falcon/fix-sync-race
  clask show falcon:app/fix`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ws, err := loadWorkspace(nil)
			if err != nil {
				return err
			}
			e, err := ws.Get(args[0])
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			switch {
			case flags.json:
				return render.JSON(out, e)
			case md:
				return render.TaskMarkdown(out, e)
			}
			return render.TaskDetail(out, e, taskfile.Today())
		},
	}
	cmd.Flags().BoolVar(&md, "md", false, "print as markdown (frontmatter + body) for piping to a renderer")
	return cmd
}
