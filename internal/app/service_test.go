package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alexmchughdev/bootwrangler/internal/profile"
)

func TestServiceHealth(t *testing.T) {
	t.Parallel()

	service := NewService()

	got := service.Health()

	if got.Status != "ok" {
		t.Fatalf("Health().Status = %q, want %q", got.Status, "ok")
	}
	if got.Version == "" {
		t.Fatal("Health().Version is empty")
	}
}

func TestServiceValidateProfile(t *testing.T) {
	t.Parallel()

	service := NewService()
	result := service.ValidateProfile(profile.Profile{})

	if result.Valid {
		t.Fatal("ValidateProfile().Valid = true, want false")
	}
	if len(result.Problems) == 0 {
		t.Fatal("ValidateProfile().Problems is empty, want validation problems")
	}
}

func TestServiceLoadAndSaveProfile(t *testing.T) {
	t.Parallel()

	service := NewService()
	source, err := profile.LoadAndValidateFile("../../examples/profiles/ubuntu-server.yaml")
	if err != nil {
		t.Fatalf("LoadAndValidateFile() error = %v", err)
	}
	path := filepath.Join(t.TempDir(), "profile.yaml")

	if err := service.SaveProfile(path, source); err != nil {
		t.Fatalf("SaveProfile() error = %v", err)
	}
	got, err := service.LoadProfile(path)
	if err != nil {
		t.Fatalf("LoadProfile() error = %v", err)
	}
	if got.Name != source.Name {
		t.Fatalf("LoadProfile().Name = %q, want %q", got.Name, source.Name)
	}
}

func TestServiceLibraryRoundtrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	svc := NewService()
	source, err := profile.LoadAndValidateFile("../../examples/profiles/ubuntu-server.yaml")
	if err != nil {
		t.Fatalf("LoadAndValidateFile() error = %v", err)
	}

	if err := svc.LibraryInit(); err != nil {
		t.Fatalf("LibraryInit() error = %v", err)
	}

	name, err := svc.LibraryAdd(source)
	if err != nil {
		t.Fatalf("LibraryAdd() error = %v", err)
	}
	if name == "" {
		t.Fatal("LibraryAdd() returned empty name")
	}

	got, err := svc.LibraryGet(source.Name)
	if err != nil {
		t.Fatalf("LibraryGet() error = %v", err)
	}
	if got.Name != source.Name {
		t.Fatalf("LibraryGet().Name = %q, want %q", got.Name, source.Name)
	}

	entries, err := svc.LibraryList()
	if err != nil {
		t.Fatalf("LibraryList() len error = %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("LibraryList() len = %d, want 1", len(entries))
	}

	if err := svc.LibraryRemove(source.Name); err != nil {
		t.Fatalf("LibraryRemove() error = %v", err)
	}
}

func TestServiceLibraryExportImportBundle(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	svc := NewService()
	source, err := profile.LoadAndValidateFile("../../examples/profiles/ubuntu-server.yaml")
	if err != nil {
		t.Fatalf("LoadAndValidateFile() error = %v", err)
	}

	if err := svc.LibraryInit(); err != nil {
		t.Fatalf("LibraryInit() error = %v", err)
	}
	if _, err := svc.LibraryAdd(source); err != nil {
		t.Fatalf("LibraryAdd() error = %v", err)
	}

	zipPath := filepath.Join(t.TempDir(), "profile.zip")
	if err := svc.LibraryExportBundle(source.Name, zipPath); err != nil {
		t.Fatalf("LibraryExportBundle() error = %v", err)
	}
	if _, err := os.Stat(zipPath); err != nil {
		t.Fatalf("exported bundle does not exist: %v", err)
	}
}

func TestServiceLibraryVersioning(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	svc := NewService()
	source, err := profile.LoadAndValidateFile("../../examples/profiles/ubuntu-server.yaml")
	if err != nil {
		t.Fatalf("LoadAndValidateFile() error = %v", err)
	}

	if err := svc.LibraryInit(); err != nil {
		t.Fatalf("LibraryInit() error = %v", err)
	}
	if _, err := svc.LibraryAdd(source); err != nil {
		t.Fatalf("LibraryAdd() error = %v", err)
	}

	if err := svc.LibraryCommit(source.Name, "initial version"); err != nil {
		t.Fatalf("LibraryCommit() error = %v", err)
	}

	entries, err := svc.LibraryHistory(source.Name)
	if err != nil {
		t.Fatalf("LibraryHistory() error = %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("LibraryHistory() returned empty, want at least one entry")
	}
	if entries[0].Message != "initial version" {
		t.Fatalf("LibraryHistory()[0].Message = %q, want %q", entries[0].Message, "initial version")
	}
}

func TestServicePlanCataloguePartitionFlashRejectsIncompatibleImageBeforeDeviceProbe(t *testing.T) {
	t.Parallel()

	svc := NewService()

	_, err := svc.PlanCataloguePartitionFlash("/dev/sdb", "/dev/sdb1", "ubuntu-server", "24.04", "x86_64")
	if err == nil {
		t.Fatal("PlanCataloguePartitionFlash() error = nil, want compatibility error")
	}
	if !strings.Contains(err.Error(), "not marked as partition-flash compatible") {
		t.Fatalf("PlanCataloguePartitionFlash() error = %q, want partition compatibility message", err.Error())
	}
}

func TestServicePlanCatalogueFlashRequiresCachedImageBeforeDeviceProbe(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	svc := NewService()

	_, err := svc.PlanCatalogueFlash("/dev/sdb", "ubuntu-server", "24.04", "x86_64")
	if err == nil {
		t.Fatal("PlanCatalogueFlash() error = nil, want cache error")
	}
	if !strings.Contains(err.Error(), "is not cached") {
		t.Fatalf("PlanCatalogueFlash() error = %q, want cache message", err.Error())
	}
}
