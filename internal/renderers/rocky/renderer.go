// Package rocky renders Rocky Linux Anaconda kickstart installer assets.
package rocky

import (
	"github.com/alexmchughdev/bootwrangler/internal/manifest"
	"github.com/alexmchughdev/bootwrangler/internal/profile"
	"github.com/alexmchughdev/bootwrangler/internal/render"
)

// Renderer generates Rocky Linux kickstart assets.
type Renderer struct{}

func (r Renderer) Name() string                     { return "rocky" }
func (r Renderer) Family() string                   { return "rocky" }
func (r Renderer) SupportedVersions() []string      { return nil }
func (r Renderer) Validate(_ profile.Profile) error { return nil }
func (r Renderer) Render(_ profile.Profile, _ render.Options) (manifest.Manifest, error) {
	return manifest.Manifest{
		OSFamily: "rocky",
		Renderer: r.Name(),
	}, nil
}

func init() {
	render.Register(Renderer{})
}
