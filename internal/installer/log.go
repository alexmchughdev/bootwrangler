// Package installer provides types for tracking installation progress
// and recording results from provisioned hosts.
package installer

import (
	"time"
)

// Phase represents a named installation phase.
type Phase string

const (
	PhaseStarted    Phase = "started"
	PhasePartition  Phase = "partition"
	PhaseInstall    Phase = "install"
	PhaseNetwork    Phase = "network"
	PhaseBootloader Phase = "bootloader"
	PhaseComplete   Phase = "complete"
	PhaseFailed     Phase = "failed"
)

// Event is a single timestamped log event from an installer.
type Event struct {
	At      time.Time
	Phase   Phase
	Message string
	Err     string // non-empty only on failure events
}

// Result is the final outcome of an installation run.
type Result struct {
	ProfileName string
	StartedAt   time.Time
	CompletedAt *time.Time
	Success     bool
	Events      []Event
	Error       string
}

// Log accumulates events for a single installation.
type Log struct {
	result Result
}

// NewLog starts a new installation log for the named profile.
func NewLog(profileName string) *Log {
	return &Log{
		result: Result{
			ProfileName: profileName,
			StartedAt:   time.Now(),
		},
	}
}

// Add appends an event to the log.
func (l *Log) Add(phase Phase, message string) {
	l.result.Events = append(l.result.Events, Event{
		At:      time.Now(),
		Phase:   phase,
		Message: message,
	})
}

// Fail records a failure event and marks the result as failed.
func (l *Log) Fail(phase Phase, errMsg string) {
	now := time.Now()
	l.result.Events = append(l.result.Events, Event{
		At:    now,
		Phase: PhaseFailed,
		Err:   errMsg,
	})
	l.result.Success = false
	l.result.CompletedAt = &now
	l.result.Error = errMsg
}

// Complete marks the installation as successful.
func (l *Log) Complete() {
	now := time.Now()
	l.result.Events = append(l.result.Events, Event{
		At:      now,
		Phase:   PhaseComplete,
		Message: "installation complete",
	})
	l.result.Success = true
	l.result.CompletedAt = &now
}

// Result returns the current log state.
func (l *Log) Result() Result {
	return l.result
}
