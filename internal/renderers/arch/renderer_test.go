package arch

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alexmchughdev/bootwrangler/internal/profile"
	"github.com/alexmchughdev/bootwrangler/internal/render"
)

func TestRenderer_Name(t *testing.T) {
	r := Renderer{}
	if got := r.Name(); got != "arch" {
		t.Errorf("Name() = %q, want \"arch\"", got)
	}
}

func TestRenderer_Family(t *testing.T) {
	r := Renderer{}
	if got := r.Family(); got != "arch" {
		t.Errorf("Family() = %q, want \"arch\"", got)
	}
}

func TestRenderer_Validate_RejectsNoUsers(t *testing.T) {
	r := Renderer{}
	p := profile.Profile{
		OS:    profile.OS{Family: "arch"},
		Users: nil,
	}
	if err := r.Validate(p); err == nil {
		t.Error("expected error for no users, got nil")
	}
}

func TestRenderer_Validate_RejectsWrongFamily(t *testing.T) {
	r := Renderer{}
	p := profile.Profile{
		OS:    profile.OS{Family: "alpine"},
		Users: []profile.User{{Name: "alice"}},
	}
	if err := r.Validate(p); err == nil {
		t.Error("expected error for non-arch family, got nil")
	}
}

func TestRenderer_Validate_AcceptsValidProfile(t *testing.T) {
	r := Renderer{}
	p := profile.Profile{
		OS:    profile.OS{Family: "arch"},
		Users: []profile.User{{Name: "alice"}},
	}
	if err := r.Validate(p); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRenderer_Render_ArchMinimal_HostnameAndTimezone(t *testing.T) {
	p, err := profile.LoadAndValidateFile("../../../examples/profiles/arch-minimal.yaml")
	if err != nil {
		t.Fatalf("load arch-minimal profile: %v", err)
	}

	outDir := t.TempDir()
	r := Renderer{}
	m, err := r.Render(p, render.Options{OutDir: outDir})
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	if m.OSFamily != "arch" {
		t.Errorf("manifest OSFamily = %q, want \"arch\"", m.OSFamily)
	}
	if m.Renderer != "arch" {
		t.Errorf("manifest Renderer = %q, want \"arch\"", m.Renderer)
	}

	installShBytes, err := os.ReadFile(filepath.Join(outDir, "install.sh"))
	if err != nil {
		t.Fatalf("read install.sh: %v", err)
	}
	installSh := string(installShBytes)

	if !strings.Contains(installSh, "arch-minimal") {
		t.Errorf("install.sh missing hostname \"arch-minimal\":\n%s", installSh)
	}

	if !strings.Contains(installSh, "Europe/London") {
		t.Errorf("install.sh missing timezone \"Europe/London\":\n%s", installSh)
	}
}

func TestRenderer_Render_SSHKey(t *testing.T) {
	p, err := profile.LoadAndValidateFile("../../../examples/profiles/arch-minimal.yaml")
	if err != nil {
		t.Fatalf("load arch-minimal profile: %v", err)
	}

	const wantKey = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJ0EkzLnXFTKfQXv2bseD0jLxoc+R6NxRHWqnY15T9i6 example-only@bootwrangler.invalid"

	outDir := t.TempDir()
	r := Renderer{}
	_, err = r.Render(p, render.Options{OutDir: outDir})
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	installShBytes, err := os.ReadFile(filepath.Join(outDir, "install.sh"))
	if err != nil {
		t.Fatalf("read install.sh: %v", err)
	}
	if !strings.Contains(string(installShBytes), wantKey) {
		t.Errorf("install.sh missing SSH key:\n%s", string(installShBytes))
	}
}

func TestRenderer_Render_DryRun(t *testing.T) {
	p, err := profile.LoadAndValidateFile("../../../examples/profiles/arch-minimal.yaml")
	if err != nil {
		t.Fatalf("load arch-minimal profile: %v", err)
	}

	outDir := t.TempDir()
	r := Renderer{}
	m, err := r.Render(p, render.Options{OutDir: outDir, DryRun: true})
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	// Manifest should still describe two files.
	if len(m.Files) != 2 {
		t.Errorf("dry-run manifest has %d files, want 2", len(m.Files))
	}

	// Verify file names are present in manifest.
	wantFiles := map[string]bool{
		"install.sh":  false,
		"custom.ipxe": false,
	}
	for _, f := range m.Files {
		wantFiles[f.Path] = true
	}
	for name, found := range wantFiles {
		if !found {
			t.Errorf("dry-run manifest missing file %q", name)
		}
	}

	// No files should have been written.
	for _, f := range m.Files {
		fullPath := filepath.Join(outDir, f.Path)
		if _, err := os.Stat(fullPath); err == nil {
			t.Errorf("dry-run wrote file %q, expected no files written", fullPath)
		}
	}

	// Dry-run manifest should contain a warning.
	if len(m.Warnings) == 0 {
		t.Error("dry-run manifest should have at least one warning")
	}
}

func TestRenderer_Validate_RejectsNoUsers_Standalone(t *testing.T) {
	r := Renderer{}
	p := profile.Profile{
		OS:    profile.OS{Family: "arch"},
		Users: []profile.User{},
	}
	if err := r.Validate(p); err == nil {
		t.Error("Validate should reject profile with no users")
	}
}

func TestRenderer_Render_ManifestWarning(t *testing.T) {
	p, err := profile.LoadAndValidateFile("../../../examples/profiles/arch-minimal.yaml")
	if err != nil {
		t.Fatalf("load arch-minimal profile: %v", err)
	}

	outDir := t.TempDir()
	r := Renderer{}
	m, err := r.Render(p, render.Options{OutDir: outDir})
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	const wantSubstr = "not unattended"
	for _, w := range m.Warnings {
		if strings.Contains(w, wantSubstr) {
			return
		}
	}
	t.Errorf("manifest warnings do not contain %q; got: %v", wantSubstr, m.Warnings)
}
