package media

import (
	"testing"
)

func TestValidate_Valid(t *testing.T) {
	r := Recipe{
		Name:   "lab-usb",
		Device: DeviceSpec{PartitionTable: PartitionTableGPT},
		Boot:   BootSpec{Mode: BootModeUEFIBIOS, Menu: BootMenuiPXE},
		Partitions: []Partition{
			{Label: "BOOT", Size: "2G", Filesystem: "fat32", Content: PartitionContent{Type: ContentBootMenu}},
			{Label: "STORAGE", Size: "remaining", Filesystem: "exfat", Content: PartitionContent{Type: ContentEmpty}},
		},
	}
	if err := Validate(r); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidate_Errors(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(*Recipe)
		wantErr string
	}{
		{
			name:    "missing name",
			mutate:  func(r *Recipe) { r.Name = "" },
			wantErr: "name is required",
		},
		{
			name:    "missing partition table",
			mutate:  func(r *Recipe) { r.Device.PartitionTable = "" },
			wantErr: "partition_table is required",
		},
		{
			name:    "unsupported partition table",
			mutate:  func(r *Recipe) { r.Device.PartitionTable = "bsd" },
			wantErr: "unsupported partition_table",
		},
		{
			name:    "no partitions",
			mutate:  func(r *Recipe) { r.Partitions = nil },
			wantErr: "at least one partition",
		},
		{
			name:    "missing partition label",
			mutate:  func(r *Recipe) { r.Partitions[0].Label = "" },
			wantErr: "label is required",
		},
		{
			name:    "duplicate partition label",
			mutate:  func(r *Recipe) { r.Partitions[1].Label = r.Partitions[0].Label },
			wantErr: "duplicate partition label",
		},
		{
			name:    "missing partition size",
			mutate:  func(r *Recipe) { r.Partitions[0].Size = "" },
			wantErr: "size is required",
		},
		{
			name:    "invalid partition size",
			mutate:  func(r *Recipe) { r.Partitions[0].Size = "notasize" },
			wantErr: "invalid size",
		},
		{
			name:    "missing filesystem",
			mutate:  func(r *Recipe) { r.Partitions[0].Filesystem = "" },
			wantErr: "filesystem is required",
		},
		{
			name:    "missing content type",
			mutate:  func(r *Recipe) { r.Partitions[0].Content.Type = "" },
			wantErr: "content type is required",
		},
		{
			name:    "unsupported content type",
			mutate:  func(r *Recipe) { r.Partitions[0].Content.Type = "unknown-type" },
			wantErr: "unsupported content type",
		},
		{
			name:    "catalogue-image missing image id",
			mutate:  func(r *Recipe) { r.Partitions[0].Content = PartitionContent{Type: ContentCatalogueImage} },
			wantErr: "requires image id",
		},
		{
			name:    "rendered-profile missing profile",
			mutate:  func(r *Recipe) { r.Partitions[0].Content = PartitionContent{Type: ContentRenderedProfile} },
			wantErr: "requires profile name",
		},
		{
			name: "multiple remaining partitions",
			mutate: func(r *Recipe) {
				r.Partitions = []Partition{
					{Label: "A", Size: "remaining", Filesystem: "fat32", Content: PartitionContent{Type: ContentEmpty}},
					{Label: "B", Size: "remaining", Filesystem: "ext4", Content: PartitionContent{Type: ContentEmpty}},
				}
			},
			wantErr: "only one partition may use size \"remaining\"",
		},
		{
			name: "remaining not last",
			mutate: func(r *Recipe) {
				r.Partitions = []Partition{
					{Label: "A", Size: "remaining", Filesystem: "fat32", Content: PartitionContent{Type: ContentEmpty}},
					{Label: "B", Size: "1G", Filesystem: "ext4", Content: PartitionContent{Type: ContentEmpty}},
				}
			},
			wantErr: "only the last partition",
		},
	}

	base := func() Recipe {
		return Recipe{
			Name:   "test",
			Device: DeviceSpec{PartitionTable: PartitionTableGPT},
			Boot:   BootSpec{Mode: BootModeUEFIBIOS, Menu: BootMenuiPXE},
			Partitions: []Partition{
				{Label: "BOOT", Size: "2G", Filesystem: "fat32", Content: PartitionContent{Type: ContentBootMenu}},
				{Label: "DATA", Size: "4G", Filesystem: "exfat", Content: PartitionContent{Type: ContentEmpty}},
			},
		}
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := base()
			tc.mutate(&r)
			err := Validate(r)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantErr)
			}
			if tc.wantErr != "" && !contains(err.Error(), tc.wantErr) {
				t.Errorf("error %q does not contain %q", err.Error(), tc.wantErr)
			}
		})
	}
}

func TestParseSize(t *testing.T) {
	cases := []struct {
		input string
		want  int64
		isErr bool
	}{
		{"2G", 2 * (1 << 30), false},
		{"500M", 500 * (1 << 20), false},
		{"1.5G", int64(1.5 * float64(1<<30)), false},
		{"remaining", -1, false},
		{"REMAINING", -1, false},
		{"", 0, true},
		{"notasize", 0, true},
		{"-1G", 0, true},
	}
	for _, tc := range cases {
		got, err := ParseSize(tc.input)
		if tc.isErr {
			if err == nil {
				t.Errorf("ParseSize(%q) expected error, got %d", tc.input, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseSize(%q) unexpected error: %v", tc.input, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ParseSize(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestLoadRecipeFromBytes(t *testing.T) {
	yaml := `
name: test-usb
device:
  partition_table: gpt
boot:
  mode: uefi-bios
  menu: ipxe
partitions:
  - label: BOOT
    size: 2G
    filesystem: fat32
    content:
      type: boot-menu
  - label: DATA
    size: remaining
    filesystem: exfat
    content:
      type: empty
`
	r, err := LoadRecipeFromBytes([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Name != "test-usb" {
		t.Errorf("Name = %q", r.Name)
	}
	if len(r.Partitions) != 2 {
		t.Errorf("partitions count = %d", len(r.Partitions))
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
