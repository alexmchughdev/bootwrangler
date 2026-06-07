//go:build windows

package usb

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type psDisk struct {
	Number       int           `json:"Number"`
	FriendlyName string        `json:"FriendlyName"`
	SerialNumber string        `json:"SerialNumber"`
	Size         int64         `json:"Size"`
	BusType      string        `json:"BusType"`
	IsSystem     bool          `json:"IsSystem"`
	IsRemovable  bool          `json:"IsRemovable"`
	BootFromDisk bool          `json:"BootFromDisk"`
	Partitions   []psPartition `json:"Partitions"`
}

type psPartition struct {
	DriveLetter string   `json:"DriveLetter"`
	Size        int64    `json:"Size"`
	Type        string   `json:"Type"`
	AccessPaths []string `json:"AccessPaths"`
}

// ListDevices uses PowerShell to enumerate disks on Windows.
func ListDevices() ([]Device, error) {
	// Query disks and their partitions with volumes in one pass.
	script := `
$disks = Get-Disk | Select-Object Number,FriendlyName,SerialNumber,Size,BusType,IsSystem,IsRemovable,BootFromDisk
$result = @()
foreach ($d in $disks) {
    $parts = @()
    Get-Partition -DiskNumber $d.Number -ErrorAction SilentlyContinue | ForEach-Object {
        $p = $_
        $vol = $p | Get-Volume -ErrorAction SilentlyContinue
        $parts += @{
            DriveLetter = if ($p.DriveLetter) { $p.DriveLetter + ":\" } else { "" }
            Size        = $p.Size
            Type        = $p.Type
            AccessPaths = $p.AccessPaths
        }
    }
    $result += @{
        Number       = $d.Number
        FriendlyName = $d.FriendlyName
        SerialNumber = $d.SerialNumber
        Size         = $d.Size
        BusType      = $d.BusType
        IsSystem     = $d.IsSystem
        IsRemovable  = $d.IsRemovable
        BootFromDisk = $d.BootFromDisk
        Partitions   = $parts
    }
}
$result | ConvertTo-Json -Depth 4
`
	out, err := exec.Command("powershell", "-NoProfile", "-Command", script).Output()
	if err != nil {
		return nil, fmt.Errorf("powershell Get-Disk: %w", err)
	}
	if len(strings.TrimSpace(string(out))) == 0 {
		return nil, nil
	}

	// PowerShell returns a single object (not array) when there's only one disk.
	raw := strings.TrimSpace(string(out))
	if !strings.HasPrefix(raw, "[") {
		raw = "[" + raw + "]"
	}

	var disks []psDisk
	if err := json.Unmarshal([]byte(raw), &disks); err != nil {
		return nil, fmt.Errorf("parse powershell output: %w", err)
	}

	var devices []Device
	for _, d := range disks {
		dev := Device{
			// Windows physical disk path: \\.\PhysicalDriveN
			Path:      fmt.Sprintf(`\\.\PhysicalDrive%d`, d.Number),
			Name:      fmt.Sprintf("PhysicalDrive%d", d.Number),
			Model:     d.FriendlyName,
			Size:      d.Size,
			SizeHuman: FormatSize(d.Size),
			Serial:    d.SerialNumber,
			Transport: strings.ToLower(d.BusType),
			Removable: d.IsRemovable,
		}

		for _, p := range d.Partitions {
			mp := p.DriveLetter
			if mp == "" && len(p.AccessPaths) > 0 {
				mp = p.AccessPaths[0]
			}
			dev.Partitions = append(dev.Partitions, Partition{
				Path:       mp,
				Name:       p.Type,
				Size:       p.Size,
				SizeHuman:  FormatSize(p.Size),
				MountPoint: mp,
			})
			// Track first mount point for safety classification.
			if dev.MountPoint == "" {
				dev.MountPoint = mp
			}
		}

		ClassifySafety(&dev)
		devices = append(devices, dev)
	}
	return devices, nil
}
