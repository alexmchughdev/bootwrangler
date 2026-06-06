package app

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alexmchughdev/bootwrangler/internal/images"
	"github.com/alexmchughdev/bootwrangler/internal/library"
	"github.com/alexmchughdev/bootwrangler/internal/media"
)

func resolveMediaBuildContent(plan *media.BuildPlan) {
	plan.Ready = true
	plan.Warnings = nil
	plan.Errors = nil

	for i := range plan.Actions {
		action := &plan.Actions[i]
		action.ContentResolution = resolveMediaPartitionContent(action.Content)
		for _, warning := range action.ContentResolution.Warnings {
			plan.Warnings = append(plan.Warnings, fmt.Sprintf("%s: %s", action.Label, warning))
		}
		for _, err := range action.ContentResolution.Errors {
			plan.Errors = append(plan.Errors, fmt.Sprintf("%s: %s", action.Label, err))
		}
		if len(action.ContentResolution.Errors) > 0 {
			plan.Ready = false
		}
	}
}

func resolveMediaPartitionContent(content media.PartitionContent) media.ContentResolution {
	switch content.Type {
	case media.ContentBootMenu:
		return readyContent("Boot menu files will be generated during media build.", "")
	case media.ContentCatalogueImage:
		return resolveCatalogueImageContent(content)
	case media.ContentCustomImage:
		return resolveCustomImageContent(content)
	case media.ContentImageFile:
		return resolvePathContent(content.Image, "image file")
	case media.ContentRenderedProfile:
		return resolveRenderedProfileContent(content.Profile)
	case media.ContentProfileBundle:
		return resolvePathContent(content.Bundle, "profile bundle")
	case media.ContentCustomFiles:
		return warningContent("Custom file content is accepted by the recipe but not expanded by the planner yet.",
			"Custom files are not resolved before build planning.")
	case media.ContentEmpty:
		return readyContent("No content will be written to this partition.", "")
	default:
		return blockedContent(fmt.Sprintf("unsupported content type %q", content.Type))
	}
}

func resolveCatalogueImageContent(content media.PartitionContent) media.ContentResolution {
	entry, version, arch, img, err := images.FindImage(images.BuiltinCatalogue(), content.Image, content.Version, "")
	if err != nil {
		return blockedContent(err.Error())
	}
	imageName := images.ImageName(entry.ID, version.Version)
	if !img.Compatibility.ISOFileBoot {
		return blockedContent(fmt.Sprintf("image %s is not marked as ISO-file boot compatible", imageName))
	}

	status := images.CheckCache(images.CacheDir(), entry.ID, version.Version, arch.Arch, img)
	if !status.Cached {
		resolution := blockedContent(fmt.Sprintf("image %s is not cached; download and verify it before media build", imageName))
		resolution.SourcePath = status.Path
		return resolution
	}

	message := fmt.Sprintf("Cached catalogue image %s for %s is available.", imageName, arch.Arch)
	resolution := readyContent(message, status.Path)
	if img.ChecksumURL != "" && !status.Verified {
		resolution.Status = media.ContentStatusWarning
		resolution.Warnings = append(resolution.Warnings, "Cached image has not been checksum verified.")
	} else if status.Verified {
		resolution.Message = fmt.Sprintf("Cached and verified catalogue image %s for %s is available.", imageName, arch.Arch)
	}
	return resolution
}

func resolveCustomImageContent(content media.PartitionContent) media.ContentResolution {
	img, err := defaultCustomImage(content.Image)
	if err != nil {
		return blockedContent(err.Error())
	}
	if !img.Compatibility.ISOFileBoot {
		return blockedContent(fmt.Sprintf("custom image %s is not marked as ISO-file boot compatible", img.ID))
	}
	if img.Source.Type == "url" {
		return blockedContent(fmt.Sprintf("custom image %s uses a URL source; download and cache support is not available for media builds yet", img.ID))
	}

	path, err := images.ResolveCustomImageFile(img)
	if err != nil {
		return blockedContent(err.Error())
	}
	message := fmt.Sprintf("Local custom image %s is available.", img.ID)
	if img.Checksum != nil {
		message = fmt.Sprintf("Local custom image %s is available and checksum verified.", img.ID)
	}
	return readyContent(message, path)
}

func resolveRenderedProfileContent(name string) media.ContentResolution {
	lib := library.New(library.DefaultDir())
	p, err := lib.Get(name)
	if err != nil {
		return blockedContent(fmt.Sprintf("profile %s is not available in the local library: %v", name, err))
	}
	return readyContent(fmt.Sprintf("Library profile %s is available for rendering.", p.Name), "")
}

func resolvePathContent(path, label string) media.ContentResolution {
	clean := filepath.Clean(path)
	info, err := os.Stat(clean)
	if err != nil {
		return blockedContent(fmt.Sprintf("%s %s is not available: %v", label, clean, err))
	}
	if info.IsDir() {
		return blockedContent(fmt.Sprintf("%s %s must be a file, not a directory", label, clean))
	}
	return readyContent(fmt.Sprintf("%s %s is available.", label, clean), clean)
}

func readyContent(message, sourcePath string) media.ContentResolution {
	return media.ContentResolution{
		Status:     media.ContentStatusReady,
		Message:    message,
		SourcePath: sourcePath,
	}
}

func warningContent(message, warning string) media.ContentResolution {
	return media.ContentResolution{
		Status:   media.ContentStatusWarning,
		Message:  message,
		Warnings: []string{warning},
	}
}

func blockedContent(message string) media.ContentResolution {
	return media.ContentResolution{
		Status:  media.ContentStatusBlocked,
		Message: message,
		Errors:  []string{message},
	}
}
