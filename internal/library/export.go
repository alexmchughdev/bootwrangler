package library

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/alexmchughdev/bootwrangler/internal/profile"
)

// ExportBundle writes a profile export ZIP to path and returns the bundle size.
func ExportBundle(p profile.Profile, path string) error {
	if err := profile.Validate(p); err != nil {
		return fmt.Errorf("export: %w", err)
	}

	profileYAML, err := profile.Encode(p)
	if err != nil {
		return fmt.Errorf("export: marshal profile: %w", err)
	}

	manifestYAML, err := marshalBundleManifest(p)
	if err != nil {
		return fmt.Errorf("export: marshal manifest: %w", err)
	}

	readme := bundleREADME(p)

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("export: create %s: %w", path, err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	for name, content := range map[string][]byte{
		"profile.yaml":  profileYAML,
		"manifest.yaml": manifestYAML,
		"README.md":     []byte(readme),
	} {
		entry, err := w.Create(name)
		if err != nil {
			return fmt.Errorf("export: zip entry %s: %w", name, err)
		}
		if _, err := entry.Write(content); err != nil {
			return fmt.Errorf("export: write %s: %w", name, err)
		}
	}

	return nil
}

// ImportBundle reads a profile export ZIP and returns the contained profile.
// The profile is validated before being returned.
func ImportBundle(path string) (profile.Profile, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return profile.Profile{}, fmt.Errorf("import: open %s: %w", path, err)
	}
	defer r.Close()

	files := map[string][]byte{}
	for _, f := range r.File {
		// Reject path traversal inside the ZIP.
		if strings.Contains(f.Name, "..") || filepath.IsAbs(f.Name) {
			return profile.Profile{}, fmt.Errorf("import: unsafe entry %q in bundle", f.Name)
		}
		rc, err := f.Open()
		if err != nil {
			return profile.Profile{}, fmt.Errorf("import: read entry %s: %w", f.Name, err)
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return profile.Profile{}, fmt.Errorf("import: read %s: %w", f.Name, err)
		}
		files[f.Name] = data
	}

	profileData, ok := files["profile.yaml"]
	if !ok {
		return profile.Profile{}, fmt.Errorf("import: bundle does not contain profile.yaml")
	}

	p, err := profile.LoadAndValidateBytes(profileData)
	if err != nil {
		return profile.Profile{}, fmt.Errorf("import: %w", err)
	}
	return p, nil
}

// bundleManifestEntry is the manifest included in an export bundle.
type bundleManifestEntry struct {
	ProfileName string    `yaml:"profile_name"`
	OSFamily    string    `yaml:"os_family"`
	OSVersion   string    `yaml:"os_version"`
	ExportedAt  time.Time `yaml:"exported_at"`
	Files       []string  `yaml:"files"`
}

func marshalBundleManifest(p profile.Profile) ([]byte, error) {
	m := bundleManifestEntry{
		ProfileName: p.Name,
		OSFamily:    p.OS.Family,
		OSVersion:   p.OS.Version,
		ExportedAt:  time.Now().UTC().Truncate(time.Second),
		Files:       []string{"profile.yaml", "manifest.yaml", "README.md"},
	}
	return yaml.Marshal(m)
}

func bundleREADME(p profile.Profile) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "# BootWrangler Profile: %s\n\n", p.Name)
	fmt.Fprintf(&sb, "OS family: %s\n", p.OS.Family)
	fmt.Fprintf(&sb, "OS version: %s\n", p.OS.Version)
	fmt.Fprintf(&sb, "Hostname: %s\n\n", p.System.Hostname)
	fmt.Fprintf(&sb, "## Contents\n\n")
	fmt.Fprintf(&sb, "- `profile.yaml` — provisioning profile\n")
	fmt.Fprintf(&sb, "- `manifest.yaml` — bundle metadata\n")
	fmt.Fprintf(&sb, "- `README.md` — this file\n\n")
	fmt.Fprintf(&sb, "## Import\n\n")
	fmt.Fprintf(&sb, "```\nbootwrangler library import <this-file>.zip\n```\n")
	return sb.String()
}
