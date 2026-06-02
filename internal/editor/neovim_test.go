package editor

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestNeovimPlan(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "profile.yaml")
	if err := os.WriteFile(path, []byte("name: example\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	plan, err := neovimPlan(path, true, func(name string) (string, error) {
		if name != "nvim" {
			t.Fatalf("lookup() name = %q, want %q", name, "nvim")
		}
		return "/usr/local/bin/nvim", nil
	})
	if err != nil {
		t.Fatalf("neovimPlan() error = %v", err)
	}

	if plan.Command != "/usr/local/bin/nvim" {
		t.Fatalf("neovimPlan().Command = %q, want %q", plan.Command, "/usr/local/bin/nvim")
	}
	wantArgs := []string{"-R", "--", path}
	if !reflect.DeepEqual(plan.Args, wantArgs) {
		t.Fatalf("neovimPlan().Args = %#v, want %#v", plan.Args, wantArgs)
	}
}

func TestNeovimPlanRejectsSymlink(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	target := filepath.Join(directory, "target.yaml")
	link := filepath.Join(directory, "link.yaml")
	if err := os.WriteFile(target, []byte("name: example\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("Symlink() error = %v", err)
	}

	_, err := neovimPlan(link, false, func(string) (string, error) {
		return "/usr/local/bin/nvim", nil
	})

	if err == nil {
		t.Fatal("neovimPlan() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "refuse to open editor symlink") {
		t.Fatalf("neovimPlan() error = %q, want symlink message", err)
	}
}

func TestNeovimPlanReportsMissingExecutable(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "profile.yaml")
	if err := os.WriteFile(path, []byte("name: example\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := neovimPlan(path, false, func(string) (string, error) {
		return "", errors.New("missing")
	})

	if err == nil {
		t.Fatal("neovimPlan() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "find Neovim executable") {
		t.Fatalf("neovimPlan() error = %q, want executable message", err)
	}
}
