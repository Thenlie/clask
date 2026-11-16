package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
	"gopkg.in/yaml.v3"

	"github.com/Thenlie/clask/internal/core"
	"github.com/Thenlie/clask/internal/taskfile"
)

var (
	dim      = color.New(color.Faint)
	bold     = color.New(color.Bold)
	red      = color.New(color.FgRed)
	yellow   = color.New(color.FgYellow)
	green    = color.New(color.FgGreen)
	magenta  = color.New(color.FgMagenta)
	cyan     = color.New(color.FgCyan)
	redBold  = color.New(color.FgRed, color.Bold)
	noStyle  = color.New()
	priority = map[int]*color.Color{1: redBold, 2: yellow}
)

// DisableColor turns off ANSI styling for the process.
func DisableColor() { color.NoColor = true }

func statusStyle(s taskfile.Status) *color.Color {
	switch s {
	case taskfile.StatusInProgress:
		return yellow
	case taskfile.StatusBlocked:
		return magenta
	case taskfile.StatusDone:
		return green
	case taskfile.StatusDropped:
		return dim
	}
	return noStyle
}

func priorityStyle(p int) *color.Color {
	if c, ok := priority[p]; ok {
		return c
	}
	return noStyle
}

// TaskTable prints entries as the standard list view.
func TaskTable(w io.Writer, entries []core.Entry, today taskfile.Date) error {
	if len(entries) == 0 {
		_, err := fmt.Fprintln(w, dim.Sprint("no tasks"))
		return err
	}
	rows := make([][]string, 0, len(entries))
	for _, e := range entries {
		due := ""
		if e.Due != nil {
			due = e.Due.String()
			if e.Overdue(today) {
				due = red.Sprint(due)
			}
		}
		rows = append(rows, []string{
			cyan.Sprint(e.Selector()),
			e.ID,
			priorityStyle(e.Priority).Sprint(e.Priority),
			statusStyle(e.Status).Sprint(string(e.Status)),
			due,
			e.Title,
		})
	}
	return Table(w, taskColumns(), rows)
}

// TaskHeaders are the column names shared by the table and delimited formats.
var TaskHeaders = []string{"project", "id", "priority", "status", "due", "title"}

func taskColumns() []Column {
	return []Column{{Header: "PROJECT"}, {Header: "ID"}, {Header: "P", Right: true}, {Header: "STATUS"}, {Header: "DUE"}, {Header: "TITLE"}}
}

// TaskRows renders entries as plain (uncolored) cells for csv/tsv output.
func TaskRows(entries []core.Entry) [][]string {
	rows := make([][]string, 0, len(entries))
	for _, e := range entries {
		due := ""
		if e.Due != nil {
			due = e.Due.String()
		}
		rows = append(rows, []string{e.Selector(), e.ID, fmt.Sprint(e.Priority), string(e.Status), due, e.Title})
	}
	return rows
}

// TaskDetail prints one task with every field and its body.
func TaskDetail(w io.Writer, e core.Entry, today taskfile.Date) error {
	var b strings.Builder
	fmt.Fprintln(&b, cyan.Sprint(e.Ref()))
	fmt.Fprintln(&b, bold.Sprint(e.Title))
	fmt.Fprintln(&b)
	field := func(k, v string) {
		if v != "" {
			fmt.Fprintf(&b, "%-11s %s\n", dim.Sprint(k), v)
		}
	}
	field("priority", priorityStyle(e.Priority).Sprint(e.Priority))
	field("status", statusStyle(e.Status).Sprint(string(e.Status)))
	field("tags", strings.Join(e.Tags, ", "))
	if e.Due != nil {
		d := e.Due.String()
		if e.Overdue(today) {
			d = red.Sprint(d + " (overdue)")
		}
		field("due", d)
	}
	if e.Created != nil {
		field("created", e.Created.String())
	}
	if e.Completed != nil {
		field("completed", e.Completed.String())
	}
	field("effort", e.Effort)
	field("blocked_by", strings.Join(e.BlockedBy, ", "))
	field("links", strings.Join(e.Links, "\n            "))
	field("file", fmt.Sprintf("%s:%d", e.Path, e.Line))
	if e.Body != "" {
		fmt.Fprintln(&b)
		fmt.Fprintln(&b, e.Body)
	}
	_, err := io.WriteString(w, b.String())
	return err
}

// TaskMarkdown prints a task as its own frontmatter block plus body, suitable
// for piping to a markdown renderer.
func TaskMarkdown(w io.Writer, e core.Entry) error {
	y, err := yaml.Marshal(e.Task)
	if err != nil {
		return err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "---\n%s---\n", y)
	fmt.Fprintf(&b, "# %s\n", e.Title)
	if e.Body != "" {
		fmt.Fprintf(&b, "\n%s\n", e.Body)
	}
	_, err = io.WriteString(w, b.String())
	return err
}

// Warnings prints load warnings, one per line.
func Warnings(w io.Writer, ws []core.Warning) {
	for _, warn := range ws {
		fmt.Fprintln(w, yellow.Sprint("warning:"), warn.String())
	}
}
