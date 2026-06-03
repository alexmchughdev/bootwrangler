package render

import (
	"fmt"
	"path/filepath"

	"github.com/alexmchughdev/bootwrangler/internal/manifest"
	"github.com/alexmchughdev/bootwrangler/internal/profile"
)

// MultiResult is the result of rendering multiple profiles.
type MultiResult struct {
	Manifests []manifest.Manifest
	Errors    []MultiError
}

// MultiError pairs a profile name with a render error.
type MultiError struct {
	ProfileName string
	Err         error
}

// RenderAll renders a slice of profiles into subdirectories of outDir.
// Each profile renders into outDir/<profile-name>/. Errors for individual
// profiles are collected rather than aborting the whole batch.
func RenderAll(profiles []profile.Profile, outDir string) MultiResult {
	result := MultiResult{}
	for _, p := range profiles {
		subDir := filepath.Join(outDir, sanitiseName(p.Name))
		r, err := Lookup(p.OS.Family)
		if err != nil {
			result.Errors = append(result.Errors, MultiError{ProfileName: p.Name, Err: err})
			continue
		}
		opts := Options{OutDir: subDir}
		m, err := r.Render(p, opts)
		if err != nil {
			result.Errors = append(result.Errors, MultiError{ProfileName: p.Name, Err: fmt.Errorf("render %s: %w", p.Name, err)})
			continue
		}
		result.Manifests = append(result.Manifests, m)
	}
	return result
}

// sanitiseName replaces characters unsafe for directory names with underscores.
func sanitiseName(name string) string {
	out := make([]byte, len(name))
	for i := range name {
		c := name[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.' {
			out[i] = c
		} else {
			out[i] = '_'
		}
	}
	return string(out)
}
