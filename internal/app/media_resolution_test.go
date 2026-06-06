package app

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alexmchughdev/bootwrangler/internal/images"
	"github.com/alexmchughdev/bootwrangler/internal/media"
	"github.com/alexmchughdev/bootwrangler/internal/profile"
)

func TestResolveMediaBuildContentBlocksUncachedCatalogueImage(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	plan := mediaPlanWithContent(media.PartitionContent{
		Type:    media.ContentCatalogueImage,
		Image:   "ubuntu-server",
		Version: "24.04",
	})

	resolveMediaBuildContent(&plan)

	if plan.Ready {
		t.Fatal("plan.Ready = true, want false")
	}
	if plan.Actions[0].ContentResolution.Status != media.ContentStatusBlocked {
		t.Fatalf("status = %q, want blocked", plan.Actions[0].ContentResolution.Status)
	}
	if !strings.Contains(plan.Actions[0].ContentResolution.Message, "not cached") {
		t.Fatalf("message = %q, want cache message", plan.Actions[0].ContentResolution.Message)
	}
}

func TestResolveMediaBuildContentBlocksURLCustomImage(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	svc := NewService()
	if err := svc.SaveCustomImage(images.CustomImage{
		ID:   "remote-os",
		Name: "Remote OS",
		Source: images.CustomSource{
			Type: "url",
			URL:  "https://example.com/remote-os.iso",
		},
		Compatibility: images.Compatibility{ISOFileBoot: true},
	}); err != nil {
		t.Fatalf("SaveCustomImage() error = %v", err)
	}

	plan := mediaPlanWithContent(media.PartitionContent{
		Type:  media.ContentCustomImage,
		Image: "remote-os",
	})

	resolveMediaBuildContent(&plan)

	if plan.Ready {
		t.Fatal("plan.Ready = true, want false")
	}
	if !strings.Contains(plan.Actions[0].ContentResolution.Message, "URL source") {
		t.Fatalf("message = %q, want URL source message", plan.Actions[0].ContentResolution.Message)
	}
}

func TestResolveMediaBuildContentAcceptsLocalCustomImageWithChecksum(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	imagePath := filepath.Join(dir, "company-os.iso")
	data := []byte("image bytes")
	if err := os.WriteFile(imagePath, data, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	sum := sha256.Sum256(data)

	svc := NewService()
	if err := svc.SaveCustomImage(images.CustomImage{
		ID:   "company-os",
		Name: "Company OS",
		Source: images.CustomSource{
			Type: "local-file",
			Path: imagePath,
		},
		Checksum: &images.Checksum{
			Type:  "sha256",
			Value: hex.EncodeToString(sum[:]),
		},
		Compatibility: images.Compatibility{ISOFileBoot: true},
	}); err != nil {
		t.Fatalf("SaveCustomImage() error = %v", err)
	}

	plan := mediaPlanWithContent(media.PartitionContent{
		Type:  media.ContentCustomImage,
		Image: "company-os",
	})

	resolveMediaBuildContent(&plan)

	if !plan.Ready {
		t.Fatalf("plan.Ready = false, errors = %v", plan.Errors)
	}
	resolution := plan.Actions[0].ContentResolution
	if resolution.Status != media.ContentStatusReady {
		t.Fatalf("status = %q, want ready", resolution.Status)
	}
	if resolution.SourcePath != imagePath {
		t.Fatalf("SourcePath = %q, want %q", resolution.SourcePath, imagePath)
	}
	if !strings.Contains(resolution.Message, "checksum verified") {
		t.Fatalf("message = %q, want checksum verification message", resolution.Message)
	}
}

func TestResolveMediaBuildContentFindsLibraryProfile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	svc := NewService()
	if err := svc.LibraryInit(); err != nil {
		t.Fatalf("LibraryInit() error = %v", err)
	}
	p := minimalMediaProfile("edge-node")
	if _, err := svc.LibraryAdd(p); err != nil {
		t.Fatalf("LibraryAdd() error = %v", err)
	}

	plan := mediaPlanWithContent(media.PartitionContent{
		Type:    media.ContentRenderedProfile,
		Profile: p.Name,
	})

	resolveMediaBuildContent(&plan)

	if !plan.Ready {
		t.Fatalf("plan.Ready = false, errors = %v", plan.Errors)
	}
	if plan.Actions[0].ContentResolution.Status != media.ContentStatusReady {
		t.Fatalf("status = %q, want ready", plan.Actions[0].ContentResolution.Status)
	}
}

func mediaPlanWithContent(content media.PartitionContent) media.BuildPlan {
	return media.BuildPlan{
		Ready: true,
		Actions: []media.PartitionAction{
			{Label: "CONTENT", Content: content},
		},
	}
}

func minimalMediaProfile(name string) profile.Profile {
	return profile.Profile{
		Name: name,
		OS: profile.OS{
			Family:       "ubuntu",
			Version:      "24.04",
			Architecture: "x86_64",
		},
		System: profile.System{
			Hostname: "edge-node",
			Timezone: "UTC",
			Locale:   "en_US.UTF-8",
			Keyboard: "us",
		},
		Network: profile.Network{Mode: "dhcp", Interface: "eth0"},
		Disk: profile.Disk{
			Mode:               "wipe",
			Target:             "auto",
			InstallMode:        "server",
			Filesystem:         "ext4",
			ConfirmDestructive: true,
		},
		SSH: profile.SSH{Enabled: true},
		Users: []profile.User{
			{
				Name:    "admin",
				Shell:   "/bin/bash",
				Sudo:    true,
				Groups:  []string{"sudo"},
				SSHKeys: []string{"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJ0EkzLnXFTKfQXv2bseD0jLxoc+R6NxRHWqnY15T9i6 example-only@bootwrangler.invalid"},
			},
		},
	}
}
