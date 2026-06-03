package media

import "fmt"

// ExecuteBuild executes the given BuildPlan.
//
// For dry-run plans it returns immediately without making any changes.
// For live runs it logs each planned action and returns an error because
// real disk partitioning requires root privileges and external tools
// (parted, mkfs.*, etc.) that are not available in all environments.
// The planning and dry-run paths are fully implemented; live execution
// is intentionally guarded behind this error.
func ExecuteBuild(plan BuildPlan) error {
	if plan.DryRun {
		return nil
	}

	for _, a := range plan.Actions {
		fmt.Printf("  partition: label=%s size=%d fs=%s content=%s\n",
			a.Label, a.SizeBytes, a.Filesystem, string(a.Content.Type))
	}

	return fmt.Errorf("media build: partitioning not available in this environment")
}
