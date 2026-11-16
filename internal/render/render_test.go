package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/fatih/color"
)

func TestTableAlignsColoredCells(t *testing.T) {
	color.NoColor = false
	defer func() { color.NoColor = true }()
	cyan := color.New(color.FgCyan)
	var buf bytes.Buffer
	err := Table(&buf, []Column{{Header: "A"}, {Header: "N", Right: true}, {Header: "LAST"}}, [][]string{
		{cyan.Sprint("xx"), cyan.Sprint("1"), "end"},
		{"longer", "10", cyan.Sprint("é")},
	})
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("lines = %q", lines)
	}
	// strip color, then every row must have identical column starts
	plain := make([]string, len(lines))
	for i, l := range lines {
		plain[i] = ansiRe.ReplaceAllString(l, "")
	}
	want := []string{
		"A        N  LAST",
		"xx       1  end",
		"longer  10  é",
	}
	for i := range want {
		if plain[i] != want[i] {
			t.Errorf("row %d = %q, want %q", i, plain[i], want[i])
		}
	}
	if strings.HasSuffix(lines[1], " ") {
		t.Error("trailing whitespace on row")
	}
}

func TestDelimited(t *testing.T) {
	var buf bytes.Buffer
	if err := Delimited(&buf, '\t', []string{"a", "b"}, [][]string{{"x", "has\ttab"}}); err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); got != "a\tb\nx\t\"has\ttab\"\n" {
		t.Fatalf("tsv = %q", got)
	}
}
