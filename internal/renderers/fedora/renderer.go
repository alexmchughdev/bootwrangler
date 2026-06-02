// Package fedora renders Fedora Server Anaconda kickstart installer assets.
package fedora

import (
	"github.com/alexmchughdev/bootwrangler/internal/manifest"
	"github.com/alexmchughdev/bootwrangler/internal/profile"
	"github.com/alexmchughdev/bootwrangler/internal/render"
)

// Renderer generates Fedora Server kickstart assets.
type Renderer struct{}

func (r Renderer) Name() string                     { return "fedora" }
func (r Renderer) Family() string                   { return "fedora" }
func (r Renderer) SupportedVersions() []string      { return nil }
func (r Renderer) Validate(_ profile.Profile) error { return nil }
func (r Renderer) Render(_ profile.Profile, _ render.Options) (manifest.Manifest, error) {
	return manifest.Manifest{
		OSFamily: "fedora",
		Renderer: r.Name(),
	}, nil
}

func init() {
	render.Register(Renderer{})
}
