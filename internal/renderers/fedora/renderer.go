// Package fedora renders Fedora Server Anaconda kickstart installer assets.
package fedora

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

// Renderer generates Fedora Server kickstart assets.
type Renderer struct{}

// Name returns the unique renderer identifier.
func (r Renderer) Name() string { return "fedora" }

// Family returns "fedora".
func (r Renderer) Family() string { return "fedora" }

// SupportedVersions returns nil, accepting any Fedora version.
func (r Renderer) SupportedVersions() []string { return nil }

// Validate performs Fedora-specific profile validation.
func (r Renderer) Validate(p profile.Profile) error {
	if p.OS.Family != "fedora" {
		return fmt.Errorf("fedora renderer: os family must be \"fedora\", got %q", p.OS.Family)
	}
	if len(p.Users) == 0 {
		return fmt.Errorf("fedora renderer: at least one user must be defined")
	}
	return nil
}

// Render generates Fedora kickstart installer assets into opts.OutDir.
func (r Renderer) Render(p profile.Profile, opts render.Options) (manifest.Manifest, error) {
	m := manifest.Manifest{
		ProfileName: p.Name,
		OSFamily:    "fedora",
		OSVersion:   p.OS.Version,
		Renderer:    r.Name(),
	}

	ksContent, err := renderKickstart(p)
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("fedora: render kickstart: %w", err)
	}

	ipxeContent, err := renderIPXE(opts.ServerBaseURL)
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("fedora: render iPXE: %w", err)
	}

	m.AddFile("ks.cfg", "Anaconda kickstart file for unattended Fedora Server installation")
	m.AddFile("custom.ipxe", "iPXE boot entry for network booting")
	m.AddWarning("Fedora kickstart targets the current Fedora release; pin the version in your PXE/ISO environment")

	if opts.DryRun {
		m.AddWarning("dry-run: no files written")
		return m, nil
	}

	if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
		return manifest.Manifest{}, fmt.Errorf("fedora: create output directory: %w", err)
	}

	if err := os.WriteFile(filepath.Join(opts.OutDir, "ks.cfg"), []byte(ksContent), 0o644); err != nil {
		return manifest.Manifest{}, fmt.Errorf("fedora: write ks.cfg: %w", err)
	}

	if err := os.WriteFile(filepath.Join(opts.OutDir, "custom.ipxe"), []byte(ipxeContent), 0o644); err != nil {
		return manifest.Manifest{}, fmt.Errorf("fedora: write custom.ipxe: %w", err)
	}

	if err := m.ComputeChecksums(opts.OutDir); err != nil {
		return manifest.Manifest{}, fmt.Errorf("fedora: compute checksums: %w", err)
	}

	return m, nil
}

// kickstartData holds data for the ks.cfg template.
type kickstartData struct {
	Locale             string
	Keyboard           string
	Timezone           string
	Hostname           string
	NetworkBootproto   string
	NetworkDevice      string
	SELinux            string
	Firewall           string
	BootloaderAppend   string
	Users              []ksUser
	SSHHardening       []sshConfigEntry
	PostInstallScripts []profile.Script
	DNFPackages        []string
}

// ksUser holds per-user data for the kickstart %post section.
type ksUser struct {
	Name    string
	Shell   string
	Groups  []string
	SSHKeys []string
}

// sshConfigEntry holds one sshd_config key/value pair.
type sshConfigEntry struct {
	Key   string
	Value string
}

const kickstartTmpl = `#version=Fedora
lang {{ .Locale }}
keyboard --vckeymap={{ .Keyboard }}
timezone {{ .Timezone }}
network --bootproto={{ .NetworkBootproto }} --device={{ .NetworkDevice }} --hostname={{ .Hostname }} --activate
rootpw --lock
services --enabled=sshd
firewall --{{ .Firewall }} --service=ssh
selinux --{{ .SELinux }}
bootloader --append="{{ .BootloaderAppend }}" --location=mbr
zerombr
clearpart --all --initlabel
autopart
%packages
@^server-product-environment
openssh-server
{{ range .DNFPackages }}{{ . }}
{{ end }}%end

%post
{{ range .Users }}
useradd -m -s {{ .Shell }}{{ if .Groups }} -G {{ join .Groups "," }}{{ end }} {{ .Name }}
{{ if .SSHKeys }}mkdir -p /home/{{ .Name }}/.ssh
chmod 700 /home/{{ .Name }}/.ssh
cat >> /home/{{ .Name }}/.ssh/authorized_keys << 'EOF'
{{ range .SSHKeys }}{{ . }}
{{ end }}EOF
chmod 600 /home/{{ .Name }}/.ssh/authorized_keys
chown -R {{ .Name }}:{{ .Name }} /home/{{ .Name }}/.ssh
{{ end }}{{ end }}{{ range .SSHHardening }}sed -i 's|^#\?{{ .Key }}.*|{{ .Key }} {{ .Value }}|' /etc/ssh/sshd_config
{{ end }}{{ range .PostInstallScripts }}
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
		keyboard = "us"
	}

	timezone := p.System.Timezone
	if timezone == "" {
		timezone = "UTC"
	}

	networkBootproto := "dhcp"
	if p.Network.Mode == "static" {
		networkBootproto = "static"
	}

	networkDevice := p.Network.Interface
	if networkDevice == "" || networkDevice == "auto" {
		networkDevice = "link"
	}

	selinux := "enforcing"
	firewall := "enabled"
	if p.OSSpecific.RHELLike != nil {
		if p.OSSpecific.RHELLike.SELinux != "" {
			selinux = p.OSSpecific.RHELLike.SELinux
		}
		if p.OSSpecific.RHELLike.Firewall != "" {
			firewall = p.OSSpecific.RHELLike.Firewall
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

	var sshHardening []sshConfigEntry
	if !p.SSH.PermitRootLogin {
		sshHardening = append(sshHardening, sshConfigEntry{Key: "PermitRootLogin", Value: "no"})
	}
	if !p.SSH.PasswordAuthentication {
		sshHardening = append(sshHardening, sshConfigEntry{Key: "PasswordAuthentication", Value: "no"})
	}

	data := kickstartData{
		Locale:             locale,
		Keyboard:           keyboard,
		Timezone:           timezone,
		Hostname:           p.System.Hostname,
		NetworkBootproto:   networkBootproto,
		NetworkDevice:      networkDevice,
		SELinux:            selinux,
		Firewall:           firewall,
		BootloaderAppend:   "rhgb quiet",
		Users:              users,
		SSHHardening:       sshHardening,
		PostInstallScripts: p.PostInstall.Scripts,
		DNFPackages:        p.Packages.Names,
	}

	funcMap := template.FuncMap{
		"join": strings.Join,
	}

	tmpl, err := template.New("ks.cfg").Funcs(funcMap).Parse(kickstartTmpl)
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
