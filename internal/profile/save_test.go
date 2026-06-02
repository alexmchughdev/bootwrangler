package profile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEncode(t *testing.T) {
	t.Parallel()

	data, err := Encode(validProfile())
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	value, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if value.Name != "test-profile" {
		t.Fatalf("Parse().Name = %q, want %q", value.Name, "test-profile")
	}
}

func TestSaveFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "profile.yaml")

	if err := SaveFile(path, validProfile()); err != nil {
		t.Fatalf("SaveFile() error = %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("saved profile mode = %o, want 600", got)
	}

	value, err := LoadAndValidateFile(path)
	if err != nil {
		t.Fatalf("LoadAndValidateFile() error = %v", err)
	}
	if value.Name != "test-profile" {
		t.Fatalf("LoadAndValidateFile().Name = %q, want %q", value.Name, "test-profile")
	}
}

func TestSaveFileRejectsSymlink(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	target := filepath.Join(directory, "target.yaml")
	link := filepath.Join(directory, "link.yaml")
	if err := os.WriteFile(target, []byte("existing"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("Symlink() error = %v", err)
	}

	err := SaveFile(link, validProfile())

	if err == nil {
		t.Fatal("SaveFile() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "refuse to overwrite profile symlink") {
		t.Fatalf("SaveFile() error = %q, want symlink message", err)
	}
}
