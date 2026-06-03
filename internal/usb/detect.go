// Package usb provides USB and block device discovery and flash operations.
package usb

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

// Device represents a top-level block device.
type Device struct {
	Path       string
	Name       string
	Model      string
	Size       int64
	SizeHuman  string
	Serial     string
	Transport  string
	Removable  bool
	MountPoint string
	Partitions []Partition
	Safe       bool
	SafetyNote string
}

// Partition represents a partition on a block device.
type Partition struct {
	Path       string
	Name       string
	Size       int64
	SizeHuman  string
	Filesystem string
	MountPoint string
	Label      string
}

// lsblkOutput is the top-level JSON structure returned by lsblk -J.
type lsblkOutput struct {
	BlockDevices []lsblkDevice `json:"blockdevices"`
}

// lsblkDevice mirrors one entry in the lsblk JSON.  All fields can be null.
type lsblkDevice struct {
	Name       *string       `json:"name"`
	Path       *string       `json:"path"`
	Model      *string       `json:"model"`
	Size       *string       `json:"size"`
	Serial     *string       `json:"serial"`
	Tran       *string       `json:"tran"`
	RM         *bool         `json:"rm"`
	MountPoint *string       `json:"mountpoint"`
	FSType     *string       `json:"fstype"`
	Label      *string       `json:"label"`
	Children   []lsblkDevice `json:"children"`
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func boolVal(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

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

// ParseLsblkOutput parses raw lsblk JSON output into a slice of Devices.
func ParseLsblkOutput(data []byte) ([]Device, error) {
	var raw lsblkOutput
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse lsblk output: %w", err)
	}

	devices := make([]Device, 0, len(raw.BlockDevices))
	for i := range raw.BlockDevices {
		dev := convertDevice(&raw.BlockDevices[i])
		ClassifySafety(&dev)
		devices = append(devices, dev)
	}
	return devices, nil
}

func convertDevice(raw *lsblkDevice) Device {
	var sizeBytes int64
	if raw.Size != nil {
		fmt.Sscanf(*raw.Size, "%d", &sizeBytes)
	}

	dev := Device{
		Path:       strVal(raw.Path),
		Name:       strVal(raw.Name),
		Model:      strVal(raw.Model),
		Size:       sizeBytes,
		SizeHuman:  FormatSize(sizeBytes),
		Serial:     strVal(raw.Serial),
		Transport:  strVal(raw.Tran),
		Removable:  boolVal(raw.RM),
		MountPoint: strVal(raw.MountPoint),
	}

	for i := range raw.Children {
		dev.Partitions = append(dev.Partitions, convertPartition(&raw.Children[i]))
	}
	return dev
}

func convertPartition(raw *lsblkDevice) Partition {
	var sizeBytes int64
	if raw.Size != nil {
		fmt.Sscanf(*raw.Size, "%d", &sizeBytes)
	}
	return Partition{
		Path:       strVal(raw.Path),
		Name:       strVal(raw.Name),
		Size:       sizeBytes,
		SizeHuman:  FormatSize(sizeBytes),
		Filesystem: strVal(raw.FSType),
		MountPoint: strVal(raw.MountPoint),
		Label:      strVal(raw.Label),
	}
}

// FormatSize formats a byte count as a human-readable string.
func FormatSize(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
		TB = 1024 * GB
	)
	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.1f TB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.0f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.0f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
