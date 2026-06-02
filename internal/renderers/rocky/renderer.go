// Package rocky renders Rocky Linux Anaconda kickstart installer assets.
package rocky

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

// Renderer generates Rocky Linux kickstart assets.
type Renderer struct{}

// Name returns the unique renderer identifier.
func (r Renderer) Name() string { return "rocky" }

// Family returns "rocky".
func (r Renderer) Family() string { return "rocky" }

// SupportedVersions returns nil, accepting any Rocky version.
func (r Renderer) SupportedVersions() []string { return nil }

// Validate performs Rocky-specific profile validation.
func (r Renderer) Validate(p profile.Profile) error {
	if p.OS.Family != "rocky" {
		return fmt.Errorf("rocky renderer: os family must be \"rocky\", got %q", p.OS.Family)
	}
	if len(p.Users) == 0 {
		return fmt.Errorf("rocky renderer: at least one user must be defined")
	}
	return nil
}

// Render generates Rocky Linux kickstart assets into opts.OutDir.
func (r Renderer) Render(p profile.Profile, opts render.Options) (manifest.Manifest, error) {
	m := manifest.Manifest{
		ProfileName: p.Name,
		OSFamily:    "rocky",
		OSVersion:   p.OS.Version,
		Renderer:    r.Name(),
	}

	ksContent, err := renderKickstart(p)
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("rocky: render kickstart: %w", err)
	}

	ipxeContent, err := renderIPXE(opts.ServerBaseURL)
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("rocky: render iPXE: %w", err)
	}

	m.AddFile("ks.cfg", "Anaconda kickstart configuration for unattended installation")
	m.AddFile("custom.ipxe", "iPXE boot entry for network booting")

	if opts.DryRun {
		m.AddWarning("dry-run: no files written")
		return m, nil
	}

	if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
		return manifest.Manifest{}, fmt.Errorf("rocky: create output directory: %w", err)
	}

	if err := os.WriteFile(filepath.Join(opts.OutDir, "ks.cfg"), []byte(ksContent), 0o644); err != nil {
		return manifest.Manifest{}, fmt.Errorf("rocky: write ks.cfg: %w", err)
	}

	if err := os.WriteFile(filepath.Join(opts.OutDir, "custom.ipxe"), []byte(ipxeContent), 0o644); err != nil {
		return manifest.Manifest{}, fmt.Errorf("rocky: write custom.ipxe: %w", err)
	}

	if err := m.ComputeChecksums(opts.OutDir); err != nil {
		return manifest.Manifest{}, fmt.Errorf("rocky: compute checksums: %w", err)
	}

	return m, nil
}

// ksUser holds per-user data for the kickstart %post template.
type ksUser struct {
	Name    string
	Shell   string
	Groups  []string
	SSHKeys []string
}

// ksData holds all data passed to the kickstart template.
type ksData struct {
	Locale                   string
	Keyboard                 string
	Timezone                 string
	Hostname                 string
	Services                 string
	SELinux                  string
	FirewallDisabled         bool
	Packages                 []string
	Users                    []ksUser
	PermitRootLoginNo        bool
	PasswordAuthenticationNo bool
	Scripts                  []profile.Script
}

const kickstartTmpl = `#version=RHEL9
lang {{ .Locale }}
keyboard --vckeymap={{ .Keyboard }}
timezone {{ .Timezone }}
network --bootproto=dhcp --device=link --hostname={{ .Hostname }} --activate
rootpw --lock
services --enabled={{ .Services }}
selinux --{{ .SELinux }}
{{ if .FirewallDisabled }}firewall --disabled
{{ else }}firewall --enabled --service=ssh
{{ end }}bootloader --append="rhgb quiet" --location=mbr
zerombr
clearpart --all --initlabel
autopart

%packages
@^minimal-environment
openssh-server
{{ range .Packages }}{{ . }}
{{ end }}%end

%post
{{ range .Users }}
# Create user {{ .Name }}
useradd -m -s {{ .Shell }}{{ if .Groups }} -G {{ join .Groups "," }}{{ end }} {{ .Name }} || true
{{ if .SSHKeys }}mkdir -p /home/{{ .Name }}/.ssh
chmod 700 /home/{{ .Name }}/.ssh
cat >> /home/{{ .Name }}/.ssh/authorized_keys << 'EOF'
{{ range .SSHKeys }}{{ . }}
{{ end }}EOF
chmod 600 /home/{{ .Name }}/.ssh/authorized_keys
chown -R {{ .Name }}:{{ .Name }} /home/{{ .Name }}/.ssh
{{ end }}{{ end }}
{{ if .PermitRootLoginNo }}sed -i 's/^#\?PermitRootLogin.*/PermitRootLogin no/' /etc/ssh/sshd_config
{{ end }}{{ if .PasswordAuthenticationNo }}sed -i 's/^#\?PasswordAuthentication.*/PasswordAuthentication no/' /etc/ssh/sshd_config
{{ end }}{{ range .Scripts }}
# {{ .Name }}
{{ .Content }}
{{ end }}%end
`

func renderKickstart(p profile.Profile) (string, error) {
	locale := p.System.Locale
	if locale == "" {
		locale = "en_GB.UTF-8"
	}

	keyboard := p.System.Keyboard
	if keyboard == "" {
		keyboard = "gb"
	}

	timezone := p.System.Timezone
	if timezone == "" {
		timezone = "UTC"
	}

	hostname := p.System.Hostname
	if hostname == "" {
		hostname = "localhost"
	}

	services := "sshd,chronyd"
	if len(p.Services.Enable) > 0 {
		services = strings.Join(p.Services.Enable, ",")
	}

	selinux := "enforcing"
	firewallDisabled := false
	if p.OSSpecific.RHELLike != nil {
		if p.OSSpecific.RHELLike.SELinux != "" {
			selinux = p.OSSpecific.RHELLike.SELinux
		}
		if p.OSSpecific.RHELLike.Firewall == "disabled" {
			firewallDisabled = true
		}
	}

	var users []ksUser
	for _, u := range p.Users {
		shell := u.Shell
		if shell == "" {
			shell = "/bin/bash"
		}
		users = append(users, ksUser{
			Name:    u.Name,
			Shell:   shell,
			Groups:  u.Groups,
			SSHKeys: u.SSHKeys,
		})
	}

	data := ksData{
		Locale:                   locale,
		Keyboard:                 keyboard,
		Timezone:                 timezone,
		Hostname:                 hostname,
		Services:                 services,
		SELinux:                  selinux,
		FirewallDisabled:         firewallDisabled,
		Packages:                 p.Packages.Names,
		Users:                    users,
		PermitRootLoginNo:        !p.SSH.PermitRootLogin,
		PasswordAuthenticationNo: !p.SSH.PasswordAuthentication,
		Scripts:                  p.PostInstall.Scripts,
	}

	funcMap := template.FuncMap{
		"join": strings.Join,
	}

	tmpl, err := template.New("kickstart").Funcs(funcMap).Parse(kickstartTmpl)
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
kernel ${base-url}/images/pxeboot/vmlinuz inst.ks=${base-url}/ks.cfg inst.repo=${base-url} quiet
initrd ${base-url}/images/pxeboot/initrd.img
boot
`

const ipxeTmplNoURL = `#!ipxe
# configure ServerBaseURL in BootWrangler settings
kernel /images/pxeboot/vmlinuz inst.ks=/ks.cfg inst.repo=/ quiet
initrd /images/pxeboot/initrd.img
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
