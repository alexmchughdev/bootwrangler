package render_test

import (
	"testing"

	"github.com/alexmchughdev/bootwrangler/internal/manifest"
	"github.com/alexmchughdev/bootwrangler/internal/profile"
	"github.com/alexmchughdev/bootwrangler/internal/render"
)

// stub is a minimal Renderer used only in tests within this package.
type stub struct{ family string }

func (s stub) Name() string                   { return "stub-" + s.family }
func (s stub) Family() string                 { return s.family }
func (s stub) SupportedVersions() []string    { return nil }
func (s stub) Validate(profile.Profile) error { return nil }
func (s stub) Render(profile.Profile, render.Options) (manifest.Manifest, error) {
	return manifest.Manifest{}, nil
}

func TestLookupRegisteredRenderer(t *testing.T) {
	render.Register(stub{family: "testfamily"})

	r, err := render.Lookup("testfamily")
	if err != nil {
		t.Fatalf("expected renderer, got error: %v", err)
	}
	if r.Family() != "testfamily" {
		t.Errorf("expected family testfamily, got %s", r.Family())
	}
}

func TestLookupUnknownFamily(t *testing.T) {
	_, err := render.Lookup("gentoo")
	if err == nil {
		t.Fatal("expected error for unknown family, got nil")
	}
}

func TestFamiliesIncludesRegistered(t *testing.T) {
	render.Register(stub{family: "familya"})
	render.Register(stub{family: "familyb"})

	families := render.Families()
	found := map[string]bool{}
	for _, f := range families {
		found[f] = true
	}
	for _, want := range []string{"testfamily", "familya", "familyb"} {
		if !found[want] {
			t.Errorf("expected family %q in list %v", want, families)
		}
	}
}
