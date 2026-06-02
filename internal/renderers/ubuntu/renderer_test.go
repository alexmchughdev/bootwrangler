package ubuntu_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/alexmchughdev/bootwrangler/internal/profile"
	"github.com/alexmchughdev/bootwrangler/internal/render"
	"github.com/alexmchughdev/bootwrangler/internal/renderers/ubuntu"
)

func loadExampleProfile(t *testing.T) profile.Profile {
	t.Helper()
	p, err := profile.LoadAndValidateFile("../../../examples/profiles/ubuntu-server.yaml")
	if err != nil {
		t.Fatalf("load ubuntu-server.yaml: %v", err)
	}
	return p
}

// TestRenderUserData verifies user-data is valid YAML and contains expected values.
func TestRenderUserData(t *testing.T) {
	p := loadExampleProfile(t)
	dir := t.TempDir()

	r := ubuntu.Renderer{}
	m, err := r.Render(p, render.Options{OutDir: dir})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	// Check manifest has expected files.
	if len(m.Files) != 3 {
		t.Fatalf("expected 3 files in manifest, got %d", len(m.Files))
	}

	// Read and parse user-data.
	userDataBytes, err := os.ReadFile(filepath.Join(dir, "user-data"))
	if err != nil {
		t.Fatalf("read user-data: %v", err)
	}

	userDataStr := string(userDataBytes)

	// Must start with the cloud-config header.
	if !strings.HasPrefix(userDataStr, "#cloud-config\n") {
		t.Errorf("user-data does not start with #cloud-config header")
	}

	// Strip the first line (comment) before YAML parsing, since yaml.v3
	// does not handle the #cloud-config directive line specially.
	lines := strings.SplitN(userDataStr, "\n", 2)
	if len(lines) < 2 {
		t.Fatal("user-data has no content after header")
	}
	yamlPart := lines[1]

	// Must parse as valid YAML.
	var parsed map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlPart), &parsed); err != nil {
		t.Fatalf("user-data is not valid YAML: %v\ncontent:\n%s", err, userDataStr)
	}

	// Verify autoinstall section exists.
	autoinstall, ok := parsed["autoinstall"].(map[string]interface{})
	if !ok {
		t.Fatalf("user-data missing 'autoinstall' key, got: %v", parsed)
	}

	// Check hostname.
	identity, ok := autoinstall["identity"].(map[string]interface{})
	if !ok {
		t.Fatalf("autoinstall missing 'identity' section")
	}
	if got := identity["hostname"]; got != p.System.Hostname {
		t.Errorf("hostname: got %q, want %q", got, p.System.Hostname)
	}

	// Check locale.
	if got := autoinstall["locale"]; got != p.System.Locale {
		t.Errorf("locale: got %q, want %q", got, p.System.Locale)
	}

	// Check SSH authorized-keys contains the first user's SSH key.
	sshSection, ok := autoinstall["ssh"].(map[string]interface{})
	if !ok {
		t.Fatalf("autoinstall missing 'ssh' section")
	}
	authorizedKeys, ok := sshSection["authorized-keys"].([]interface{})
	if !ok {
		t.Fatalf("ssh missing 'authorized-keys' list")
	}
	if len(authorizedKeys) == 0 {
		t.Fatal("authorized-keys is empty")
	}
	wantKey := p.Users[0].SSHKeys[0]
	if got := authorizedKeys[0].(string); got != wantKey {
		t.Errorf("authorized-keys[0]: got %q, want %q", got, wantKey)
	}
}

// TestRenderMetaData verifies meta-data contains the expected instance-id and local-hostname.
func TestRenderMetaData(t *testing.T) {
	p := loadExampleProfile(t)
	dir := t.TempDir()

	r := ubuntu.Renderer{}
	if _, err := r.Render(p, render.Options{OutDir: dir}); err != nil {
		t.Fatalf("Render: %v", err)
	}

	metaDataBytes, err := os.ReadFile(filepath.Join(dir, "meta-data"))
	if err != nil {
		t.Fatalf("read meta-data: %v", err)
	}

	var parsed map[string]interface{}
	if err := yaml.Unmarshal(metaDataBytes, &parsed); err != nil {
		t.Fatalf("meta-data is not valid YAML: %v", err)
	}

	wantInstanceID := "iid-" + p.Name
	if got := parsed["instance-id"]; got != wantInstanceID {
		t.Errorf("instance-id: got %q, want %q", got, wantInstanceID)
	}

	if got := parsed["local-hostname"]; got != p.System.Hostname {
		t.Errorf("local-hostname: got %q, want %q", got, p.System.Hostname)
	}
}

// TestDryRunDoesNotWriteFiles verifies DryRun returns a complete manifest without writing files.
func TestDryRunDoesNotWriteFiles(t *testing.T) {
	p := loadExampleProfile(t)
	dir := t.TempDir()

	r := ubuntu.Renderer{}
	m, err := r.Render(p, render.Options{OutDir: dir, DryRun: true})
	if err != nil {
		t.Fatalf("Render (dry-run): %v", err)
	}

	// Manifest must still list the three expected files.
	if len(m.Files) != 3 {
		t.Fatalf("expected 3 files in manifest, got %d", len(m.Files))
	}

	wantFiles := map[string]bool{
		"user-data":   false,
		"meta-data":   false,
		"custom.ipxe": false,
	}
	for _, f := range m.Files {
		if _, known := wantFiles[f.Path]; !known {
			t.Errorf("unexpected file in manifest: %s", f.Path)
		}
		wantFiles[f.Path] = true
	}
	for name, found := range wantFiles {
		if !found {
			t.Errorf("expected file %q missing from manifest", name)
		}
	}

	// No files must have been written to disk.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read output dir: %v", err)
	}
	if len(entries) != 0 {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("DryRun wrote files to disk: %v", names)
	}
}

// TestValidateRejectsNoUsers verifies Validate rejects a profile with no users.
func TestValidateRejectsNoUsers(t *testing.T) {
	p := loadExampleProfile(t)
	p.Users = nil

	r := ubuntu.Renderer{}
	err := r.Validate(p)
	if err == nil {
		t.Fatal("expected error for profile with no users, got nil")
	}
	if !strings.Contains(err.Error(), "at least one user") {
		t.Errorf("expected 'at least one user' in error, got: %v", err)
	}
}

// TestValidateRejectsFirstUserWithNoSSHKeys verifies Validate rejects a profile
// where the first user has no SSH keys.
func TestValidateRejectsFirstUserWithNoSSHKeys(t *testing.T) {
	p := loadExampleProfile(t)
	// Keep the user but remove their SSH keys.
	p.Users[0].SSHKeys = nil

	r := ubuntu.Renderer{}
	err := r.Validate(p)
	if err == nil {
		t.Fatal("expected error for first user with no SSH keys, got nil")
	}
	if !strings.Contains(err.Error(), "SSH key") {
		t.Errorf("expected 'SSH key' in error, got: %v", err)
	}
}
