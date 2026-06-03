package app

import (
	"fmt"
	"sync"

	"github.com/alexmchughdev/bootwrangler/internal/lab"
)

// labMu protects labRuns independently from the main service mutex so lab
// operations do not block unrelated service calls.
var labMu sync.Mutex
var labRuns = map[string]*lab.Run{}

// LabStart creates and starts a lab VM from a profile name.
func (s *Service) LabStart(profileName string, memorymb, cpus int) (lab.Run, error) {
	opts := lab.StartOptions{
		ProfileName: profileName,
		MemoryMB:    memorymb,
		CPUs:        cpus,
	}
	run, err := lab.NewRun(opts)
	if err != nil {
		return lab.Run{}, fmt.Errorf("create lab run: %w", err)
	}

	if err := lab.Start(run); err != nil {
		return lab.Run{}, fmt.Errorf("start lab VM: %w", err)
	}

	labMu.Lock()
	labRuns[run.ID] = run
	labMu.Unlock()

	return *run, nil
}

// LabStop stops a running lab VM.
func (s *Service) LabStop(runID string) error {
	labMu.Lock()
	run, ok := labRuns[runID]
	labMu.Unlock()
	if !ok {
		return fmt.Errorf("lab run not found: %s", runID)
	}
	return lab.Stop(run)
}

// LabStatus returns the run state of a lab VM.
func (s *Service) LabStatus(runID string) (lab.Run, error) {
	labMu.Lock()
	run, ok := labRuns[runID]
	labMu.Unlock()
	if !ok {
		return lab.Run{}, fmt.Errorf("lab run not found: %s", runID)
	}
	return *run, nil
}

// LabSerialLog returns the full serial console log.
func (s *Service) LabSerialLog(runID string) (string, error) {
	labMu.Lock()
	run, ok := labRuns[runID]
	labMu.Unlock()
	if !ok {
		return "", fmt.Errorf("lab run not found: %s", runID)
	}
	return lab.ReadSerialLog(run)
}

// LabSSHCommand returns the SSH command string for a lab VM.
func (s *Service) LabSSHCommand(runID, user string) (string, error) {
	labMu.Lock()
	run, ok := labRuns[runID]
	labMu.Unlock()
	if !ok {
		return "", fmt.Errorf("lab run not found: %s", runID)
	}
	return lab.SSHCommand(run, user), nil
}

// LabPlanQEMU returns the planned QEMU command for inspection.
func (s *Service) LabPlanQEMU(profileName string, memorymb, cpus int) ([]string, error) {
	opts := lab.StartOptions{
		ProfileName: profileName,
		MemoryMB:    memorymb,
		CPUs:        cpus,
		// Use fixed ports so PlanQEMU is deterministic; actual ports are
		// assigned by NewRun.
		SSHPort: 12200,
		VNCPort: 15900,
	}
	run, err := lab.NewRun(opts)
	if err != nil {
		return nil, fmt.Errorf("create run for plan: %w", err)
	}
	return lab.PlanQEMU(run)
}

// LabCreateSnapshot creates a snapshot of a stopped lab VM.
func (s *Service) LabCreateSnapshot(runID, name string) error {
	labMu.Lock()
	run, ok := labRuns[runID]
	labMu.Unlock()
	if !ok {
		return fmt.Errorf("lab run not found: %s", runID)
	}
	return lab.CreateSnapshot(run, name)
}

// LabListSnapshots lists snapshots for a lab VM.
func (s *Service) LabListSnapshots(runID string) ([]string, error) {
	labMu.Lock()
	run, ok := labRuns[runID]
	labMu.Unlock()
	if !ok {
		return nil, fmt.Errorf("lab run not found: %s", runID)
	}
	return lab.ListSnapshots(run)
}

// LabListRuns returns all active lab VM runs tracked in this session.
func (s *Service) LabListRuns() []lab.Run {
	labMu.Lock()
	defer labMu.Unlock()
	runs := make([]lab.Run, 0, len(labRuns))
	for _, r := range labRuns {
		runs = append(runs, *r)
	}
	return runs
}
