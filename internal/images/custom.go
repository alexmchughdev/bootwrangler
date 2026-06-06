package images

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var hexPattern = regexp.MustCompile(`(?i)^[a-f0-9]+$`)

// CustomImageCatalogue is the user-managed custom image definition file.
type CustomImageCatalogue struct {
	Images []CustomImage `yaml:"images"`
}

// CustomImagesPath returns the default user custom-image catalogue path.
func CustomImagesPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".bootwrangler", "catalogue", "custom-images.yaml")
	}
	return filepath.Join(home, ".bootwrangler", "catalogue", "custom-images.yaml")
}

// LoadDefaultCustomImages loads the user custom-image catalogue if present.
func LoadDefaultCustomImages() ([]CustomImage, error) {
	return LoadCustomImages(CustomImagesPath())
}

// LoadCustomImages loads custom image definitions from a YAML file.
func LoadCustomImages(path string) ([]CustomImage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []CustomImage{}, nil
		}
		return nil, fmt.Errorf("custom images: read %s: %w", path, err)
	}
	return LoadCustomImagesFromBytes(data)
}

// LoadCustomImagesFromBytes parses and validates custom image definitions.
func LoadCustomImagesFromBytes(data []byte) ([]CustomImage, error) {
	if strings.TrimSpace(string(data)) == "" {
		return []CustomImage{}, nil
	}
	var cat CustomImageCatalogue
	if err := yaml.Unmarshal(data, &cat); err != nil {
		return nil, fmt.Errorf("custom images: parse: %w", err)
	}
	if err := ValidateCustomImages(cat.Images); err != nil {
		return nil, err
	}
	return cat.Images, nil
}

// SaveCustomImages validates and writes custom image definitions atomically.
func SaveCustomImages(path string, list []CustomImage) error {
	if err := ValidateCustomImages(list); err != nil {
		return err
	}
	data, err := yaml.Marshal(CustomImageCatalogue{Images: list})
	if err != nil {
		return fmt.Errorf("custom images: encode: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("custom images: mkdir: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".custom-images-*")
	if err != nil {
		return fmt.Errorf("custom images: temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("custom images: write: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("custom images: close: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("custom images: replace: %w", err)
	}
	return nil
}

// UpsertCustomImage validates and writes one custom image by ID.
func UpsertCustomImage(path string, img CustomImage) error {
	list, err := LoadCustomImages(path)
	if err != nil {
		return err
	}
	replaced := false
	for i := range list {
		if list[i].ID == img.ID {
			list[i] = img
			replaced = true
			break
		}
	}
	if !replaced {
		list = append(list, img)
	}
	return SaveCustomImages(path, list)
}

// FindCustomImage returns the custom image with the requested ID.
func FindCustomImage(list []CustomImage, id string) (CustomImage, error) {
	for _, img := range list {
		if img.ID == id {
			return img, nil
		}
	}
	return CustomImage{}, fmt.Errorf("custom image not found: %s", id)
}

// ResolveCustomImageFile validates a local custom-image source and checksum.
func ResolveCustomImageFile(img CustomImage) (string, error) {
	if img.Source.Type != "local-file" {
		return "", fmt.Errorf("custom image %s uses source type %q; download it before flashing",
			img.ID, img.Source.Type)
	}
	info, err := os.Stat(img.Source.Path)
	if err != nil {
		return "", fmt.Errorf("custom image %s: stat %s: %w", img.ID, img.Source.Path, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("custom image %s: source path is a directory: %s", img.ID, img.Source.Path)
	}
	if err := verifyCustomImageChecksum(img.Source.Path, img); err != nil {
		return "", err
	}
	return img.Source.Path, nil
}

// ValidateCustomImages checks user-provided custom image metadata.
func ValidateCustomImages(list []CustomImage) error {
	seen := map[string]bool{}
	for i, img := range list {
		if img.ID == "" {
			return fmt.Errorf("custom images: image[%d]: id is required", i)
		}
		if seen[img.ID] {
			return fmt.Errorf("custom images: duplicate image id %q", img.ID)
		}
		seen[img.ID] = true
		if img.Name == "" {
			return fmt.Errorf("custom images: image %q: name is required", img.ID)
		}
		if err := validateCustomSource(img); err != nil {
			return err
		}
		if !img.Compatibility.WholeDrive && !img.Compatibility.Partition && !img.Compatibility.ISOFileBoot {
			return fmt.Errorf("custom images: image %q: at least one compatibility mode is required", img.ID)
		}
		if err := validateCustomChecksum(img); err != nil {
			return err
		}
	}
	return nil
}

func validateCustomSource(img CustomImage) error {
	switch img.Source.Type {
	case "local-file":
		if img.Source.Path == "" {
			return fmt.Errorf("custom images: image %q: local-file source requires path", img.ID)
		}
	case "url":
		if img.Source.URL == "" {
			return fmt.Errorf("custom images: image %q: url source requires url", img.ID)
		}
		parsed, err := url.Parse(img.Source.URL)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return fmt.Errorf("custom images: image %q: url source must be absolute", img.ID)
		}
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			return fmt.Errorf("custom images: image %q: url source must use http or https", img.ID)
		}
	default:
		return fmt.Errorf("custom images: image %q: unsupported source type %q", img.ID, img.Source.Type)
	}
	return nil
}

func validateCustomChecksum(img CustomImage) error {
	if img.Checksum == nil {
		return nil
	}
	var wantLen int
	switch strings.ToLower(img.Checksum.Type) {
	case "sha256":
		wantLen = 64
	case "md5":
		wantLen = 32
	default:
		return fmt.Errorf("custom images: image %q: unsupported checksum type %q", img.ID, img.Checksum.Type)
	}
	if len(img.Checksum.Value) != wantLen || !hexPattern.MatchString(img.Checksum.Value) {
		return fmt.Errorf("custom images: image %q: invalid %s checksum", img.ID, img.Checksum.Type)
	}
	return nil
}

func verifyCustomImageChecksum(path string, img CustomImage) error {
	if img.Checksum == nil {
		return nil
	}
	switch strings.ToLower(img.Checksum.Type) {
	case "sha256":
		return VerifySHA256(path, img.Checksum.Value)
	case "md5":
		return verifyMD5(path, img.Checksum.Value)
	default:
		return fmt.Errorf("custom image %s: unsupported checksum type %q", img.ID, img.Checksum.Type)
	}
}

func verifyMD5(path, expectedHex string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("verify: open %s: %w", path, err)
	}
	defer f.Close()

	h := md5.New() //nolint:gosec // MD5 is supported only for matching legacy published checksums.
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("verify: hash %s: %w", path, err)
	}
	got := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(got, expectedHex) {
		return fmt.Errorf("verify: checksum mismatch for %s: expected %s, got %s",
			filepath.Base(path), expectedHex, got)
	}
	return nil
}
