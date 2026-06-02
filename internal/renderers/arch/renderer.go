// Package arch renders Arch Linux scripted install assets.
package arch

import (
	"github.com/alexmchughdev/bootwrangler/internal/manifest"
	"github.com/alexmchughdev/bootwrangler/internal/profile"
	"github.com/alexmchughdev/bootwrangler/internal/render"
)

// Renderer generates Arch Linux scripted install assets.
type Renderer struct{}

func (r Renderer) Name() string                     { return "arch" }
func (r Renderer) Family() string                   { return "arch" }
func (r Renderer) SupportedVersions() []string      { return nil }
func (r Renderer) Validate(_ profile.Profile) error { return nil }
func (r Renderer) Render(_ profile.Profile, _ render.Options) (manifest.Manifest, error) {
	return manifest.Manifest{
		OSFamily: "arch",
		Renderer: r.Name(),
	}, nil
}

func init() {
	render.Register(Renderer{})
}
