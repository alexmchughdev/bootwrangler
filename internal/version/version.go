// Package version reports the BootForge build version.
package version

// Value can be replaced at build time with -ldflags.
var Value = "dev"

// Current returns the BootForge build version.
func Current() string {
	return Value
}
