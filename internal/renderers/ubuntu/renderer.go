// Package ubuntu renders Ubuntu Server Subiquity autoinstall assets.
package ubuntu

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/alexmchughdev/bootwrangler/internal/manifest"
	"github.com/alexmchughdev/bootwrangler/internal/profile"
	"github.com/alexmchughdev/bootwrangler/internal/render"
)

// Renderer generates Ubuntu autoinstall assets.
type Renderer struct{}

func (r Renderer) Name() string                { return "ubuntu" }
func (r Renderer) Family() string              { return "ubuntu" }
func (r Renderer) SupportedVersions() []string { return nil }

// Validate performs Ubuntu-specific profile validation.
func (r Renderer) Validate(p profile.Profile) error {
	var problems []string

	if p.OS.Family != "ubuntu" {
		problems = append(problems, fmt.Sprintf("ubuntu renderer requires os.family \"ubuntu\", got %q", p.OS.Family))
	}

	if len(p.Users) == 0 {
		problems = append(problems, "ubuntu renderer requires at least one user")
	} else if len(p.Users[0].SSHKeys) == 0 {
		problems = append(problems, "ubuntu renderer requires the first user to have at least one SSH key")
	}

	if len(problems) > 0 {
		return fmt.Errorf("ubuntu renderer validation: %s", strings.Join(problems, "; "))
	}
	return nil
}

// Render generates Ubuntu autoinstall assets into opts.OutDir.
func (r Renderer) Render(p profile.Profile, opts render.Options) (manifest.Manifest, error) {
	m := manifest.Manifest{
		ProfileName: p.Name,
		OSFamily:    "ubuntu",
		OSVersion:   p.OS.Version,
		Renderer:    r.Name(),
	}

	userData, err := renderUserData(p)
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("ubuntu: render user-data: %w", err)
	}

	metaData, err := renderMetaData(p)
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("ubuntu: render meta-data: %w", err)
	}

	customIPXE, err := renderCustomIPXE(opts)
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("ubuntu: render custom.ipxe: %w", err)
	}

	m.AddFile("user-data", "Ubuntu Subiquity autoinstall configuration")
	m.AddFile("meta-data", "cloud-init metadata")
	m.AddFile("custom.ipxe", "iPXE boot entry")

	if !opts.DryRun {
		if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
			return manifest.Manifest{}, fmt.Errorf("ubuntu: create output directory: %w", err)
		}

		type fileSpec struct {
			name    string
			content []byte
		}
		for _, f := range []fileSpec{
			{"user-data", userData},
			{"meta-data", metaData},
			{"custom.ipxe", customIPXE},
		} {
			path := filepath.Join(opts.OutDir, f.name)
			if err := os.WriteFile(path, f.content, 0o644); err != nil {
				return manifest.Manifest{}, fmt.Errorf("ubuntu: write %s: %w", f.name, err)
			}
		}

		if err := m.ComputeChecksums(opts.OutDir); err != nil {
			return manifest.Manifest{}, fmt.Errorf("ubuntu: compute checksums: %w", err)
		}
	}

	return m, nil
}

func init() {
	render.Register(Renderer{})
}

// ---- templates ---------------------------------------------------------------

const userDataTmpl = `#cloud-config
autoinstall:
  version: {{ .Version }}
  locale: {{ .Locale }}
  keyboard:
    layout: {{ .Keyboard }}
  identity:
    hostname: {{ .Hostname }}
    username: {{ .PrimaryUser.Name }}
    password: ""
  ssh:
    install-server: true
    allow-pw: {{ .AllowPassword }}
    authorized-keys:{{ range .PrimaryUser.SSHKeys }}
      - {{ . }}{{ end }}
  storage:
    layout:
      name: {{ .StorageLayout }}
  packages:{{ if .Packages }}{{ range .Packages }}
    - {{ . }}{{ end }}{{ else }} []{{ end }}
  user-data:
    timezone: {{ .Timezone }}{{ if not .PermitRoot }}
    disable_root: true{{ end }}{{ if .AdditionalUsers }}
    users:{{ range .AdditionalUsers }}
      - name: {{ .Name }}{{ if .Shell }}
        shell: {{ .Shell }}{{ end }}{{ if .Groups }}
        groups:{{ range .Groups }}
          - {{ . }}{{ end }}{{ end }}{{ if .Sudo }}
        sudo: ALL=(ALL) NOPASSWD:ALL{{ end }}{{ if .SSHKeys }}
        ssh_authorized_keys:{{ range .SSHKeys }}
          - {{ . }}{{ end }}{{ end }}{{ end }}{{ end }}{{ if .LateCommands }}
  late-commands:{{ range .LateCommands }}
    - curtin in-target -- bash -c {{ escapeShell .Content }}{{ end }}{{ end }}
`

