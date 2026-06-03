package library

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// VersionEntry is one entry in a profile's local history.
type VersionEntry struct {
	Hash      string    `yaml:"hash"`
	Message   string    `yaml:"message"`
	Timestamp time.Time `yaml:"timestamp"`
}

// VersionedLibrary wraps Library with Git-backed version history.
// Version history is maintained in a bare Git repository inside the workspace.
type VersionedLibrary struct {
	*Library
}

// NewVersioned creates a VersionedLibrary backed by workspaceDir.
func NewVersioned(workspaceDir string) *VersionedLibrary {
	return &VersionedLibrary{Library: New(workspaceDir)}
}

// Init initialises the workspace and the embedded Git repository.
func (v *VersionedLibrary) Init() error {
	if err := v.Library.Init(); err != nil {
		return err
	}
	return v.initGit()
}

func (v *VersionedLibrary) initGit() error {
	gitDir := filepath.Join(v.workspaceDir, "profiles")
	if _, err := os.Stat(filepath.Join(gitDir, ".git")); err == nil {
		return nil
	}
	if err := gitRun(gitDir, "init"); err != nil {
		return fmt.Errorf("version history: git init: %w", err)
	}
	if err := gitRun(gitDir, "config", "user.email", "bootwrangler@localhost"); err != nil {
		return fmt.Errorf("version history: git config: %w", err)
	}
	if err := gitRun(gitDir, "config", "user.name", "BootWrangler"); err != nil {
		return fmt.Errorf("version history: git config: %w", err)
	}
	if err := gitRun(gitDir, "config", "commit.gpgsign", "false"); err != nil {
		return fmt.Errorf("version history: git config: %w", err)
	}
	return nil
}

// CommitProfile saves the profile YAML and commits it to the embedded Git repo.
func (v *VersionedLibrary) CommitProfile(filename, message string) error {
	dir := ProfilesDir(v.workspaceDir)
	if err := gitRun(dir, "add", filename); err != nil {
		return fmt.Errorf("version history: git add: %w", err)
	}
	// Check if there is anything to commit.
	out, err := gitOutput(dir, "status", "--porcelain")
	if err != nil {
		return fmt.Errorf("version history: git status: %w", err)
	}
	if strings.TrimSpace(out) == "" {
		return nil
	}
	if err := gitRun(dir, "commit", "-m", message); err != nil {
		return fmt.Errorf("version history: git commit: %w", err)
	}
	return nil
}

// History returns the version log for one profile file.
func (v *VersionedLibrary) History(filename string) ([]VersionEntry, error) {
	dir := ProfilesDir(v.workspaceDir)
	out, err := gitOutput(dir, "log", "--format=%H\x1f%s\x1f%aI", "--", filename)
	if err != nil {
		return nil, fmt.Errorf("version history: git log: %w", err)
	}
	var entries []VersionEntry
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\x1f", 3)
		if len(parts) != 3 {
			continue
		}
		ts, _ := time.Parse(time.RFC3339, parts[2])
		entries = append(entries, VersionEntry{
			Hash:      parts[0],
			Message:   parts[1],
			Timestamp: ts,
		})
	}
	return entries, nil
}

// Diff returns the unified diff between two commits for one profile file.
func (v *VersionedLibrary) Diff(filename, fromHash, toHash string) (string, error) {
	dir := ProfilesDir(v.workspaceDir)
	out, err := gitOutput(dir, "diff", fromHash+".."+toHash, "--", filename)
	if err != nil {
		return "", fmt.Errorf("version history: git diff: %w", err)
	}
	return out, nil
}

// Restore checks out a specific commit version of a profile file.
func (v *VersionedLibrary) Restore(filename, hash string) error {
	dir := ProfilesDir(v.workspaceDir)
	if err := gitRun(dir, "checkout", hash, "--", filename); err != nil {
		return fmt.Errorf("version history: restore %s@%s: %w", filename, hash, err)
	}
	return nil
}

func gitRun(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
