//go:build linux

package usb

import (
	"fmt"
	"os/exec"
)

// ListDevices runs lsblk and returns discovered block devices.
func ListDevices() ([]Device, error) {
	out, err := exec.Command(
		"lsblk", "-J", "-b",
		"-o", "NAME,PATH,MODEL,SIZE,SERIAL,TRAN,RM,MOUNTPOINT,FSTYPE,LABEL",
	).Output()
	if err != nil {
		return nil, fmt.Errorf("lsblk: %w", err)
	}
	return ParseLsblkOutput(out)
}
