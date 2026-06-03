package lab

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"
)

// SSHReady reports whether the SSH daemon is accepting connections on the
// run's forwarded port. It attempts a single TCP dial with a short timeout.
func SSHReady(run *Run) bool {
	addr := fmt.Sprintf("localhost:%d", run.SSHPort)
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// SSHCommand returns the ssh command string for connecting to a lab VM.
func SSHCommand(run *Run, user string) string {
	return fmt.Sprintf("ssh -p %d -o StrictHostKeyChecking=no %s@localhost",
		run.SSHPort, user)
}

// RunSSHCommand executes a single command over SSH using the system ssh binary
// and returns its combined stdout. The private key at privateKeyPath is used
// for authentication.
func RunSSHCommand(run *Run, user, command, privateKeyPath string) (string, error) {
	args := []string{
		"-p", fmt.Sprintf("%d", run.SSHPort),
		"-o", "StrictHostKeyChecking=no",
		"-i", privateKeyPath,
		fmt.Sprintf("%s@localhost", user),
		command,
	}
	out, err := exec.Command("ssh", args...).Output()
	if err != nil {
		return "", fmt.Errorf("ssh command %q: %w", command, err)
	}
	return strings.TrimRight(string(out), "\n"), nil
}
