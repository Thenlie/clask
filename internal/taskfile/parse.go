package taskfile

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// File is a parsed TASKS.md.
type File struct {
	Meta     FileMeta
	MetaSpan Span
	Tasks    []*Task
	// Problems are per-task errors: the block exists but could not be used.
	// The file itself is still usable; callers decide whether to fail.
	Problems []Problem
	// Data is the original bytes, retained so the writer can patch ranges.
	Data []byte
}

// Problem describes a task block that was skipped.
type Problem struct {
	Line int
	Err  error
}

func (p Problem) Error() string { return fmt.Sprintf("line %d: %v", p.Line, p.Err) }

// Options tune parsing.
type Options struct {
	// DefaultPriority is used when neither the task nor the file sets one.
	DefaultPriority int
}

// ErrNoFileBlock is returned when the file does not start with a --- block.
var ErrNoFileBlock = errors.New("file must start with a '---' block (it may be empty)")

const hrHint = "a '---' line inside a task body starts a new block; use '***' for horizontal rules"

// Parse decodes a TASKS.md. Structural errors (unclosed block, no leading
// file block, bad file metadata) return an error. Per-task errors are
// collected in File.Problems and the task is omitted from File.Tasks.
func Parse(data []byte, opts Options) (*File, error) {
	spans, err := splitBlocks(data)
	if err != nil {
		return nil, err
	}
	f := &File{Data: data, MetaSpan: spans[0]}

	if err := decodeStrict(data[spans[0].YAMLStart:spans[0].YAMLEnd], &f.Meta); err != nil {
		line, err := absoluteLine(spans[0].Line, err)
		return nil, fmt.Errorf("file block (line %d): %w", line, err)
	}
	if f.Meta.DefaultPriority != 0 && (f.Meta.DefaultPriority < 1 || f.Meta.DefaultPriority > 5) {
		return nil, fmt.Errorf("file block (line %d): default_priority must be 1-5", spans[0].Line)
	}
	defPri := f.Meta.DefaultPriority
	if defPri == 0 {
		defPri = opts.DefaultPriority
	}
	if defPri == 0 {
		defPri = 3
	}

	seen := map[string]int{}
	for _, sp := range spans[1:] {
		t := &Task{Span: sp}
		yamlText := data[sp.YAMLStart:sp.YAMLEnd]
		if err := decodeStrict(yamlText, t); err != nil {
			line, err := absoluteLine(sp.Line, err)
			if !looksLikeMapping(yamlText) {
				err = fmt.Errorf("%w (%s)", err, hrHint)
			}
			f.Problems = append(f.Problems, Problem{Line: line, Err: err})
			continue
		}
		if t.Priority == 0 {
			t.Priority = defPri
		}
		if t.Status == "" {
			t.Status = StatusTodo
		}
		if err := t.Validate(); err != nil {
			f.Problems = append(f.Problems, Problem{Line: sp.Line, Err: err})
			continue
		}
		if first, dup := seen[t.ID]; dup {
			f.Problems = append(f.Problems, Problem{Line: sp.Line, Err: fmt.Errorf("duplicate id %q (first defined at line %d)", t.ID, first)})
			continue
		}
		seen[t.ID] = sp.Line
		t.Body = trimBody(data[sp.BodyStart:sp.BodyEnd])
		f.Tasks = append(f.Tasks, t)
	}
	return f, nil
}

// Task returns the task with the given id.
func (f *File) Task(id string) (*Task, bool) {
	for _, t := range f.Tasks {
		if t.ID == id {
			return t, true
		}
	}
	return nil, false
}

// decodeStrict unmarshals YAML rejecting unknown keys. Empty input is valid
// and leaves v untouched.
func decodeStrict(b []byte, v any) error {
	dec := yaml.NewDecoder(bytes.NewReader(b))
	dec.KnownFields(true)
	if err := dec.Decode(v); err != nil && !errors.Is(err, io.EOF) {
		return cleanYAMLError(err)
	}
	return nil
}

// cleanYAMLError rewrites yaml.v3 messages for users: drops the "yaml: "
// prefix and replaces Go type names with the schema they map to.
func cleanYAMLError(err error) error {
	msg := strings.TrimPrefix(err.Error(), "yaml: ")
	msg = strings.TrimPrefix(msg, "unmarshal errors:\n  ")
	msg = strings.ReplaceAll(msg, " not found in type taskfile.Task", " is not a task key")
	msg = strings.ReplaceAll(msg, " not found in type taskfile.FileMeta", " is not a file key")
	msg = strings.ReplaceAll(msg, "into taskfile.Task", "into a task")
	msg = strings.ReplaceAll(msg, "into taskfile.FileMeta", "into file metadata")
	return errors.New(msg)
}

var yamlLineRe = regexp.MustCompile(`^line (\d+): `)

// absoluteLine converts a block-relative "line N: ..." YAML error into the
// file line and strips the prefix. Errors without a line keep blockLine.
func absoluteLine(blockLine int, err error) (int, error) {
	msg := err.Error()
	m := yamlLineRe.FindStringSubmatch(msg)
	if m == nil {
		return blockLine, err
	}
	rel, _ := strconv.Atoi(m[1])
	return blockLine + rel, errors.New(strings.TrimPrefix(msg, m[0]))
}

var keyLineRe = regexp.MustCompile(`(?m)^[a-z_]+:`)

// looksLikeMapping reports whether the YAML text has at least one key line.
// Prose captured by a stray --- in a body does not.
func looksLikeMapping(b []byte) bool { return keyLineRe.Match(b) }

func trimBody(b []byte) string {
	return strings.Trim(string(b), "\r\n")
}

// isDelimiter reports whether a line (without its newline) is exactly ---,
// tolerating a trailing carriage return.
func isDelimiter(line []byte) bool {
	line = bytes.TrimSuffix(line, []byte("\r"))
	return bytes.Equal(line, []byte("---"))
}

// splitBlocks finds every "---"-delimited block and the body after it.
func splitBlocks(data []byte) ([]Span, error) {
	var spans []Span
	inYAML := false
	offset := 0
	lineNo := 0
	for offset < len(data) {
		end := bytes.IndexByte(data[offset:], '\n')
		var line []byte
		next := len(data)
		if end >= 0 {
			line = data[offset : offset+end]
			next = offset + end + 1
		} else {
			line = data[offset:]
		}
		lineNo++

		switch {
		case isDelimiter(line) && !inYAML:
			if len(spans) == 0 {
				if len(bytes.TrimSpace(data[:offset])) != 0 {
					return nil, ErrNoFileBlock
				}
			} else {
				spans[len(spans)-1].BodyEnd = offset
			}
			spans = append(spans, Span{Line: lineNo, YAMLStart: next})
			inYAML = true
		case isDelimiter(line) && inYAML:
			cur := &spans[len(spans)-1]
			cur.YAMLEnd = offset
			cur.BodyStart = next
			inYAML = false
		case len(spans) == 0 && len(bytes.TrimSpace(line)) != 0:
			return nil, ErrNoFileBlock
		}
		offset = next
	}
	if len(spans) == 0 {
		return nil, ErrNoFileBlock
	}
	if inYAML {
		return nil, fmt.Errorf("block opened at line %d is never closed with '---' (%s)", spans[len(spans)-1].Line, hrHint)
	}
	spans[len(spans)-1].BodyEnd = len(data)
	return spans, nil
}
