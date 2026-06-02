package alpine

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
	if got := r.Name(); got != "alpine" {
		t.Errorf("Name() = %q, want \"alpine\"", got)
	}
}

func TestRenderer_Family(t *testing.T) {
	r := Renderer{}
	if got := r.Family(); got != "alpine" {
		t.Errorf("Family() = %q, want \"alpine\"", got)
	}
}

func TestRenderer_Validate(t *testing.T) {
	t.Run("rejects non-alpine family", func(t *testing.T) {
		r := Renderer{}
		p := profile.Profile{
			OS:    profile.OS{Family: "ubuntu"},
			Users: []profile.User{{Name: "alice"}},
		}
		if err := r.Validate(p); err == nil {
			t.Error("expected error for non-alpine family, got nil")
		}
	})

	t.Run("rejects profile with no users", func(t *testing.T) {
		r := Renderer{}
		p := profile.Profile{
			OS:    profile.OS{Family: "alpine"},
			Users: nil,
		}
		if err := r.Validate(p); err == nil {
			t.Error("expected error for no users, got nil")
		}
	})

	t.Run("accepts valid alpine profile", func(t *testing.T) {
		r := Renderer{}
		p := profile.Profile{
			OS:    profile.OS{Family: "alpine"},
			Users: []profile.User{{Name: "alice"}},
		}
		if err := r.Validate(p); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestRenderer_Render_AlpineMinimal(t *testing.T) {
	p, err := profile.LoadAndValidateFile("../../../examples/profiles/alpine-minimal.yaml")
	if err != nil {
		t.Fatalf("load alpine-minimal profile: %v", err)
	}

	outDir := t.TempDir()
	r := Renderer{}
	m, err := r.Render(p, render.Options{OutDir: outDir})
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	if m.OSFamily != "alpine" {
		t.Errorf("manifest OSFamily = %q, want \"alpine\"", m.OSFamily)
	}
	if m.Renderer != "alpine" {
		t.Errorf("manifest Renderer = %q, want \"alpine\"", m.Renderer)
	}
	if len(m.Files) != 3 {
		t.Errorf("manifest has %d files, want 3", len(m.Files))
	}

	// Read generated answerfile and verify key fields.
	answerfileBytes, err := os.ReadFile(filepath.Join(outDir, "answerfile"))
	if err != nil {
		t.Fatalf("read answerfile: %v", err)
	}
	answerfile := string(answerfileBytes)

	// Verify hostname
	if !strings.Contains(answerfile, "alpine-minimal") {
		t.Errorf("answerfile missing hostname \"alpine-minimal\":\n%s", answerfile)
	}

	// Verify timezone
	if !strings.Contains(answerfile, "Europe/London") {
		t.Errorf("answerfile missing timezone \"Europe/London\":\n%s", answerfile)
	}

	// Verify keyboard
	if !strings.Contains(answerfile, "gb gb") {
		t.Errorf("answerfile missing keyboard \"gb gb\":\n%s", answerfile)
	}
}

func TestRenderer_Render_DHCPNetwork(t *testing.T) {
	p, err := profile.LoadAndValidateFile("../../../examples/profiles/alpine-minimal.yaml")
	if err != nil {
		t.Fatalf("load alpine-minimal profile: %v", err)
	}
	// Ensure the profile uses DHCP (alpine-minimal does by default).
	p.Network.Mode = "dhcp"

	outDir := t.TempDir()
	r := Renderer{}
	_, err = r.Render(p, render.Options{OutDir: outDir})
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	answerfileBytes, err := os.ReadFile(filepath.Join(outDir, "answerfile"))
	if err != nil {
		t.Fatalf("read answerfile: %v", err)
	}
	if !strings.Contains(string(answerfileBytes), "dhcp") {
		t.Errorf("answerfile missing \"dhcp\" for DHCP network:\n%s", string(answerfileBytes))
	}
}

func TestRenderer_Render_SSHKey(t *testing.T) {
	p, err := profile.LoadAndValidateFile("../../../examples/profiles/alpine-minimal.yaml")
	if err != nil {
		t.Fatalf("load alpine-minimal profile: %v", err)
	}

	const wantKey = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJ0EkzLnXFTKfQXv2bseD0jLxoc+R6NxRHWqnY15T9i6 example-only@bootwrangler.invalid"

	outDir := t.TempDir()
	r := Renderer{}
	_, err = r.Render(p, render.Options{OutDir: outDir})
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	postInstallBytes, err := os.ReadFile(filepath.Join(outDir, "post-install.sh"))
	if err != nil {
		t.Fatalf("read post-install.sh: %v", err)
	}
	if !strings.Contains(string(postInstallBytes), wantKey) {
		t.Errorf("post-install.sh missing SSH key:\n%s", string(postInstallBytes))
	}
}

func TestRenderer_Render_DryRun(t *testing.T) {
	p, err := profile.LoadAndValidateFile("../../../examples/profiles/alpine-minimal.yaml")
	if err != nil {
		t.Fatalf("load alpine-minimal profile: %v", err)
	}

	outDir := t.TempDir()
	r := Renderer{}
	m, err := r.Render(p, render.Options{OutDir: outDir, DryRun: true})
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	// Manifest should still describe three files.
	if len(m.Files) != 3 {
		t.Errorf("dry-run manifest has %d files, want 3", len(m.Files))
	}

	// Verify file names are present in manifest.
	wantFiles := map[string]bool{
		"answerfile":      false,
		"post-install.sh": false,
		"custom.ipxe":     false,
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

func TestRenderer_Render_Checksums(t *testing.T) {
	p, err := profile.LoadAndValidateFile("../../../examples/profiles/alpine-minimal.yaml")
	if err != nil {
		t.Fatalf("load alpine-minimal profile: %v", err)
	}

	outDir := t.TempDir()
	r := Renderer{}
	m, err := r.Render(p, render.Options{OutDir: outDir})
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	for _, f := range m.Files {
		if f.SHA256 == "" {
			t.Errorf("file %q has empty SHA256", f.Path)
		}
	}
}

func TestRenderer_Render_IPXEWithURL(t *testing.T) {
	p, err := profile.LoadAndValidateFile("../../../examples/profiles/alpine-minimal.yaml")
	if err != nil {
		t.Fatalf("load alpine-minimal profile: %v", err)
	}

	outDir := t.TempDir()
	r := Renderer{}
	_, err = r.Render(p, render.Options{
		OutDir:        outDir,
		ServerBaseURL: "http://192.168.1.100:8080",
	})
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	ipxeBytes, err := os.ReadFile(filepath.Join(outDir, "custom.ipxe"))
	if err != nil {
		t.Fatalf("read custom.ipxe: %v", err)
	}
	if !strings.Contains(string(ipxeBytes), "http://192.168.1.100:8080") {
		t.Errorf("custom.ipxe missing server base URL:\n%s", string(ipxeBytes))
	}
}

func TestRenderer_Render_IPXEWithoutURL(t *testing.T) {
	p, err := profile.LoadAndValidateFile("../../../examples/profiles/alpine-minimal.yaml")
	if err != nil {
		t.Fatalf("load alpine-minimal profile: %v", err)
	}

	outDir := t.TempDir()
	r := Renderer{}
	_, err = r.Render(p, render.Options{OutDir: outDir, ServerBaseURL: ""})
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	ipxeBytes, err := os.ReadFile(filepath.Join(outDir, "custom.ipxe"))
	if err != nil {
		t.Fatalf("read custom.ipxe: %v", err)
	}
	if !strings.Contains(string(ipxeBytes), "configure ServerBaseURL") {
		t.Errorf("custom.ipxe missing placeholder comment:\n%s", string(ipxeBytes))
	}
}
