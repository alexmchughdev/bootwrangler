// Package render defines the renderer interface and render options used by all
// distribution-specific renderer implementations.
package render

import (
	"github.com/alexmchughdev/bootwrangler/internal/manifest"
	"github.com/alexmchughdev/bootwrangler/internal/profile"
)

// Options controls renderer behaviour for one render run.
type Options struct {
	// OutDir is the directory where generated files are written.
	OutDir string

	// ServerBaseURL is the local provisioning server base URL that the
	// generated iPXE/netboot entries should reference.
	ServerBaseURL string

	// DryRun causes renderers to plan without writing files.
	DryRun bool
}

// Renderer generates unattended installer assets from a profile.
type Renderer interface {
	// Name returns the unique renderer identifier.
	Name() string

	// Family returns the OS family this renderer handles (e.g. "alpine").
	Family() string

	// SupportedVersions returns the OS versions this renderer accepts.
	// An empty slice means any version is accepted.
	SupportedVersions() []string

	// Validate performs renderer-specific validation of the profile in
	// addition to the common validation already applied by profile.Validate.
	Validate(profile.Profile) error

	// Render generates installer assets into opts.OutDir and returns the
	// resulting manifest. It must not perform destructive disk operations.
	Render(profile.Profile, Options) (manifest.Manifest, error)
}
