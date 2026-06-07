package lab

import (
	"fmt"
	"os/exec"
	"strings"
)

// CreateSnapshot creates a named internal snapshot of the run's disk image.
// The VM does not need to be stopped for snapshot creation, though it is
// recommended.
func CreateSnapshot(run *Run, name string) error {
	out, err := exec.Command(qemuImgPath(), "snapshot", "-c", name, run.DiskPath).CombinedOutput()
	if err != nil {
		return fmt.Errorf("create snapshot %q: %w: %s", name, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// ListSnapshots returns the names of all snapshots stored in the disk image.
func ListSnapshots(run *Run) ([]string, error) {
	out, err := exec.Command(qemuImgPath(), "snapshot", "-l", run.DiskPath).Output()
	if err != nil {
		return nil, fmt.Errorf("list snapshots: %w", err)
	}
	return parseSnapshotList(string(out)), nil
}

// RevertSnapshot applies a named snapshot, restoring the disk to that state.
// The VM must be stopped before calling this function.
func RevertSnapshot(run *Run, name string) error {
	out, err := exec.Command(qemuImgPath(), "snapshot", "-a", name, run.DiskPath).CombinedOutput()
	if err != nil {
		return fmt.Errorf("revert snapshot %q: %w: %s", name, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// DeleteSnapshot removes a named snapshot from the disk image.
func DeleteSnapshot(run *Run, name string) error {
	out, err := exec.Command(qemuImgPath(), "snapshot", "-d", name, run.DiskPath).CombinedOutput()
	if err != nil {
		return fmt.Errorf("delete snapshot %q: %w: %s", name, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// parseSnapshotList parses the output of `qemu-img snapshot -l` and returns
// the list of snapshot names (second column).
//
// Example output:
//
//	Snapshot list:
//	ID        TAG                     VM SIZE                DATE       VM CLOCK
//	1         snap1                   0 B 2024-01-01 00:00:00   00:00:00.000
func parseSnapshotList(output string) []string {
	var names []string
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Snapshot list") || strings.HasPrefix(line, "ID") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			names = append(names, fields[1])
		}
	}
	return names
}
