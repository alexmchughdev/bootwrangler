// Package app provides the backend methods exposed to the BootWrangler desktop
// application.
package app

import (
	"context"

	"github.com/alexmchughdev/bootwrangler/internal/version"
)

// HealthStatus describes desktop backend availability.
type HealthStatus struct {
	Status  string `json:"status"`
	Version string `json:"version"`
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
