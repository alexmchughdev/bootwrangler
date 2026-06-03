// Package presets expands named preset strings into OS-family-specific package lists.
package presets

import (
	"fmt"
	"strings"
)

// PackagePreset describes one named set of packages, keyed by OS family.
type PackagePreset struct {
	Name        string
	Description string
	Packages    map[string][]string // key: os family, value: package list
}

// allPackagePresets is the master list of built-in package presets.
var allPackagePresets = []PackagePreset{
	{
		Name:        "minimal",
		Description: "Minimal base packages for a bootable, remotely accessible system",
		Packages: map[string][]string{
			"alpine":   {"busybox", "openssh"},
			"ubuntu":   {"openssh-server", "ca-certificates"},
			"debian":   {"openssh-server", "ca-certificates"},
			"rocky":    {"openssh-server", "ca-certificates"},
			"fedora":   {"openssh-server", "ca-certificates"},
			"arch":     {"openssh", "ca-certificates"},
			"opensuse": {"openssh", "ca-certificates"},
		},
	},
	{
		Name:        "remote-admin",
		Description: "Common remote administration and system inspection utilities",
		Packages: map[string][]string{
			"alpine":   {"curl", "wget", "nano", "htop", "tmux"},
			"ubuntu":   {"curl", "wget", "vim", "htop", "tmux"},
			"debian":   {"curl", "wget", "vim", "htop", "tmux"},
			"rocky":    {"curl", "wget", "vim", "htop", "tmux"},
			"fedora":   {"curl", "wget", "vim", "htop", "tmux"},
			"arch":     {"curl", "wget", "vim", "htop", "tmux"},
			"opensuse": {"curl", "wget", "vim", "htop", "tmux"},
		},
	},
	{
		Name:        "vm-guest",
		Description: "Guest agent for hypervisor integration",
		Packages: map[string][]string{
			"alpine":   {"qemu-guest-agent"},
			"ubuntu":   {"qemu-guest-agent"},
			"debian":   {"qemu-guest-agent"},
			"rocky":    {"qemu-guest-agent"},
			"fedora":   {"qemu-guest-agent"},
			"arch":     {"qemu-guest-agent"},
			"opensuse": {"qemu-guest-agent"},
		},
	},
	{
		Name:        "container-host",
		Description: "Container runtime and CLI tools",
		Packages: map[string][]string{
			"alpine":   {"docker", "docker-cli"},
			"ubuntu":   {"docker.io", "docker-compose"},
			"debian":   {"docker.io", "docker-compose"},
			"rocky":    {"docker-ce", "docker-ce-cli"},
			"fedora":   {"docker-ce", "docker-ce-cli"},
			"arch":     {"docker", "docker-compose"},
			"opensuse": {"docker", "docker-compose"},
		},
	},
	{
		Name:        "developer",
		Description: "Common development and build tools",
		Packages: map[string][]string{
			"alpine":   {"git", "build-base", "go", "python3"},
			"ubuntu":   {"git", "build-essential", "golang", "python3"},
			"debian":   {"git", "build-essential", "golang", "python3"},
			"rocky":    {"git", "gcc", "make", "golang", "python3"},
			"fedora":   {"git", "gcc", "make", "golang", "python3"},
			"arch":     {"git", "base-devel", "go", "python"},
			"opensuse": {"git", "gcc", "make", "go", "python3"},
		},
	},
	{
		Name:        "security-basic",
		Description: "Basic host firewall and intrusion prevention",
		Packages: map[string][]string{
			"alpine":   {"fail2ban"},
			"ubuntu":   {"ufw", "fail2ban"},
			"debian":   {"ufw", "fail2ban"},
			"rocky":    {"firewalld", "fail2ban"},
			"fedora":   {"firewalld", "fail2ban"},
			"arch":     {"ufw", "fail2ban"},
			"opensuse": {"firewalld", "fail2ban"},
		},
	},
}

// AllPackagePresets returns all defined package presets.
func AllPackagePresets() []PackagePreset {
	result := make([]PackagePreset, len(allPackagePresets))
	copy(result, allPackagePresets)
	return result
}

// ExpandPreset returns the package list for the given preset name and OS family.
// An error is returned if the preset name is unknown.
func ExpandPreset(presetName, osFamily string) ([]string, error) {
	for _, p := range allPackagePresets {
		if p.Name == presetName {
			pkgs, ok := p.Packages[strings.ToLower(osFamily)]
			if !ok {
				return nil, fmt.Errorf("preset %q has no entry for OS family %q", presetName, osFamily)
			}
			result := make([]string, len(pkgs))
			copy(result, pkgs)
			return result, nil
		}
	}
	return nil, fmt.Errorf("unknown package preset %q", presetName)
}

// ExpandPresets expands multiple presets for the given OS family and deduplicates
// the resulting package list. An error is returned if any preset name is unknown.
func ExpandPresets(presetNames []string, osFamily string) ([]string, error) {
	seen := make(map[string]struct{})
	var result []string
	for _, name := range presetNames {
		pkgs, err := ExpandPreset(name, osFamily)
		if err != nil {
			return nil, err
		}
		for _, pkg := range pkgs {
			if _, exists := seen[pkg]; !exists {
				seen[pkg] = struct{}{}
				result = append(result, pkg)
			}
		}
	}
	return result, nil
}

// ValidatePreset returns an error if name does not correspond to a known preset.
func ValidatePreset(name string) error {
	for _, p := range allPackagePresets {
		if p.Name == name {
			return nil
		}
	}
	return fmt.Errorf("unknown package preset %q", name)
}
