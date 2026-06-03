package render_test

import (
	"os"
	"testing"

	"github.com/alexmchughdev/bootwrangler/internal/profile"
	"github.com/alexmchughdev/bootwrangler/internal/render"
	_ "github.com/alexmchughdev/bootwrangler/internal/renderers/all"
)

func makeTestProfile(name, family, version string) profile.Profile {
	return profile.Profile{
		Name: name,
		OS:   profile.OS{Family: family, Version: version},
		System: profile.System{
			Hostname: name,
			Timezone: "UTC",
			Keyboard: "us",
			Locale:   "en_US.UTF-8",
		},
		Network: profile.Network{Mode: "dhcp"},
		Disk:    profile.Disk{Mode: "wipe", Target: "auto", InstallMode: "server", Filesystem: "ext4", ConfirmDestructive: true},
		SSH:     profile.SSH{Enabled: true},
		Users:   []profile.User{{Name: "admin", Shell: "/bin/bash", Sudo: true}},
	}
}

func TestRenderAll_Empty(t *testing.T) {
	result := render.RenderAll(nil, t.TempDir())
	if len(result.Manifests) != 0 || len(result.Errors) != 0 {
		t.Error("empty input should produce empty result")
	}
}

func TestRenderAll_MixedSuccess(t *testing.T) {
	dir := t.TempDir()
	profiles := []profile.Profile{
		makeTestProfile("node-01", "ubuntu", "24.04"),
		{Name: "bad", OS: profile.OS{Family: "unknowndistro", Version: "1.0"}},
	}
	result := render.RenderAll(profiles, dir)
	if len(result.Manifests) != 1 {
		t.Errorf("expected 1 manifest, got %d", len(result.Manifests))
	}
	if len(result.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(result.Errors))
	}
	if result.Errors[0].ProfileName != "bad" {
		t.Errorf("error profile name = %q", result.Errors[0].ProfileName)
	}
}

func TestRenderAll_OutputDirs(t *testing.T) {
	dir := t.TempDir()
	profiles := []profile.Profile{
		makeTestProfile("server-01", "ubuntu", "24.04"),
		makeTestProfile("server-02", "ubuntu", "24.04"),
	}
	result := render.RenderAll(profiles, dir)
	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors[0].Err)
	}
	for _, name := range []string{"server-01", "server-02"} {
		if _, err := os.Stat(dir + "/" + name); err != nil {
			t.Errorf("output dir for %q not created: %v", name, err)
		}
	}
}
