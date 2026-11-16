// Package taskfile defines the TASKS.md format: a YAML multi-document stream
// where the first block describes the file and every later block is a task
// followed by an optional markdown body.
package taskfile

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Status is the lifecycle state of a task.
type Status string

// Valid statuses, in display order.
const (
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in-progress"
	StatusBlocked    Status = "blocked"
	StatusDone       Status = "done"
	StatusDropped    Status = "dropped"
)

// Statuses lists every valid status in display order.
var Statuses = []Status{StatusTodo, StatusInProgress, StatusBlocked, StatusDone, StatusDropped}

// Valid reports whether s is a known status.
func (s Status) Valid() bool {
	for _, v := range Statuses {
		if s == v {
			return true
		}
	}
	return false
}

// Closed reports whether the task needs no further work.
func (s Status) Closed() bool { return s == StatusDone || s == StatusDropped }

// ParseStatus accepts a status string, returning an error listing valid values.
func ParseStatus(s string) (Status, error) {
	st := Status(s)
	if !st.Valid() {
		return "", fmt.Errorf("invalid status %q (want one of %s)", s, joinStatuses())
	}
	return st, nil
}

func joinStatuses() string {
	parts := make([]string, len(Statuses))
	for i, s := range Statuses {
		parts[i] = string(s)
	}
	return strings.Join(parts, ", ")
}

// Date is a calendar day serialized as YYYY-MM-DD in both YAML and JSON.
type Date struct{ time.Time }

// DateLayout is the only accepted date format.
const DateLayout = time.DateOnly

// ParseDate parses YYYY-MM-DD.
func ParseDate(s string) (Date, error) {
	t, err := time.Parse(DateLayout, s)
	if err != nil {
		return Date{}, fmt.Errorf("invalid date %q (want YYYY-MM-DD)", s)
	}
	return Date{t}, nil
}

// Today returns the current local date.
func Today() Date {
	y, m, d := time.Now().Date()
	return Date{time.Date(y, m, d, 0, 0, 0, 0, time.UTC)}
}

func (d Date) String() string { return d.Format(DateLayout) }

// Before reports whether d is strictly earlier than o.
func (d Date) Before(o Date) bool { return d.Time.Before(o.Time) }

// UnmarshalYAML implements yaml.Unmarshaler.
func (d *Date) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind != yaml.ScalarNode {
		return fmt.Errorf("line %d: date must be a scalar", n.Line)
	}
	p, err := ParseDate(n.Value)
	if err != nil {
		return fmt.Errorf("line %d: %w", n.Line, err)
	}
	*d = p
	return nil
}

// MarshalYAML implements yaml.Marshaler.
func (d Date) MarshalYAML() (any, error) { return d.String(), nil }

// MarshalJSON implements json.Marshaler.
func (d Date) MarshalJSON() ([]byte, error) { return json.Marshal(d.String()) }

// UnmarshalJSON implements json.Unmarshaler.
func (d *Date) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	p, err := ParseDate(s)
	if err != nil {
		return err
	}
	*d = p
	return nil
}

// FileMeta is the first block of a TASKS.md.
type FileMeta struct {
	Name            string `yaml:"name,omitempty" json:"name,omitempty"`
	DefaultPriority int    `yaml:"default_priority,omitempty" json:"default_priority,omitempty"`
}

// Task is one task block plus its markdown body. Field order here is the
// order the writer emits keys.
type Task struct {
	ID        string   `yaml:"id" json:"id"`
	Title     string   `yaml:"title" json:"title"`
	Priority  int      `yaml:"priority" json:"priority"`
	Status    Status   `yaml:"status" json:"status"`
	Tags      []string `yaml:"tags,omitempty,flow" json:"tags,omitempty"`
	Due       *Date    `yaml:"due,omitempty" json:"due,omitempty"`
	Created   *Date    `yaml:"created,omitempty" json:"created,omitempty"`
	Completed *Date    `yaml:"completed,omitempty" json:"completed,omitempty"`
	Effort    string   `yaml:"effort,omitempty" json:"effort,omitempty"`
	BlockedBy []string `yaml:"blocked_by,omitempty,flow" json:"blocked_by,omitempty"`
	Links     []string `yaml:"links,omitempty" json:"links,omitempty"`

	// Body is the markdown after the block, leading/trailing blank lines removed.
	Body string `yaml:"-" json:"body,omitempty"`
	// Span locates the block in the source file. Not serialized.
	Span Span `yaml:"-" json:"-"`
}

// Span records where a block sits in the file. Offsets are byte positions
// into the original data; Line is the 1-based line of the opening "---".
type Span struct {
	Line      int
	YAMLStart int // first byte after the opening --- line
	YAMLEnd   int // first byte of the closing --- line
	BodyStart int // first byte after the closing --- line
	BodyEnd   int // first byte of the next block's --- line, or len(data)
}

// Summary returns the first paragraph of the body on one line.
func (t *Task) Summary() string {
	body := strings.TrimSpace(t.Body)
	if body == "" {
		return ""
	}
	para := body
	if i := strings.Index(body, "\n\n"); i >= 0 {
		para = body[:i]
	}
	return strings.Join(strings.Fields(para), " ")
}

// Overdue reports whether the task is open with a due date before today.
func (t *Task) Overdue(today Date) bool {
	return t.Due != nil && !t.Status.Closed() && t.Due.Before(today)
}

var idPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,39}$`)

// ValidID reports whether id matches the required slug shape.
func ValidID(id string) bool { return idPattern.MatchString(id) }

// Validate checks field rules that the YAML schema cannot express.
func (t *Task) Validate() error {
	switch {
	case t.ID == "":
		return fmt.Errorf("id is required")
	case !ValidID(t.ID):
		return fmt.Errorf("id %q must match %s", t.ID, idPattern)
	case strings.TrimSpace(t.Title) == "":
		return fmt.Errorf("title is required")
	case strings.ContainsAny(t.Title, "\n\r"):
		return fmt.Errorf("title must be a single line")
	case t.Priority < 1 || t.Priority > 5:
		return fmt.Errorf("priority must be 1-5, got %d", t.Priority)
	case !t.Status.Valid():
		return fmt.Errorf("invalid status %q (want one of %s)", t.Status, joinStatuses())
	}
	for _, ref := range t.BlockedBy {
		if ref == "" {
			return fmt.Errorf("blocked_by entries must not be empty")
		}
	}
	return nil
}

// Slug derives an id from a title: lowercase, runs of non-alphanumerics
// collapsed to "-", trimmed, at most 40 characters.
func Slug(title string) string {
	var b strings.Builder
	lastDash := true // suppress leading dash
	for _, r := range strings.ToLower(title) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		default:
			if !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	s := strings.TrimRight(b.String(), "-")
	if len(s) > 40 {
		cut := s[:40]
		if s[40] != '-' {
			if i := strings.LastIndexByte(cut, '-'); i > 0 {
				cut = cut[:i]
			}
		}
		s = strings.TrimRight(cut, "-")
	}
	if s == "" {
		return "task"
	}
	return s
}
