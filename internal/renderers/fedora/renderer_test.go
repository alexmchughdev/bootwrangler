package fedora

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
	if got := r.Name(); got != "fedora" {
		t.Errorf("Name() = %q, want \"fedora\"", got)
	}
}

func TestRenderer_Family(t *testing.T) {
	r := Renderer{}
	if got := r.Family(); got != "fedora" {
		t.Errorf("Family() = %q, want \"fedora\"", got)
	}
}

func TestRenderer_Validate_RejectsNonFedora(t *testing.T) {
	r := Renderer{}
	p := profile.Profile{
		OS:    profile.OS{Family: "rocky"},
		Users: []profile.User{{Name: "admin"}},
	}
	if err := r.Validate(p); err == nil {
		t.Error("expected error for non-fedora family, got nil")
	}
}

func TestRenderer_Validate_RejectsNoUsers(t *testing.T) {
	r := Renderer{}
	p := profile.Profile{
		OS:    profile.OS{Family: "fedora"},
		Users: nil,
	}
	if err := r.Validate(p); err == nil {
		t.Error("expected error for no users, got nil")
	}
}

func TestRenderer_Validate_AcceptsValidProfile(t *testing.T) {
	r := Renderer{}
	p := profile.Profile{
		OS:    profile.OS{Family: "fedora"},
		Users: []profile.User{{Name: "admin"}},
	}
	if err := r.Validate(p); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRenderer_Render_FedoraServer(t *testing.T) {
	p, err := profile.LoadAndValidateFile("../../../examples/profiles/fedora-server.yaml")
	if err != nil {
		t.Fatalf("load fedora-server profile: %v", err)
	}

	outDir := t.TempDir()
	r := Renderer{}
	m, err := r.Render(p, render.Options{OutDir: outDir})
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	if m.OSFamily != "fedora" {
		t.Errorf("manifest OSFamily = %q, want \"fedora\"", m.OSFamily)
	}
	if m.Renderer != "fedora" {
		t.Errorf("manifest Renderer = %q, want \"fedora\"", m.Renderer)
	}
	if len(m.Files) != 2 {
		t.Errorf("manifest has %d files, want 2", len(m.Files))
	}

	ksBytes, err := os.ReadFile(filepath.Join(outDir, "ks.cfg"))
	if err != nil {
		t.Fatalf("read ks.cfg: %v", err)
	}
	ksCfg := string(ksBytes)

	// Verify hostname appears.
	if !strings.Contains(ksCfg, "fedora-server") {
		t.Errorf("ks.cfg missing hostname \"fedora-server\":\n%s", ksCfg)
	}

	// Verify timezone appears.
	if !strings.Contains(ksCfg, "Europe/London") {
		t.Errorf("ks.cfg missing timezone \"Europe/London\":\n%s", ksCfg)
	}

	// Verify Fedora-specific environment group.
	if !strings.Contains(ksCfg, "@^server-product-environment") {
		t.Errorf("ks.cfg missing \"@^server-product-environment\":\n%s", ksCfg)
	}
}

func TestRenderer_Render_SSHKey(t *testing.T) {
	p, err := profile.LoadAndValidateFile("../../../examples/profiles/fedora-server.yaml")
	if err != nil {
		t.Fatalf("load fedora-server profile: %v", err)
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
	p, err := profile.LoadAndValidateFile("../../../examples/profiles/fedora-server.yaml")
	if err != nil {
		t.Fatalf("load fedora-server profile: %v", err)
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

	// Verify expected file names appear in manifest.
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

	// No actual files should have been written.
	for _, f := range m.Files {
		fullPath := filepath.Join(outDir, f.Path)
		if _, err := os.Stat(fullPath); err == nil {
			t.Errorf("dry-run wrote file %q, expected no files written", fullPath)
		}
	}

	// Dry-run manifest must include a warning.
	if len(m.Warnings) == 0 {
		t.Error("dry-run manifest should have at least one warning")
	}
}

func TestRenderer_Render_ManifestFedoraWarning(t *testing.T) {
	p, err := profile.LoadAndValidateFile("../../../examples/profiles/fedora-server.yaml")
	if err != nil {
		t.Fatalf("load fedora-server profile: %v", err)
	}

	outDir := t.TempDir()
	r := Renderer{}
	m, err := r.Render(p, render.Options{OutDir: outDir})
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	const wantFragment = "Fedora kickstart targets the current Fedora release"
	for _, w := range m.Warnings {
		if strings.Contains(w, wantFragment) {
			return
		}
	}
	t.Errorf("manifest warnings missing Fedora version pin notice; got: %v", m.Warnings)
}

func TestRenderer_Render_SSHHardening(t *testing.T) {
	p, err := profile.LoadAndValidateFile("../../../examples/profiles/fedora-server.yaml")
	if err != nil {
		t.Fatalf("load fedora-server profile: %v", err)
	}
	// fedora-server.yaml has permit_root_login: false and password_authentication: false
	p.SSH.PermitRootLogin = false
	p.SSH.PasswordAuthentication = false

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
	ksCfg := string(ksBytes)

	if !strings.Contains(ksCfg, "PermitRootLogin no") {
		t.Errorf("ks.cfg missing PermitRootLogin hardening:\n%s", ksCfg)
	}
	if !strings.Contains(ksCfg, "PasswordAuthentication no") {
		t.Errorf("ks.cfg missing PasswordAuthentication hardening:\n%s", ksCfg)
	}
}

func TestRenderer_Render_Checksums(t *testing.T) {
	p, err := profile.LoadAndValidateFile("../../../examples/profiles/fedora-server.yaml")
	if err != nil {
		t.Fatalf("load fedora-server profile: %v", err)
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
	p, err := profile.LoadAndValidateFile("../../../examples/profiles/fedora-server.yaml")
	if err != nil {
		t.Fatalf("load fedora-server profile: %v", err)
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
	ipxe := string(ipxeBytes)
	if !strings.Contains(ipxe, "http://192.168.1.100:8080") {
		t.Errorf("custom.ipxe missing server base URL:\n%s", ipxe)
	}
	if !strings.Contains(ipxe, "inst.ks=") {
		t.Errorf("custom.ipxe missing inst.ks parameter:\n%s", ipxe)
	}
}

func TestRenderer_Render_IPXEWithoutURL(t *testing.T) {
	p, err := profile.LoadAndValidateFile("../../../examples/profiles/fedora-server.yaml")
	if err != nil {
		t.Fatalf("load fedora-server profile: %v", err)
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

func TestRenderer_Render_DNFPackages(t *testing.T) {
	p, err := profile.LoadAndValidateFile("../../../examples/profiles/fedora-server.yaml")
	if err != nil {
		t.Fatalf("load fedora-server profile: %v", err)
	}
	// fedora-server.yaml includes "git" in packages.names
	p.Packages.Names = []string{"git", "htop"}

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
	ksCfg := string(ksBytes)
	for _, pkg := range []string{"git", "htop"} {
		if !strings.Contains(ksCfg, pkg) {
			t.Errorf("ks.cfg missing package %q:\n%s", pkg, ksCfg)
		}
	}
}
