package bootmenu

import (
	"bytes"
	"fmt"
	"text/template"
)

const grubTemplate = `set timeout=10
set default=0

menuentry "{{ .Title }}" {
  true
}
{{ range $i, $e := .Entries }}
menuentry "{{ $e.Label }}" {
  {{ grubEntryBlock $e -}}
}
{{ end }}
`

// RenderGRUB renders a Menu as a GRUB configuration.
func RenderGRUB(m Menu) (string, error) {
	funcs := template.FuncMap{
		"grubEntryBlock": grubEntryBlock,
	}
	tmpl, err := template.New("grub").Funcs(funcs).Parse(grubTemplate)
	if err != nil {
		return "", fmt.Errorf("grub menu: parse template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, m); err != nil {
		return "", fmt.Errorf("grub menu: render: %w", err)
	}
	return buf.String(), nil
}

func grubEntryBlock(e Entry) string {
	switch e.Kind {
	case EntryNetboot:
		return fmt.Sprintf("chainloader %s", e.URL)
	case EntryLocalProfile, EntryLocalImage:
		if e.Kernel != "" {
			s := fmt.Sprintf("linux %s", e.Kernel)
			if e.Cmdline != "" {
				s += " " + e.Cmdline
			}
			if e.Initrd != "" {
				s += fmt.Sprintf("\n  initrd %s", e.Initrd)
			}
			return s
		}
		return fmt.Sprintf("echo 'No boot method for %q'", e.Label)
	case EntryShell:
		return "terminal_input console\n  terminal_output console"
	case EntryReboot:
		return "reboot"
	default:
		return fmt.Sprintf("echo 'Unknown entry'")
	}
}
