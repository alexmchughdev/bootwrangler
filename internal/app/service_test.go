package app

import (
	"path/filepath"
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
