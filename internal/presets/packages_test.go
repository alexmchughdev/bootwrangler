package presets

import (
	"testing"
)

func TestExpandPreset_Known(t *testing.T) {
	pkgs, err := ExpandPreset("remote-admin", "ubuntu")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) == 0 {
		t.Fatal("expected non-empty package list")
	}
	// ubuntu remote-admin should include curl
	found := false
	for _, p := range pkgs {
		if p == "curl" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'curl' in remote-admin ubuntu packages, got %v", pkgs)
	}
}

func TestExpandPreset_Unknown(t *testing.T) {
	_, err := ExpandPreset("does-not-exist", "ubuntu")
	if err == nil {
		t.Fatal("expected error for unknown preset, got nil")
	}
}

func TestExpandPresets_Dedup(t *testing.T) {
	// Both minimal and remote-admin may share packages on some families;
	// craft a case that definitely overlaps: expand the same preset twice.
	pkgs, err := ExpandPresets([]string{"minimal", "minimal"}, "ubuntu")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
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

func TestAllPackagePresets_Complete(t *testing.T) {
	families := []string{"alpine", "ubuntu", "debian", "rocky", "fedora", "arch", "opensuse"}
	for _, preset := range AllPackagePresets() {
		for _, family := range families {
			pkgs, ok := preset.Packages[family]
			if !ok {
				t.Errorf("preset %q is missing entry for OS family %q", preset.Name, family)
				continue
			}
			if len(pkgs) == 0 {
				t.Errorf("preset %q has empty package list for OS family %q", preset.Name, family)
			}
		}
	}
}
