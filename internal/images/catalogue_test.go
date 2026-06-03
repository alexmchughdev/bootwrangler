package images

import (
	"strings"
	"testing"
)

const validCatalogueYAML = `
entries:
  - id: test-os
    name: Test OS
    family: debian
    versions:
      - version: "12"
        architectures:
          - arch: x86_64
            images:
              - type: iso
                url: https://example.com/test.iso
                compatibility:
                  whole_drive: true
                  partition: false
                  iso_file_boot: true
                boot_mode:
                  - bios
                  - uefi
`

func TestLoadCatalogueFromBytes(t *testing.T) {
	cat, err := LoadCatalogueFromBytes([]byte(validCatalogueYAML))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cat.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(cat.Entries))
	}
	e := cat.Entries[0]
	if e.ID != "test-os" {
		t.Errorf("expected id test-os, got %s", e.ID)
	}
	if e.Family != "debian" {
		t.Errorf("expected family debian, got %s", e.Family)
	}
}

func TestLoadCatalogueFromBytes_Invalid(t *testing.T) {
	cases := []struct {
		name string
		yaml string
	}{
		{
			name: "missing id",
			yaml: `entries: [{name: "X", family: "debian", versions: []}]`,
		},
		{
			name: "missing name",
			yaml: `entries: [{id: "x", family: "debian", versions: []}]`,
		},
		{
			name: "missing family",
			yaml: `entries: [{id: "x", name: "X", versions: []}]`,
		},
		{
			name: "duplicate id",
			yaml: `entries:
  - id: dup
    name: A
    family: ubuntu
    versions: []
  - id: dup
    name: B
    family: ubuntu
    versions: []`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := LoadCatalogueFromBytes([]byte(tc.yaml))
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestFindImage(t *testing.T) {
	cat, err := LoadCatalogueFromBytes([]byte(validCatalogueYAML))
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	entry, ver, arch, img, err := FindImage(cat, "test-os", "12", "x86_64")
	if err != nil {
		t.Fatalf("FindImage: %v", err)
	}
	if entry.ID != "test-os" {
		t.Errorf("entry.ID = %s", entry.ID)
	}
	if ver.Version != "12" {
		t.Errorf("ver.Version = %s", ver.Version)
	}
	if arch.Arch != "x86_64" {
		t.Errorf("arch.Arch = %s", arch.Arch)
	}
	if img.Type != ImageTypeISO {
		t.Errorf("img.Type = %s", img.Type)
	}
}

func TestFindImage_NotFound(t *testing.T) {
	cat, err := LoadCatalogueFromBytes([]byte(validCatalogueYAML))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	_, _, _, _, err = FindImage(cat, "nonexistent", "", "")
	if err == nil {
		t.Error("expected error for missing image")
	}
}

func TestRequirePartitionCompatibleRejectsWholeDriveOnlyImage(t *testing.T) {
	img := ArchImage{
		Compatibility: Compatibility{
			WholeDrive:  true,
			Partition:   false,
			ISOFileBoot: true,
		},
	}

	err := RequirePartitionCompatible("ubuntu-server", "24.04", img)
	if err == nil {
		t.Fatal("RequirePartitionCompatible() error = nil, want error")
	}
	want := "image ubuntu-server-24.04 is not marked as partition-flash compatible"
	if got := err.Error(); !strings.HasPrefix(got, want) {
		t.Fatalf("RequirePartitionCompatible() error = %q, want prefix %q", got, want)
	}
}

func TestRequirePartitionCompatibleAllowsMarkedImage(t *testing.T) {
	img := ArchImage{
		Compatibility: Compatibility{
			Partition: true,
		},
	}

	if err := RequirePartitionCompatible("company-os", "1.0", img); err != nil {
		t.Fatalf("RequirePartitionCompatible() error = %v, want nil", err)
	}
}

func TestRequireWholeDriveCompatibleRejectsPartitionOnlyImage(t *testing.T) {
	img := ArchImage{
		Compatibility: Compatibility{
			WholeDrive: false,
			Partition:  true,
		},
	}

	err := RequireWholeDriveCompatible("partition-os", "1.0", img)
	if err == nil {
		t.Fatal("RequireWholeDriveCompatible() error = nil, want error")
	}
	want := "image partition-os-1.0 is not marked as whole-drive compatible"
	if got := err.Error(); !strings.HasPrefix(got, want) {
		t.Fatalf("RequireWholeDriveCompatible() error = %q, want prefix %q", got, want)
	}
}

func TestBuiltinCatalogue(t *testing.T) {
	cat := BuiltinCatalogue()
	if len(cat.Entries) == 0 {
		t.Fatal("builtin catalogue has no entries")
	}
	if err := ValidateCatalogue(cat); err != nil {
		t.Fatalf("builtin catalogue invalid: %v", err)
	}

	requiredIDs := []string{
		"ubuntu-server", "ubuntu-desktop", "debian-netinst",
		"fedora-server", "rocky-linux", "alpine-standard",
		"arch-linux", "opensuse-leap",
	}
	idSet := map[string]bool{}
	for _, e := range cat.Entries {
		idSet[e.ID] = true
	}
	for _, id := range requiredIDs {
		if !idSet[id] {
			t.Errorf("builtin catalogue missing entry: %s", id)
		}
	}
}
