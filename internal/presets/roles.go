package presets

import "fmt"

// RolePreset combines package presets, services, groups, and advisory notes into
// a single named role that can be applied to a provisioning profile.
type RolePreset struct {
	Name        string
	Description string
	Packages    []string // preset names or explicit package names
	Services    []string // service names to enable
	Groups      []string // system groups to add users to
	Notes       []string // human-readable notes shown in the GUI
	Warning     string   // shown to the operator before applying the role
}

// allRolePresets is the master list of built-in role presets.
var allRolePresets = []RolePreset{
	{
		Name:        "minimal-ssh-server",
		Description: "Minimal installation with SSH access and basic admin tools",
		Packages:    []string{"minimal", "remote-admin"},
		Services:    []string{"ssh", "sshd"},
		Notes:       []string{"Installs SSH server with admin tools"},
	},
	{
		Name:        "docker-host",
		Description: "Docker container host with SSH access and VM guest integration",
		Packages:    []string{"minimal", "remote-admin", "container-host", "vm-guest"},
		Services:    []string{"docker", "ssh", "sshd"},
		Groups:      []string{"docker"},
		Notes:       []string{"Installs Docker and adds users to docker group"},
		Warning:     "Adding users to the docker group grants root-equivalent access",
	},
	{
		Name:        "kubernetes-node",
		Description: "Base Kubernetes node with container runtime",
		Packages:    []string{"minimal", "remote-admin", "container-host"},
		Services:    []string{"ssh", "sshd"},
		Notes: []string{
			"Base Kubernetes node — install kubeadm separately",
			"Swap must be disabled",
		},
	},
	{
		Name:        "proxmox-guest",
		Description: "Optimised guest image for running inside Proxmox VE",
		Packages:    []string{"minimal", "remote-admin", "vm-guest"},
		Services:    []string{"ssh", "sshd", "qemu-guest-agent"},
		Notes:       []string{"Optimised for running inside Proxmox VE"},
	},
	{
		Name:        "edge-agent",
		Description: "Minimal edge compute node with container runtime",
		Packages:    []string{"minimal", "remote-admin", "container-host"},
		Services:    []string{"docker", "ssh", "sshd"},
		Notes:       []string{"Minimal edge compute node with container runtime"},
	},
	{
		Name:        "security-lab",
		Description: "Security testing and research base image",
		Packages:    []string{"minimal", "remote-admin", "security-basic", "developer"},
		Services:    []string{"ssh", "sshd"},
		Notes:       []string{"Security testing and research base image"},
		Warning:     "Review installed security tools before deploying",
	},
	{
		Name:        "developer-workstation",
		Description: "Development environment with common build tools",
		Packages:    []string{"minimal", "remote-admin", "developer"},
		Services:    []string{"ssh", "sshd"},
		Notes:       []string{"Development environment with common build tools"},
	},
	{
		Name:        "ci-runner",
		Description: "CI/CD runner with Docker and build tools",
		Packages:    []string{"minimal", "remote-admin", "developer", "container-host"},
		Services:    []string{"docker", "ssh", "sshd"},
		Notes:       []string{"CI/CD runner with Docker and build tools"},
	},
}

// AllRolePresets returns all defined role presets.
func AllRolePresets() []RolePreset {
	result := make([]RolePreset, len(allRolePresets))
	copy(result, allRolePresets)
	return result
}

// GetRole returns the RolePreset with the given name, or an error if not found.
func GetRole(name string) (RolePreset, error) {
	for _, r := range allRolePresets {
		if r.Name == name {
			return r, nil
		}
	}
	return RolePreset{}, fmt.Errorf("unknown role preset %q", name)
}

// ExpandRole expands the role's package preset names into concrete package names
// for the given OS family, deduplicating the result.
// It also returns the role's services and groups unchanged.
func ExpandRole(role RolePreset, osFamily string) (packages []string, services []string, groups []string, err error) {
	pkgs, err := ExpandPresets(role.Packages, osFamily)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("expand role %q: %w", role.Name, err)
	}
	svcs := make([]string, len(role.Services))
	copy(svcs, role.Services)
	grps := make([]string, len(role.Groups))
	copy(grps, role.Groups)
	return pkgs, svcs, grps, nil
}

// ValidateRole returns an error if name does not correspond to a known role preset.
func ValidateRole(name string) error {
	for _, r := range allRolePresets {
		if r.Name == name {
			return nil
		}
	}
	return fmt.Errorf("unknown role preset %q", name)
}
