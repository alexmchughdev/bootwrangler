package opensuse

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alexmchughdev/bootwrangler/internal/profile"
	"github.com/alexmchughdev/bootwrangler/internal/render"
)

// TestRenderer_Name verifies the renderer identifier.
func TestRenderer_Name(t *testing.T) {
	r := Renderer{}
	if got := r.Name(); got != "opensuse" {
		t.Errorf("Name() = %q, want \"opensuse\"", got)
	}
}

// TestRenderer_Family verifies the OS family.
func TestRenderer_Family(t *testing.T) {
	r := Renderer{}
	if got := r.Family(); got != "opensuse" {
		t.Errorf("Family() = %q, want \"opensuse\"", got)
	}
}

// TestRenderer_Validate checks validation rules.
func TestRenderer_Validate(t *testing.T) {
	t.Run("rejects non-opensuse family", func(t *testing.T) {
		r := Renderer{}
		p := profile.Profile{
			OS:    profile.OS{Family: "alpine"},
			Users: []profile.User{{Name: "alice"}},
		}
		if err := r.Validate(p); err == nil {
			t.Error("expected error for non-opensuse family, got nil")
		}
	})

	t.Run("rejects profile with no users", func(t *testing.T) {
		r := Renderer{}
		p := profile.Profile{
			OS:    profile.OS{Family: "opensuse"},
			Users: nil,
		}
		if err := r.Validate(p); err == nil {
			t.Error("expected error for no users, got nil")
		}
	})

	t.Run("accepts valid opensuse profile", func(t *testing.T) {
		r := Renderer{}
		p := profile.Profile{
			OS:    profile.OS{Family: "opensuse"},
			Users: []profile.User{{Name: "alice"}},
		}
		if err := r.Validate(p); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

// TestRenderer_Render_OpenSUSEServer renders the example profile and verifies
// autoinst.xml is valid XML and contains expected values.
func TestRenderer_Render_OpenSUSEServer(t *testing.T) {
	p, err := profile.LoadAndValidateFile("../../../examples/profiles/opensuse-server.yaml")
	if err != nil {
		t.Fatalf("load opensuse-server profile: %v", err)
	}

	outDir := t.TempDir()
	r := Renderer{}
	m, err := r.Render(p, render.Options{OutDir: outDir})
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	if m.OSFamily != "opensuse" {
		t.Errorf("manifest OSFamily = %q, want \"opensuse\"", m.OSFamily)
	}
	if m.Renderer != "opensuse" {
		t.Errorf("manifest Renderer = %q, want \"opensuse\"", m.Renderer)
	}
	if len(m.Files) != 2 {
		t.Errorf("manifest has %d files, want 2", len(m.Files))
	}

	// Read and parse autoinst.xml.
	autoinstBytes, err := os.ReadFile(filepath.Join(outDir, "autoinst.xml"))
	if err != nil {
		t.Fatalf("read autoinst.xml: %v", err)
	}

	// Verify it is valid XML by unmarshaling into a generic structure.
	var xmlDoc any
	// Strip the DOCTYPE declaration which xml.Unmarshal does not handle.
	xmlContent := stripDoctype(string(autoinstBytes))
	if err := xml.Unmarshal([]byte(xmlContent), &struct {
		XMLName xml.Name
		Inner   []byte `xml:",innerxml"`
	}{}); err != nil {
		t.Errorf("autoinst.xml is not valid XML: %v\n%s", err, autoinstBytes)
	}
	_ = xmlDoc

	autoinst := string(autoinstBytes)

	// Verify timezone.
	if !strings.Contains(autoinst, "Europe/London") {
		t.Errorf("autoinst.xml missing timezone \"Europe/London\":\n%s", autoinst)
	}

	// Verify hostname is not required in AutoYaST XML directly, but verify
	// the profile loaded correctly.
	if p.System.Hostname != "opensuse-server" {
		t.Errorf("profile hostname = %q, want \"opensuse-server\"", p.System.Hostname)
	}

	// Verify locale.
	if !strings.Contains(autoinst, "en_GB") {
		t.Errorf("autoinst.xml missing locale \"en_GB\":\n%s", autoinst)
	}

	// Verify disk target defaulted from "auto".
	if !strings.Contains(autoinst, "/dev/sda") {
		t.Errorf("autoinst.xml missing disk target \"/dev/sda\":\n%s", autoinst)
	}

	// Verify network interface defaulted from "auto".
	if !strings.Contains(autoinst, "eth0") {
		t.Errorf("autoinst.xml missing interface \"eth0\":\n%s", autoinst)
	}

	// Verify user is listed.
	if !strings.Contains(autoinst, "deploy") {
		t.Errorf("autoinst.xml missing user \"deploy\":\n%s", autoinst)
	}
}

// TestRenderer_Render_SSHKeyHandling verifies that SSH keys appear in the
// post-scripts section of autoinst.xml.
func TestRenderer_Render_SSHKeyHandling(t *testing.T) {
	p, err := profile.LoadAndValidateFile("../../../examples/profiles/opensuse-server.yaml")
	if err != nil {
		t.Fatalf("load opensuse-server profile: %v", err)
	}

	const wantKey = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJ0EkzLnXFTKfQXv2bseD0jLxoc+R6NxRHWqnY15T9i6 example-only@bootwrangler.invalid"

	outDir := t.TempDir()
	r := Renderer{}
	_, err = r.Render(p, render.Options{OutDir: outDir})
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	autoinstBytes, err := os.ReadFile(filepath.Join(outDir, "autoinst.xml"))
	if err != nil {
		t.Fatalf("read autoinst.xml: %v", err)
	}

	// The SSH key should appear inside the post-scripts source element.
	if !strings.Contains(string(autoinstBytes), wantKey) {
		t.Errorf("autoinst.xml post-scripts missing SSH key:\n%s", string(autoinstBytes))
	}

	// Verify sshd_config hardening commands are present.
	if !strings.Contains(string(autoinstBytes), "PermitRootLogin") {
		t.Errorf("autoinst.xml post-scripts missing PermitRootLogin hardening")
	}
	if !strings.Contains(string(autoinstBytes), "PasswordAuthentication") {
		t.Errorf("autoinst.xml post-scripts missing PasswordAuthentication hardening")
	}
}

// TestRenderer_Render_DryRun verifies dry-run returns manifest without writing files.
func TestRenderer_Render_DryRun(t *testing.T) {
	p, err := profile.LoadAndValidateFile("../../../examples/profiles/opensuse-server.yaml")
	if err != nil {
		t.Fatalf("load opensuse-server profile: %v", err)
	}

	outDir := t.TempDir()
	r := Renderer{}
	m, err := r.Render(p, render.Options{OutDir: outDir, DryRun: true})
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	// Manifest should describe two files.
	if len(m.Files) != 2 {
		t.Errorf("dry-run manifest has %d files, want 2", len(m.Files))
	}

	// Verify expected file names are in the manifest.
	wantFiles := map[string]bool{
		"autoinst.xml": false,
		"custom.ipxe":  false,
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

// TestRenderer_Validate_RejectsNoUsers verifies that Validate rejects profiles
// with no users defined.
func TestRenderer_Validate_RejectsNoUsers(t *testing.T) {
	r := Renderer{}
	p := profile.Profile{
		OS:    profile.OS{Family: "opensuse"},
		Users: nil,
	}
	if err := r.Validate(p); err == nil {
		t.Error("Validate() expected error for profile with no users, got nil")
	}
}

// stripDoctype removes the XML declaration and DOCTYPE lines so that
// encoding/xml can parse the document body without choking on DOCTYPE.
func stripDoctype(s string) string {
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "<?xml") || strings.HasPrefix(trimmed, "<!DOCTYPE") {
			continue
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}
