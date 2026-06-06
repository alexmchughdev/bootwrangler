package media

import (
	"fmt"
	"strings"
)

const (
	// ContentStatusPlanned means the content is structurally valid but has not
	// been resolved against local cache or library state.
	ContentStatusPlanned = "planned"
	ContentStatusReady   = "ready"
	ContentStatusWarning = "warning"
	ContentStatusBlocked = "blocked"
)

// ContentResolution describes whether partition content can be assembled from
// local inputs before any destructive media build path runs.
type ContentResolution struct {
	Status     string
	Message    string
	SourcePath string
	Warnings   []string
	Errors     []string
}

// PartitionAction is one planned step in the build.
type PartitionAction struct {
	Label             string
	SizeBytes         int64
	Filesystem        string
	Content           PartitionContent
	ContentResolution ContentResolution
	DevicePath        string // filled in at build time
}

// BuildPlan is the validated, ordered sequence of operations.
type BuildPlan struct {
	RecipeName string
	DevicePath string
	DeviceSize int64
	TotalBytes int64
	Actions    []PartitionAction
	DryRun     bool
	Ready      bool
	Warnings   []string
	Errors     []string
}

// PlanBuild validates the recipe against the device and returns a BuildPlan.
func PlanBuild(recipe Recipe, devicePath string, deviceSizeBytes int64, dryRun bool) (BuildPlan, error) {
	if err := Validate(recipe); err != nil {
		return BuildPlan{}, err
	}

	// Calculate total fixed size (all non-"remaining" partitions).
	var totalFixed int64
	hasRemaining := false
	for _, p := range recipe.Partitions {
		sz, err := ParseSize(p.Size)
		if err != nil {
			return BuildPlan{}, fmt.Errorf("plan build: partition %q: %w", p.Label, err)
		}
		if sz == -1 {
			hasRemaining = true
		} else {
			totalFixed += sz
		}
	}

	if totalFixed > deviceSizeBytes {
		return BuildPlan{}, fmt.Errorf("recipe requires %d but device is %d", totalFixed, deviceSizeBytes)
	}

	remainingSize := deviceSizeBytes - totalFixed
	if hasRemaining && remainingSize <= 0 {
		return BuildPlan{}, fmt.Errorf("no space for remaining partition")
	}

	actions := make([]PartitionAction, 0, len(recipe.Partitions))
	var totalBytes int64
	for _, p := range recipe.Partitions {
		sz, _ := ParseSize(p.Size)
		if sz == -1 {
			sz = remainingSize
		}
		totalBytes += sz
		actions = append(actions, PartitionAction{
			Label:      p.Label,
			SizeBytes:  sz,
			Filesystem: p.Filesystem,
			Content:    p.Content,
			ContentResolution: ContentResolution{
				Status:  ContentStatusPlanned,
				Message: "Content source resolution is pending.",
			},
		})
	}

	return BuildPlan{
		RecipeName: recipe.Name,
		DevicePath: devicePath,
		DeviceSize: deviceSizeBytes,
		TotalBytes: totalBytes,
		Actions:    actions,
		DryRun:     dryRun,
		Ready:      true,
	}, nil
}

// FormatBuildPlan returns a human-readable summary of the build plan.
func FormatBuildPlan(plan BuildPlan) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Build plan: %s\n", plan.RecipeName)
	fmt.Fprintf(&sb, "  device:      %s (%s)\n", plan.DevicePath, formatBytes(plan.DeviceSize))
	fmt.Fprintf(&sb, "  total used:  %s\n", formatBytes(plan.TotalBytes))
	if plan.DryRun {
		fmt.Fprintf(&sb, "  mode:        dry-run\n")
	}
	if !plan.Ready {
		fmt.Fprintf(&sb, "  content:     blocked\n")
	} else if len(plan.Warnings) > 0 {
		fmt.Fprintf(&sb, "  content:     ready with warnings\n")
	}
	fmt.Fprintf(&sb, "  partitions:\n")
	for i, a := range plan.Actions {
		status := a.ContentResolution.Status
		if status == "" {
			status = ContentStatusPlanned
		}
		fmt.Fprintf(&sb, "    [%d] %-16s  %-10s  %-8s  %-8s  %s\n",
			i+1, a.Label, formatBytes(a.SizeBytes), a.Filesystem, status, string(a.Content.Type))
		if a.ContentResolution.Message != "" {
			fmt.Fprintf(&sb, "        %s\n", a.ContentResolution.Message)
		}
		if a.ContentResolution.SourcePath != "" {
			fmt.Fprintf(&sb, "        source: %s\n", a.ContentResolution.SourcePath)
		}
		for _, warning := range a.ContentResolution.Warnings {
			fmt.Fprintf(&sb, "        warning: %s\n", warning)
		}
		for _, err := range a.ContentResolution.Errors {
			fmt.Fprintf(&sb, "        error: %s\n", err)
		}
	}
	for _, warning := range plan.Warnings {
		fmt.Fprintf(&sb, "  warning: %s\n", warning)
	}
	for _, err := range plan.Errors {
		fmt.Fprintf(&sb, "  error: %s\n", err)
	}
	return sb.String()
}

// formatBytes formats a byte count with appropriate unit suffix.
func formatBytes(n int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
		TB = 1024 * GB
	)
	switch {
	case n >= TB:
		return fmt.Sprintf("%.1f TB", float64(n)/TB)
	case n >= GB:
		return fmt.Sprintf("%.1f GB", float64(n)/GB)
	case n >= MB:
		return fmt.Sprintf("%.0f MB", float64(n)/MB)
	case n >= KB:
		return fmt.Sprintf("%.0f KB", float64(n)/KB)
	default:
		return fmt.Sprintf("%d B", n)
	}
}
