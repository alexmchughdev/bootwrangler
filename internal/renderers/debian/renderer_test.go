package debian

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alexmchughdev/bootwrangler/internal/profile"
	"github.com/alexmchughdev/bootwrangler/internal/render"
)

// loadExample loads the debian-server.yaml example profile for tests.
func loadExample(t *testing.T) profile.Profile {
	t.Helper()
	p, err := profile.LoadAndValidateFile("../../../examples/profiles/debian-server.yaml")
	if err != nil {
		t.Fatalf("load example profile: %v", err)
	}
	return p
}

func TestRender_PreseedContainsLocaleTimezoneHostname(t *testing.T) {
	p := loadExample(t)
	outDir := t.TempDir()
	r := Renderer{}

	m, err := r.Render(p, render.Options{
		OutDir:        outDir,
		ServerBaseURL: "http://boot.example.com",
	})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	// Verify manifest has three files
	if len(m.Files) != 3 {
		t.Errorf("expected 3 files in manifest, got %d", len(m.Files))
	}

	preseedBytes, err := os.ReadFile(filepath.Join(outDir, "preseed.cfg"))
	if err != nil {
		t.Fatalf("read preseed.cfg: %v", err)
	}
	preseed := string(preseedBytes)

	checks := []struct {
		label string
		want  string
	}{
		{"locale", "d-i debian-installer/locale string en_GB.UTF-8"},
		{"timezone", "d-i time/zone string Europe/London"},
		{"hostname", "d-i netcfg/get_hostname string debian-server"},
		{"keyboard", "d-i keyboard-configuration/xkb-keymap select gb"},
		{"root-login disabled", "d-i passwd/root-login boolean false"},
		{"first user", "d-i passwd/username string deploy"},
		{"package git", "d-i pkgsel/include string git"},
		{"late_command URL", "http://boot.example.com/late-command.sh"},
	}

	for _, tc := range checks {
		t.Run(tc.label, func(t *testing.T) {
			if !strings.Contains(preseed, tc.want) {
				t.Errorf("preseed.cfg missing %q\ngot:\n%s", tc.want, preseed)
			}
		})
	}
}

func TestRender_LateCommandContainsSSHKey(t *testing.T) {
	p := loadExample(t)
	outDir := t.TempDir()
	r := Renderer{}

	_, err := r.Render(p, render.Options{
		OutDir:        outDir,
		ServerBaseURL: "http://boot.example.com",
	})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	lateBytes, err := os.ReadFile(filepath.Join(outDir, "late-command.sh"))
	if err != nil {
		t.Fatalf("read late-command.sh: %v", err)
	}
	late := string(lateBytes)

	expectedKey := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJ0EkzLnXFTKfQXv2bseD0jLxoc+R6NxRHWqnY15T9i6 example-only@bootwrangler.invalid"
	if !strings.Contains(late, expectedKey) {
		t.Errorf("late-command.sh missing expected SSH key\ngot:\n%s", late)
	}

	// Also verify sshd hardening directives
	if !strings.Contains(late, "PermitRootLogin no") {
		t.Errorf("late-command.sh missing PermitRootLogin no\ngot:\n%s", late)
	}
	if !strings.Contains(late, "PasswordAuthentication no") {
		t.Errorf("late-command.sh missing PasswordAuthentication no\ngot:\n%s", late)
	}
}

