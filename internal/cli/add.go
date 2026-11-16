package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/Thenlie/clask/internal/core"
	"github.com/Thenlie/clask/internal/registry"
	"github.com/Thenlie/clask/internal/render"
	"github.com/Thenlie/clask/internal/taskfile"
)

// currentSelector returns "project[:scope]" for the nearest registered
// task file above the working directory.
func currentSelector() (string, error) {
	cfg, _, err := loadConfig()
	if err != nil {
		return "", err
	}
	sels, err := resolveSelectors(cfg, []string{"."})
	if err != nil {
		return "", err
	}
	return sels[0], nil
}

func newAddCmd() *cobra.Command {
	var (
		project, due, effort, body string
		priority                   int
		tags, links, blockedBy     []string
		edit                       bool
	)
	cmd := &cobra.Command{
		Use:   "add <title>",
		Short: "Add a task",
		Long: `Append a task to a task file. The target is --project (project or
project:scope) or, by default, the nearest task file above the current
directory. The id is derived from the title.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sel := project
			if sel == "" || sel == "." {
				s, err := currentSelector()
				if err != nil {
					return fmt.Errorf("%w; pass --project", err)
				}
				sel = s
			}
			nt := core.NewTask{Title: args[0], Priority: priority, Tags: tags, Effort: effort, Links: links, BlockedBy: blockedBy, Body: body}
			if due != "" {
				d, err := taskfile.ParseDate(due)
				if err != nil {
					return err
				}
				nt.Due = &d
			}
			ws, err := loadWorkspace([]string{sel})
			if err != nil {
				return err
			}
			e, err := ws.Add(sel, nt)
			if err != nil {
				return err
			}
			if flags.json {
				return render.JSON(cmd.OutOrStdout(), e)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "added %s\n", e.Ref())
			if edit {
				return openEditor(e.Path, e.Line)
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVarP(&project, "project", "p", "", "target project or project:scope (default: nearest task file)")
	f.IntVar(&priority, "priority", 0, "priority 1-5 (default: file or config default)")
	f.StringArrayVarP(&tags, "tag", "t", nil, "tag; repeatable")
	f.StringVar(&due, "due", "", "due date YYYY-MM-DD")
	f.StringVar(&effort, "effort", "", "free-form effort, e.g. s, m, l, 2h")
	f.StringArrayVar(&links, "link", nil, "related URL; repeatable")
	f.StringArrayVar(&blockedBy, "blocked-by", nil, "id of a blocking task; repeatable")
	f.StringVar(&body, "body", "", "markdown body")
	f.BoolVar(&edit, "edit", false, "open the new task in $EDITOR")
	return cmd
}

// openEditor runs the user's editor on path at line, attached to the
// terminal.
func openEditor(path string, line int) error {
	argv := core.EditorCommand(path, line)
	c := exec.Command(argv[0], argv[1:]...)
	c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("editor %q: %w", argv[0], err)
	}
	return nil
}

func newStatusCmd(use, short string, status taskfile.Status) *cobra.Command {
	return &cobra.Command{
		Use:   use + " <ref>...",
		Short: short,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ws, err := loadWorkspace(nil)
			if err != nil {
				return err
			}
			var out []core.Entry
			for _, ref := range args {
				e, err := ws.SetStatus(ref, status)
				if err != nil {
					return err
				}
				out = append(out, e)
			}
			if flags.json {
				return render.JSON(cmd.OutOrStdout(), out)
			}
			for _, e := range out {
				fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", e.Ref(), e.Status)
			}
			return nil
		},
	}
}

func newSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <ref> <key=value>...",
		Short: "Change task fields",
		Long: `Set one or more fields on a task.

Keys: title, id, priority, status, due, created, completed, effort,
tags, blocked_by, links. Dates are YYYY-MM-DD; an empty value clears
them. List keys take comma-separated values and support += and -=:

  clask set fix-sync priority=1 due=2026-12-01
  clask set fix-sync tags+=infra tags-=old
  clask set fix-sync links=https://a,https://b`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ws, err := loadWorkspace(nil)
			if err != nil {
				return err
			}
			e, err := ws.Update(args[0], func(t *taskfile.Task) error {
				return core.Assign(t, args[1:])
			})
			if err != nil {
				return err
			}
			if flags.json {
				return render.JSON(cmd.OutOrStdout(), e)
			}
			return render.TaskDetail(cmd.OutOrStdout(), e, taskfile.Today())
		},
	}
}

func newEditCmd() *cobra.Command {
	var project string
	cmd := &cobra.Command{
		Use:   "edit [ref]",
		Short: "Open a task file in $EDITOR",
		Long: `Open a task file in your editor. With a task ref the editor jumps to that
task's block. With --project (or neither) the nearest task file above the
current directory is opened at the top.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				ws, err := loadWorkspace(nil)
				if err != nil {
					return err
				}
				e, err := ws.Get(args[0])
				if err != nil {
					return err
				}
				return openEditor(e.Path, e.Line)
			}
			sel := project
			if sel == "" || sel == "." {
				s, err := currentSelector()
				if err != nil {
					return fmt.Errorf("%w; pass a ref or --project", err)
				}
				sel = s
			}
			cfg, _, err := loadConfig()
			if err != nil {
				return err
			}
			p, f, err := registry.ResolveScope(cfg, sel)
			if err != nil {
				return err
			}
			if f == nil {
				switch len(p.Files) {
				case 0:
					return fmt.Errorf("project %q has no task files; run \"clask init\"", p.Name)
				case 1:
					f = &p.Files[0]
				default:
					return fmt.Errorf("project %q has several task files; pass project:scope", p.Name)
				}
			}
			path, err := p.FileAbsPath(*f)
			if err != nil {
				return err
			}
			return openEditor(path, 0)
		},
	}
	cmd.Flags().StringVarP(&project, "project", "p", "", "project or project:scope to open")
	return cmd
}
