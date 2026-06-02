// Package arch renders Arch Linux scripted install assets.
package arch

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

// Renderer generates Arch Linux scripted install assets.
type Renderer struct{}

// Name returns the unique renderer identifier.
func (r Renderer) Name() string { return "arch" }

// Family returns "arch".
func (r Renderer) Family() string { return "arch" }

// SupportedVersions returns nil, accepting any Arch version.
func (r Renderer) SupportedVersions() []string { return nil }

// Validate performs Arch-specific profile validation.
func (r Renderer) Validate(p profile.Profile) error {
	if p.OS.Family != "arch" {
		return fmt.Errorf("arch renderer: os family must be \"arch\", got %q", p.OS.Family)
	}
	if len(p.Users) == 0 {
		return fmt.Errorf("arch renderer: at least one user must be defined")
	}
	return nil
}

// Render generates Arch Linux installer assets into opts.OutDir.
func (r Renderer) Render(p profile.Profile, opts render.Options) (manifest.Manifest, error) {
	m := manifest.Manifest{
		ProfileName: p.Name,
		OSFamily:    "arch",
		OSVersion:   p.OS.Version,
		Renderer:    r.Name(),
	}

	installShContent, diskWarning, err := renderInstallSh(p)
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("arch: render install.sh: %w", err)
	}

	ipxeContent, err := renderIPXE(opts.ServerBaseURL)
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("arch: render iPXE: %w", err)
	}

	m.AddFile("install.sh", "Arch Linux scripted install script")
	m.AddFile("custom.ipxe", "iPXE boot entry for network booting")

	m.AddWarning("Arch Linux scripted install is not unattended by default; review install.sh before executing.")
	if diskWarning != "" {
		m.AddWarning(diskWarning)
	}

	if opts.DryRun {
		m.AddWarning("dry-run: no files written")
		return m, nil
	}

	if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
		return manifest.Manifest{}, fmt.Errorf("arch: create output directory: %w", err)
	}

	if err := os.WriteFile(filepath.Join(opts.OutDir, "install.sh"), []byte(installShContent), 0o755); err != nil {
		return manifest.Manifest{}, fmt.Errorf("arch: write install.sh: %w", err)
	}

	if err := os.WriteFile(filepath.Join(opts.OutDir, "custom.ipxe"), []byte(ipxeContent), 0o644); err != nil {
		return manifest.Manifest{}, fmt.Errorf("arch: write custom.ipxe: %w", err)
	}

	if err := m.ComputeChecksums(opts.OutDir); err != nil {
		return manifest.Manifest{}, fmt.Errorf("arch: compute checksums: %w", err)
	}

	return m, nil
}

// installShData holds all template variables for install.sh.
type installShData struct {
	Timezone      string
	Locale        string
	Hostname      string
	Keyboard      string
	Disk          string
	ExtraPackages string
	ExtraServices string
	Users         []installShUser
	SSHConfig     []sshConfigEntry
	Scripts       []profile.Script
}

type installShUser struct {
	Name    string
	Shell   string
	Groups  []string
	Sudo    bool
	SSHKeys []string
}

type sshConfigEntry struct {
	Key   string
	Value string
}

