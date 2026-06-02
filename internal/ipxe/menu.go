// Package ipxe generates custom iPXE menu scripts for BootWrangler builds.
package ipxe

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// Entry represents one boot menu entry.
type Entry struct {
	Label       string // display label
	KernelLine  string // full kernel/boot line(s)
	Description string // optional second line
}

// MenuOptions controls menu generation.
type MenuOptions struct {
	Title         string  // menu title (default: "BootWrangler")
	ServerBaseURL string  // base URL for assets
	Entries       []Entry // explicit entries (overrides auto-discovery)
	AddRescue     bool    // include a rescue shell entry
	NetbootURL    string  // optional netboot.xyz chain URL
}

// indexedEntry pairs an Entry with its zero-based index for template use.
type indexedEntry struct {
	Index int
	Entry Entry
}

// menuData is the data passed to the menu template.
type menuData struct {
	Title      string
	Entries    []indexedEntry
	AddRescue  bool
	NetbootURL string
}

var menuTemplate = template.Must(template.New("ipxe-menu").Funcs(template.FuncMap{
	"kernelLines": kernelLines,
}).Parse(menuTmplStr))

const menuTmplStr = `#!ipxe

set menu-timeout 5000
set menu-default boot-local

:menu
menu {{ .Title }} Boot Menu
item --gap -- {{ .Title }}
item boot-local  Boot from local disk
{{ if .AddRescue -}}
item shell       Rescue shell
{{ end -}}
{{ if .NetbootURL -}}
item --gap --
item netboot     netboot.xyz
{{ end -}}
item --gap --
{{ range .Entries -}}
item profile-{{ .Index }}  {{ .Entry.Label }}
{{ end -}}
item --gap --
choose --timeout ${menu-timeout} --default ${menu-default} target && goto ${target}

:boot-local
sanboot --no-describe --drive 0x80
goto menu
{{ if .AddRescue }}
:shell
shell
goto menu
{{ end -}}
{{ if .NetbootURL }}
:netboot
chain {{ .NetbootURL }}
goto menu
{{ end -}}
{{ range .Entries }}
:profile-{{ .Index }}
{{ kernelLines .Entry }}
{{ end -}}
`

var singleTemplate = template.Must(template.New("ipxe-single").Funcs(template.FuncMap{
	"kernelLines": kernelLines,
}).Parse(singleTmplStr))

const singleTmplStr = `#!ipxe

{{ kernelLines .Entry }}
`

// kernelLines produces the iPXE kernel/initrd/boot lines for an entry.
// The KernelLine field may contain multiple newline-separated lines; this
// function emits them verbatim followed by a final "boot" directive.
func kernelLines(e Entry) string {
	lines := strings.Split(strings.TrimSpace(e.KernelLine), "\n")
	var out strings.Builder
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		out.WriteString(l)
		out.WriteByte('\n')
	}
	out.WriteString("boot")
	return out.String()
}

// RenderMenu generates a complete iPXE menu script from opts.
func RenderMenu(opts MenuOptions) (string, error) {
	title := opts.Title
	if title == "" {
		title = "BootWrangler"
	}

	indexed := make([]indexedEntry, len(opts.Entries))
	for i, e := range opts.Entries {
		indexed[i] = indexedEntry{Index: i, Entry: e}
	}

	data := menuData{
		Title:      title,
		Entries:    indexed,
		AddRescue:  opts.AddRescue,
		NetbootURL: opts.NetbootURL,
	}

	var buf bytes.Buffer
	if err := menuTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("ipxe: render menu: %w", err)
	}
	return buf.String(), nil
}

// RenderSingleEntry generates a minimal single-profile iPXE script with no
// menu — just the kernel/boot directives for e.
func RenderSingleEntry(entry Entry, serverBaseURL string) (string, error) {
	_ = serverBaseURL // reserved for future URL rewriting

	data := struct{ Entry Entry }{Entry: entry}

	var buf bytes.Buffer
	if err := singleTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("ipxe: render single entry: %w", err)
	}
	return buf.String(), nil
}
