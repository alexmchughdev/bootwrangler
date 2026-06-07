package lab

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestResolvePrefersBundled(t *testing.T) {
	// Point the bundled-dir override at a temp dir containing fake binaries.
	dir := t.TempDir()
	sys := filepath.Join(dir, systemBinName())
	img := filepath.Join(dir, imgBinName())
	for _, p := range []string{sys, img} {
		if err := os.WriteFile(p, []byte("#!/bin/sh\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	t.Setenv("BOOTWRANGLER_QEMU_DIR", dir)

	gotSys, gotImg, src := resolve()
	if src != SourceBundled {
		t.Fatalf("source = %q, want bundled", src)
	}
	if gotSys != sys || gotImg != img {
		t.Errorf("resolved paths = %q,%q want %q,%q", gotSys, gotImg, sys, img)
	}
}

func TestResolveNoneWhenAbsent(t *testing.T) {
	// Override bundled dir to an empty location and neutralise PATH so the
	// system lookup and app-data lookup both miss.
	t.Setenv("BOOTWRANGLER_QEMU_DIR", t.TempDir())
	t.Setenv("PATH", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", t.TempDir())
	}

	if _, _, src := resolve(); src != SourceNone {
		t.Fatalf("source = %q, want none", src)
	}
}

func TestAvailabilityReportsMethod(t *testing.T) {
	a := CheckAvailability()
	if a.Method == "" {
		t.Error("expected a non-empty install method for the current platform")
	}
	if a.Hint == "" {
		t.Error("expected a non-empty manual hint")
	}
}

func TestToolsAvailable(t *testing.T) {
	if (Tools{Source: SourceNone}).Available() {
		t.Error("none source should not be available")
	}
	if !(Tools{Source: SourceSystem, SystemBinary: "a", ImgBinary: "b"}).Available() {
		t.Error("system source with both binaries should be available")
	}
}
