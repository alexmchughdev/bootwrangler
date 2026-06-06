package images

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const validCustomImagesYAML = `
images:
  - id: company-os
    name: Company OS
    source:
      type: local-file
      path: /tmp/company-os.iso
    checksum:
      type: sha256
      value: 0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
    compatibility:
      whole_drive: true
      partition: false
      iso_file_boot: true
`

func TestLoadCustomImagesFromBytes(t *testing.T) {
	got, err := LoadCustomImagesFromBytes([]byte(validCustomImagesYAML))
	if err != nil {
		t.Fatalf("LoadCustomImagesFromBytes() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("LoadCustomImagesFromBytes() len = %d, want 1", len(got))
	}
	if got[0].ID != "company-os" {
		t.Fatalf("LoadCustomImagesFromBytes()[0].ID = %q, want company-os", got[0].ID)
	}
	if got[0].Source.Type != "local-file" {
		t.Fatalf("LoadCustomImagesFromBytes()[0].Source.Type = %q, want local-file", got[0].Source.Type)
	}
}

func TestLoadCustomImagesMissingFileReturnsEmpty(t *testing.T) {
	got, err := LoadCustomImages(filepath.Join(t.TempDir(), "missing.yaml"))
	if err != nil {
		t.Fatalf("LoadCustomImages() error = %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("LoadCustomImages() len = %d, want 0", len(got))
	}
}

func TestSaveAndUpsertCustomImages(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".bootwrangler", "catalogue", "custom-images.yaml")
	first := validCustomImage("company-os")
	first.Name = "Company OS"
	if err := UpsertCustomImage(path, first); err != nil {
		t.Fatalf("UpsertCustomImage() first error = %v", err)
	}

	second := validCustomImage("rescue-os")
	second.Name = "Rescue OS"
	if err := UpsertCustomImage(path, second); err != nil {
		t.Fatalf("UpsertCustomImage() second error = %v", err)
	}

	replacement := validCustomImage("company-os")
	replacement.Name = "Company OS Updated"
	replacement.Compatibility = Compatibility{WholeDrive: true, ISOFileBoot: true}
	if err := UpsertCustomImage(path, replacement); err != nil {
		t.Fatalf("UpsertCustomImage() replacement error = %v", err)
	}

	got, err := LoadCustomImages(path)
	if err != nil {
		t.Fatalf("LoadCustomImages() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("LoadCustomImages() len = %d, want 2", len(got))
	}
	if got[0].ID != "company-os" || got[0].Name != "Company OS Updated" {
		t.Fatalf("first custom image = %#v, want updated company-os", got[0])
	}
	if got[1].ID != "rescue-os" {
		t.Fatalf("second custom image id = %q, want rescue-os", got[1].ID)
	}
}

func TestFindCustomImage(t *testing.T) {
	list := []CustomImage{
		validCustomImage("company-os"),
		validCustomImage("rescue-os"),
	}

	got, err := FindCustomImage(list, "rescue-os")
	if err != nil {
		t.Fatalf("FindCustomImage() error = %v", err)
	}
	if got.ID != "rescue-os" {
		t.Fatalf("FindCustomImage().ID = %q, want rescue-os", got.ID)
	}
}

func TestResolveCustomImageFileVerifiesChecksum(t *testing.T) {
	path := filepath.Join(t.TempDir(), "company-os.iso")
	content := []byte("custom image bytes")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	sum := sha256.Sum256(content)
	img := validCustomImage("company-os")
	img.Source.Path = path
	img.Checksum = &Checksum{Type: "sha256", Value: hex.EncodeToString(sum[:])}

	got, err := ResolveCustomImageFile(img)
	if err != nil {
		t.Fatalf("ResolveCustomImageFile() error = %v", err)
	}
	if got != path {
		t.Fatalf("ResolveCustomImageFile() = %q, want %q", got, path)
	}
}

func TestResolveCustomImageFileVerifiesMD5Checksum(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.img")
	content := []byte("legacy image bytes")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	sum := md5.Sum(content)
	img := validCustomImage("legacy-os")
	img.Source.Path = path
	img.Checksum = &Checksum{Type: "md5", Value: hex.EncodeToString(sum[:])}

	if _, err := ResolveCustomImageFile(img); err != nil {
		t.Fatalf("ResolveCustomImageFile() error = %v", err)
	}
}

func TestResolveCustomImageFileRejectsChecksumMismatch(t *testing.T) {
	path := filepath.Join(t.TempDir(), "company-os.iso")
	if err := os.WriteFile(path, []byte("custom image bytes"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	img := validCustomImage("company-os")
	img.Source.Path = path
	img.Checksum = &Checksum{
		Type:  "sha256",
		Value: "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
	}

	_, err := ResolveCustomImageFile(img)
	if err == nil {
		t.Fatal("ResolveCustomImageFile() error = nil, want checksum mismatch")
	}
	if !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("ResolveCustomImageFile() error = %q, want mismatch message", err.Error())
	}
}

func TestValidateCustomImagesRejectsDuplicateID(t *testing.T) {
	list := []CustomImage{
		validCustomImage("company-os"),
		validCustomImage("company-os"),
	}

	err := ValidateCustomImages(list)
	if err == nil {
		t.Fatal("ValidateCustomImages() error = nil, want duplicate id error")
	}
	if !strings.Contains(err.Error(), "duplicate image id") {
		t.Fatalf("ValidateCustomImages() error = %q, want duplicate id message", err.Error())
	}
}

func TestValidateCustomImagesRejectsInvalidSource(t *testing.T) {
	img := validCustomImage("bad-source")
	img.Source = CustomSource{Type: "url", URL: "ftp://example.com/image.iso"}

	err := ValidateCustomImages([]CustomImage{img})
	if err == nil {
		t.Fatal("ValidateCustomImages() error = nil, want invalid source error")
	}
	if !strings.Contains(err.Error(), "http or https") {
		t.Fatalf("ValidateCustomImages() error = %q, want http/https message", err.Error())
	}
}

func TestValidateCustomImagesRequiresCompatibility(t *testing.T) {
	img := validCustomImage("no-compatibility")
	img.Compatibility = Compatibility{}

	err := ValidateCustomImages([]CustomImage{img})
	if err == nil {
		t.Fatal("ValidateCustomImages() error = nil, want compatibility error")
	}
	if !strings.Contains(err.Error(), "at least one compatibility mode") {
		t.Fatalf("ValidateCustomImages() error = %q, want compatibility message", err.Error())
	}
}

func TestValidateCustomImagesRejectsInvalidChecksum(t *testing.T) {
	img := validCustomImage("bad-checksum")
	img.Checksum = &Checksum{Type: "sha256", Value: "not-a-hash"}

	err := ValidateCustomImages([]CustomImage{img})
	if err == nil {
		t.Fatal("ValidateCustomImages() error = nil, want checksum error")
	}
	if !strings.Contains(err.Error(), "invalid sha256 checksum") {
		t.Fatalf("ValidateCustomImages() error = %q, want checksum message", err.Error())
	}
}

func validCustomImage(id string) CustomImage {
	return CustomImage{
		ID:   id,
		Name: "Company OS",
		Source: CustomSource{
			Type: "local-file",
			Path: "/tmp/company-os.iso",
		},
		Compatibility: Compatibility{
			WholeDrive: true,
		},
	}
}
