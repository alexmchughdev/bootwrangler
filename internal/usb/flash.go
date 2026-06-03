package usb

import (
	"fmt"
	"os"
	"os/exec"
)

// FlashOptions configures a flash operation.
type FlashOptions struct {
	ImagePath  string
	DevicePath string
	DryRun     bool
}

// FlashPlan describes a validated flash operation ready to execute.
type FlashPlan struct {
	ImagePath       string
	DevicePath      string
	ImageSize       int64
	ImageSizeHuman  string
	DeviceSize      int64
	DeviceSizeHuman string
	Command         string
	DryRun          bool
}

// PlanFlash validates the device and image and returns an executable plan.
// It does not write anything to disk.
func PlanFlash(dev Device, imagePath string, dryRun bool) (FlashPlan, error) {
	if !dev.Safe {
		return FlashPlan{}, fmt.Errorf("unsafe device: %s", dev.SafetyNote)
	}

	info, err := os.Stat(imagePath)
	if err != nil {
		return FlashPlan{}, fmt.Errorf("image file: %w", err)
	}

	imageSize := info.Size()
	if imageSize > dev.Size {
		return FlashPlan{}, fmt.Errorf("image size (%s) exceeds device capacity (%s)",
			FormatSize(imageSize), FormatSize(dev.Size))
	}

	cmd := fmt.Sprintf("dd if=%s of=%s bs=4M status=progress", imagePath, dev.Path)

	return FlashPlan{
		ImagePath:       imagePath,
		DevicePath:      dev.Path,
		ImageSize:       imageSize,
		ImageSizeHuman:  FormatSize(imageSize),
		DeviceSize:      dev.Size,
		DeviceSizeHuman: dev.SizeHuman,
		Command:         cmd,
		DryRun:          dryRun,
	}, nil
}

// ExecuteFlash runs the dd command described by plan unless plan.DryRun is true.
func ExecuteFlash(plan FlashPlan) error {
	if plan.DryRun {
		return nil
	}
	if err := exec.Command("dd",
		"if="+plan.ImagePath,
		"of="+plan.DevicePath,
		"bs=4M",
		"status=progress",
	).Run(); err != nil {
		return fmt.Errorf("flash: dd: %w", err)
	}
	return nil
}
