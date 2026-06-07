//go:build windows

package lab

import (
	"os"
	"os/exec"
)

// setProcessGroup is a no-op on Windows.
func setProcessGroup(_ *exec.Cmd) {}

// terminateProcess kills the process on Windows (no SIGTERM).
func terminateProcess(proc *os.Process) error {
	return proc.Kill()
}

// processAlive checks if the process is still running on Windows.
func processAlive(proc *os.Process) bool {
	// On Windows, FindProcess always succeeds; use a no-op signal check.
	state, err := proc.Wait()
	if err != nil {
		return true // still running if Wait errors without exiting
	}
	return !state.Exited()
}
