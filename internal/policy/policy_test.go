package policy

import (
	"strings"
	"testing"
)

func baseSnap() ProfileSnapshot {
	return ProfileSnapshot{
		OSFamily:               "ubuntu",
		SSHPasswordAuth:        false,
		SSHPermitRootLogin:     false,
		DiskConfirmDestructive: true,
		Users:                  []string{"admin"},
		Packages:               []string{"curl", "vim"},
		PackagePresets:         []string{"minimal"},
	}
}

func TestCheck_Pass(t *testing.T) {
	p := Policy{
		RequireSSHKeyOnly:     true,
		ForbidRootLogin:       true,
		RequireDiskConfirm:    true,
		AllowedDistros:        []string{"ubuntu", "debian"},
		MinUsers:              1,
		RequirePackagePresets: []string{"minimal"},
	}
	result := Check(p, baseSnap())
	if !result.Passed {
		t.Errorf("expected pass, got violations: %v", result.Violations)
	}
}

func TestCheck_SSHPasswordFail(t *testing.T) {
	p := Policy{RequireSSHKeyOnly: true}
	snap := baseSnap()
	snap.SSHPasswordAuth = true

	result := Check(p, snap)
	if result.Passed {
		t.Fatal("expected failure, got pass")
	}
	if !containsRule(result.Violations, "require_ssh_key_only") {
		t.Errorf("expected require_ssh_key_only violation, got %v", result.Violations)
	}
}

func TestCheck_RootLoginFail(t *testing.T) {
	p := Policy{ForbidRootLogin: true}
	snap := baseSnap()
	snap.SSHPermitRootLogin = true

	result := Check(p, snap)
	if result.Passed {
		t.Fatal("expected failure, got pass")
	}
	if !containsRule(result.Violations, "forbid_root_login") {
		t.Errorf("expected forbid_root_login violation, got %v", result.Violations)
	}
}

func TestCheck_AllowedDistroFail(t *testing.T) {
	p := Policy{AllowedDistros: []string{"debian", "rocky"}}
	snap := baseSnap()
	snap.OSFamily = "arch"

	result := Check(p, snap)
	if result.Passed {
		t.Fatal("expected failure, got pass")
	}
	if !containsRule(result.Violations, "allowed_distros") {
		t.Errorf("expected allowed_distros violation, got %v", result.Violations)
	}
}

func TestCheck_ForbidPackageFail(t *testing.T) {
	p := Policy{ForbidPackages: []string{"telnet", "vim"}}
	snap := baseSnap()
	// snap.Packages includes "vim"

	result := Check(p, snap)
	if result.Passed {
		t.Fatal("expected failure, got pass")
	}
	if !containsRule(result.Violations, "forbid_packages") {
		t.Errorf("expected forbid_packages violation, got %v", result.Violations)
	}
}

func TestCheck_MinUsersFail(t *testing.T) {
	p := Policy{MinUsers: 3}
	snap := baseSnap()
	snap.Users = []string{"admin"}

	result := Check(p, snap)
	if result.Passed {
		t.Fatal("expected failure, got pass")
	}
	if !containsRule(result.Violations, "min_users") {
		t.Errorf("expected min_users violation, got %v", result.Violations)
	}
	if !strings.Contains(result.Violations[0].Message, "1") {
		t.Errorf("expected message to mention user count, got %q", result.Violations[0].Message)
	}
}

func TestLoadPolicyFromBytes(t *testing.T) {
	raw := []byte(`
require_ssh_key_only: true
forbid_root_login: true
allowed_distros:
  - ubuntu
  - debian
min_users: 2
`)
	p, err := LoadPolicyFromBytes(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !p.RequireSSHKeyOnly {
		t.Error("expected RequireSSHKeyOnly to be true")
	}
	if !p.ForbidRootLogin {
		t.Error("expected ForbidRootLogin to be true")
	}
	if len(p.AllowedDistros) != 2 {
		t.Errorf("expected 2 allowed distros, got %d", len(p.AllowedDistros))
	}
	if p.MinUsers != 2 {
		t.Errorf("expected MinUsers=2, got %d", p.MinUsers)
	}
}

// containsRule returns true if any violation has the given rule name.
func containsRule(violations []Violation, rule string) bool {
	for _, v := range violations {
		if v.Rule == rule {
			return true
		}
	}
	return false
}
