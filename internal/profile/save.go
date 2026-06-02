package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Encode validates and serializes a profile deterministically.
func Encode(value Profile) ([]byte, error) {
	if err := Validate(value); err != nil {
		return nil, err
	}

	data, err := yaml.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode profile YAML: %w", err)
	}
	return data, nil
}

// SaveFile atomically writes a validated YAML profile.
func SaveFile(path string, value Profile) error {
	if path == "" {
		return fmt.Errorf("profile path is required")
	}
	extension := strings.ToLower(filepath.Ext(path))
	if extension != ".yaml" && extension != ".yml" {
		return fmt.Errorf("profile path %q must use a .yaml or .yml extension", path)
	}

	data, err := Encode(value)
	if err != nil {
		return err
	}

	if info, statErr := os.Lstat(path); statErr == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("refuse to overwrite profile symlink %q", path)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("profile path %q is not a regular file", path)
		}
	} else if !os.IsNotExist(statErr) {
		return fmt.Errorf("inspect profile path %q: %w", path, statErr)
	}

	temp, err := os.CreateTemp(filepath.Dir(path), ".bootwrangler-profile-*")
	if err != nil {
		return fmt.Errorf("create temporary profile next to %q: %w", path, err)
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)

	if err := temp.Chmod(0o600); err != nil {
		temp.Close()
		return fmt.Errorf("set temporary profile permissions: %w", err)
	}
	if _, err := temp.Write(data); err != nil {
		temp.Close()
		return fmt.Errorf("write temporary profile: %w", err)
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return fmt.Errorf("sync temporary profile: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close temporary profile: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("replace profile %q: %w", path, err)
	}

	return nil
}