const installShTmpl = `#!/bin/bash
set -euo pipefail

# Network: assumed up via archiso (DHCP by default)

# Set timezone
ln -sf /usr/share/zoneinfo/{{ .Timezone }} /etc/localtime
hwclock --systohc

# Set locale
sed -i 's/^#\?{{ .Locale }}/{{ .Locale }}/' /etc/locale.gen
locale-gen
echo "LANG={{ .Locale }}" > /etc/locale.conf

# Set hostname
echo "{{ .Hostname }}" > /etc/hostname

# Set keymap
echo "KEYMAP={{ .Keyboard }}" > /etc/vconsole.conf

# Partition disk (auto mode: simple GPT layout)
# WARNING: This will DESTROY data on {{ .Disk }}
DISK={{ .Disk }}
sgdisk -Z ${DISK}
sgdisk -n 1:0:+512M -t 1:ef00 -c 1:EFI ${DISK}
sgdisk -n 2:0:0 -t 2:8300 -c 2:ROOT ${DISK}
mkfs.fat -F32 ${DISK}1
mkfs.ext4 -F ${DISK}2
mount ${DISK}2 /mnt
mkdir -p /mnt/boot/efi
mount ${DISK}1 /mnt/boot/efi

# Bootstrap base system
pacstrap /mnt base linux linux-firmware

# Generate fstab
genfstab -U /mnt >> /mnt/etc/fstab

# Chroot configuration
arch-chroot /mnt /bin/bash << 'CHROOT'
ln -sf /usr/share/zoneinfo/{{ .Timezone }} /etc/localtime
hwclock --systohc
echo "{{ .Hostname }}" > /etc/hostname
echo "LANG={{ .Locale }}" > /etc/locale.conf
echo "KEYMAP={{ .Keyboard }}" > /etc/vconsole.conf
sed -i 's/^#\?{{ .Locale }}/{{ .Locale }}/' /etc/locale.gen
locale-gen

# Install packages
pacman -Sy --noconfirm openssh{{ if .ExtraPackages }} {{ .ExtraPackages }}{{ end }}

# Enable services
systemctl enable sshd{{ if .ExtraServices }} {{ .ExtraServices }}{{ end }}

# Create users
{{ range .Users -}}
useradd -m -s {{ .Shell }}{{ if .Groups }} -G {{ join .Groups "," }}{{ end }} {{ .Name }} || true
{{ if .Sudo -}}
echo "{{ .Name }} ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/{{ .Name }}
{{ end -}}
{{ if .SSHKeys -}}
install -d -m 0700 /home/{{ .Name }}/.ssh
cat > /home/{{ .Name }}/.ssh/authorized_keys << 'SSHEOF'
{{ range .SSHKeys }}{{ . }}
{{ end -}}
SSHEOF
chmod 0600 /home/{{ .Name }}/.ssh/authorized_keys
chown -R {{ .Name }}:{{ .Name }} /home/{{ .Name }}/.ssh
{{ end -}}
{{ end }}
# SSH hardening
{{ range .SSHConfig -}}
sed -i 's|^#\?{{ .Key }}.*|{{ .Key }} {{ .Value }}|' /etc/ssh/sshd_config
{{ end }}
# Post-install scripts
{{ range .Scripts -}}
# {{ .Name }}
{{ .Content }}
{{ end }}
CHROOT

# Install bootloader
arch-chroot /mnt bootctl install
cat > /mnt/boot/loader/loader.conf << 'EOF'
default arch
timeout 4
editor no
EOF
cat > /mnt/boot/loader/entries/arch.conf << 'EOF'
title   Arch Linux
linux   /vmlinuz-linux
initrd  /initramfs-linux.img
options root=PARTUUID=$(blkid -s PARTUUID -o value ${DISK}2) rw quiet
EOF
`

func renderInstallSh(p profile.Profile) (string, string, error) {
	timezone := p.System.Timezone
	if timezone == "" {
		timezone = "UTC"
	}

	locale := p.System.Locale
	if locale == "" {
		locale = "en_GB.UTF-8"
	}

	keyboard := p.System.Keyboard
	if keyboard == "" {
		keyboard = "us"
	}

	var diskWarning string
	disk := p.Disk.Target
	if disk == "" || disk == "auto" {
		disk = "/dev/sda"
		diskWarning = "Disk target was \"auto\"; defaulting to /dev/sda — verify this is correct before running install.sh."
	}

	var extraPackages string
	if len(p.Packages.Names) > 0 {
		extraPackages = strings.Join(p.Packages.Names, " ")
	}

	var extraServices []string
	for _, svc := range p.Services.Enable {
		if svc != "sshd" {
			extraServices = append(extraServices, svc)
		}
	}
	extraServicesStr := strings.Join(extraServices, " ")

	var users []installShUser
	for _, u := range p.Users {
		shell := u.Shell
		if shell == "" {
			shell = "/bin/bash"
		}
		users = append(users, installShUser{
			Name:    u.Name,
			Shell:   shell,
			Groups:  u.Groups,
			Sudo:    u.Sudo,
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

	data := installShData{
		Timezone:      timezone,
		Locale:        locale,
		Hostname:      p.System.Hostname,
		Keyboard:      keyboard,
		Disk:          disk,
		ExtraPackages: extraPackages,
		ExtraServices: extraServicesStr,
		Users:         users,
		SSHConfig:     sshConfig,
		Scripts:       p.PostInstall.Scripts,
	}

	funcMap := template.FuncMap{
		"join": strings.Join,
	}

	tmpl, err := template.New("install.sh").Funcs(funcMap).Parse(installShTmpl)
	if err != nil {
		return "", diskWarning, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", diskWarning, err
	}

	return buf.String(), diskWarning, nil
}

const ipxeTmplWithURL = `#!ipxe
set base-url %s
# Arch Linux netboot
kernel ${base-url}/arch/boot/x86_64/vmlinuz-linux archiso_http_srv=${base-url}/ archisobasedir=arch quiet
initrd ${base-url}/arch/boot/x86_64/initramfs-linux.img
boot
`

const ipxeTmplNoURL = `#!ipxe
# configure ServerBaseURL in BootWrangler settings
# Arch Linux netboot
kernel /arch/boot/x86_64/vmlinuz-linux archisobasedir=arch quiet
initrd /arch/boot/x86_64/initramfs-linux.img
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
