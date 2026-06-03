package usb

import (
	"fmt"
	"os"
	"path/filepath"
)

// ContainedPlan describes the steps to write assets to a USB partition.
type ContainedPlan struct {
	SourceDir   string // rendered profile output dir
	TargetMount string // mounted target partition path
	Files       []ContainedFile
	DryRun      bool
}

// ContainedFile describes one file copy operation.
type ContainedFile struct {
	Source string
	Dest   string
}

// PlanContained builds a plan to copy rendered assets onto a mounted USB partition.
// sourceDir is the rendered profile directory (contains user-data, meta-data, etc.).
// targetMount is the mount point of the target USB partition.
func PlanContained(sourceDir, targetMount string, dryRun bool) (ContainedPlan, error) {
	// sourceDir must exist
	if _, err := os.Stat(sourceDir); err != nil {
		return ContainedPlan{}, fmt.Errorf("source dir not found: %w", err)
	}
	// targetMount must exist (caller is responsible for mounting)
	if _, err := os.Stat(targetMount); err != nil {
		return ContainedPlan{}, fmt.Errorf("target mount not found: %w", err)
	}

	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return ContainedPlan{}, fmt.Errorf("read source dir: %w", err)
	}

	plan := ContainedPlan{
		SourceDir:   sourceDir,
		TargetMount: targetMount,
		DryRun:      dryRun,
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		plan.Files = append(plan.Files, ContainedFile{
			Source: filepath.Join(sourceDir, e.Name()),
			Dest:   filepath.Join(targetMount, e.Name()),
		})
	}
	return plan, nil
}

// ExecuteContained copies files according to the plan.
// If DryRun is true, it validates sources exist and returns without copying.
func ExecuteContained(plan ContainedPlan) error {
	for _, f := range plan.Files {
		if _, err := os.Stat(f.Source); err != nil {
			return fmt.Errorf("source missing %s: %w", f.Source, err)
		}
		if plan.DryRun {
			continue
		}
		data, err := os.ReadFile(f.Source)
		if err != nil {
			return fmt.Errorf("read %s: %w", f.Source, err)
		}
		if err := os.WriteFile(f.Dest, data, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", f.Dest, err)
		}
	}
	return nil
}
