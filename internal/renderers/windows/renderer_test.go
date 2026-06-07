package windows

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alexmchughdev/bootwrangler/internal/profile"
	"github.com/alexmchughdev/bootwrangler/internal/render"
)

func baseProfile() profile.Profile {
	return profile.Profile{
		Name: "win",
		OS:   profile.OS{Family: "windows", Version: "11", Architecture: "x86_64"},
		System: profile.System{
			Hostname: "WIN-EDGE-01", Timezone: "Europe/London", Keyboard: "gb", Locale: "en-GB",
		},
		Network: profile.Network{Mode: "dhcp", Interface: "auto"},
		Disk: profile.Disk{
			Mode: "wipe", Target: "auto", InstallMode: "image",
			Filesystem: "ntfs", ConfirmDestructive: true,
		},
		Users: []profile.User{{Name: "admin", Groups: []string{"Administrators"}}},
	}
}

func render_(t *testing.T, p profile.Profile) (string, string, render.Options, func()) {
	t.Helper()
	dir := t.TempDir()
	opts := render.Options{OutDir: dir}
	if _, err := (Renderer{}).Render(p, opts); err != nil {
		t.Fatalf("Render: %v", err)
	}
	unattend, err := os.ReadFile(filepath.Join(dir, "autounattend.xml"))
	if err != nil {
		t.Fatal(err)
	}
	ps1, err := os.ReadFile(filepath.Join(dir, "FirstLogon.ps1"))
	if err != nil {
		t.Fatal(err)
	}
	return string(unattend), string(ps1), opts, func() {}
}

func TestValidate(t *testing.T) {
	if err := (Renderer{}).Validate(baseProfile()); err != nil {
		t.Errorf("expected valid windows profile, got %v", err)
	}
	bad := baseProfile()
	bad.OS.Family = "ubuntu"
	if err := (Renderer{}).Validate(bad); err == nil {
		t.Error("expected error for non-windows family")
	}
	noUser := baseProfile()
	noUser.Users = nil
	if err := (Renderer{}).Validate(noUser); err == nil {
		t.Error("expected error when no users defined")
	}
}

func TestAutounattendIsValidXML(t *testing.T) {
	unattend, _, _, _ := render_(t, baseProfile())
	var v any
	if err := xml.Unmarshal([]byte(unattend), &v); err != nil {
		t.Fatalf("autounattend.xml is not well-formed XML: %v", err)
	}
	for _, want := range []string{"WIN-EDGE-01", "Administrators", "en-GB", "FirstLogon.ps1"} {
		if !strings.Contains(unattend, want) {
			t.Errorf("autounattend.xml missing %q", want)
		}
	}
}

func TestOfflineDomainJoinEmbedsBlob(t *testing.T) {
	p := baseProfile()
	p.OSSpecific.Windows = &profile.WindowsOverrides{
		DomainJoin: &profile.DomainJoin{
			Domain: "corp.example.com", Method: "offline-djoin", ProvisionBlob: "QUJDREVG",
		},
	}
	unattend, _, _, _ := render_(t, p)
	if !strings.Contains(unattend, "offlineServicing") {
		t.Error("expected offlineServicing pass for offline-djoin")
	}
	if !strings.Contains(unattend, "QUJDREVG") {
		t.Error("expected provisioning blob embedded in answer file")
	}
}

func TestFirstLogonJoinUsesAddComputer(t *testing.T) {
	p := baseProfile()
	p.OSSpecific.Windows = &profile.WindowsOverrides{
		DomainJoin: &profile.DomainJoin{
			Domain: "corp.example.com", Method: "first-logon", JoinUser: "joiner",
		},
	}
	_, ps1, _, _ := render_(t, p)
	if !strings.Contains(ps1, "Add-Computer") {
		t.Error("expected Add-Computer in FirstLogon.ps1 for first-logon join")
	}
	if !strings.Contains(ps1, "corp.example.com") {
		t.Error("expected domain name in FirstLogon.ps1")
	}
}

func TestPackagesAndSSHInFirstLogon(t *testing.T) {
	p := baseProfile()
	p.SSH.Enabled = false
	p.Packages.Names = []string{"Google.Chrome"}
	_, ps1, _, _ := render_(t, p)
	if !strings.Contains(ps1, "winget install") || !strings.Contains(ps1, "Google.Chrome") {
		t.Error("expected winget install of Google.Chrome")
	}
}

func TestDryRunWritesNothing(t *testing.T) {
	dir := t.TempDir()
	m, err := (Renderer{}).Render(baseProfile(), render.Options{OutDir: dir, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Files) != 2 {
		t.Errorf("expected 2 manifest files, got %d", len(m.Files))
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 0 {
		t.Errorf("dry-run wrote %d files, want 0", len(entries))
	}
}

func TestHostnameTruncatedTo15(t *testing.T) {
	p := baseProfile()
	p.System.Hostname = "THIS-NAME-IS-WAY-TOO-LONG"
	unattend, _, _, _ := render_(t, p)
	if !strings.Contains(unattend, "<ComputerName>THIS-NAME-IS-WA</ComputerName>") {
		t.Error("expected computer name truncated to 15 chars")
	}
}
