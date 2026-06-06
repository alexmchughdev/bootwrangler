package images

import (
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
