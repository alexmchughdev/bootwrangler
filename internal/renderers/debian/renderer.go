// Package debian renders Debian preseed unattended installer assets.
package debian

import (
	"github.com/alexmchughdev/bootwrangler/internal/manifest"
	"github.com/alexmchughdev/bootwrangler/internal/profile"
	"github.com/alexmchughdev/bootwrangler/internal/render"
)

// Renderer generates Debian preseed assets.
type Renderer struct{}

func (r Renderer) Name() string                     { return "debian" }
func (r Renderer) Family() string                   { return "debian" }
func (r Renderer) SupportedVersions() []string      { return nil }
func (r Renderer) Validate(_ profile.Profile) error { return nil }
func (r Renderer) Render(_ profile.Profile, _ render.Options) (manifest.Manifest, error) {
	return manifest.Manifest{
		OSFamily: "debian",
		Renderer: r.Name(),
	}, nil
}

func init() {
	render.Register(Renderer{})
}
