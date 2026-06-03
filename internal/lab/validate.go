package lab

import (
	"fmt"
	"os/exec"
	"strings"
)

// ValidationCheck describes a single SSH-based validation step.
type ValidationCheck struct {
	Name           string
	Command        string
	ExpectExitCode int
	ExpectOutput   string // substring match; empty means no output check
}

// ValidationResult holds the outcome of a single validation check.
type ValidationResult struct {
	Check  ValidationCheck
	Passed bool
	Output string
	Error  string
}

// RunValidation runs each check against the VM via SSH and returns results.
// It continues after individual check failures so callers receive a full
// picture of what passed and what failed.
func RunValidation(run *Run, user, privateKeyPath string, checks []ValidationCheck) ([]ValidationResult, error) {
	results := make([]ValidationResult, 0, len(checks))

	for _, check := range checks {
		result := runCheck(run, user, privateKeyPath, check)
		results = append(results, result)
	}
	return results, nil
}

func runCheck(run *Run, user, privateKeyPath string, check ValidationCheck) ValidationResult {
	args := []string{
		"-p", fmt.Sprintf("%d", run.SSHPort),
		"-o", "StrictHostKeyChecking=no",
		"-i", privateKeyPath,
		fmt.Sprintf("%s@localhost", user),
		check.Command,
	}

	cmd := exec.Command("ssh", args...)
	out, err := cmd.Output()
	output := strings.TrimRight(string(out), "\n")

	// Determine the exit code.
	exitCode := 0
	if err != nil {
		var exitErr *exec.ExitError
		if asExitError(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			// Non-exit error (connection failure, etc.).
			return ValidationResult{
				Check:  check,
				Passed: false,
				Output: output,
				Error:  err.Error(),
			}
		}
	}

	passed := exitCode == check.ExpectExitCode
	if passed && check.ExpectOutput != "" {
		passed = strings.Contains(output, check.ExpectOutput)
	}

	errMsg := ""
	if !passed {
		parts := []string{fmt.Sprintf("exit code %d (expected %d)", exitCode, check.ExpectExitCode)}
		if check.ExpectOutput != "" && !strings.Contains(output, check.ExpectOutput) {
			parts = append(parts, fmt.Sprintf("output does not contain %q", check.ExpectOutput))
		}
		errMsg = strings.Join(parts, "; ")
	}

	return ValidationResult{
		Check:  check,
		Passed: passed,
		Output: output,
		Error:  errMsg,
	}
}

// asExitError is a helper to avoid importing errors in the hot path.
func asExitError(err error, target **exec.ExitError) bool {
	ee, ok := err.(*exec.ExitError)
	if ok {
		*target = ee
	}
	return ok
}
