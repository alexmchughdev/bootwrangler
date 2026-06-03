// Package bootmenu generates boot menus for multi-profile and multi-image media.
package bootmenu

// Entry is one item in a boot menu.
type Entry struct {
	Label   string
	Kind    EntryKind
	URL     string // iPXE URL or path
	Kernel  string
	Initrd  string
	Cmdline string
}

// EntryKind describes what a menu entry does.
type EntryKind string

const (
	EntryNetboot      EntryKind = "netboot"
	EntryLocalProfile EntryKind = "local-profile"
	EntryLocalImage   EntryKind = "local-image"
	EntryShell        EntryKind = "shell"
	EntryReboot       EntryKind = "reboot"
)

// Menu is a complete boot menu definition.
type Menu struct {
	Title   string
	Entries []Entry
}

// MenuOptions configures boot menu generation.
type MenuOptions struct {
	Title      string
	BaseURL    string // HTTP base URL for served assets
	NetbootURL string // netboot.xyz URL; defaults to public if empty
	AddShell   bool   // include a rescue shell entry
	AddNetboot bool   // include a netboot.xyz entry
	AddReboot  bool   // include a reboot entry
}

// DefaultNetbootURL is the public netboot.xyz iPXE URL.
const DefaultNetbootURL = "https://boot.netboot.xyz"

func defaultStr(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

// Build constructs a Menu from options and a set of profile/image entries.
func Build(opts MenuOptions, entries []Entry) Menu {
	title := defaultStr(opts.Title, "BootWrangler")
	all := make([]Entry, 0, len(entries)+3)

	if opts.AddNetboot {
		all = append(all, Entry{
			Label: "netboot.xyz",
			Kind:  EntryNetboot,
			URL:   defaultStr(opts.NetbootURL, DefaultNetbootURL),
		})
	}

	all = append(all, entries...)

	if opts.AddShell {
		all = append(all, Entry{
			Label: "Shell",
			Kind:  EntryShell,
		})
	}
	if opts.AddReboot {
		all = append(all, Entry{
			Label: "Reboot",
			Kind:  EntryReboot,
		})
	}

	return Menu{Title: title, Entries: all}
}
