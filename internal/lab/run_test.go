package lab

import (
	"strings"
	"testing"
)

func TestNewRun(t *testing.T) {
	t.Parallel()

	opts := StartOptions{
		ProfileName: "test-profile",
		LabDir:      t.TempDir(),
		DiskSizeGB:  10,
		MemoryMB:    1024,
		CPUs:        1,
		SSHPort:     12255,
		VNCPort:     15902,
	}
	run, err := NewRun(opts)
	if err != nil {
		t.Fatalf("NewRun() error = %v", err)
	}

	if run.ID == "" {
		t.Fatal("NewRun().ID is empty")
	}
	if run.State != RunStatePending {
		t.Fatalf("NewRun().State = %q, want %q", run.State, RunStatePending)
	}
	if run.WorkDir == "" {
		t.Fatal("NewRun().WorkDir is empty")
	}
	if !strings.Contains(run.WorkDir, run.ID) {
		t.Fatalf("NewRun().WorkDir %q does not contain ID %q", run.WorkDir, run.ID)
	}
	if run.SSHPort == 0 {
		t.Fatal("NewRun().SSHPort is 0, want an assigned port")
	}
	if run.VNCPort == 0 {
		t.Fatal("NewRun().VNCPort is 0, want an assigned port")
	}
}

func TestPlanQEMU(t *testing.T) {
	t.Parallel()

	opts := StartOptions{
		ProfileName: "myprofile",
		LabDir:      t.TempDir(),
		MemoryMB:    4096,
		CPUs:        4,
		SSHPort:     12250,
		VNCPort:     15901,
	}
	run, err := NewRun(opts)
	if err != nil {
		t.Fatalf("NewRun() error = %v", err)
	}

	args, err := PlanQEMU(run)
	if err != nil {
		t.Fatalf("PlanQEMU() error = %v", err)
	}
	if len(args) == 0 {
		t.Fatal("PlanQEMU() returned empty args")
	}

	joined := strings.Join(args, " ")

	checks := []struct {
		name string
		want string
	}{
		{"name flag", "-name"},
		{"bootwrangler-lab prefix", "bootwrangler-lab-"},
		{"memory flag", "-m"},
		{"4096 MB", "4096"},
		{"smp flag", "-smp"},
		{"4 CPUs", "4"},
		{"drive flag", "-drive"},
		{"qcow2 format", "format=qcow2"},
		{"serial flag", "-serial"},
		{"serial.log", "serial.log"},
		{"hostfwd SSH port", "hostfwd=tcp::12250-:22"},
		{"vnc flag", "-vnc"},
		{"vnc display 1", ":1"},
		{"nographic flag", "-nographic"},
		{"no-reboot flag", "-no-reboot"},
	}

	for _, c := range checks {
		if !strings.Contains(joined, c.want) {
			t.Errorf("PlanQEMU() args missing %s: want %q in %q", c.name, c.want, joined)
		}
	}
}

func TestSSHCommand(t *testing.T) {
	t.Parallel()

	run := &Run{
		SSHPort: 12255,
	}

	cmd := SSHCommand(run, "ubuntu")
	if !strings.Contains(cmd, "ssh") {
		t.Errorf("SSHCommand() = %q, want to contain \"ssh\"", cmd)
	}
	if !strings.Contains(cmd, "12255") {
		t.Errorf("SSHCommand() = %q, want to contain port 12255", cmd)
	}
	if !strings.Contains(cmd, "ubuntu@localhost") {
		t.Errorf("SSHCommand() = %q, want to contain \"ubuntu@localhost\"", cmd)
	}
	if !strings.Contains(cmd, "StrictHostKeyChecking=no") {
		t.Errorf("SSHCommand() = %q, want to contain StrictHostKeyChecking=no", cmd)
	}
}

func TestWorkDir(t *testing.T) {
	t.Parallel()

	id := "abc-123"
	wd := WorkDir(id)

	if !strings.Contains(wd, id) {
		t.Errorf("WorkDir(%q) = %q, want to contain the ID", id, wd)
	}
	if !strings.Contains(wd, ".bootwrangler") {
		t.Errorf("WorkDir(%q) = %q, want to contain .bootwrangler", id, wd)
	}
	if !strings.Contains(wd, "lab") {
		t.Errorf("WorkDir(%q) = %q, want to contain lab", id, wd)
	}
}
