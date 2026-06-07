package profile

import (
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestLoadFileExamples(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob("../../examples/profiles/*.yaml")
	if err != nil {
		t.Fatalf("Glob() error = %v", err)
	}

	var names []string
	for _, file := range files {
		value, err := LoadAndValidateFile(file)
		if err != nil {
			t.Fatalf("LoadAndValidateFile(%q) error = %v", file, err)
		}
		names = append(names, value.Name)
	}
	sort.Strings(names)

	want := []string{
		"alpine-minimal",
		"arch-minimal",
		"debian-server",
		"fedora-server",
		"opensuse-server",
		"rocky-server",
		"ubuntu-server",
		"windows-workstation",
	}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("example profile names = %#v, want %#v", names, want)
	}
}

func TestParseRejectsUnknownField(t *testing.T) {
	t.Parallel()

	_, err := Parse([]byte("name: example\nunknown: true\n"))

	if err == nil {
		t.Fatal("Parse() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "field unknown not found") {
		t.Fatalf("Parse() error = %q, want unknown-field message", err)
	}
}

func TestParseRejectsMultipleDocuments(t *testing.T) {
	t.Parallel()

	_, err := Parse([]byte("name: first\n---\nname: second\n"))

	if err == nil {
		t.Fatal("Parse() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "multiple documents") {
		t.Fatalf("Parse() error = %q, want multiple-documents message", err)
	}
}
