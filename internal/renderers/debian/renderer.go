// Package debian renders Debian preseed unattended installer assets.
package debian

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

// Renderer generates Debian preseed assets.
type Renderer struct{}

func (r Renderer) Name() string                { return "debian" }
func (r Renderer) Family() string              { return "debian" }
func (r Renderer) SupportedVersions() []string { return nil }

// Validate performs Debian-specific profile validation.
func (r Renderer) Validate(p profile.Profile) error {
	var problems []string
	if p.OS.Family != "debian" {
		problems = append(problems, fmt.Sprintf("os.family must be \"debian\", got %q", p.OS.Family))
	}
	if len(p.Users) == 0 {
		problems = append(problems, "at least one user is required")
	}
	if len(problems) > 0 {
		return fmt.Errorf("debian renderer: %s", strings.Join(problems, "; "))
	}
	return nil
}

// Render generates preseed.cfg, late-command.sh, and custom.ipxe.
func (r Renderer) Render(p profile.Profile, opts render.Options) (manifest.Manifest, error) {
	m := manifest.Manifest{
		ProfileName: p.Name,
		OSFamily:    "debian",
		OSVersion:   p.OS.Version,
		Renderer:    r.Name(),
	}

	preseedContent, err := renderPreseed(p, opts)
	if err != nil {
		return m, fmt.Errorf("debian: render preseed.cfg: %w", err)
	}

	lateContent, err := renderLateCommand(p)
	if err != nil {
		return m, fmt.Errorf("debian: render late-command.sh: %w", err)
	}

	ipxeContent, err := renderIPXE(opts)
	if err != nil {
		return m, fmt.Errorf("debian: render custom.ipxe: %w", err)
	}

	if opts.ServerBaseURL == "" {
		m.AddWarning("ServerBaseURL is not set: late-command.sh will not be fetched automatically; configure ServerBaseURL in BootWrangler settings")
	}

	m.AddFile("preseed.cfg", "Debian d-i preseed configuration")
	m.AddFile("late-command.sh", "Post-install configuration script")
	m.AddFile("custom.ipxe", "iPXE boot entry")

	if opts.DryRun {
		return m, nil
	}

	if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
		return m, fmt.Errorf("debian: create output directory: %w", err)
	}

	type fileSpec struct {
		name    string
		content []byte
		perm    os.FileMode
	}

	files := []fileSpec{
		{"preseed.cfg", preseedContent, 0o644},
		{"late-command.sh", lateContent, 0o755},
		{"custom.ipxe", ipxeContent, 0o644},
	}

	for _, f := range files {
		if err := os.WriteFile(filepath.Join(opts.OutDir, f.name), f.content, f.perm); err != nil {
			return m, fmt.Errorf("debian: write %s: %w", f.name, err)
		}
	}

	if err := m.ComputeChecksums(opts.OutDir); err != nil {
		return m, fmt.Errorf("debian: compute checksums: %w", err)
	}

	return m, nil
}

func init() {
	render.Register(Renderer{})
}

// ---------------------------------------------------------------------------
// Template data types
// ---------------------------------------------------------------------------

type preseedData struct {
	Locale           string
	Keyboard         string
	Hostname         string
	Timezone         string
	PartmanMethod    string
	PermitRootLogin  bool
	FirstUserName    string
	PackageNames     string
	ServerBaseURL    string
	HasServerBaseURL bool
}

type lateCommandData struct {
	Users    []profile.User
	SSH      profile.SSH
	Services profile.Services
	Scripts  []profile.Script
}

type ipxeData struct {
	ServerBaseURL    string
	HasServerBaseURL bool
}

// ---------------------------------------------------------------------------
// Templates
// ---------------------------------------------------------------------------

const preseedTmpl = `d-i debian-installer/locale string {{ .Locale }}
d-i keyboard-configuration/xkb-keymap select {{ .Keyboard }}
d-i netcfg/get_hostname string {{ .Hostname }}
d-i netcfg/get_domain string local
d-i netcfg/choose_interface select auto
d-i mirror/country string manual
d-i mirror/http/hostname string deb.debian.org
d-i mirror/http/directory string /debian
d-i mirror/http/proxy string
d-i clock-setup/utc boolean true
d-i time/zone string {{ .Timezone }}
d-i partman-auto/method string {{ .PartmanMethod }}
d-i partman-auto/choose_recipe select atomic
d-i partman/confirm_write_new_label boolean true
d-i partman/choose_partition select finish
d-i partman/confirm boolean true
d-i partman/confirm_nooverwrite boolean true
d-i passwd/root-login boolean {{ if .PermitRootLogin }}true{{ else }}false{{ end }}
d-i passwd/make-user boolean true
d-i passwd/user-fullname string {{ .FirstUserName }}
d-i passwd/username string {{ .FirstUserName }}
d-i passwd/user-password-crypted password !
d-i passwd/user-default-groups string sudo
d-i apt-setup/use_mirror boolean true
tasksel tasksel/first multiselect standard, ssh-server
d-i pkgsel/include string {{ .PackageNames }}
d-i grub-installer/only_debian boolean true
d-i grub-installer/with_other_os boolean true
d-i grub-installer/bootdev string default
d-i finish-install/reboot_in_progress note
{{ if .HasServerBaseURL -}}
d-i preseed/late_command string wget -O /tmp/late.sh {{ .ServerBaseURL }}/late-command.sh && chmod +x /tmp/late.sh && /tmp/late.sh
{{- end }}
`

