package rocky

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
	if got := r.Name(); got != "rocky" {
		t.Errorf("Name() = %q, want \"rocky\"", got)
	}
}

func TestRenderer_Family(t *testing.T) {
	r := Renderer{}
	if got := r.Family(); got != "rocky" {
		t.Errorf("Family() = %q, want \"rocky\"", got)
	}
}

func TestRenderer_Validate_RejectsNonRockyFamily(t *testing.T) {
	r := Renderer{}
	p := profile.Profile{
		OS:    profile.OS{Family: "ubuntu"},
		Users: []profile.User{{Name: "alice"}},
	}
	if err := r.Validate(p); err == nil {
		t.Error("expected error for non-rocky family, got nil")
	}
}

func TestRenderer_Validate_RejectsNoUsers(t *testing.T) {
	r := Renderer{}
	p := profile.Profile{
		OS:    profile.OS{Family: "rocky"},
		Users: nil,
	}
	if err := r.Validate(p); err == nil {
		t.Error("expected error for no users, got nil")
	}
}

func TestRenderer_Validate_AcceptsValidProfile(t *testing.T) {
	r := Renderer{}
	p := profile.Profile{
		OS:    profile.OS{Family: "rocky"},
		Users: []profile.User{{Name: "alice"}},
	}
	if err := r.Validate(p); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRenderer_Render_RockyServer(t *testing.T) {
	p, err := profile.LoadAndValidateFile("../../../examples/profiles/rocky-server.yaml")
	if err != nil {
		t.Fatalf("load rocky-server profile: %v", err)
	}

	outDir := t.TempDir()
	r := Renderer{}
	m, err := r.Render(p, render.Options{OutDir: outDir})
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	if m.OSFamily != "rocky" {
		t.Errorf("manifest OSFamily = %q, want \"rocky\"", m.OSFamily)
	}
	if m.Renderer != "rocky" {
		t.Errorf("manifest Renderer = %q, want \"rocky\"", m.Renderer)
	}
	if len(m.Files) != 2 {
		t.Errorf("manifest has %d files, want 2", len(m.Files))
	}

	ksBytes, err := os.ReadFile(filepath.Join(outDir, "ks.cfg"))
	if err != nil {
		t.Fatalf("read ks.cfg: %v", err)
	}
	ks := string(ksBytes)

	// Verify hostname appears in ks.cfg
	if !strings.Contains(ks, "rocky-server") {
		t.Errorf("ks.cfg missing hostname \"rocky-server\":\n%s", ks)
	}

	// Verify timezone appears in ks.cfg
	if !strings.Contains(ks, "Europe/London") {
		t.Errorf("ks.cfg missing timezone \"Europe/London\":\n%s", ks)
	}

	// Verify at least one package appears in ks.cfg
	if !strings.Contains(ks, "git") {
		t.Errorf("ks.cfg missing package \"git\":\n%s", ks)
	}
}

func TestRenderer_Render_SSHKey(t *testing.T) {
	p, err := profile.LoadAndValidateFile("../../../examples/profiles/rocky-server.yaml")
	if err != nil {
		t.Fatalf("load rocky-server profile: %v", err)
	}

	const wantKey = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJ0EkzLnXFTKfQXv2bseD0jLxoc+R6NxRHWqnY15T9i6 example-only@bootwrangler.invalid"

	outDir := t.TempDir()
	r := Renderer{}
	_, err = r.Render(p, render.Options{OutDir: outDir})
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	ksBytes, err := os.ReadFile(filepath.Join(outDir, "ks.cfg"))
	if err != nil {
		t.Fatalf("read ks.cfg: %v", err)
	}
	if !strings.Contains(string(ksBytes), wantKey) {
		t.Errorf("ks.cfg missing SSH key:\n%s", string(ksBytes))
	}
}

func TestRenderer_Render_DryRun(t *testing.T) {
	p, err := profile.LoadAndValidateFile("../../../examples/profiles/rocky-server.yaml")
	if err != nil {
		t.Fatalf("load rocky-server profile: %v", err)
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
		"ks.cfg":      false,
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

func TestRenderer_Validate_RejectsProfileWithNoUsers(t *testing.T) {
	r := Renderer{}
	p := profile.Profile{
		OS:    profile.OS{Family: "rocky"},
		Users: []profile.User{},
	}
	if err := r.Validate(p); err == nil {
		t.Error("expected error for empty users slice, got nil")
	}
}
