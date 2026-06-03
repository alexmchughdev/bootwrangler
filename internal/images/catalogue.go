package images

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadCatalogueFromBytes parses a catalogue from YAML bytes.
func LoadCatalogueFromBytes(data []byte) (Catalogue, error) {
	var cat Catalogue
	if err := yaml.Unmarshal(data, &cat); err != nil {
		return Catalogue{}, fmt.Errorf("catalogue: parse: %w", err)
	}
	if err := ValidateCatalogue(cat); err != nil {
		return Catalogue{}, err
	}
	return cat, nil
}

// LoadCatalogue parses a catalogue from a YAML file.
func LoadCatalogue(path string) (Catalogue, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Catalogue{}, fmt.Errorf("catalogue: read %s: %w", path, err)
	}
	return LoadCatalogueFromBytes(data)
}

// ValidateCatalogue checks catalogue entries for required fields.
func ValidateCatalogue(cat Catalogue) error {
	seen := map[string]bool{}
	for i, entry := range cat.Entries {
		if entry.ID == "" {
			return fmt.Errorf("catalogue: entry[%d]: id is required", i)
		}
		if entry.Name == "" {
			return fmt.Errorf("catalogue: entry[%d] %q: name is required", i, entry.ID)
		}
		if entry.Family == "" {
			return fmt.Errorf("catalogue: entry[%d] %q: family is required", i, entry.ID)
		}
		if seen[entry.ID] {
			return fmt.Errorf("catalogue: duplicate entry id %q", entry.ID)
		}
		seen[entry.ID] = true
		for j, v := range entry.Versions {
			if v.Version == "" {
				return fmt.Errorf("catalogue: entry %q version[%d]: version is required", entry.ID, j)
			}
			for k, a := range v.Architectures {
				if a.Arch == "" {
					return fmt.Errorf("catalogue: entry %q version %s arch[%d]: arch is required", entry.ID, v.Version, k)
				}
			}
		}
	}
	return nil
}

// FindImage returns the first matching ArchImage and its entry/version/arch.
// Returns an error if no match is found.
func FindImage(cat Catalogue, id, version, arch string) (CatalogueEntry, VersionEntry, ArchEntry, ArchImage, error) {
	for _, entry := range cat.Entries {
		if entry.ID != id {
			continue
		}
		for _, v := range entry.Versions {
			if version != "" && v.Version != version {
				continue
			}
			for _, a := range v.Architectures {
				if arch != "" && a.Arch != arch {
					continue
				}
				if len(a.Images) == 0 {
					continue
				}
				return entry, v, a, a.Images[0], nil
			}
		}
	}
	return CatalogueEntry{}, VersionEntry{}, ArchEntry{}, ArchImage{},
		fmt.Errorf("image not found: id=%s version=%s arch=%s", id, version, arch)
}

