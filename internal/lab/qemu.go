package lab

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// PlanQEMU returns the argument list for qemu-system-x86_64 without executing.
// It does not require QEMU to be installed.
func PlanQEMU(run *Run) ([]string, error) {
	vncDisplay := run.VNCPort - 5900
	args := []string{
		"-name", fmt.Sprintf("bootwrangler-lab-%s", run.ID),
		"-m", strconv.Itoa(run.MemoryMB()),
		"-smp", strconv.Itoa(run.CPUs()),
		"-drive", fmt.Sprintf("file=%s,format=qcow2", run.DiskPath),
		"-serial", fmt.Sprintf("file:%s", run.SerialLog),
		"-net", "nic",
		"-net", fmt.Sprintf("user,hostfwd=tcp::%d-:22", run.SSHPort),
		"-vnc", fmt.Sprintf(":%d", vncDisplay),
		"-nographic",
		"-no-reboot",
	}
	return args, nil
}

// Start creates a disk image with qemu-img and starts qemu-system-x86_64 as
// a background process. It writes the PID to WorkDir/qemu.pid.
func Start(run *Run) error {
	tools := Locate()
	if !tools.Available() {
		return fmt.Errorf("QEMU is not set up: open the Lab and choose Set Up Lab to install it automatically")
	}

	// Create qcow2 disk image.
	imgArgs := []string{
		"create", "-f", "qcow2",
		run.DiskPath,
		fmt.Sprintf("%dG", run.DiskSizeGB()),
	}
	imgCmd := exec.Command(tools.ImgBinary, imgArgs...)
	if out, err := imgCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("qemu-img create: %w: %s", err, strings.TrimSpace(string(out)))
	}

	// Build QEMU args.
	args, err := PlanQEMU(run)
	if err != nil {
		return fmt.Errorf("plan QEMU: %w", err)
	}

	// Launch QEMU as a background process.
	cmd := exec.Command(tools.SystemBinary, args...)
	setProcessGroup(cmd)
	if err := cmd.Start(); err != nil {
		run.State = RunStateFailed
		run.Error = err.Error()
		return fmt.Errorf("start qemu: %w", err)
	}

	// Persist PID.
	pidPath := pidFile(run)
	pidData := strconv.Itoa(cmd.Process.Pid)
	if err := os.WriteFile(pidPath, []byte(pidData), 0o644); err != nil {
		// Best-effort: kill the process so we don't leak it.
		_ = cmd.Process.Kill()
		run.State = RunStateFailed
		run.Error = err.Error()
		return fmt.Errorf("write pid file: %w", err)
	}

	now := timeNow()
	run.State = RunStateRunning
	run.StartedAt = &now

	return nil
}

// Stop sends SIGTERM to the QEMU process and updates the run state.
func Stop(run *Run) error {
	pid, err := readPID(run)
	if err != nil {
		return fmt.Errorf("read pid: %w", err)
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find process %d: %w", pid, err)
	}

	if err := terminateProcess(proc); err != nil {
		return fmt.Errorf("signal qemu: %w", err)
	}

	now := timeNow()
	run.State = RunStateStopped
	run.StoppedAt = &now
	return nil
}

// IsRunning reports whether the QEMU process for run is still alive.
func IsRunning(run *Run) bool {
	pid, err := readPID(run)
	if err != nil {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return processAlive(proc)
}

// pidFile returns the path of the PID file for a run.
func pidFile(run *Run) string {
	return fmt.Sprintf("%s/qemu.pid", run.WorkDir)
}

// readPID reads and parses the PID file for a run.
func readPID(run *Run) (int, error) {
	data, err := os.ReadFile(pidFile(run))
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, fmt.Errorf("invalid pid %q: %w", string(data), err)
	}
	return pid, nil
}
