// Package images manages OS image catalogue entries, downloads, and local caching.
package images

// ImageType is the kind of image file.
type ImageType string

const (
	ImageTypeISO   ImageType = "iso"
	ImageTypeRaw   ImageType = "raw"
	ImageTypeQCOW2 ImageType = "qcow2"
	ImageTypeOther ImageType = "other"
)

// Compatibility describes how an image can be flashed.
type Compatibility struct {
	WholeDrive  bool `yaml:"whole_drive"`
	Partition   bool `yaml:"partition"`
	ISOFileBoot bool `yaml:"iso_file_boot"`
}

// BootMode is a supported boot firmware type.
type BootMode string

const (
	BootModeBIOS BootMode = "bios"
	BootModeUEFI BootMode = "uefi"
)

// ArchImage is one architecture-specific image download.
type ArchImage struct {
	Type             ImageType     `yaml:"type"`
	URL              string        `yaml:"url"`
	ChecksumURL      string        `yaml:"checksum_url,omitempty"`
	SignatureURL     string        `yaml:"signature_url,omitempty"`
	Compatibility    Compatibility `yaml:"compatibility"`
	BootMode         []BootMode    `yaml:"boot_mode"`
	NetbootKernelURL string        `yaml:"netboot_kernel_url,omitempty"`
	NetbootInitrdURL string        `yaml:"netboot_initrd_url,omitempty"`
}

// ArchEntry is one architecture for a version.
type ArchEntry struct {
	Arch   string      `yaml:"arch"`
	Images []ArchImage `yaml:"images"`
}

// VersionEntry is one OS version.
type VersionEntry struct {
	Version       string      `yaml:"version"`
	Architectures []ArchEntry `yaml:"architectures"`
}

// CatalogueEntry is one OS entry in the catalogue.
type CatalogueEntry struct {
	ID       string         `yaml:"id"`
	Name     string         `yaml:"name"`
	Family   string         `yaml:"family"`
	Versions []VersionEntry `yaml:"versions"`
}

// CustomSource is the source for a custom image.
type CustomSource struct {
	Type string `yaml:"type"` // "local-file" | "url"
	Path string `yaml:"path,omitempty"`
	URL  string `yaml:"url,omitempty"`
}

// Checksum is an expected checksum for image verification.
type Checksum struct {
	Type  string `yaml:"type"` // "sha256" | "md5"
	Value string `yaml:"value"`
}

// CustomImage is a user-defined image (local or remote).
type CustomImage struct {
	ID            string        `yaml:"id"`
	Name          string        `yaml:"name"`
	Source        CustomSource  `yaml:"source"`
	Checksum      *Checksum     `yaml:"checksum,omitempty"`
	Compatibility Compatibility `yaml:"compatibility"`
}

// Catalogue is the loaded collection of official images.
type Catalogue struct {
	Entries []CatalogueEntry `yaml:"entries"`
}
