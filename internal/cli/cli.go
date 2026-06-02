package cli

import (
	"fmt"
	"io"

	"github.com/alexmchughdev/bootwrangler/internal/editor"
	"github.com/alexmchughdev/bootwrangler/internal/profile"
	"github.com/alexmchughdev/bootwrangler/internal/render"
	"github.com/alexmchughdev/bootwrangler/internal/version"
)

const helpText = `BootWrangler is a GUI-first Linux provisioning and boot media studio.

Usage:
  bootwrangler <command>

Commands:
  profile     Manage provisioning profiles
  render      Render a profile into unattended installer assets
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
	case "render":
		return runRender(args[1:], stdout, stderr)
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
	if len(args) == 0 {
		printProfileUsage(stderr)
		return 2
	}

	switch args[0] {
	case "edit":
		path, err := parseProfileEditArgs(args[1:])
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
		plan, err := editor.NeovimPlan(path, false)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if err := editor.Start(plan); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintf(stdout, "opened profile in Neovim: %s\n", path)
		return 0
	case "validate":
		if len(args) != 2 {
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
	default:
		printProfileUsage(stderr)
		return 2
	}
}

func parseProfileEditArgs(args []string) (string, error) {
	if len(args) != 3 || args[1] != "--editor" || args[2] != "nvim" {
		return "", fmt.Errorf("usage: bootwrangler profile edit <profile.yaml> --editor nvim")
	}
	return args[0], nil
}

func runRender(args []string, stdout, stderr io.Writer) int {
	outDir := "."
	profilePath := ""

	i := 0
	for i < len(args) {
		switch args[i] {
		case "--out":
			if i+1 >= len(args) {
				fmt.Fprintln(stderr, "render: --out requires a directory argument")
				return 2
			}
			outDir = args[i+1]
			i += 2
		default:
			if profilePath != "" {
				fmt.Fprintln(stderr, "render: unexpected argument:", args[i])
				return 2
			}
			profilePath = args[i]
			i++
		}
	}
	if profilePath == "" {
		fmt.Fprintln(stderr, "usage: bootwrangler render <profile.yaml> [--out <dir>]")
		return 2
	}

	p, err := profile.LoadAndValidateFile(profilePath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	r, err := render.Lookup(p.OS.Family)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	opts := render.Options{OutDir: outDir}
	m, err := r.Render(p, opts)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	fmt.Fprintf(stdout, "rendered %s (%s %s) via %s renderer into %s\n",
		m.ProfileName, m.OSFamily, m.OSVersion, m.Renderer, outDir)
	for _, w := range m.Warnings {
		fmt.Fprintf(stdout, "warning: %s\n", w)
	}
	return 0
}

func printProfileUsage(writer io.Writer) {
	fmt.Fprintln(writer, "usage:")
	fmt.Fprintln(writer, "  bootwrangler profile edit <profile.yaml> --editor nvim")
	fmt.Fprintln(writer, "  bootwrangler profile validate <profile.yaml>")
}
