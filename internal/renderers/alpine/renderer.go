// Package alpine renders Alpine Linux setup-alpine unattended installer assets.
package alpine

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

// Renderer generates Alpine Linux installer assets.
type Renderer struct{}

// Name returns the unique renderer identifier.
func (r Renderer) Name() string { return "alpine" }

// Family returns "alpine".
func (r Renderer) Family() string { return "alpine" }

// SupportedVersions returns nil, accepting any Alpine version.
func (r Renderer) SupportedVersions() []string { return nil }

// Validate performs Alpine-specific profile validation.
func (r Renderer) Validate(p profile.Profile) error {
	if p.OS.Family != "alpine" {
		return fmt.Errorf("alpine renderer: os family must be \"alpine\", got %q", p.OS.Family)
	}
	if len(p.Users) == 0 {
		return fmt.Errorf("alpine renderer: at least one user must be defined")
	}
	return nil
}

// Render generates Alpine installer assets into opts.OutDir.
func (r Renderer) Render(p profile.Profile, opts render.Options) (manifest.Manifest, error) {
	m := manifest.Manifest{
		ProfileName: p.Name,
		OSFamily:    "alpine",
		OSVersion:   p.OS.Version,
		Renderer:    r.Name(),
	}

	answerfileContent, err := renderAnswerfile(p)
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("alpine: render answerfile: %w", err)
	}

	postInstallContent, err := renderPostInstall(p)
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("alpine: render post-install script: %w", err)
	}

	ipxeContent, err := renderIPXE(opts.ServerBaseURL)
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("alpine: render iPXE: %w", err)
	}

	m.AddFile("answerfile", "Alpine setup-alpine answerfile for unattended installation")
	m.AddFile("post-install.sh", "Post-installation script for users, SSH keys, and services")
	m.AddFile("custom.ipxe", "iPXE boot entry for network booting")

	if opts.DryRun {
		m.AddWarning("dry-run: no files written")
		return m, nil
	}

	if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
		return manifest.Manifest{}, fmt.Errorf("alpine: create output directory: %w", err)
	}

	if err := os.WriteFile(filepath.Join(opts.OutDir, "answerfile"), []byte(answerfileContent), 0o644); err != nil {
		return manifest.Manifest{}, fmt.Errorf("alpine: write answerfile: %w", err)
	}

	if err := os.WriteFile(filepath.Join(opts.OutDir, "post-install.sh"), []byte(postInstallContent), 0o755); err != nil {
		return manifest.Manifest{}, fmt.Errorf("alpine: write post-install.sh: %w", err)
	}

	if err := os.WriteFile(filepath.Join(opts.OutDir, "custom.ipxe"), []byte(ipxeContent), 0o644); err != nil {
		return manifest.Manifest{}, fmt.Errorf("alpine: write custom.ipxe: %w", err)
	}

	if err := m.ComputeChecksums(opts.OutDir); err != nil {
		return manifest.Manifest{}, fmt.Errorf("alpine: compute checksums: %w", err)
	}

	return m, nil
}

// answerfileData holds data for the answerfile template.
type answerfileData struct {
	Keyboard    string
	Hostname    string
	Timezone    string
	Interface   string
	NetworkMode string
	Address     string
	Gateway     string
	DNS         string
	RepoOpts    string
	DiskTarget  string
	DiskMode    string
	InstallMode string
}

const answerfileTmpl = `KEYMAPOPTS="{{ .Keyboard }} {{ .Keyboard }}"
HOSTNAMEOPTS="-n {{ .Hostname }}"
INTERFACESOPTS="{{ .InterfacesOpts }}"
{{ if .DNS }}DNSOPTS="-d {{ .DNS }}"
{{ end }}TIMEZONEOPTS="-z {{ .Timezone }}"
PROXYOPTS="none"
APKREPOSOPTS="{{ .RepoOpts }}"
SSHDOPTS="-c openssh"
NTPOPTS="-c chrony"
DISKOPTS="-m {{ .InstallMode }} {{ .DiskTarget }}"
LBUOPTS="none"
APKCACHEOPTS="/var/cache/apk"
`

type answerfileTmplData struct {
	Keyboard       string
	Hostname       string
	Timezone       string
	InterfacesOpts string
	DNS            string
	RepoOpts       string
	DiskTarget     string
	InstallMode    string
}

