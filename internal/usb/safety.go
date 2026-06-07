package usb

import (
	"fmt"
	"strings"
)

// unixSystemMounts is the set of mount points that indicate a system disk on Linux/macOS.
var unixSystemMounts = map[string]bool{
	"/":         true,
	"/boot":     true,
	"/boot/efi": true,
	"/home":     true,
	"/usr":      true,
	"/var":      true,
	"/etc":      true,
}

// isSystemMount returns true if the mount point looks like a system location
// on any supported platform.
func isSystemMount(mp string) bool {
	if mp == "" {
		return false
	}
	if unixSystemMounts[mp] {
		return true
	}
	// Windows: C:\ is the typical system drive.
	upper := strings.ToUpper(strings.TrimSuffix(mp, "/"))
	return upper == `C:\` || upper == "C:"
}

// ClassifySafety sets dev.Safe and dev.SafetyNote based on transport, removability,
// and mount points.
func ClassifySafety(dev *Device) {
	if isSystemMount(dev.MountPoint) {
		dev.Safe = false
		dev.SafetyNote = fmt.Sprintf("system disk: partition mounted at %s", dev.MountPoint)
		return
	}

	for _, p := range dev.Partitions {
		if isSystemMount(p.MountPoint) {
			dev.Safe = false
			dev.SafetyNote = fmt.Sprintf("system disk: partition mounted at %s", p.MountPoint)
			return
		}
	}

	// Non-removable, non-USB devices are unsafe.
	if dev.Transport != "usb" && !dev.Removable {
		dev.Safe = false
		dev.SafetyNote = fmt.Sprintf("non-removable %s device", dev.Transport)
		return
	}

	dev.Safe = true
	dev.SafetyNote = ""
}
