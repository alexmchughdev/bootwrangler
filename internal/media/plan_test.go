package media

import (
	"testing"
)

func baseRecipe() Recipe {
	return Recipe{
		Name:   "test-usb",
		Device: DeviceSpec{PartitionTable: PartitionTableGPT},
		Boot:   BootSpec{Mode: BootModeUEFIBIOS, Menu: BootMenuiPXE},
		Partitions: []Partition{
			{Label: "BOOT", Size: "2G", Filesystem: "fat32", Content: PartitionContent{Type: ContentBootMenu}},
			{Label: "DATA", Size: "4G", Filesystem: "exfat", Content: PartitionContent{Type: ContentEmpty}},
		},
	}
}

func TestPlanBuild_Valid(t *testing.T) {
	t.Parallel()

	r := baseRecipe()
	const deviceSize = 16 * (1 << 30) // 16 GB

	plan, err := PlanBuild(r, "/dev/sdb", deviceSize, false)
	if err != nil {
		t.Fatalf("PlanBuild() error = %v, want nil", err)
	}

	if plan.RecipeName != "test-usb" {
		t.Errorf("RecipeName = %q, want %q", plan.RecipeName, "test-usb")
	}
	if plan.DevicePath != "/dev/sdb" {
		t.Errorf("DevicePath = %q, want %q", plan.DevicePath, "/dev/sdb")
	}
	if plan.DeviceSize != deviceSize {
		t.Errorf("DeviceSize = %d, want %d", plan.DeviceSize, deviceSize)
	}
	if len(plan.Actions) != 2 {
		t.Fatalf("len(Actions) = %d, want 2", len(plan.Actions))
	}

	want2G := int64(2 * (1 << 30))
	want4G := int64(4 * (1 << 30))

	if plan.Actions[0].SizeBytes != want2G {
		t.Errorf("Actions[0].SizeBytes = %d, want %d", plan.Actions[0].SizeBytes, want2G)
	}
	if plan.Actions[1].SizeBytes != want4G {
		t.Errorf("Actions[1].SizeBytes = %d, want %d", plan.Actions[1].SizeBytes, want4G)
	}

	wantTotal := want2G + want4G
	if plan.TotalBytes != wantTotal {
		t.Errorf("TotalBytes = %d, want %d", plan.TotalBytes, wantTotal)
	}
}

func TestPlanBuild_TooLarge(t *testing.T) {
	t.Parallel()

	r := baseRecipe()
	// Device is only 4 GB but recipe needs at least 6 GB.
	const deviceSize = 4 * (1 << 30)

	_, err := PlanBuild(r, "/dev/sdb", deviceSize, false)
	if err == nil {
		t.Fatal("PlanBuild() error = nil, want error for oversized recipe")
	}
	if !contains(err.Error(), "recipe requires") {
		t.Errorf("error %q does not contain 'recipe requires'", err.Error())
	}
}

func TestPlanBuild_Remaining(t *testing.T) {
	t.Parallel()

	r := Recipe{
		Name:   "remaining-test",
		Device: DeviceSpec{PartitionTable: PartitionTableGPT},
		Boot:   BootSpec{Mode: BootModeUEFIBIOS, Menu: BootMenuiPXE},
		Partitions: []Partition{
			{Label: "BOOT", Size: "2G", Filesystem: "fat32", Content: PartitionContent{Type: ContentBootMenu}},
			{Label: "STORAGE", Size: "remaining", Filesystem: "exfat", Content: PartitionContent{Type: ContentEmpty}},
		},
	}
	const deviceSize = 8 * (1 << 30) // 8 GB

	plan, err := PlanBuild(r, "/dev/sdb", deviceSize, false)
	if err != nil {
		t.Fatalf("PlanBuild() error = %v, want nil", err)
	}

	if len(plan.Actions) != 2 {
		t.Fatalf("len(Actions) = %d, want 2", len(plan.Actions))
	}

	want2G := int64(2 * (1 << 30))
	wantRemaining := deviceSize - want2G // 6 GB

	if plan.Actions[0].SizeBytes != want2G {
		t.Errorf("Actions[0].SizeBytes = %d, want %d", plan.Actions[0].SizeBytes, want2G)
	}
	if plan.Actions[1].SizeBytes != wantRemaining {
		t.Errorf("Actions[1].SizeBytes = %d, want %d (remaining)", plan.Actions[1].SizeBytes, wantRemaining)
	}
	if plan.TotalBytes != deviceSize {
		t.Errorf("TotalBytes = %d, want %d (full device)", plan.TotalBytes, deviceSize)
	}
}
