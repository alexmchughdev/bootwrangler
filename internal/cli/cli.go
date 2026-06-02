package cli

import (
	"fmt"
	"io"

	"github.com/alexmchughdev/bootwrangler/internal/profile"
	"github.com/alexmchughdev/bootwrangler/internal/version"
)

const helpText = `BootWrangler is a GUI-first Linux provisioning and boot media studio.

Usage:
  bootwrangler <command>

Commands:
  profile     Manage provisioning profiles
  version     Print the BootWrangler version

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
	case "profile":
		return runProfile(args[1:], stdout, stderr)
	case "version":
		if len(args) != 1 {
			fmt.Fprintln(stderr, "usage: bootwrangler version")
			return 2
		}
		fmt.Fprintf(stdout, "BootWrangler %s\n", version.Current())
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command %q\n", args[0])
		fmt.Fprintln(stderr, "run \"bootwrangler --help\" for usage")
		return 2
	}
}

func runProfile(args []string, stdout, stderr io.Writer) int {
	if len(args) != 2 || args[0] != "validate" {
		fmt.Fprintln(stderr, "usage: bootwrangler profile validate <profile.yaml>")
		return 2
	}

	value, err := profile.LoadAndValidateFile(args[1])
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	fmt.Fprintf(stdout, "profile valid: %s\n", value.Name)
	return 0
}
