package usb

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPlanContained_Valid(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	// Create a few source files.
	files := []string{"user-data", "meta-data", "vendor-data"}
	for _, name := range files {
		if err := os.WriteFile(filepath.Join(src, name), []byte("content of "+name), 0o644); err != nil {
			t.Fatalf("setup: write %s: %v", name, err)
		}
	}
	// Also create a subdirectory that should be skipped.
	if err := os.Mkdir(filepath.Join(src, "subdir"), 0o755); err != nil {
		t.Fatalf("setup: mkdir: %v", err)
	}

	plan, err := PlanContained(src, dst, false)
	if err != nil {
		t.Fatalf("PlanContained() error = %v", err)
	}
	if len(plan.Files) != len(files) {
		t.Errorf("len(plan.Files) = %d, want %d", len(plan.Files), len(files))
	}
}

func TestPlanContained_MissingSource(t *testing.T) {
	dst := t.TempDir()
	_, err := PlanContained("/nonexistent/source/dir", dst, false)
	if err == nil {
		t.Fatal("PlanContained() expected error for missing source, got nil")
	}
}

func TestPlanContained_MissingTarget(t *testing.T) {
	src := t.TempDir()
	_, err := PlanContained(src, "/nonexistent/target/mount", false)
	if err == nil {
		t.Fatal("PlanContained() expected error for missing target, got nil")
	}
}

func TestExecuteContained_DryRun(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	names := []string{"user-data", "meta-data"}
	for _, name := range names {
		if err := os.WriteFile(filepath.Join(src, name), []byte("data"), 0o644); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}

	plan, err := PlanContained(src, dst, true)
	if err != nil {
		t.Fatalf("PlanContained() error = %v", err)
	}

	if err := ExecuteContained(plan); err != nil {
		t.Fatalf("ExecuteContained() dry-run error = %v", err)
	}

	// Confirm no files were written to target.
	entries, err := os.ReadDir(dst)
	if err != nil {
		t.Fatalf("ReadDir target: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("dry run wrote %d files to target, want 0", len(entries))
	}
}
