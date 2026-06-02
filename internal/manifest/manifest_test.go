package manifest_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alexmchughdev/bootwrangler/internal/manifest"
)

func TestManifestAddFileAndWarning(t *testing.T) {
	m := &manifest.Manifest{
		ProfileName: "test",
		OSFamily:    "alpine",
		OSVersion:   "3.20",
		Renderer:    "alpine",
	}
	m.AddFile("answerfile", "Alpine setup-alpine answer file")
	m.AddWarning("example warning")

	if len(m.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(m.Files))
	}
	if m.Files[0].Path != "answerfile" {
		t.Errorf("unexpected file path: %s", m.Files[0].Path)
	}
	if len(m.Warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(m.Warnings))
	}
}

func TestComputeChecksumsAndWrite(t *testing.T) {
	dir := t.TempDir()
	content := []byte("hello renderer\n")
	if err := os.WriteFile(filepath.Join(dir, "user-data"), content, 0o644); err != nil {
		t.Fatal(err)
	}

	m := &manifest.Manifest{
		ProfileName: "p",
		OSFamily:    "ubuntu",
		OSVersion:   "24.04",
		Renderer:    "ubuntu",
	}
	m.AddFile("user-data", "ubuntu autoinstall config")

	if err := m.ComputeChecksums(dir); err != nil {
		t.Fatalf("ComputeChecksums: %v", err)
	}
	if m.Files[0].SHA256 == "" {
		t.Error("expected SHA256 to be populated")
	}

	if err := m.WriteSHA256SUMS(dir); err != nil {
		t.Fatalf("WriteSHA256SUMS: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "SHA256SUMS")); err != nil {
		t.Errorf("SHA256SUMS file missing: %v", err)
	}

	if err := m.WriteManifestFile(dir); err != nil {
		t.Fatalf("WriteManifestFile: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "manifest.yaml")); err != nil {
		t.Errorf("manifest.yaml missing: %v", err)
	}
}
