package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Thenlie/clask/internal/core"
	"github.com/Thenlie/clask/internal/render"
)

// errCheckFailed is returned by check when issues were found; Execute
// prints nothing extra for it since the issues were already listed.
type errCheckFailed struct{ n int }

func (e errCheckFailed) Error() string {
	return fmt.Sprintf("%d issue%s found", e.n, plural(e.n, "", "s"))
}

func newCheckCmd() *cobra.Command {
	var projects []string
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Lint task files",
		Long: `Validate task files: structure, unknown keys, bad values, duplicate ids,
dangling blocked_by references and inconsistent completed dates. Exits 1
when anything is wrong, so it works as a git hook.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := loadConfig()
			if err != nil {
				return err
			}
			sels, err := resolveSelectors(cfg, projects)
			if err != nil {
				return err
			}
			ws, err := core.Load(cfg, sels...)
			if err != nil {
				return err
			}
			issues := ws.Check()
			out := cmd.OutOrStdout()
			if flags.json {
				if issues == nil {
					issues = []core.Issue{}
				}
				if err := render.JSON(out, issues); err != nil {
					return err
				}
			} else {
				for _, i := range issues {
					fmt.Fprintln(out, i)
				}
				if len(issues) == 0 {
					fmt.Fprintf(out, "ok: %d file(s), %d task(s)\n", len(ws.Files), len(ws.Entries()))
				}
			}
			if len(issues) > 0 {
				return errCheckFailed{len(issues)}
			}
			return nil
		},
	}
	cmd.Flags().StringArrayVarP(&projects, "project", "p", nil, "limit to project or project:scope (\".\" = current directory); repeatable")
	return cmd
}
