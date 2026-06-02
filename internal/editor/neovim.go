// Package editor plans guarded launches of external editors.
package editor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// LaunchPlan describes one direct process invocation.
type LaunchPlan struct {
	Command string
	Args    []string
}

// NeovimPlan plans a direct Neovim process launch for one regular file.
func NeovimPlan(path string, readOnly bool) (LaunchPlan, error) {
	return neovimPlan(path, readOnly, exec.LookPath)
}

func neovimPlan(path string, readOnly bool, lookup func(string) (string, error)) (LaunchPlan, error) {
	if path == "" {
		return LaunchPlan{}, fmt.Errorf("editor path is required")
	}

	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return LaunchPlan{}, fmt.Errorf("resolve editor path %q: %w", path, err)
	}
	info, err := os.Lstat(absolutePath)
	if err != nil {
		return LaunchPlan{}, fmt.Errorf("inspect editor path %q: %w", absolutePath, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return LaunchPlan{}, fmt.Errorf("refuse to open editor symlink %q", absolutePath)
	}
	if !info.Mode().IsRegular() {
		return LaunchPlan{}, fmt.Errorf("editor path %q is not a regular file", absolutePath)
	}

	command, err := lookup("nvim")
	if err != nil {
		return LaunchPlan{}, fmt.Errorf("find Neovim executable: %w", err)
	}
	args := make([]string, 0, 3)
	if readOnly {
		args = append(args, "-R")
	}
	args = append(args, "--", absolutePath)

	return LaunchPlan{
		Command: command,
		Args:    args,
	}, nil
}

// Start launches an editor process without shell interpolation.
func Start(plan LaunchPlan) error {
	command := exec.Command(plan.Command, plan.Args...)
	if err := command.Start(); err != nil {
		return fmt.Errorf("start editor: %w", err)
	}
	go func() {
		_ = command.Wait()
	}()

	return nil
}
