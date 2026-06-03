package cli

import (
	"fmt"
	"io"

	"github.com/alexmchughdev/bootwrangler/internal/editor"
	"github.com/alexmchughdev/bootwrangler/internal/images"
	"github.com/alexmchughdev/bootwrangler/internal/library"
	"github.com/alexmchughdev/bootwrangler/internal/profile"
	"github.com/alexmchughdev/bootwrangler/internal/render"
	"github.com/alexmchughdev/bootwrangler/internal/version"
)

const helpText = `BootWrangler is a GUI-first Linux provisioning and boot media studio.

Usage:
  bootwrangler <command>

Commands:
  images      Browse and manage OS image catalogue
  library     Manage the local profile library
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
	case "images":
		return runImages(args[1:], stdout, stderr)
	case "library":
		return runLibrary(args[1:], stdout, stderr)
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

func runLibrary(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage:")
		fmt.Fprintln(stderr, "  bootwrangler library init")
		fmt.Fprintln(stderr, "  bootwrangler library list")
		fmt.Fprintln(stderr, "  bootwrangler library add <profile.yaml>")
		fmt.Fprintln(stderr, "  bootwrangler library show <name>")
		fmt.Fprintln(stderr, "  bootwrangler library remove <name>")
		return 2
	}

	lib := library.New(library.DefaultDir())

	switch args[0] {
	case "init":
		if err := lib.Init(); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintf(stdout, "initialised library at %s\n", library.DefaultDir())
		return 0

	case "list":
		entries, err := lib.List()
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if len(entries) == 0 {
			fmt.Fprintln(stdout, "no profiles in library")
			return 0
		}
		for _, e := range entries {
			fmt.Fprintf(stdout, "%-30s %s %s\n", e.Name, e.OSFamily, e.OSVersion)
		}
		return 0

	case "add":
		if len(args) != 2 {
			fmt.Fprintln(stderr, "usage: bootwrangler library add <profile.yaml>")
			return 2
		}
		p, err := profile.LoadAndValidateFile(args[1])
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		filename, err := lib.Add(p)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintf(stdout, "added profile %q as %s\n", p.Name, filename)
		return 0

	case "show":
		if len(args) != 2 {
			fmt.Fprintln(stderr, "usage: bootwrangler library show <name>")
			return 2
		}
		p, err := lib.Get(args[1])
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintf(stdout, "name:       %s\n", p.Name)
		fmt.Fprintf(stdout, "os_family:  %s\n", p.OS.Family)
		fmt.Fprintf(stdout, "os_version: %s\n", p.OS.Version)
		fmt.Fprintf(stdout, "hostname:   %s\n", p.System.Hostname)
		fmt.Fprintf(stdout, "users:      %d\n", len(p.Users))
		return 0

	case "remove":
		if len(args) != 2 {
			fmt.Fprintln(stderr, "usage: bootwrangler library remove <name>")
			return 2
		}
		if err := lib.Remove(args[1]); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintf(stdout, "removed profile %q\n", args[1])
		return 0

	case "export":
		if len(args) < 2 {
			fmt.Fprintln(stderr, "usage: bootwrangler library export <name> --out <file.zip>")
			return 2
		}
		name := args[0]
		outPath := name + ".zip"
		for i := 1; i < len(args)-1; i++ {
			if args[i] == "--out" {
				outPath = args[i+1]
			}
		}
		p, err := lib.Get(name)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if err := library.ExportBundle(p, outPath); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintf(stdout, "exported %q to %s\n", name, outPath)
		return 0

	case "import":
		if len(args) != 2 {
			fmt.Fprintln(stderr, "usage: bootwrangler library import <file.zip>")
			return 2
		}
		p, err := library.ImportBundle(args[1])
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		filename, err := lib.Add(p)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintf(stdout, "imported profile %q as %s\n", p.Name, filename)
		return 0

	default:
		fmt.Fprintf(stderr, "unknown library command %q\n", args[0])
		return 2
	}
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

func runImages(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage:")
		fmt.Fprintln(stderr, "  bootwrangler images list")
		fmt.Fprintln(stderr, "  bootwrangler images show <id>")
		return 2
	}

	cat := images.BuiltinCatalogue()

	switch args[0] {
	case "list":
		for _, e := range cat.Entries {
			fmt.Fprintf(stdout, "%-22s  %-18s  %s\n", e.ID, e.Family, e.Name)
		}
		return 0

	case "show":
		if len(args) != 2 {
			fmt.Fprintln(stderr, "usage: bootwrangler images show <id>")
			return 2
		}
		id := args[1]
		for _, e := range cat.Entries {
			if e.ID != id {
				continue
			}
			fmt.Fprintf(stdout, "id:     %s\n", e.ID)
			fmt.Fprintf(stdout, "name:   %s\n", e.Name)
			fmt.Fprintf(stdout, "family: %s\n", e.Family)
			for _, v := range e.Versions {
				for _, a := range v.Architectures {
					for _, img := range a.Images {
						fmt.Fprintf(stdout, "version: %-10s  arch: %-8s  type: %-6s  url: %s\n",
							v.Version, a.Arch, img.Type, img.URL)
					}
				}
			}
			return 0
		}
		fmt.Fprintf(stderr, "image not found: %s\n", id)
		return 1

	default:
		fmt.Fprintf(stderr, "unknown images command %q\n", args[0])
		return 2
	}
}

func printProfileUsage(writer io.Writer) {
	fmt.Fprintln(writer, "usage:")
	fmt.Fprintln(writer, "  bootwrangler profile edit <profile.yaml> --editor nvim")
	fmt.Fprintln(writer, "  bootwrangler profile validate <profile.yaml>")
}
