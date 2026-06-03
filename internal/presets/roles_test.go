package presets

import (
	"testing"
)

func TestGetRole_Known(t *testing.T) {
	role, err := GetRole("docker-host")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if role.Name != "docker-host" {
		t.Errorf("expected name 'docker-host', got %q", role.Name)
	}
}

func TestGetRole_Unknown(t *testing.T) {
	_, err := GetRole("does-not-exist")
	if err == nil {
		t.Fatal("expected error for unknown role, got nil")
	}
}

func TestExpandRole_DockerHost(t *testing.T) {
	role, err := GetRole("docker-host")
	if err != nil {
		t.Fatalf("GetRole: %v", err)
	}

	pkgs, svcs, groups, err := ExpandRole(role, "ubuntu")
	if err != nil {
		t.Fatalf("ExpandRole: %v", err)
	}

	if len(pkgs) == 0 {
		t.Error("expected non-empty package list")
	}

	// ubuntu docker-host should include docker.io (from container-host preset)
	foundDocker := false
	for _, p := range pkgs {
		if p == "docker.io" {
			foundDocker = true
			break
		}
	}
	if !foundDocker {
		t.Errorf("expected 'docker.io' in docker-host ubuntu packages, got %v", pkgs)
	}

	// should enable docker service
	foundDockerSvc := false
	for _, s := range svcs {
		if s == "docker" {
			foundDockerSvc = true
			break
		}
	}
	if !foundDockerSvc {
		t.Errorf("expected 'docker' in services, got %v", svcs)
	}

	// should include docker group
	foundGroup := false
	for _, g := range groups {
		if g == "docker" {
			foundGroup = true
			break
		}
	}
	if !foundGroup {
		t.Errorf("expected 'docker' in groups, got %v", groups)
	}

	// packages must be deduplicated
	seen := make(map[string]int)
	for _, p := range pkgs {
		seen[p]++
	}
	for pkg, count := range seen {
		if count > 1 {
			t.Errorf("duplicate package %q appeared %d times", pkg, count)
		}
	}
}

func TestAllRoles_HaveDescriptions(t *testing.T) {
	for _, role := range AllRolePresets() {
		if role.Description == "" {
			t.Errorf("role %q has no description", role.Name)
		}
	}
}