func renderAnswerfile(p profile.Profile) (string, error) {
	iface := p.Network.Interface
	if iface == "" || iface == "auto" {
		iface = "eth0"
	}

	var ifaceOpts string
	switch p.Network.Mode {
	case "static":
		ifaceOpts = fmt.Sprintf(
			`auto lo\niface lo inet loopback\n\nauto %s\niface %s inet static\n\taddress %s\n\tgateway %s\n`,
			iface, iface, p.Network.Address, p.Network.Gateway,
		)
	default: // dhcp
		ifaceOpts = fmt.Sprintf(
			`auto lo\niface lo inet loopback\n\nauto %s\niface %s inet dhcp\n\t hostname %s`,
			iface, iface, p.System.Hostname,
		)
	}

	var dnsStr string
	if len(p.Network.DNS) > 0 {
		dnsStr = strings.Join(p.Network.DNS, " ")
	}

	repoOpts := "-1"
	if p.OSSpecific.Alpine != nil && len(p.OSSpecific.Alpine.Repositories) > 0 {
		repoOpts = strings.Join(p.OSSpecific.Alpine.Repositories, " ")
	}

	installMode := "sys"
	if p.Disk.InstallMode != "" {
		installMode = p.Disk.InstallMode
	}
	if p.OSSpecific.Alpine != nil && p.OSSpecific.Alpine.InstallMode != "" {
		installMode = p.OSSpecific.Alpine.InstallMode
	}

	diskTarget := p.Disk.Target
	if diskTarget == "" || diskTarget == "auto" {
		diskTarget = "/dev/sda"
	}

	keyboard := p.System.Keyboard
	if keyboard == "" {
		keyboard = "us"
	}

	data := answerfileTmplData{
		Keyboard:       keyboard,
		Hostname:       p.System.Hostname,
		Timezone:       p.System.Timezone,
		InterfacesOpts: ifaceOpts,
		DNS:            dnsStr,
		RepoOpts:       repoOpts,
		DiskTarget:     diskTarget,
		InstallMode:    installMode,
	}

	tmpl, err := template.New("answerfile").Parse(answerfileTmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

const postInstallTmpl = `#!/bin/sh
set -eu
{{ range .Users }}
# Create user {{ .Name }}
adduser -D -s {{ .Shell }}{{ if .Groups }} -G {{ join .Groups "," }}{{ end }} {{ .Name }} || true
{{ if .SSHKeys }}
install -d -m 0700 /home/{{ .Name }}/.ssh
cat > /home/{{ .Name }}/.ssh/authorized_keys << 'SSHEOF'
{{ range .SSHKeys }}{{ . }}
{{ end }}SSHEOF
chmod 0600 /home/{{ .Name }}/.ssh/authorized_keys
chown -R {{ .Name }}:{{ .Name }} /home/{{ .Name }}/.ssh
{{ end }}{{ end }}
{{ if .Packages }}
# Install packages
apk add{{ range .Packages }} {{ . }}{{ end }}
{{ end }}
{{ if .Services }}
# Enable services
{{ range .Services }}rc-update add {{ . }} default
{{ end }}{{ end }}
{{ if .SSHConfig }}
# Configure sshd
{{ range .SSHConfig }}sed -i 's|^#\?{{ .Key }}.*|{{ .Key }} {{ .Value }}|' /etc/ssh/sshd_config
{{ end }}{{ end }}
{{ range .Scripts }}
# {{ .Name }}
{{ .Content }}
{{ end }}`

type sshConfigEntry struct {
	Key   string
	Value string
}

type postInstallData struct {
	Users     []postInstallUser
	Packages  []string
	Services  []string
	SSHConfig []sshConfigEntry
	Scripts   []profile.Script
}

type postInstallUser struct {
	Name    string
	Shell   string
	Groups  []string
	SSHKeys []string
}

func renderPostInstall(p profile.Profile) (string, error) {
	var users []postInstallUser
	for _, u := range p.Users {
		shell := u.Shell
		if shell == "" {
			shell = "/bin/sh"
		}
		users = append(users, postInstallUser{
			Name:    u.Name,
			Shell:   shell,
			Groups:  u.Groups,
			SSHKeys: u.SSHKeys,
		})
	}

	var sshConfig []sshConfigEntry
	if !p.SSH.PermitRootLogin {
		sshConfig = append(sshConfig, sshConfigEntry{Key: "PermitRootLogin", Value: "no"})
	}
	if !p.SSH.PasswordAuthentication {
		sshConfig = append(sshConfig, sshConfigEntry{Key: "PasswordAuthentication", Value: "no"})
	}

	data := postInstallData{
		Users:     users,
		Packages:  p.Packages.Names,
		Services:  p.Services.Enable,
		SSHConfig: sshConfig,
		Scripts:   p.PostInstall.Scripts,
	}

	funcMap := template.FuncMap{
		"join": strings.Join,
	}

	tmpl, err := template.New("post-install").Funcs(funcMap).Parse(postInstallTmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

const ipxeTmplWithURL = `#!ipxe
set base-url %s
kernel ${base-url}/boot/vmlinuz-lts modules=loop,squashfs quiet
initrd ${base-url}/boot/initramfs-lts
boot
`

const ipxeTmplNoURL = `#!ipxe
# configure ServerBaseURL in BootWrangler settings
kernel /boot/vmlinuz-lts modules=loop,squashfs quiet
initrd /boot/initramfs-lts
boot
`

func renderIPXE(serverBaseURL string) (string, error) {
	if serverBaseURL == "" {
		return ipxeTmplNoURL, nil
	}
	return fmt.Sprintf(ipxeTmplWithURL, serverBaseURL), nil
}

func init() {
	render.Register(Renderer{})
}
