// Package render formats clask output for the terminal and for machines.
package render

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"regexp"
	"strings"
	"unicode/utf8"
)

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// visibleWidth counts the characters a terminal would show, ignoring ANSI
// color sequences.
func visibleWidth(s string) int {
	return utf8.RuneCountInString(ansiRe.ReplaceAllString(s, ""))
}

// Column controls how one table column is laid out.
type Column struct {
	Header string
	Right  bool // right-align (numbers)
}

// Table writes rows aligned in columns, measuring visible width so colored
// cells line up with the header. Columns may be nil for header-less output.
func Table(w io.Writer, cols []Column, rows [][]string) error {
	n := len(cols)
	for _, r := range rows {
		if len(r) > n {
			n = len(r)
		}
	}
	widths := make([]int, n)
	for i, c := range cols {
		widths[i] = visibleWidth(c.Header)
	}
	for _, r := range rows {
		for i, cell := range r {
			if wd := visibleWidth(cell); wd > widths[i] {
				widths[i] = wd
			}
		}
	}
	var b strings.Builder
	writeRow := func(cells []string) {
		for i := 0; i < n; i++ {
			cell := ""
			if i < len(cells) {
				cell = cells[i]
			}
			pad := widths[i] - visibleWidth(cell)
			last := i == n-1
			right := i < len(cols) && cols[i].Right
			switch {
			case right:
				b.WriteString(strings.Repeat(" ", pad))
				b.WriteString(cell)
			case last:
				b.WriteString(cell) // no trailing spaces
			default:
				b.WriteString(cell)
				b.WriteString(strings.Repeat(" ", pad))
			}
			if !last {
				b.WriteString("  ")
			}
		}
		b.WriteByte('\n')
	}
	if len(cols) > 0 {
		headers := make([]string, len(cols))
		for i, c := range cols {
			headers[i] = c.Header
		}
		writeRow(headers)
	}
	for _, r := range rows {
		writeRow(r)
	}
	_, err := io.WriteString(w, b.String())
	return err
}

// Headers is a convenience for plain left-aligned columns.
func Headers(names ...string) []Column {
	cols := make([]Column, len(names))
	for i, n := range names {
		cols[i] = Column{Header: n}
	}
	return cols
}

// JSON writes v as indented JSON followed by a newline.
func JSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// JSONLines writes each item as one compact JSON object per line.
func JSONLines[T any](w io.Writer, items []T) error {
	enc := json.NewEncoder(w)
	for _, it := range items {
		if err := enc.Encode(it); err != nil {
			return err
		}
	}
	return nil
}

// Delimited writes a header row and rows as CSV (comma) or TSV (tab).
func Delimited(w io.Writer, delim rune, headers []string, rows [][]string) error {
	cw := csv.NewWriter(w)
	cw.Comma = delim
	if err := cw.Write(headers); err != nil {
		return err
	}
	for _, r := range rows {
		if err := cw.Write(r); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}
