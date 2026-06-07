//go:build linux || darwin

package lab

import (
	"os"
	"os/exec"
	"syscall"
)

// setProcessGroup places cmd in its own process group so SIGKILL on the
// parent does not automatically propagate to QEMU child processes.
func setProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// terminateProcess sends SIGTERM to the process, ignoring "already finished".
func terminateProcess(proc *os.Process) error {
	err := proc.Signal(syscall.SIGTERM)
	if err != nil && err.Error() == "os: process already finished" {
		return nil
	}
	return err
}

// processAlive reports whether the process is still running using signal 0.
func processAlive(proc *os.Process) bool {
	return proc.Signal(syscall.Signal(0)) == nil
}
