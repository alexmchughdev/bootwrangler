// Package opensuse renders openSUSE AutoYaST installer assets.
package opensuse

import (
	"github.com/alexmchughdev/bootwrangler/internal/manifest"
	"github.com/alexmchughdev/bootwrangler/internal/profile"
	"github.com/alexmchughdev/bootwrangler/internal/render"
)

// Renderer generates openSUSE AutoYaST assets.
type Renderer struct{}

func (r Renderer) Name() string                     { return "opensuse" }
func (r Renderer) Family() string                   { return "opensuse" }
func (r Renderer) SupportedVersions() []string      { return nil }
func (r Renderer) Validate(_ profile.Profile) error { return nil }
func (r Renderer) Render(_ profile.Profile, _ render.Options) (manifest.Manifest, error) {
	return manifest.Manifest{
		OSFamily: "opensuse",
		Renderer: r.Name(),
	}, nil
}

func init() {
	render.Register(Renderer{})
}
