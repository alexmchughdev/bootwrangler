// Package app provides the backend methods exposed to the BootWrangler desktop
// application.
package app

import (
	"context"
	"errors"

	"github.com/alexmchughdev/bootwrangler/internal/editor"
	"github.com/alexmchughdev/bootwrangler/internal/library"
	"github.com/alexmchughdev/bootwrangler/internal/manifest"
	"github.com/alexmchughdev/bootwrangler/internal/profile"
	"github.com/alexmchughdev/bootwrangler/internal/render"
	"github.com/alexmchughdev/bootwrangler/internal/secrets"
	"github.com/alexmchughdev/bootwrangler/internal/version"
)

// HealthStatus describes desktop backend availability.
type HealthStatus struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// ValidationResult describes backend validation feedback for the GUI.
type ValidationResult struct {
	Valid    bool     `json:"valid"`
	Problems []string `json:"problems"`
}

// Service exposes application operations to the desktop frontend.
type Service struct{}

// NewService creates a desktop application service.
func NewService() *Service {
	return &Service{}
}

// Startup is called when the desktop window is created.
func (s *Service) Startup(context.Context) {}

// Version returns the BootWrangler build version.
func (s *Service) Version() string {
	return version.Current()
}

// Health returns the current backend health status.
func (s *Service) Health() HealthStatus {
	return HealthStatus{
		Status:  "ok",
		Version: s.Version(),
	}
}

// ValidateProfile validates one profile without side effects.
func (s *Service) ValidateProfile(value profile.Profile) ValidationResult {
	return validationResult(profile.Validate(value))
}

// ValidateSSHPublicKey validates one SSH public key without storing it.
func (s *Service) ValidateSSHPublicKey(value string) ValidationResult {
	return validationResult(secrets.ValidateSSHPublicKey(value))
}

// LoadProfile loads and validates one profile YAML file.
func (s *Service) LoadProfile(path string) (profile.Profile, error) {
	return profile.LoadAndValidateFile(path)
}

// SaveProfile validates and atomically saves one profile YAML file.
func (s *Service) SaveProfile(path string, value profile.Profile) error {
	return profile.SaveFile(path, value)
}

// AvailableRenderers returns the OS families for which a renderer is registered.
func (s *Service) AvailableRenderers() []string {
	return render.Families()
}

// RenderProfile renders one profile into outDir and returns the manifest.
func (s *Service) RenderProfile(value profile.Profile, outDir string) (manifest.Manifest, error) {
	r, err := render.Lookup(value.OS.Family)
	if err != nil {
		return manifest.Manifest{}, err
	}
	opts := render.Options{OutDir: outDir}
	return r.Render(value, opts)
}

// LibraryInit initialises the default BootWrangler workspace.
func (s *Service) LibraryInit() error {
	lib := library.New(library.DefaultDir())
	return lib.Init()
}

// LibraryAdd validates and saves the profile to the local library.
func (s *Service) LibraryAdd(value profile.Profile) (string, error) {
	lib := library.New(library.DefaultDir())
	return lib.Add(value)
}

// LibraryGet loads a profile by name from the local library.
func (s *Service) LibraryGet(name string) (profile.Profile, error) {
	lib := library.New(library.DefaultDir())
	return lib.Get(name)
}

// LibraryList returns all profiles stored in the local library.
func (s *Service) LibraryList() ([]library.Entry, error) {
	lib := library.New(library.DefaultDir())
	return lib.List()
}

// LibraryRemove deletes a profile from the local library by name.
func (s *Service) LibraryRemove(name string) error {
	lib := library.New(library.DefaultDir())
	return lib.Remove(name)
}

// OpenInNeovim opens one regular file in Neovim without shell interpolation.
func (s *Service) OpenInNeovim(path string, readOnly bool) error {
	plan, err := editor.NeovimPlan(path, readOnly)
	if err != nil {
		return err
	}
	return editor.Start(plan)
}

func validationResult(err error) ValidationResult {
	if err == nil {
		return ValidationResult{
			Valid:    true,
			Problems: []string{},
		}
	}

	var validationError *profile.ValidationError
	if errors.As(err, &validationError) {
		return ValidationResult{
			Problems: append([]string(nil), validationError.Problems...),
		}
	}

	return ValidationResult{
		Problems: []string{err.Error()},
	}
}
