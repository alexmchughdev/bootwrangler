package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunHelp(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"--help"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0", exitCode)
	}
	if !strings.Contains(stdout.String(), "Usage:") {
		t.Fatalf("Run() stdout = %q, want usage", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("Run() stderr = %q, want empty", stderr.String())
	}
}

func TestRunVersion(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"version"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0", exitCode)
	}
	if stdout.String() != "BootWrangler dev\n" {
		t.Fatalf("Run() stdout = %q, want %q", stdout.String(), "BootWrangler dev\n")
	}
	if stderr.Len() != 0 {
		t.Fatalf("Run() stderr = %q, want empty", stderr.String())
	}
}

func TestRunUnknownCommand(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"unknown"}, &stdout, &stderr)

	if exitCode != 2 {
		t.Fatalf("Run() exit code = %d, want 2", exitCode)
	}
	if stdout.Len() != 0 {
		t.Fatalf("Run() stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), `unknown command "unknown"`) {
		t.Fatalf("Run() stderr = %q, want unknown command error", stderr.String())
	}
}

func TestRunProfileValidate(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"profile", "validate", "../../examples/profiles/alpine-minimal.yaml"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0; stderr = %q", exitCode, stderr.String())
	}
	if stdout.String() != "profile valid: alpine-minimal\n" {
		t.Fatalf("Run() stdout = %q, want %q", stdout.String(), "profile valid: alpine-minimal\n")
	}
	if stderr.Len() != 0 {
		t.Fatalf("Run() stderr = %q, want empty", stderr.String())
	}
}

func TestParseProfileEditArgs(t *testing.T) {
	t.Parallel()

	path, err := parseProfileEditArgs([]string{"profile.yaml", "--editor", "nvim"})
	if err != nil {
		t.Fatalf("parseProfileEditArgs() error = %v, want nil", err)
	}
	if path != "profile.yaml" {
		t.Fatalf("parseProfileEditArgs() path = %q, want %q", path, "profile.yaml")
	}
}

func TestParseProfileEditArgsRejectsUnsupportedEditor(t *testing.T) {
	t.Parallel()

	_, err := parseProfileEditArgs([]string{"profile.yaml", "--editor", "vim"})

	if err == nil {
		t.Fatal("parseProfileEditArgs() error = nil, want error")
	}
}
