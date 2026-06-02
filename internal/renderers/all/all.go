// Package all imports all renderer implementations so their init() functions
// register them with the render registry. Import this package in main and in
// any binary that needs the full renderer set.
package all

import (
	_ "github.com/alexmchughdev/bootwrangler/internal/renderers/alpine"
	_ "github.com/alexmchughdev/bootwrangler/internal/renderers/arch"
	_ "github.com/alexmchughdev/bootwrangler/internal/renderers/debian"
	_ "github.com/alexmchughdev/bootwrangler/internal/renderers/fedora"
	_ "github.com/alexmchughdev/bootwrangler/internal/renderers/opensuse"
	_ "github.com/alexmchughdev/bootwrangler/internal/renderers/rocky"
	_ "github.com/alexmchughdev/bootwrangler/internal/renderers/ubuntu"
)
