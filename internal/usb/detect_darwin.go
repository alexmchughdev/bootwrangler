//go:build darwin

package usb

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// diskutilList is the top-level output of `diskutil list -plist`.
type diskutilList struct {
	AllDisksAndPartitions []diskutilDisk `json:"AllDisksAndPartitions"`
}

type diskutilDisk struct {
	DeviceIdentifier string              `json:"DeviceIdentifier"`
	Partitions       []diskutilPartition `json:"Partitions"`
}

type diskutilPartition struct {
	DeviceIdentifier string `json:"DeviceIdentifier"`
}

// diskutilInfo is the output of `diskutil info -plist /dev/diskN`.
type diskutilInfo struct {
	DeviceIdentifier    string `json:"DeviceIdentifier"`
	DeviceNode          string `json:"DeviceNode"`
	MediaName           string `json:"MediaName"`
	IORegistryEntryName string `json:"IORegistryEntryName"`
	TotalSize           int64  `json:"TotalSize"`
	Removable           bool   `json:"RemovableMedia"`
	Internal            bool   `json:"Internal"`
	MountPoint          string `json:"MountPoint"`
	FilesystemType      string `json:"FilesystemType"`
	VolumeName          string `json:"VolumeName"`
	BusProtocol         string `json:"BusProtocol"`
}

// ListDevices uses diskutil to enumerate block devices on macOS.
func ListDevices() ([]Device, error) {
	listOut, err := exec.Command("diskutil", "list", "-plist").Output()
	if err != nil {
		return nil, fmt.Errorf("diskutil list: %w", err)
	}

	var listing diskutilList
	if err := json.Unmarshal(listOut, &listing); err != nil {
		return nil, fmt.Errorf("diskutil list parse: %w", err)
	}

	var devices []Device
	for _, d := range listing.AllDisksAndPartitions {
		info, err := getDiskInfo(d.DeviceIdentifier)
		if err != nil {
			continue
		}

		dev := Device{
			Path:       "/dev/" + d.DeviceIdentifier,
			Name:       d.DeviceIdentifier,
			Model:      info.MediaName,
			Size:       info.TotalSize,
			SizeHuman:  FormatSize(info.TotalSize),
			Transport:  strings.ToLower(info.BusProtocol),
			Removable:  info.Removable,
			MountPoint: info.MountPoint,
		}

		for _, p := range d.Partitions {
			pinfo, err := getDiskInfo(p.DeviceIdentifier)
			if err != nil {
				continue
			}
			dev.Partitions = append(dev.Partitions, Partition{
				Path:       "/dev/" + p.DeviceIdentifier,
				Name:       p.DeviceIdentifier,
				Size:       pinfo.TotalSize,
				SizeHuman:  FormatSize(pinfo.TotalSize),
				Filesystem: pinfo.FilesystemType,
				MountPoint: pinfo.MountPoint,
				Label:      pinfo.VolumeName,
			})
		}

		ClassifySafety(&dev)
		devices = append(devices, dev)
	}
	return devices, nil
}

func getDiskInfo(identifier string) (diskutilInfo, error) {
	out, err := exec.Command("diskutil", "info", "-plist", "/dev/"+identifier).Output()
	if err != nil {
		return diskutilInfo{}, fmt.Errorf("diskutil info %s: %w", identifier, err)
	}
	var info diskutilInfo
	if err := json.Unmarshal(out, &info); err != nil {
		return diskutilInfo{}, fmt.Errorf("diskutil info parse %s: %w", identifier, err)
	}
	return info, nil
}