// BuiltinCatalogue returns the built-in catalogue of official OS images.
func BuiltinCatalogue() Catalogue {
	return Catalogue{
		Entries: []CatalogueEntry{
			{
				ID:     "ubuntu-server",
				Name:   "Ubuntu Server",
				Family: "ubuntu",
				Versions: []VersionEntry{
					{
						Version: "24.04",
						Architectures: []ArchEntry{
							{
								Arch: "x86_64",
								Images: []ArchImage{
									{
										Type:        ImageTypeISO,
										URL:         "https://releases.ubuntu.com/24.04/ubuntu-24.04-live-server-amd64.iso",
										ChecksumURL: "https://releases.ubuntu.com/24.04/SHA256SUMS",
										Compatibility: Compatibility{
											WholeDrive:  true,
											Partition:   false,
											ISOFileBoot: true,
										},
										BootMode: []BootMode{BootModeBIOS, BootModeUEFI},
									},
								},
							},
						},
					},
					{
						Version: "22.04",
						Architectures: []ArchEntry{
							{
								Arch: "x86_64",
								Images: []ArchImage{
									{
										Type:        ImageTypeISO,
										URL:         "https://releases.ubuntu.com/22.04/ubuntu-22.04.4-live-server-amd64.iso",
										ChecksumURL: "https://releases.ubuntu.com/22.04/SHA256SUMS",
										Compatibility: Compatibility{
											WholeDrive:  true,
											Partition:   false,
											ISOFileBoot: true,
										},
										BootMode: []BootMode{BootModeBIOS, BootModeUEFI},
									},
								},
							},
						},
					},
				},
			},
			{
				ID:     "ubuntu-desktop",
				Name:   "Ubuntu Desktop",
				Family: "ubuntu",
				Versions: []VersionEntry{
					{
						Version: "24.04",
						Architectures: []ArchEntry{
							{
								Arch: "x86_64",
								Images: []ArchImage{
									{
										Type:        ImageTypeISO,
										URL:         "https://releases.ubuntu.com/24.04/ubuntu-24.04-desktop-amd64.iso",
										ChecksumURL: "https://releases.ubuntu.com/24.04/SHA256SUMS",
										Compatibility: Compatibility{
											WholeDrive:  true,
											Partition:   false,
											ISOFileBoot: true,
										},
										BootMode: []BootMode{BootModeBIOS, BootModeUEFI},
									},
								},
							},
						},
					},
				},
			},
			{
				ID:     "debian-netinst",
				Name:   "Debian Netinstall",
				Family: "debian",
				Versions: []VersionEntry{
					{
						Version: "12",
						Architectures: []ArchEntry{
							{
								Arch: "x86_64",
								Images: []ArchImage{
									{
										Type:        ImageTypeISO,
										URL:         "https://cdimage.debian.org/debian-cd/current/amd64/iso-cd/debian-12.5.0-amd64-netinst.iso",
										ChecksumURL: "https://cdimage.debian.org/debian-cd/current/amd64/iso-cd/SHA256SUMS",
										Compatibility: Compatibility{
											WholeDrive:  true,
											Partition:   false,
											ISOFileBoot: true,
										},
										BootMode: []BootMode{BootModeBIOS, BootModeUEFI},
									},
								},
							},
						},
					},
				},
			},
			{
				ID:     "fedora-server",
				Name:   "Fedora Server",
				Family: "fedora",
				Versions: []VersionEntry{
					{
						Version: "40",
						Architectures: []ArchEntry{
							{
								Arch: "x86_64",
								Images: []ArchImage{
									{
										Type:        ImageTypeISO,
										URL:         "https://download.fedoraproject.org/pub/fedora/linux/releases/40/Server/x86_64/iso/Fedora-Server-dvd-x86_64-40-1.14.iso",
										ChecksumURL: "https://download.fedoraproject.org/pub/fedora/linux/releases/40/Server/x86_64/iso/Fedora-Server-40-1.14-x86_64-CHECKSUM",
										Compatibility: Compatibility{
											WholeDrive:  true,
											Partition:   false,
											ISOFileBoot: true,
										},
										BootMode: []BootMode{BootModeBIOS, BootModeUEFI},
									},
								},
							},
						},
					},
				},
			},
			{
				ID:     "rocky-linux",
				Name:   "Rocky Linux",
				Family: "rocky",
				Versions: []VersionEntry{
					{
						Version: "9",
						Architectures: []ArchEntry{
							{
								Arch: "x86_64",
								Images: []ArchImage{
									{
										Type:        ImageTypeISO,
										URL:         "https://download.rockylinux.org/pub/rocky/9/isos/x86_64/Rocky-9.3-x86_64-dvd.iso",
										ChecksumURL: "https://download.rockylinux.org/pub/rocky/9/isos/x86_64/CHECKSUM",
										Compatibility: Compatibility{
											WholeDrive:  true,
											Partition:   false,
											ISOFileBoot: true,
										},
										BootMode: []BootMode{BootModeBIOS, BootModeUEFI},
									},
								},
							},
						},
					},
				},
			},
			{
				ID:     "alpine-standard",
				Name:   "Alpine Linux",
				Family: "alpine",
				Versions: []VersionEntry{
					{
						Version: "3.19",
						Architectures: []ArchEntry{
							{
								Arch: "x86_64",
								Images: []ArchImage{
									{
										Type:        ImageTypeISO,
										URL:         "https://dl-cdn.alpinelinux.org/alpine/v3.19/releases/x86_64/alpine-standard-3.19.1-x86_64.iso",
										ChecksumURL: "https://dl-cdn.alpinelinux.org/alpine/v3.19/releases/x86_64/alpine-standard-3.19.1-x86_64.iso.sha256",
										Compatibility: Compatibility{
											WholeDrive:  true,
											Partition:   false,
											ISOFileBoot: true,
										},
										BootMode: []BootMode{BootModeBIOS, BootModeUEFI},
									},
								},
							},
						},
					},
				},
			},
			{
				ID:     "arch-linux",
				Name:   "Arch Linux",
				Family: "arch",
				Versions: []VersionEntry{
					{
						Version: "2024.01",
						Architectures: []ArchEntry{
							{
								Arch: "x86_64",
								Images: []ArchImage{
									{
										Type:        ImageTypeISO,
										URL:         "https://mirror.rackspace.com/archlinux/iso/2024.01.01/archlinux-2024.01.01-x86_64.iso",
										ChecksumURL: "https://mirror.rackspace.com/archlinux/iso/2024.01.01/sha256sums.txt",
										Compatibility: Compatibility{
											WholeDrive:  true,
											Partition:   false,
											ISOFileBoot: true,
										},
										BootMode: []BootMode{BootModeBIOS, BootModeUEFI},
									},
								},
							},
						},
					},
				},
			},
			{
				ID:     "opensuse-leap",
				Name:   "openSUSE Leap",
				Family: "opensuse",
				Versions: []VersionEntry{
					{
						Version: "15.5",
						Architectures: []ArchEntry{
							{
								Arch: "x86_64",
								Images: []ArchImage{
									{
										Type:        ImageTypeISO,
										URL:         "https://download.opensuse.org/distribution/leap/15.5/iso/openSUSE-Leap-15.5-DVD-x86_64-Media.iso",
										ChecksumURL: "https://download.opensuse.org/distribution/leap/15.5/iso/openSUSE-Leap-15.5-DVD-x86_64-Media.iso.sha256",
										Compatibility: Compatibility{
											WholeDrive:  true,
											Partition:   false,
											ISOFileBoot: true,
										},
										BootMode: []BootMode{BootModeBIOS, BootModeUEFI},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
