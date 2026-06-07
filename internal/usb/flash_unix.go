//go:build linux || darwin

package usb

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// ExecuteFlash writes the image to the device using dd.
// On macOS the device is unmounted first and the raw-disk path (/dev/rdiskN)
// is used for speed. On Linux the path is used as-is with bs=4M.
// Returns nil immediately when plan.DryRun is true.
func ExecuteFlash(plan FlashPlan) error {
	if plan.DryRun {
		return nil
	}

	target := plan.DevicePath
	bs := "4M"

	if runtime.GOOS == "darwin" {
		bs = "4m"
		// Use the raw device path for faster throughput on macOS.
		target = toRawDevice(plan.DevicePath)
		// Unmount all partitions on the disk before writing.
		if err := exec.Command("diskutil", "unmountDisk", plan.DevicePath).Run(); err != nil {
			return fmt.Errorf("flash: unmount %s: %w", plan.DevicePath, err)
		}
	}

	if err := exec.Command("dd",
		"if="+plan.ImagePath,
		"of="+target,
		"bs="+bs,
		"status=progress",
	).Run(); err != nil {
		return fmt.Errorf("flash: dd: %w", err)
	}
	return nil
}

// toRawDevice converts /dev/diskN to /dev/rdiskN for macOS raw-disk access.
// On Linux this is a no-op.
func toRawDevice(path string) string {
	if runtime.GOOS != "darwin" {
		return path
	}
	// /dev/disk2 → /dev/rdisk2
	const prefix = "/dev/disk"
	if strings.HasPrefix(path, prefix) {
		return "/dev/r" + path[len("/dev/"):]
	}
	return path
}
