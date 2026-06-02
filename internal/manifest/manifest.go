// Package manifest defines the build record produced by a renderer run.
package manifest

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// FileEntry records one output file in a renderer manifest.
type FileEntry struct {
	Path    string `json:"path" yaml:"path"`
	Purpose string `json:"purpose" yaml:"purpose"`
	SHA256  string `json:"sha256" yaml:"sha256,omitempty"`
}

// Manifest describes the complete output of a renderer run.
type Manifest struct {
	ProfileName string      `json:"profile_name" yaml:"profile_name"`
	OSFamily    string      `json:"os_family" yaml:"os_family"`
	OSVersion   string      `json:"os_version" yaml:"os_version"`
	Renderer    string      `json:"renderer" yaml:"renderer"`
	Files       []FileEntry `json:"files" yaml:"files"`
	Warnings    []string    `json:"warnings" yaml:"warnings"`
}

// AddFile appends a file entry to the manifest.
func (m *Manifest) AddFile(path, purpose string) {
	m.Files = append(m.Files, FileEntry{
		Path:    path,
		Purpose: purpose,
	})
}

// AddWarning appends a warning message to the manifest.
func (m *Manifest) AddWarning(msg string) {
	m.Warnings = append(m.Warnings, msg)
}

// ComputeChecksums populates the SHA256 field for each file relative to outDir.
func (m *Manifest) ComputeChecksums(outDir string) error {
	for i, entry := range m.Files {
		sum, err := sha256File(filepath.Join(outDir, entry.Path))
		if err != nil {
			return fmt.Errorf("manifest: checksum %s: %w", entry.Path, err)
		}
		m.Files[i].SHA256 = sum
	}
	return nil
}

// WriteSHA256SUMS writes a SHA256SUMS file inside outDir.
func (m *Manifest) WriteSHA256SUMS(outDir string) error {
	var sb strings.Builder
	for _, entry := range m.Files {
		if entry.SHA256 != "" {
			fmt.Fprintf(&sb, "%s  %s\n", entry.SHA256, entry.Path)
		}
	}
	return os.WriteFile(filepath.Join(outDir, "SHA256SUMS"), []byte(sb.String()), 0o644)
}

// WriteManifestFile serialises the manifest as YAML inside outDir.
func (m *Manifest) WriteManifestFile(outDir string) error {
	data, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("manifest: marshal: %w", err)
	}
	return os.WriteFile(filepath.Join(outDir, "manifest.yaml"), data, 0o644)
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
