// Package media defines the media recipe schema for multi-partition USB and
// disk layouts.
package media

// ContentType describes what is written to a partition.
type ContentType string

const (
	ContentBootMenu        ContentType = "boot-menu"
	ContentCatalogueImage  ContentType = "catalogue-image"
	ContentCustomImage     ContentType = "custom-image"
	ContentImageFile       ContentType = "image-file"
	ContentRenderedProfile ContentType = "rendered-profile"
	ContentProfileBundle   ContentType = "profile-bundle"
	ContentEmpty           ContentType = "empty"
	ContentCustomFiles     ContentType = "custom-files"
)

// PartitionTable is the disk label type.
type PartitionTable string

const (
	PartitionTableGPT PartitionTable = "gpt"
	PartitionTableMBR PartitionTable = "mbr"
)

// BootMenuType selects which boot menu generator to use.
type BootMenuType string

const (
	BootMenuiPXE     BootMenuType = "ipxe"
	BootMenuGRUB     BootMenuType = "grub"
	BootMenuSyslinux BootMenuType = "syslinux"
)

// BootMode is the supported firmware boot mode for the medium.
type BootMode string

const (
	BootModeUEFIBIOS BootMode = "uefi-bios"
	BootModeUEFI     BootMode = "uefi"
	BootModeBIOS     BootMode = "bios"
)

// DeviceSpec describes the target device layout.
type DeviceSpec struct {
	PartitionTable PartitionTable `yaml:"partition_table"`
}

// BootSpec describes the boot configuration for the medium.
type BootSpec struct {
	Mode BootMode     `yaml:"mode"`
	Menu BootMenuType `yaml:"menu"`
}

// PartitionContent describes the content assigned to a partition.
type PartitionContent struct {
	Type    ContentType `yaml:"type"`
	Image   string      `yaml:"image,omitempty"`
	Version string      `yaml:"version,omitempty"`
	Profile string      `yaml:"profile,omitempty"`
	Bundle  string      `yaml:"bundle,omitempty"`
}

// Partition defines one partition in a media recipe.
type Partition struct {
	Label      string           `yaml:"label"`
	Size       string           `yaml:"size"` // e.g. "2G", "500M", "remaining"
	Filesystem string           `yaml:"filesystem"`
	Content    PartitionContent `yaml:"content"`
}

// Recipe is the top-level media recipe document.
type Recipe struct {
	Name       string      `yaml:"name"`
	Device     DeviceSpec  `yaml:"device"`
	Boot       BootSpec    `yaml:"boot"`
	Partitions []Partition `yaml:"partitions"`
}
