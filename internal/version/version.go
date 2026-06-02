// Package version reports the BootWrangler build version.
package version

// Value can be replaced at build time with -ldflags.
var Value = "dev"

// Current returns the BootWrangler build version.
func Current() string {
	return Value
}
