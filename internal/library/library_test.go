package library_test

import (
	"testing"

	"github.com/alexmchughdev/bootwrangler/internal/library"
	"github.com/alexmchughdev/bootwrangler/internal/profile"
)

const exampleSSHKey = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJ0EkzLnXFTKfQXv2bseD0jLxoc+R6NxRHWqnY15T9i6 example-only@bootwrangler.invalid"

func minimalProfile(name string) profile.Profile {
	return profile.Profile{
		Name: name,
		OS: profile.OS{
			Family:       "alpine",
			Version:      "3.20",
			Architecture: "x86_64",
		},
		System: profile.System{
			Hostname: "testhost",
			Timezone: "UTC",
			Keyboard: "us",
			Locale:   "en_US.UTF-8",
		},
		Network: profile.Network{Mode: "dhcp", Interface: "eth0"},
		Disk: profile.Disk{
			Mode:               "wipe",
			Target:             "/dev/sda",
			InstallMode:        "sys",
			Filesystem:         "ext4",
			ConfirmDestructive: true,
		},
		Users: []profile.User{
			{
				Name:    "admin",
				Shell:   "/bin/sh",
				SSHKeys: []string{exampleSSHKey},
			},
		},
		SSH: profile.SSH{Enabled: true},
	}
}

func TestLibraryAddAndGet(t *testing.T) {
	dir := t.TempDir()
	lib := library.New(dir)
	if err := lib.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	p := minimalProfile("test-profile")
	filename, err := lib.Add(p)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if filename == "" {
		t.Fatal("expected non-empty filename")
	}

	got, err := lib.Get("test-profile")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Name != "test-profile" {
		t.Errorf("expected name %q, got %q", "test-profile", got.Name)
	}
}

func TestLibraryList(t *testing.T) {
	dir := t.TempDir()
	lib := library.New(dir)
	_ = lib.Init()

	for _, name := range []string{"alpha", "beta", "gamma"} {
		if _, err := lib.Add(minimalProfile(name)); err != nil {
			t.Fatalf("Add %q: %v", name, err)
		}
	}

	entries, err := lib.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}
}

func TestLibraryRemove(t *testing.T) {
	dir := t.TempDir()
	lib := library.New(dir)
	_ = lib.Init()

	p := minimalProfile("removeme")
	if _, err := lib.Add(p); err != nil {
		t.Fatal(err)
	}
	if err := lib.Remove("removeme"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	_, err := lib.Get("removeme")
	if err == nil {
		t.Error("expected error after removal, got nil")
	}
}

func TestLibraryListEmpty(t *testing.T) {
	dir := t.TempDir()
	lib := library.New(dir)
	entries, err := lib.List()
	if err != nil {
		t.Fatalf("List on empty workspace: %v", err)
	}
	if entries != nil {
		t.Errorf("expected nil entries on empty workspace, got %v", entries)
	}
}

func TestLibraryRejectsInvalidProfile(t *testing.T) {
	dir := t.TempDir()
	lib := library.New(dir)
	_ = lib.Init()

	bad := profile.Profile{Name: "bad"}
	_, err := lib.Add(bad)
	if err == nil {
		t.Error("expected Add to reject invalid profile, got nil error")
	}
}

func TestLibraryWriteIndex(t *testing.T) {
	dir := t.TempDir()
	lib := library.New(dir)
	_ = lib.Init()

	if _, err := lib.Add(minimalProfile("idx-test")); err != nil {
		t.Fatal(err)
	}
	if err := lib.WriteIndex(); err != nil {
		t.Fatalf("WriteIndex: %v", err)
	}
}
