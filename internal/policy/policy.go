// Package policy defines optional rules that provisioning profiles must satisfy.
package policy

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Policy describes a set of rules that a ProfileSnapshot must satisfy.
type Policy struct {
	RequireSSHKeyOnly     bool     `yaml:"require_ssh_key_only"`
	ForbidRootLogin       bool     `yaml:"forbid_root_login"`
	RequireDiskConfirm    bool     `yaml:"require_disk_confirm"`
	RequireUsers          []string `yaml:"require_users"`
	AllowedDistros        []string `yaml:"allowed_distros"`
	ForbidPackages        []string `yaml:"forbid_packages"`
	RequirePackagePresets []string `yaml:"require_package_presets"`
	MinUsers              int      `yaml:"min_users"`
}

// Violation records a single policy rule failure.
type Violation struct {
	Rule    string
	Message string
}

// CheckResult summarises the outcome of running policy checks against a profile.
type CheckResult struct {
	Passed     bool
	Violations []Violation
}

// ProfileSnapshot carries the profile fields needed by policy checks, avoiding
// a direct import cycle with the profile package.
type ProfileSnapshot struct {
	OSFamily               string
	SSHPasswordAuth        bool
	SSHPermitRootLogin     bool
	DiskConfirmDestructive bool
	Users                  []string
	Packages               []string
	PackagePresets         []string
}

// LoadPolicy reads and parses a YAML policy file from disk.
func LoadPolicy(path string) (Policy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Policy{}, fmt.Errorf("read policy file: %w", err)
	}
	return LoadPolicyFromBytes(data)
}

// LoadPolicyFromBytes parses a YAML-encoded policy.
func LoadPolicyFromBytes(data []byte) (Policy, error) {
	var p Policy
	if err := yaml.Unmarshal(data, &p); err != nil {
		return Policy{}, fmt.Errorf("parse policy: %w", err)
	}
	return p, nil
}

// Check evaluates all policy rules against snap and returns the combined result.
func Check(p Policy, snap ProfileSnapshot) CheckResult {
	var violations []Violation

	add := func(rule, msg string) {
		violations = append(violations, Violation{Rule: rule, Message: msg})
	}

	if p.RequireSSHKeyOnly && snap.SSHPasswordAuth {
		add("require_ssh_key_only",
			"SSH password authentication is enabled; policy requires key-only")
	}

	if p.ForbidRootLogin && snap.SSHPermitRootLogin {
		add("forbid_root_login",
			"Root SSH login is permitted; policy forbids it")
	}

	if p.RequireDiskConfirm && !snap.DiskConfirmDestructive {
		add("require_disk_confirm",
			"Disk destructive mode requires confirm_destructive: true")
	}

	if len(p.RequireUsers) > 0 {
		found := false
		for _, required := range p.RequireUsers {
			for _, u := range snap.Users {
				if u == required {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			add("require_users",
				fmt.Sprintf("Profile must include at least one of: %s",
					strings.Join(p.RequireUsers, ", ")))
		}
	}

	if len(p.AllowedDistros) > 0 {
		allowed := false
		for _, d := range p.AllowedDistros {
			if strings.EqualFold(d, snap.OSFamily) {
				allowed = true
				break
			}
		}
		if !allowed {
			add("allowed_distros",
				fmt.Sprintf("OS family %q is not in allowed distros: %s",
					snap.OSFamily, strings.Join(p.AllowedDistros, ", ")))
		}
	}

	for _, forbidden := range p.ForbidPackages {
		for _, pkg := range snap.Packages {
			if pkg == forbidden {
				add("forbid_packages",
					fmt.Sprintf("Package %q is forbidden by policy", pkg))
				break
			}
		}
	}

	for _, required := range p.RequirePackagePresets {
		found := false
		for _, preset := range snap.PackagePresets {
			if preset == required {
				found = true
				break
			}
		}
		if !found {
			add("require_package_presets",
				fmt.Sprintf("Policy requires package preset %q", required))
		}
	}

	if p.MinUsers > 0 && len(snap.Users) < p.MinUsers {
		add("min_users",
			fmt.Sprintf("Profile has %d users; policy requires at least %d",
				len(snap.Users), p.MinUsers))
	}

	return CheckResult{
		Passed:     len(violations) == 0,
		Violations: violations,
	}
}
