package lab

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// RunState represents the lifecycle state of a lab VM run.
type RunState string

const (
	RunStatePending RunState = "pending"
	RunStateRunning RunState = "running"
	RunStateStopped RunState = "stopped"
	RunStateFailed  RunState = "failed"
)

// Run holds the runtime state for a single lab VM instance.
type Run struct {
	ID          string
	ProfileName string
	State       RunState
	WorkDir     string // ~/.bootwrangler/lab/runs/<id>/
	DiskPath    string // WorkDir/disk.qcow2
	SerialLog   string // WorkDir/serial.log
	SSHPort     int
	VNCPort     int
	CreatedAt   time.Time
	StartedAt   *time.Time
	StoppedAt   *time.Time
	Error       string

	// creation-time options stored for QEMU planning
	memoryMB   int
	cpus       int
	diskSizeGB int
}

// StartOptions configures a new lab VM run.
type StartOptions struct {
	ProfileName string
	DiskSizeGB  int // default 20
	MemoryMB    int // default 2048
	CPUs        int // default 2
	SSHPort     int // 0 = auto-assign in range 12200-12299
	VNCPort     int // 0 = auto-assign in range 15900-15999
}

// DefaultLabDir returns the root lab directory under the user home.
func DefaultLabDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".bootwrangler", "lab")
	}
	return filepath.Join(home, ".bootwrangler", "lab")
}

// WorkDir returns the working directory for a run with the given ID.
func WorkDir(id string) string {
	return filepath.Join(DefaultLabDir(), "runs", id)
}

// NewRun creates the workspace directory and a Run ready to be started.
// It assigns SSH and VNC ports (auto or from opts) and does NOT start QEMU.
func NewRun(opts StartOptions) (*Run, error) {
	// Apply defaults.
	if opts.DiskSizeGB <= 0 {
		opts.DiskSizeGB = 20
	}
	if opts.MemoryMB <= 0 {
		opts.MemoryMB = 2048
	}
	if opts.CPUs <= 0 {
		opts.CPUs = 2
	}

	id := uuid.New().String()
	wd := WorkDir(id)
	if err := os.MkdirAll(wd, 0o755); err != nil {
		return nil, fmt.Errorf("create run workspace: %w", err)
	}

	sshPort := opts.SSHPort
	if sshPort == 0 {
		p, err := assignPort(12200, 12299)
		if err != nil {
			return nil, fmt.Errorf("assign SSH port: %w", err)
		}
		sshPort = p
	}

	vncPort := opts.VNCPort
	if vncPort == 0 {
		p, err := assignPort(15900, 15999)
		if err != nil {
			return nil, fmt.Errorf("assign VNC port: %w", err)
		}
		vncPort = p
	}

	run := &Run{
		ID:          id,
		ProfileName: opts.ProfileName,
		State:       RunStatePending,
		WorkDir:     wd,
		DiskPath:    filepath.Join(wd, "disk.qcow2"),
		SerialLog:   filepath.Join(wd, "serial.log"),
		SSHPort:     sshPort,
		VNCPort:     vncPort,
		CreatedAt:   time.Now(),
		memoryMB:    opts.MemoryMB,
		cpus:        opts.CPUs,
		diskSizeGB:  opts.DiskSizeGB,
	}
	return run, nil
}

// assignPort finds a free TCP port in [low, high] by trying random candidates.
func assignPort(low, high int) (int, error) {
	span := high - low + 1
	perm := rand.Perm(span)
	for _, offset := range perm {
		port := low + offset
		addr := fmt.Sprintf("127.0.0.1:%d", port)
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			continue
		}
		ln.Close()
		return port, nil
	}
	return 0, fmt.Errorf("no free port found in range %d-%d", low, high)
}