const metaDataTmpl = `instance-id: iid-{{ .ProfileName }}
local-hostname: {{ .Hostname }}
`

const customIPXETmpl = `#!ipxe
{{ if .ServerBaseURL -}}
set base-url {{ .ServerBaseURL }}
kernel ${base-url}/casper/vmlinuz autoinstall ds=nocloud-net;s=${base-url}/ quiet ---
initrd ${base-url}/casper/initrd
boot
{{- else -}}
# configure ServerBaseURL in BootWrangler settings
{{- end }}
`

// ---- template data structures -----------------------------------------------

type userDataData struct {
	Version         int
	Locale          string
	Keyboard        string
	Hostname        string
	PrimaryUser     profile.User
	AllowPassword   bool
	StorageLayout   string
	Packages        []string
	Timezone        string
	PermitRoot      bool
	AdditionalUsers []profile.User
	LateCommands    []profile.Script
}

type metaDataData struct {
	ProfileName string
	Hostname    string
}

type customIPXEData struct {
	ServerBaseURL string
}

// ---- rendering helpers -------------------------------------------------------

func renderUserData(p profile.Profile) ([]byte, error) {
	version := 1
	if p.OSSpecific.Ubuntu != nil && p.OSSpecific.Ubuntu.AutoinstallVersion > 0 {
		version = p.OSSpecific.Ubuntu.AutoinstallVersion
	}

	storageLayout := "direct"
	if p.Disk.Mode != "wipe" {
		storageLayout = p.Disk.Mode
	}

	var additionalUsers []profile.User
	if len(p.Users) > 1 {
		additionalUsers = p.Users[1:]
	}

	data := userDataData{
		Version:         version,
		Locale:          p.System.Locale,
		Keyboard:        p.System.Keyboard,
		Hostname:        p.System.Hostname,
		PrimaryUser:     p.Users[0],
		AllowPassword:   p.SSH.PasswordAuthentication,
		StorageLayout:   storageLayout,
		Packages:        p.Packages.Names,
		Timezone:        p.System.Timezone,
		PermitRoot:      p.SSH.PermitRootLogin,
		AdditionalUsers: additionalUsers,
		LateCommands:    p.PostInstall.Scripts,
	}

	funcMap := template.FuncMap{
		"escapeShell": escapeShell,
		"not":         func(b bool) bool { return !b },
	}

	tmpl, err := template.New("user-data").Funcs(funcMap).Parse(userDataTmpl)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func renderMetaData(p profile.Profile) ([]byte, error) {
	data := metaDataData{
		ProfileName: p.Name,
		Hostname:    p.System.Hostname,
	}

	tmpl, err := template.New("meta-data").Parse(metaDataTmpl)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func renderCustomIPXE(opts render.Options) ([]byte, error) {
	data := customIPXEData{
		ServerBaseURL: opts.ServerBaseURL,
	}

	tmpl, err := template.New("custom.ipxe").Parse(customIPXETmpl)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// escapeShell wraps a shell string in single quotes, escaping any embedded
// single quotes so the result is safe to pass as a bash argument.
func escapeShell(s string) string {
	escaped := strings.ReplaceAll(s, "'", `'\''`)
	// Trim trailing newline inside the script content before wrapping.
	escaped = strings.TrimRight(escaped, "\n")
	return "'" + escaped + "'"
}
