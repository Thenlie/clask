package cli

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// confirm asks a yes/no question on the command's streams. It returns true
// without asking when yes is set. EOF or anything but y/yes is false.
func confirm(cmd *cobra.Command, yes bool, prompt string) bool {
	if yes {
		return true
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%s [y/N] ", prompt)
	line, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
	if err != nil && line == "" {
		fmt.Fprintln(cmd.OutOrStdout())
		return false
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true
	}
	return false
}
