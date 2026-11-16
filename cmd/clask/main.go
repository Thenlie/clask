// Command clask aggregates TASKS.md files across registered projects.
package main

import (
	"os"

	"github.com/Thenlie/clask/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
