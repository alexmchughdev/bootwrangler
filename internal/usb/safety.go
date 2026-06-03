package usb

import "fmt"

// systemMounts is the set of mount points that indicate a system disk.
var systemMounts = map[string]bool{
	"/":         true,
	"/boot":     true,
	"/boot/efi": true,
	"/home":     true,
	"/usr":      true,
	"/var":      true,
	"/etc":      true,
}

// ClassifySafety sets dev.Safe and dev.SafetyNote based on transport, removability,
// and mount points.
func ClassifySafety(dev *Device) {
	// Check the device's own mount point.
	if systemMounts[dev.MountPoint] {
		dev.Safe = false
		dev.SafetyNote = fmt.Sprintf("system disk: partition mounted at %s", dev.MountPoint)
		return
	}

	// Check each partition's mount point.
	for _, p := range dev.Partitions {
		if systemMounts[p.MountPoint] {
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