func TestRender_DryRunDoesNotWriteFiles(t *testing.T) {
	p := loadExample(t)
	outDir := t.TempDir()
	r := Renderer{}

	m, err := r.Render(p, render.Options{
		OutDir:        outDir,
		ServerBaseURL: "http://boot.example.com",
		DryRun:        true,
	})
	if err != nil {
		t.Fatalf("Render (dry-run): %v", err)
	}

	// Manifest must still list all three files
	if len(m.Files) != 3 {
		t.Errorf("dry-run manifest: expected 3 files, got %d", len(m.Files))
	}

	wantFiles := []string{"preseed.cfg", "late-command.sh", "custom.ipxe"}
	for _, name := range wantFiles {
		found := false
		for _, f := range m.Files {
			if f.Path == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("dry-run manifest missing entry for %q", name)
		}
	}

	// No files should be written to disk
	for _, name := range wantFiles {
		path := filepath.Join(outDir, name)
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("dry-run: file %q should not exist on disk", name)
		}
	}

	// SHA256 checksums should be empty because files were not written
	for _, f := range m.Files {
		if f.SHA256 != "" {
			t.Errorf("dry-run: file %q should have empty SHA256, got %q", f.Path, f.SHA256)
		}
	}
}

func TestValidate_RejectsNoUsers(t *testing.T) {
	r := Renderer{}
	p := profile.Profile{
		OS: profile.OS{Family: "debian"},
		// No users
	}
	err := r.Validate(p)
	if err == nil {
		t.Fatal("expected Validate to return an error for profile with no users")
	}
	if !strings.Contains(err.Error(), "at least one user") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestValidate_RejectsWrongFamily(t *testing.T) {
	r := Renderer{}
	p := profile.Profile{
		OS: profile.OS{Family: "alpine"},
		Users: []profile.User{
			{Name: "admin"},
		},
	}
	err := r.Validate(p)
	if err == nil {
		t.Fatal("expected Validate to return an error for non-debian family")
	}
	if !strings.Contains(err.Error(), "os.family") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestValidate_AcceptsValidDebianProfile(t *testing.T) {
	r := Renderer{}
	p := loadExample(t)
	if err := r.Validate(p); err != nil {
		t.Errorf("Validate on valid debian profile returned error: %v", err)
	}
}

func TestRender_NoServerBaseURL_AddsWarning(t *testing.T) {
	p := loadExample(t)
	outDir := t.TempDir()
	r := Renderer{}

	m, err := r.Render(p, render.Options{
		OutDir:        outDir,
		ServerBaseURL: "", // intentionally empty
	})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	if len(m.Warnings) == 0 {
		t.Fatal("expected a warning when ServerBaseURL is empty, got none")
	}

	found := false
	for _, w := range m.Warnings {
		if strings.Contains(w, "ServerBaseURL") && strings.Contains(w, "late-command") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected warning about ServerBaseURL and late-command, got: %v", m.Warnings)
	}

	// Preseed should not contain late_command line
	preseedBytes, err := os.ReadFile(filepath.Join(outDir, "preseed.cfg"))
	if err != nil {
		t.Fatalf("read preseed.cfg: %v", err)
	}
	if strings.Contains(string(preseedBytes), "late_command") {
		t.Errorf("preseed.cfg should not contain late_command when ServerBaseURL is empty")
	}

	// custom.ipxe should contain fallback comment
	ipxeBytes, err := os.ReadFile(filepath.Join(outDir, "custom.ipxe"))
	if err != nil {
		t.Fatalf("read custom.ipxe: %v", err)
	}
	if !strings.Contains(string(ipxeBytes), "configure ServerBaseURL") {
		t.Errorf("custom.ipxe should contain fallback comment when ServerBaseURL is empty\ngot:\n%s", string(ipxeBytes))
	}
}

func TestRender_ChecksumsPopulated(t *testing.T) {
	p := loadExample(t)
	outDir := t.TempDir()
	r := Renderer{}

	m, err := r.Render(p, render.Options{
		OutDir:        outDir,
		ServerBaseURL: "http://boot.example.com",
	})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	for _, f := range m.Files {
		if f.SHA256 == "" {
			t.Errorf("file %q has empty SHA256 checksum", f.Path)
		}
		if len(f.SHA256) != 64 {
			t.Errorf("file %q SHA256 should be 64 hex chars, got %d: %q", f.Path, len(f.SHA256), f.SHA256)
		}
	}
}
