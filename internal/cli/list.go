package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Thenlie/clask/internal/core"
	"github.com/Thenlie/clask/internal/render"
	"github.com/Thenlie/clask/internal/taskfile"
)

type listFlags struct {
	projects  []string
	statuses  []string
	tags      []string
	priority  int
	dueBefore string
	overdue   bool
	all       bool
	sortKey   string
	format    string
}

func (lf *listFlags) bind(cmd *cobra.Command) {
	f := cmd.Flags()
	f.StringArrayVarP(&lf.projects, "project", "p", nil, "limit to project or project:scope (\".\" = current directory); repeatable")
	f.StringArrayVarP(&lf.statuses, "status", "s", nil, "limit to status (todo, in-progress, blocked, done, dropped); repeatable")
	f.StringArrayVarP(&lf.tags, "tag", "t", nil, "require tag; repeatable, all must match")
	f.IntVar(&lf.priority, "priority", 0, "limit to priority 1-5")
	f.StringVar(&lf.dueBefore, "due-before", "", "only tasks due before YYYY-MM-DD")
	f.BoolVar(&lf.overdue, "overdue", false, "only open tasks past their due date")
	f.BoolVarP(&lf.all, "all", "a", false, "include done and dropped tasks")
	f.StringVar(&lf.sortKey, "sort", "priority", "sort by priority, due, created or project")
	f.StringVarP(&lf.format, "format", "o", "table", "output format: table, tsv, csv, jsonl (or --json)")
}

func (lf *listFlags) filter() (core.Filter, core.SortKey, error) {
	var fl core.Filter
	for _, s := range lf.statuses {
		st, err := taskfile.ParseStatus(s)
		if err != nil {
			return fl, "", err
		}
		fl.Statuses = append(fl.Statuses, st)
	}
	fl.Tags = lf.tags
	if lf.priority != 0 && (lf.priority < 1 || lf.priority > 5) {
		return fl, "", fmt.Errorf("--priority must be 1-5")
	}
	fl.Priority = lf.priority
	if lf.dueBefore != "" {
		d, err := taskfile.ParseDate(lf.dueBefore)
		if err != nil {
			return fl, "", err
		}
		fl.DueBefore = &d
	}
	fl.Overdue = lf.overdue
	fl.IncludeClosed = lf.all
	fl.Today = taskfile.Today()
	key, err := core.ParseSortKey(lf.sortKey)
	return fl, key, err
}

type listOutput struct {
	Tasks    []core.Entry   `json:"tasks"`
	Warnings []core.Warning `json:"warnings"`
}

func runList(cmd *cobra.Command, lf *listFlags, limit int) error {
	fl, key, err := lf.filter()
	if err != nil {
		return err
	}
	ws, err := loadWorkspace(lf.projects)
	if err != nil {
		return err
	}
	entries := ws.List(fl, key)
	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}
	out := cmd.OutOrStdout()
	if entries == nil {
		entries = []core.Entry{}
	}
	if flags.json {
		warns := ws.Warnings
		if warns == nil {
			warns = []core.Warning{}
		}
		return render.JSON(out, listOutput{Tasks: entries, Warnings: warns})
	}
	switch lf.format {
	case "table":
		return render.TaskTable(out, entries, fl.Today)
	case "tsv":
		return render.Delimited(out, '\t', render.TaskHeaders, render.TaskRows(entries))
	case "csv":
		return render.Delimited(out, ',', render.TaskHeaders, render.TaskRows(entries))
	case "jsonl":
		return render.JSONLines(out, entries)
	}
	return fmt.Errorf("invalid format %q (want table, tsv, csv or jsonl)", lf.format)
}

func newListCmd() *cobra.Command {
	lf := &listFlags{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks across registered projects",
		Long: `List open tasks from every registered task file, highest priority first.
Done and dropped tasks are hidden unless --all or an explicit --status is given.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, lf, 0)
		},
	}
	lf.bind(cmd)
	return cmd
}

func newNextCmd() *cobra.Command {
	lf := &listFlags{}
	cmd := &cobra.Command{
		Use:   "next",
		Short: "Show the single highest-priority open task",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, lf, 1)
		},
	}
	lf.bind(cmd)
	return cmd
}
