// Command gendocs writes the CLI reference for the docs site: one
// Starlight markdown page per clask command, generated from cobra's help
// text so --help and the site never disagree. It is a dev tool and is not
// part of the shipped binary.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Thenlie/clask/internal/cli"
)

func main() {
	out := flag.String("out", "docs/src/content/docs/reference/cli", "output directory (replaced)")
	base := flag.String("base", "/clask/reference/cli/", "URL prefix for links between pages")
	flag.Parse()
	n, err := generate(cli.NewRootCmd(), *out, *base)
	if err != nil {
		fmt.Fprintln(os.Stderr, "gendocs:", err)
		os.Exit(1)
	}
	fmt.Printf("gendocs: wrote %d pages to %s\n", n, *out)
}
