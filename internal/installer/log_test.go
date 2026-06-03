package installer

import (
	"testing"
)

func TestNewLog(t *testing.T) {
	l := NewLog("test-profile")
	if l == nil {
		t.Fatal("NewLog() returned nil")
	}
	r := l.Result()
	if r.ProfileName != "test-profile" {
		t.Errorf("ProfileName = %q, want %q", r.ProfileName, "test-profile")
	}
	if r.StartedAt.IsZero() {
		t.Error("StartedAt is zero")
	}
	if r.CompletedAt != nil {
		t.Error("CompletedAt should be nil for a new log")
	}
	if r.Success {
		t.Error("Success should be false for a new log")
	}
	if len(r.Events) != 0 {
		t.Errorf("Events len = %d, want 0", len(r.Events))
	}
}

func TestLog_Add(t *testing.T) {
	l := NewLog("profile")
	l.Add(PhaseInstall, "installing packages")
	l.Add(PhaseNetwork, "configuring network")

	r := l.Result()
	if len(r.Events) != 2 {
		t.Fatalf("len(Events) = %d, want 2", len(r.Events))
	}
	if r.Events[0].Phase != PhaseInstall {
		t.Errorf("Events[0].Phase = %q, want %q", r.Events[0].Phase, PhaseInstall)
	}
	if r.Events[0].Message != "installing packages" {
		t.Errorf("Events[0].Message = %q, want %q", r.Events[0].Message, "installing packages")
	}
	if r.Events[1].Phase != PhaseNetwork {
		t.Errorf("Events[1].Phase = %q, want %q", r.Events[1].Phase, PhaseNetwork)
	}
	if r.Events[0].At.IsZero() {
		t.Error("Events[0].At is zero")
	}
}

func TestLog_Fail(t *testing.T) {
	l := NewLog("profile")
	l.Add(PhasePartition, "partitioning disk")
	l.Fail(PhasePartition, "disk not found")

	r := l.Result()
	if r.Success {
		t.Error("Success should be false after Fail()")
	}
	if r.Error != "disk not found" {
		t.Errorf("Error = %q, want %q", r.Error, "disk not found")
	}
	if r.CompletedAt == nil {
		t.Error("CompletedAt should be set after Fail()")
	}

	// Last event should be the failure event.
	last := r.Events[len(r.Events)-1]
	if last.Phase != PhaseFailed {
		t.Errorf("last event Phase = %q, want %q", last.Phase, PhaseFailed)
	}
	if last.Err != "disk not found" {
		t.Errorf("last event Err = %q, want %q", last.Err, "disk not found")
	}
}

func TestLog_Complete(t *testing.T) {
	l := NewLog("profile")
	l.Add(PhaseInstall, "installing")
	l.Complete()

	r := l.Result()
	if !r.Success {
		t.Error("Success should be true after Complete()")
	}
	if r.CompletedAt == nil {
		t.Error("CompletedAt should be set after Complete()")
	}
	if r.Error != "" {
		t.Errorf("Error should be empty after Complete(), got %q", r.Error)
	}

	// Last event should be the complete event.
	last := r.Events[len(r.Events)-1]
	if last.Phase != PhaseComplete {
		t.Errorf("last event Phase = %q, want %q", last.Phase, PhaseComplete)
	}
	if last.Message != "installation complete" {
		t.Errorf("last event Message = %q, want %q", last.Message, "installation complete")
	}
}

func TestLog_Result_IsCopy(t *testing.T) {
	l := NewLog("profile")
	r1 := l.Result()
	l.Add(PhaseInstall, "step")
	r2 := l.Result()

	if len(r1.Events) == len(r2.Events) && len(r2.Events) > 0 {
		t.Error("Result() should return independent snapshots; r1 was mutated by subsequent Add()")
	}
}

func TestPhaseConstants(t *testing.T) {
	phases := []Phase{
		PhaseStarted, PhasePartition, PhaseInstall,
		PhaseNetwork, PhaseBootloader, PhaseComplete, PhaseFailed,
	}
	seen := make(map[Phase]bool)
	for _, p := range phases {
		if seen[p] {
			t.Errorf("duplicate phase value: %q", p)
		}
		seen[p] = true
		if p == "" {
			t.Error("empty phase constant")
		}
	}
}
