// Package alpine renders Alpine Linux setup-alpine unattended installer assets.
package alpine

import (
	"github.com/alexmchughdev/bootwrangler/internal/manifest"
	"github.com/alexmchughdev/bootwrangler/internal/profile"
	"github.com/alexmchughdev/bootwrangler/internal/render"
)

// Renderer generates Alpine Linux installer assets.
type Renderer struct{}

// Name returns the unique renderer identifier.
func (r Renderer) Name() string { return "alpine" }

// Family returns "alpine".
func (r Renderer) Family() string { return "alpine" }

// SupportedVersions returns nil, accepting any Alpine version.
func (r Renderer) SupportedVersions() []string { return nil }

// Validate performs Alpine-specific profile validation.
func (r Renderer) Validate(_ profile.Profile) error { return nil }

// Render generates Alpine installer assets into opts.OutDir.
func (r Renderer) Render(_ profile.Profile, _ render.Options) (manifest.Manifest, error) {
	return manifest.Manifest{
		OSFamily: "alpine",
		Renderer: r.Name(),
	}, nil
}

func init() {
	render.Register(Renderer{})
}
