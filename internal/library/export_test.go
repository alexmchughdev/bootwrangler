package library_test

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexmchughdev/bootwrangler/internal/library"
)

func TestExportAndImportRoundtrip(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "export.zip")

	p := minimalProfile("export-test")
	if err := library.ExportBundle(p, zipPath); err != nil {
		t.Fatalf("ExportBundle: %v", err)
	}

	got, err := library.ImportBundle(zipPath)
	if err != nil {
		t.Fatalf("ImportBundle: %v", err)
	}
	if got.Name != p.Name {
		t.Errorf("name mismatch: got %q, want %q", got.Name, p.Name)
	}
	if got.OS.Family != p.OS.Family {
		t.Errorf("os.family mismatch: got %q, want %q", got.OS.Family, p.OS.Family)
	}
}

func TestExportRejectsInvalidProfile(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "bad.zip")

	p := minimalProfile("export-test")
	p.Name = ""
	err := library.ExportBundle(p, zipPath)
	if err == nil {
		t.Error("expected ExportBundle to reject profile with empty name")
	}
}

func TestImportRejectsMissingProfileYAML(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "no-profile.zip")

	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	entry, _ := w.Create("README.md")
	_, _ = entry.Write([]byte("hello"))
	_ = w.Close()
	_ = f.Close()

	_, err = library.ImportBundle(zipPath)
	if err == nil {
		t.Error("expected ImportBundle to fail on missing profile.yaml")
	}
}

func TestImportRejectsInvalidProfile(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "invalid-profile.zip")

	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	entry, _ := w.Create("profile.yaml")
	_, _ = entry.Write([]byte("name: \nos:\n  family: alpine\n"))
	_ = w.Close()
	_ = f.Close()

	_, err = library.ImportBundle(zipPath)
	if err == nil {
		t.Error("expected ImportBundle to fail on invalid profile")
	}
}
