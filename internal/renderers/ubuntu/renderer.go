// Package ubuntu renders Ubuntu Server Subiquity autoinstall assets.
package ubuntu

import (
	"github.com/alexmchughdev/bootwrangler/internal/manifest"
	"github.com/alexmchughdev/bootwrangler/internal/profile"
	"github.com/alexmchughdev/bootwrangler/internal/render"
)

// Renderer generates Ubuntu autoinstall assets.
type Renderer struct{}

func (r Renderer) Name() string                     { return "ubuntu" }
func (r Renderer) Family() string                   { return "ubuntu" }
func (r Renderer) SupportedVersions() []string      { return nil }
func (r Renderer) Validate(_ profile.Profile) error { return nil }
func (r Renderer) Render(_ profile.Profile, _ render.Options) (manifest.Manifest, error) {
	return manifest.Manifest{
		OSFamily: "ubuntu",
		Renderer: r.Name(),
	}, nil
}

func init() {
	render.Register(Renderer{})
}
