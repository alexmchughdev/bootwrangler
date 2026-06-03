package bootmenu

import (
	"bytes"
	"fmt"
	"text/template"
)

const ipxeTemplate = `#!ipxe

set menu-timeout 5000
set menu-default {{ .DefaultEntry }}

:start
menu {{ .Title }}
{{ range $i, $e := .Entries -}}
item --key {{ inc $i }} entry-{{ $i }} {{ $e.Label }}
{{ end -}}
choose --timeout ${menu-timeout} --default ${menu-default} selected
goto ${selected}
{{ range $i, $e := .Entries }}
:entry-{{ $i }}
{{ entryScript $e -}}
goto start
{{ end }}
`

// RenderIPXE renders a Menu as an iPXE script.
func RenderIPXE(m Menu) (string, error) {
	type tmplData struct {
		Title        string
		Entries      []Entry
		DefaultEntry string
	}

	defaultEntry := "entry-0"
	if len(m.Entries) == 0 {
		defaultEntry = ""
	}

	funcs := template.FuncMap{
		"inc": func(i int) int { return i + 1 },
		"entryScript": func(e Entry) string {
			return ipxeEntryScript(e)
		},
	}

	tmpl, err := template.New("ipxe").Funcs(funcs).Parse(ipxeTemplate)
	if err != nil {
		return "", fmt.Errorf("ipxe menu: parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, tmplData{
		Title:        m.Title,
		Entries:      m.Entries,
		DefaultEntry: defaultEntry,
	}); err != nil {
		return "", fmt.Errorf("ipxe menu: render: %w", err)
	}
	return buf.String(), nil
}

func ipxeEntryScript(e Entry) string {
	switch e.Kind {
	case EntryNetboot:
		return fmt.Sprintf("chain %s\n", e.URL)
	case EntryLocalProfile, EntryLocalImage:
		if e.URL != "" {
			return fmt.Sprintf("chain %s\n", e.URL)
		}
		if e.Kernel != "" {
			s := fmt.Sprintf("kernel %s", e.Kernel)
			if e.Initrd != "" {
				s += fmt.Sprintf("\ninitrd %s", e.Initrd)
			}
			if e.Cmdline != "" {
				s += fmt.Sprintf(" %s", e.Cmdline)
			}
			return s + "\nboot\n"
		}
		return fmt.Sprintf("echo No boot method configured for %q\nprompt\n", e.Label)
	case EntryShell:
		return "shell\n"
	case EntryReboot:
		return "reboot\n"
	default:
		return fmt.Sprintf("echo Unknown entry kind %q\nprompt\n", e.Kind)
	}
}