const lateCommandTmpl = `#!/bin/sh
set -e
{{ range .Users }}
# --- user: {{ .Name }} ---
if ! id -u {{ .Name }} >/dev/null 2>&1; then
    useradd -m{{ if .Shell }} -s {{ .Shell }}{{ end }}{{ range .Groups }} -G {{ . }}{{ end }} {{ .Name }}
fi
{{ if .SSHKeys -}}
mkdir -p /home/{{ .Name }}/.ssh
chmod 700 /home/{{ .Name }}/.ssh
cat > /home/{{ .Name }}/.ssh/authorized_keys <<'SSHEOF'
{{ range .SSHKeys -}}
{{ . }}
{{ end -}}
SSHEOF
chmod 600 /home/{{ .Name }}/.ssh/authorized_keys
chown -R {{ .Name }}:{{ .Name }} /home/{{ .Name }}/.ssh
{{- end }}
{{ end -}}
# --- sshd configuration ---
SSHD_CONFIG=/etc/ssh/sshd_config
{{ if not .SSH.PermitRootLogin -}}
sed -i 's/^#\?PermitRootLogin .*/PermitRootLogin no/' "$SSHD_CONFIG" || echo 'PermitRootLogin no' >> "$SSHD_CONFIG"
{{ end -}}
{{ if not .SSH.PasswordAuthentication -}}
sed -i 's/^#\?PasswordAuthentication .*/PasswordAuthentication no/' "$SSHD_CONFIG" || echo 'PasswordAuthentication no' >> "$SSHD_CONFIG"
{{ end -}}
{{ if .Services.Enable -}}
# --- enable services ---
{{ range .Services.Enable -}}
systemctl enable {{ . }}
{{ end -}}
{{ end -}}
{{ if .Scripts -}}
# --- post-install scripts ---
{{ range .Scripts -}}
# script: {{ .Name }}
{{ .Content }}
{{ end -}}
{{ end -}}
`

const ipxeTmpl = `#!ipxe
{{ if .HasServerBaseURL -}}
set base-url {{ .ServerBaseURL }}
kernel ${base-url}/linux auto=true priority=critical url=${base-url}/preseed.cfg --- quiet
initrd ${base-url}/initrd.gz
boot
{{- else -}}
# configure ServerBaseURL in BootWrangler settings
{{- end }}
`

// ---------------------------------------------------------------------------
// Render helpers
// ---------------------------------------------------------------------------

var lateCommandFuncs = template.FuncMap{
	"not": func(b bool) bool { return !b },
}

func renderPreseed(p profile.Profile, opts render.Options) ([]byte, error) {
	firstUserName := ""
	if len(p.Users) > 0 {
		firstUserName = p.Users[0].Name
	}

	pkgNames := strings.Join(p.Packages.Names, " ")
	if pkgNames == "" {
		pkgNames = "openssh-server"
	}

	data := preseedData{
		Locale:           p.System.Locale,
		Keyboard:         p.System.Keyboard,
		Hostname:         p.System.Hostname,
		Timezone:         p.System.Timezone,
		PartmanMethod:    "regular",
		PermitRootLogin:  p.SSH.PermitRootLogin,
		FirstUserName:    firstUserName,
		PackageNames:     pkgNames,
		ServerBaseURL:    opts.ServerBaseURL,
		HasServerBaseURL: opts.ServerBaseURL != "",
	}

	return execTemplate("preseed.cfg", preseedTmpl, nil, data)
}

func renderLateCommand(p profile.Profile) ([]byte, error) {
	data := lateCommandData{
		Users:    p.Users,
		SSH:      p.SSH,
		Services: p.Services,
		Scripts:  p.PostInstall.Scripts,
	}
	return execTemplate("late-command.sh", lateCommandTmpl, lateCommandFuncs, data)
}

func renderIPXE(opts render.Options) ([]byte, error) {
	data := ipxeData{
		ServerBaseURL:    opts.ServerBaseURL,
		HasServerBaseURL: opts.ServerBaseURL != "",
	}
	return execTemplate("custom.ipxe", ipxeTmpl, nil, data)
}

func execTemplate(name, tmplText string, funcs template.FuncMap, data any) ([]byte, error) {
	tmpl := template.New(name)
	if funcs != nil {
		tmpl = tmpl.Funcs(funcs)
	}
	tmpl, err := tmpl.Parse(tmplText)
	if err != nil {
		return nil, fmt.Errorf("parse template %s: %w", name, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute template %s: %w", name, err)
	}
	return buf.Bytes(), nil
}
