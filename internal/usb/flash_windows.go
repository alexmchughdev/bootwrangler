//go:build windows

package usb

import (
	"fmt"
	"io"
	"os"
)

// ExecuteFlash writes the image file directly to a physical disk on Windows.
// Windows does not have dd, so we stream via Go's os.File.
// The target must be a physical disk path such as \\.\PhysicalDrive2.
// Returns nil immediately when plan.DryRun is true.
func ExecuteFlash(plan FlashPlan) error {
	if plan.DryRun {
		return nil
	}

	src, err := os.Open(plan.ImagePath)
	if err != nil {
		return fmt.Errorf("flash: open image: %w", err)
	}
	defer src.Close()

	// Open the physical disk with write access.
	// On Windows this requires elevation (Administrator).
	dst, err := os.OpenFile(plan.DevicePath, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("flash: open device %s (run as Administrator): %w", plan.DevicePath, err)
	}
	defer dst.Close()

	const bufSize = 4 * 1024 * 1024 // 4 MiB chunks
	buf := make([]byte, bufSize)
	if _, err := io.CopyBuffer(dst, src, buf); err != nil {
		return fmt.Errorf("flash: write: %w", err)
	}
	return nil
}
