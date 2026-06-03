package usb

import (
	"testing"
)

const sampleLsblk = `{"blockdevices":[{"name":"sda","path":"/dev/sda","model":"SanDisk Ultra","size":"32010928128","serial":"4C531001450304101173","tran":"usb","rm":true,"mountpoint":null,"fstype":null,"label":null,"children":[{"name":"sda1","path":"/dev/sda1","size":"536870912","mountpoint":null,"fstype":"vfat","label":"BOOT","tran":null,"rm":null,"model":null,"serial":null},{"name":"sda2","path":"/dev/sda2","size":"31473193984","mountpoint":null,"fstype":"ext4","label":"ROOT","tran":null,"rm":null,"model":null,"serial":null}]},{"name":"nvme0n1","path":"/dev/nvme0n1","model":"Samsung 970 EVO","size":"1000204886016","serial":"S466NX0M123456","tran":"nvme","rm":false,"mountpoint":null,"fstype":null,"label":null,"children":[{"name":"nvme0n1p1","path":"/dev/nvme0n1p1","size":"536870912","mountpoint":"/boot/efi","fstype":"vfat","label":"EFI","tran":null,"rm":null,"model":null,"serial":null},{"name":"nvme0n1p2","path":"/dev/nvme0n1p2","size":"999666487296","mountpoint":"/","fstype":"ext4","label":null,"tran":null,"rm":null,"model":null,"serial":null}]}]}`

func TestParseLsblkOutput(t *testing.T) {
	devices, err := ParseLsblkOutput([]byte(sampleLsblk))
	if err != nil {
		t.Fatalf("ParseLsblkOutput error: %v", err)
	}
	if len(devices) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(devices))
	}

	usb := devices[0]
	if usb.Path != "/dev/sda" {
		t.Errorf("sda path: got %q", usb.Path)
	}
	if usb.Model != "SanDisk Ultra" {
		t.Errorf("sda model: got %q", usb.Model)
	}
	if usb.Transport != "usb" {
		t.Errorf("sda transport: got %q", usb.Transport)
	}
	if !usb.Removable {
		t.Error("sda should be removable")
	}
	if len(usb.Partitions) != 2 {
		t.Fatalf("sda: expected 2 partitions, got %d", len(usb.Partitions))
	}
	if usb.Partitions[0].Filesystem != "vfat" {
		t.Errorf("sda1 fstype: got %q", usb.Partitions[0].Filesystem)
	}
	if usb.Partitions[0].Label != "BOOT" {
		t.Errorf("sda1 label: got %q", usb.Partitions[0].Label)
	}

	nvme := devices[1]
	if nvme.Path != "/dev/nvme0n1" {
		t.Errorf("nvme path: got %q", nvme.Path)
	}
	if nvme.Transport != "nvme" {
		t.Errorf("nvme transport: got %q", nvme.Transport)
	}
	if nvme.Removable {
		t.Error("nvme should not be removable")
	}
	if len(nvme.Partitions) != 2 {
		t.Fatalf("nvme: expected 2 partitions, got %d", len(nvme.Partitions))
	}
}

func TestClassifySafety(t *testing.T) {
	tests := []struct {
		name       string
		dev        Device
		wantSafe   bool
		wantNoteOK func(string) bool
	}{
		{
			name: "usb removable no mounts",
			dev: Device{
				Transport: "usb",
				Removable: true,
				Partitions: []Partition{
					{MountPoint: ""},
				},
			},
			wantSafe:   true,
			wantNoteOK: func(s string) bool { return s == "" },
		},
		{
			name: "usb removable with non-system mount",
			dev: Device{
				Transport: "usb",
				Removable: true,
				Partitions: []Partition{
					{MountPoint: "/mnt/data"},
				},
			},
			wantSafe:   true,
			wantNoteOK: func(s string) bool { return s == "" },
		},
		{
			name: "partition mounted at /",
			dev: Device{
				Transport: "nvme",
				Removable: false,
				Partitions: []Partition{
					{MountPoint: "/"},
				},
			},
			wantSafe:   false,
			wantNoteOK: func(s string) bool { return s == "system disk: partition mounted at /" },
		},
		{
			name: "partition mounted at /boot/efi",
			dev: Device{
				Transport: "nvme",
				Removable: false,
				Partitions: []Partition{
					{MountPoint: "/boot/efi"},
				},
			},
			wantSafe:   false,
			wantNoteOK: func(s string) bool { return s == "system disk: partition mounted at /boot/efi" },
		},
		{
			name: "non-removable sata no system mounts",
			dev: Device{
				Transport: "sata",
				Removable: false,
				Partitions: []Partition{
					{MountPoint: "/mnt/backup"},
				},
			},
			wantSafe:   false,
			wantNoteOK: func(s string) bool { return s == "non-removable sata device" },
		},
		{
			name: "device own mountpoint is system",
			dev: Device{
				Transport:  "usb",
				Removable:  true,
				MountPoint: "/home",
			},
			wantSafe:   false,
			wantNoteOK: func(s string) bool { return s == "system disk: partition mounted at /home" },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dev := tt.dev
			ClassifySafety(&dev)
			if dev.Safe != tt.wantSafe {
				t.Errorf("Safe: got %v, want %v (note: %q)", dev.Safe, tt.wantSafe, dev.SafetyNote)
			}
			if !tt.wantNoteOK(dev.SafetyNote) {
				t.Errorf("SafetyNote: got %q", dev.SafetyNote)
			}
		})
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1 KB"},
		{1536, "2 KB"},
		{1048576, "1 MB"},
		{536870912, "512 MB"},
		{1073741824, "1.0 GB"},
		{17179869184, "16.0 GB"},
		{32010928128, "29.8 GB"},
		{1099511627776, "1.0 TB"},
	}

	for _, tt := range tests {
		got := FormatSize(tt.bytes)
		if got != tt.want {
			t.Errorf("FormatSize(%d) = %q, want %q", tt.bytes, got, tt.want)
		}
	}
}

func TestParseLsblkOutputSafety(t *testing.T) {
	devices, err := ParseLsblkOutput([]byte(sampleLsblk))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// sda is USB removable with no system mounts → safe
	if !devices[0].Safe {
		t.Errorf("sda should be safe, note: %q", devices[0].SafetyNote)
	}

	// nvme has / and /boot/efi mounted → unsafe
	if devices[1].Safe {
		t.Errorf("nvme0n1 should be unsafe")
	}
}
