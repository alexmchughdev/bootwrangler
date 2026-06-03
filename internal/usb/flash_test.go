package usb

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlanFlash_Success(t *testing.T) {
	t.Parallel()

	// Create a temporary image file with some content.
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "test.img")
	if err := os.WriteFile(imgPath, make([]byte, 1024), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	dev := Device{
		Path:      "/dev/sdb",
		Name:      "sdb",
		Safe:      true,
		Size:      1024 * 1024, // 1 MB — larger than image
		SizeHuman: "1 MB",
		Transport: "usb",
		Removable: true,
	}

	plan, err := PlanFlash(dev, imgPath, false)
	if err != nil {
		t.Fatalf("PlanFlash() error = %v, want nil", err)
	}
	if plan.ImagePath != imgPath {
		t.Errorf("plan.ImagePath = %q, want %q", plan.ImagePath, imgPath)
	}
	if plan.DevicePath != "/dev/sdb" {
		t.Errorf("plan.DevicePath = %q, want %q", plan.DevicePath, "/dev/sdb")
	}
	if plan.ImageSize != 1024 {
		t.Errorf("plan.ImageSize = %d, want 1024", plan.ImageSize)
	}
	if plan.DeviceSize != 1024*1024 {
		t.Errorf("plan.DeviceSize = %d, want %d", plan.DeviceSize, 1024*1024)
	}
	if plan.DryRun {
		t.Error("plan.DryRun = true, want false")
	}
	if !strings.Contains(plan.Command, "dd") {
		t.Errorf("plan.Command = %q, expected dd command", plan.Command)
	}
}

func TestPlanFlash_DryRun(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	imgPath := filepath.Join(dir, "test.img")
	if err := os.WriteFile(imgPath, make([]byte, 512), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	dev := Device{
		Path:      "/dev/sdb",
		Name:      "sdb",
		Safe:      true,
		Size:      1024 * 1024,
		SizeHuman: "1 MB",
		Transport: "usb",
		Removable: true,
	}

	plan, err := PlanFlash(dev, imgPath, true)
	if err != nil {
		t.Fatalf("PlanFlash() error = %v, want nil", err)
	}
	if !plan.DryRun {
		t.Error("plan.DryRun = false, want true")
	}
	// ExecuteFlash on a dry-run plan is a no-op.
	if err := ExecuteFlash(plan); err != nil {
		t.Fatalf("ExecuteFlash(dryRun) error = %v, want nil", err)
	}
}

func TestPlanFlash_UnsafeDevice(t *testing.T) {
	t.Parallel()

	dev := Device{
		Path:       "/dev/sda",
		Name:       "sda",
		Safe:       false,
		SafetyNote: "system disk: partition mounted at /",
	}

	_, err := PlanFlash(dev, "/some/image.img", false)
	if err == nil {
		t.Fatal("PlanFlash() error = nil, want error for unsafe device")
	}
	if !strings.Contains(err.Error(), "unsafe device") {
		t.Errorf("PlanFlash() error = %q, want 'unsafe device' in message", err.Error())
	}
}

func TestPlanFlash_ImageNotFound(t *testing.T) {
	t.Parallel()

	dev := Device{
		Path:      "/dev/sdb",
		Name:      "sdb",
		Safe:      true,
		Size:      1024 * 1024 * 1024,
		Transport: "usb",
		Removable: true,
	}

	_, err := PlanFlash(dev, "/nonexistent/image.img", false)
	if err == nil {
		t.Fatal("PlanFlash() error = nil, want error for missing image")
	}
}

func TestPlanFlash_ImageTooLarge(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	imgPath := filepath.Join(dir, "large.img")
	// Write a 2-byte file to an image.
	if err := os.WriteFile(imgPath, []byte{0, 0}, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Device with Size=1 so any real file will exceed it.
	dev := Device{
		Path:      "/dev/sdb",
		Name:      "sdb",
		Safe:      true,
		Size:      1,
		SizeHuman: "1 B",
		Transport: "usb",
		Removable: true,
	}

	_, err := PlanFlash(dev, imgPath, false)
	if err == nil {
		t.Fatal("PlanFlash() error = nil, want error for oversized image")
	}
	if !strings.Contains(err.Error(), "exceeds device capacity") {
		t.Errorf("PlanFlash() error = %q, want 'exceeds device capacity' in message", err.Error())
	}
}
