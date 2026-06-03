package lab

import (
	"os"
	"strings"
)

// SerialLogPath returns the path to the serial console log file for a run.
func SerialLogPath(run *Run) string {
	return run.SerialLog
}

// ReadSerialLog reads the entire serial console log for a run.
func ReadSerialLog(run *Run) (string, error) {
	data, err := os.ReadFile(SerialLogPath(run))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

// TailSerialLog returns the last n lines of the serial console log.
// If n <= 0 all lines are returned.
func TailSerialLog(run *Run, lines int) (string, error) {
	content, err := ReadSerialLog(run)
	if err != nil {
		return "", err
	}
	if content == "" || lines <= 0 {
		return content, nil
	}

	all := strings.Split(content, "\n")
	// Remove a trailing empty element caused by a final newline.
	if len(all) > 0 && all[len(all)-1] == "" {
		all = all[:len(all)-1]
	}

	if lines >= len(all) {
		return strings.Join(all, "\n") + "\n", nil
	}
	tail := all[len(all)-lines:]
	return strings.Join(tail, "\n") + "\n", nil
}
