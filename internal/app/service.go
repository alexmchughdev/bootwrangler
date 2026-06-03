// Package app provides the backend methods exposed to the BootWrangler desktop
// application.
package app

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/alexmchughdev/bootwrangler/internal/editor"
	"github.com/alexmchughdev/bootwrangler/internal/images"
	"github.com/alexmchughdev/bootwrangler/internal/library"
	"github.com/alexmchughdev/bootwrangler/internal/manifest"
	"github.com/alexmchughdev/bootwrangler/internal/media"
	"github.com/alexmchughdev/bootwrangler/internal/policy"
	"github.com/alexmchughdev/bootwrangler/internal/profile"
	"github.com/alexmchughdev/bootwrangler/internal/render"
	"github.com/alexmchughdev/bootwrangler/internal/secrets"
	"github.com/alexmchughdev/bootwrangler/internal/server"
	"github.com/alexmchughdev/bootwrangler/internal/usb"
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
type Service struct {
	mu  sync.Mutex
	srv *server.Server
}

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

// ServerStart starts the provisioning server serving root at addr.
// It returns the actual listening address.
func (s *Service) ServerStart(root, addr string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.srv != nil {
		return "", errors.New("server is already running")
	}
	srv, err := server.New(root, addr)
	if err != nil {
		return "", err
	}
	if err := srv.Start(); err != nil {
		return "", err
	}
	s.srv = srv
	return srv.Addr(), nil
}

// ServerStop stops the running provisioning server.
func (s *Service) ServerStop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.srv == nil {
		return nil
	}
	err := s.srv.Stop()
	s.srv = nil
	return err
}

// ServerAddr returns the address of the running server.
func (s *Service) ServerAddr() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.srv == nil {
		return ""
	}
	return s.srv.Addr()
}

// ServerLogs returns the request log of the running server.
func (s *Service) ServerLogs() []server.RequestLog {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.srv == nil {
		return nil
	}
	return s.srv.Logs()
}

// ServerEvents returns the install callback events of the running server.
func (s *Service) ServerEvents() []server.CallbackEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.srv == nil {
		return nil
	}
	return s.srv.Events()
}

// OpenInNeovim opens one regular file in Neovim without shell interpolation.
func (s *Service) OpenInNeovim(path string, readOnly bool) error {
	plan, err := editor.NeovimPlan(path, readOnly)
	if err != nil {
		return err
	}
	return editor.Start(plan)
}

// ListImages returns all entries in the built-in image catalogue.
func (s *Service) ListImages() []images.CatalogueEntry {
	return images.BuiltinCatalogue().Entries
}

// GetImage returns one catalogue entry by ID, or an error if not found.
func (s *Service) GetImage(id string) (images.CatalogueEntry, error) {
	cat := images.BuiltinCatalogue()
	for _, e := range cat.Entries {
		if e.ID == id {
			return e, nil
		}
	}
	return images.CatalogueEntry{}, fmt.Errorf("image not found: %s", id)
}

// ImageCacheStatus returns the cache status for a specific image version/arch.
func (s *Service) ImageCacheStatus(id, version, arch string) images.CacheStatus {
	cat := images.BuiltinCatalogue()
	_, _, _, img, err := images.FindImage(cat, id, version, arch)
	if err != nil {
		return images.CacheStatus{}
	}
	return images.CheckCache(images.CacheDir(), id, version, arch, img)
}

// ListDevices returns the currently connected block devices with safety annotations.
func (s *Service) ListDevices() ([]usb.Device, error) {
	return usb.ListDevices()
}

// GetDevice returns the block device with the given path, or an error if not found.
func (s *Service) GetDevice(path string) (*usb.Device, error) {
	devices, err := usb.ListDevices()
	if err != nil {
		return nil, err
	}
	for i := range devices {
		if devices[i].Path == path {
			return &devices[i], nil
		}
	}
	return nil, fmt.Errorf("device not found: %s", path)
}

// PlanFlash validates a flash operation and returns the plan without executing.
func (s *Service) PlanFlash(devicePath, imagePath string) (usb.FlashPlan, error) {
	devices, err := usb.ListDevices()
	if err != nil {
		return usb.FlashPlan{}, fmt.Errorf("list devices: %w", err)
	}
	for _, dev := range devices {
		if dev.Path == devicePath {
			return usb.PlanFlash(dev, imagePath, false)
		}
	}
	return usb.FlashPlan{}, fmt.Errorf("device not found: %s", devicePath)
}

// PlanPartitionFlash validates flashing an image to a specific partition.
func (s *Service) PlanPartitionFlash(devicePath, partitionPath, imagePath string) (usb.FlashPlan, error) {
	devices, err := usb.ListDevices()
	if err != nil {
		return usb.FlashPlan{}, fmt.Errorf("list devices: %w", err)
	}
	for _, dev := range devices {
		if dev.Path != devicePath {
			continue
		}
		for _, part := range dev.Partitions {
			if part.Path == partitionPath {
				return usb.PlanPartitionFlash(dev, part, imagePath, false)
			}
		}
		return usb.FlashPlan{}, fmt.Errorf("partition not found: %s", partitionPath)
	}
	return usb.FlashPlan{}, fmt.Errorf("device not found: %s", devicePath)
}

// ExecuteFlash runs a previously validated flash plan.
// Returns error if device is no longer safe at execution time.
func (s *Service) ExecuteFlash(plan usb.FlashPlan) error {
	devices, err := usb.ListDevices()
	if err != nil {
		return fmt.Errorf("list devices: %w", err)
	}
	for _, dev := range devices {
		if dev.Path == plan.DevicePath {
			if !dev.Safe {
				return fmt.Errorf("device is no longer safe: %s", dev.SafetyNote)
			}
			return usb.ExecuteFlash(plan)
		}
	}
	return fmt.Errorf("device not found: %s", plan.DevicePath)
}

// PlanMediaBuild parses the recipe YAML, looks up the device size, and returns a BuildPlan.
func (s *Service) PlanMediaBuild(recipeYAML string, devicePath string) (media.BuildPlan, error) {
	r, err := media.LoadRecipeFromBytes([]byte(recipeYAML))
	if err != nil {
		return media.BuildPlan{}, err
	}

	devices, err := usb.ListDevices()
	if err != nil {
		return media.BuildPlan{}, fmt.Errorf("list devices: %w", err)
	}

	var deviceSize int64
	for _, d := range devices {
		if d.Path == devicePath {
			deviceSize = d.Size
			break
		}
	}
	if deviceSize == 0 {
		return media.BuildPlan{}, fmt.Errorf("device not found: %s", devicePath)
	}

	return media.PlanBuild(r, devicePath, deviceSize, false)
}

// FormatMediaBuildPlan returns a human-readable summary of a BuildPlan.
func (s *Service) FormatMediaBuildPlan(plan media.BuildPlan) string {
	return media.FormatBuildPlan(plan)
}

// CheckPolicy validates a profile against a YAML policy string.
func (s *Service) CheckPolicy(policyYAML string, snap policy.ProfileSnapshot) (policy.CheckResult, error) {
	p, err := policy.LoadPolicyFromBytes([]byte(policyYAML))
	if err != nil {
		return policy.CheckResult{}, err
	}
	return policy.Check(p, snap), nil
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
