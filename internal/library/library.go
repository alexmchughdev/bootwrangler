// Package library manages the local BootWrangler profile library stored under
// ~/.bootwrangler/profiles/.
package library

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/alexmchughdev/bootwrangler/internal/profile"
)

// DefaultDir returns the default BootWrangler workspace directory.
func DefaultDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".bootwrangler"
	}
	return filepath.Join(home, ".bootwrangler")
}

// ProfilesDir returns the profiles subdirectory within the workspace.
func ProfilesDir(workspaceDir string) string {
	return filepath.Join(workspaceDir, "profiles")
}

// Entry is a library index record for one stored profile.
type Entry struct {
	Name      string    `yaml:"name"`
	Filename  string    `yaml:"filename"`
	OSFamily  string    `yaml:"os_family"`
	OSVersion string    `yaml:"os_version"`
	UpdatedAt time.Time `yaml:"updated_at"`
}

// Library manages the local profile store.
type Library struct {
	workspaceDir string
}

// New creates a Library backed by the given workspace directory.
// Call Init to ensure the directory structure exists.
func New(workspaceDir string) *Library {
	return &Library{workspaceDir: workspaceDir}
}

// Init creates the workspace directory structure.
func (l *Library) Init() error {
	dirs := []string{
		l.workspaceDir,
		ProfilesDir(l.workspaceDir),
		filepath.Join(l.workspaceDir, "cache", "images"),
		filepath.Join(l.workspaceDir, "exports"),
		filepath.Join(l.workspaceDir, "lab"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("library init: %w", err)
		}
	}
	return nil
}

// Add saves the profile to the library and returns its filename.
// The profile must pass validation before being added.
func (l *Library) Add(p profile.Profile) (string, error) {
	if err := profile.Validate(p); err != nil {
		return "", fmt.Errorf("library add: %w", err)
	}
	if err := os.MkdirAll(ProfilesDir(l.workspaceDir), 0o755); err != nil {
		return "", fmt.Errorf("library add: %w", err)
	}

	filename := safeFilename(p.Name) + ".yaml"
	path := filepath.Join(ProfilesDir(l.workspaceDir), filename)
	if err := profile.SaveFile(path, p); err != nil {
		return "", fmt.Errorf("library add: %w", err)
	}
	return filename, nil
}

// Get loads one profile by name from the library.
func (l *Library) Get(name string) (profile.Profile, error) {
	filename := safeFilename(name) + ".yaml"
	path := filepath.Join(ProfilesDir(l.workspaceDir), filename)
	p, err := profile.LoadAndValidateFile(path)
	if err != nil {
		return profile.Profile{}, fmt.Errorf("library get %q: %w", name, err)
	}
	return p, nil
}

// Remove deletes one profile from the library by name.
func (l *Library) Remove(name string) error {
	filename := safeFilename(name) + ".yaml"
	path := filepath.Join(ProfilesDir(l.workspaceDir), filename)
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("library remove %q: %w", name, err)
	}
	return nil
}

// List returns all profile entries in the library, sorted by filename.
func (l *Library) List() ([]Entry, error) {
	dir := ProfilesDir(l.workspaceDir)
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("library list: %w", err)
	}

	var out []Entry
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		p, err := profile.LoadAndValidateFile(path)
		if err != nil {
			continue
		}
		info, _ := e.Info()
		entry := Entry{
			Name:      p.Name,
			Filename:  e.Name(),
			OSFamily:  p.OS.Family,
			OSVersion: p.OS.Version,
		}
		if info != nil {
			entry.UpdatedAt = info.ModTime()
		}
		out = append(out, entry)
	}
	return out, nil
}

// IndexPath returns the path to the library index file.
func (l *Library) IndexPath() string {
	return filepath.Join(l.workspaceDir, "index.yaml")
}

// WriteIndex serialises the current library listing to index.yaml.
func (l *Library) WriteIndex() error {
	entries, err := l.List()
	if err != nil {
		return err
	}
	data, err := yaml.Marshal(entries)
	if err != nil {
		return fmt.Errorf("library index: %w", err)
	}
	return os.WriteFile(l.IndexPath(), data, 0o644)
}

func safeFilename(name string) string {
	var sb strings.Builder
	for _, r := range strings.ToLower(name) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' {
			sb.WriteRune(r)
		} else if r == ' ' || r == '_' {
			sb.WriteRune('-')
		}
	}
	result := sb.String()
	if result == "" {
		return "profile"
	}
	return result
}
