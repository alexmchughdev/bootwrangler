package cli

import (
	"fmt"
	"io"

	"github.com/alexmchughdev/bootforge/internal/version"
)

const helpText = `BootForge is a GUI-first Linux provisioning and boot media studio.

Usage:
  bootforge <command>

Commands:
  version     Print the BootForge version

Options:
  -h, --help  Show this help
`

// Run executes the CLI and returns a process exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Fprint(stdout, helpText)
		return 0
	}

	switch args[0] {
	case "version":
		if len(args) != 1 {
			fmt.Fprintln(stderr, "usage: bootforge version")
			return 2
		}
		fmt.Fprintf(stdout, "BootForge %s\n", version.Current())
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command %q\n", args[0])
		fmt.Fprintln(stderr, "run \"bootforge --help\" for usage")
		return 2
	}
}
