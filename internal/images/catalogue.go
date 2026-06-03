package images

import (
	_ "embed"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

//go:embed catalogue.yaml
var builtinCatalogueYAML []byte

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
// Data is sourced from the embedded catalogue.yaml, which is updated by the
// update-catalogue GitHub Action to keep URLs current.
func BuiltinCatalogue() Catalogue {
	cat, err := LoadCatalogueFromBytes(builtinCatalogueYAML)
	if err != nil {
		panic(fmt.Sprintf("builtin catalogue: %v", err))
	}
	return cat
}

// NetbootImage is a distro entry that has direct PXE kernel/initrd URLs.
type NetbootImage struct {
	ID        string
	Name      string
	Family    string
	Version   string
	Arch      string
	KernelURL string
	InitrdURL string
}

// ListNetbootImages returns catalogue entries that have direct netboot kernel/initrd URLs.
func ListNetbootImages(cat Catalogue) []NetbootImage {
	var result []NetbootImage
	for _, entry := range cat.Entries {
		for _, ver := range entry.Versions {
			for _, arch := range ver.Architectures {
				for _, img := range arch.Images {
					if img.NetbootKernelURL == "" {
						continue
					}
					result = append(result, NetbootImage{
						ID:        entry.ID,
						Name:      entry.Name,
						Family:    entry.Family,
						Version:   ver.Version,
						Arch:      arch.Arch,
						KernelURL: img.NetbootKernelURL,
						InitrdURL: img.NetbootInitrdURL,
					})
				}
			}
		}
	}
	return result
}
