package ipxe

import "fmt"

// NetbootMode selects how assets are served to the iPXE client.
type NetbootMode string

const (
	// NetbootPublic uses the public netboot.xyz CDN with a custom chain.
	NetbootPublic NetbootMode = "public"
	// NetbootSelfHosted serves all assets from a local HTTP server.
	NetbootSelfHosted NetbootMode = "self-hosted"
)

// NetbootOptions configures the netboot.xyz integration.
type NetbootOptions struct {
	Mode        NetbootMode
	BaseURL     string // HTTP base URL for self-hosted mode
	NetbootURL  string // override the public netboot.xyz URL
	ProfileName string // used to name the generated menu file
}

const publicNetbootURL = "https://boot.netboot.xyz"

// NetbootAssetPath returns the relative URL path for a rendered profile asset.
// In self-hosted mode this is served by the local provisioning server.
func NetbootAssetPath(opts NetbootOptions, filename string) string {
	base := opts.BaseURL
	if base == "" {
		return filename
	}
	return fmt.Sprintf("%s/%s", base, filename)
}

// NetbootChainURL returns the URL the iPXE client should chain to.
func NetbootChainURL(opts NetbootOptions) string {
	switch opts.Mode {
	case NetbootSelfHosted:
		if opts.BaseURL == "" {
			return publicNetbootURL
		}
		return fmt.Sprintf("%s/custom.ipxe", opts.BaseURL)
	default:
		if opts.NetbootURL != "" {
			return opts.NetbootURL
		}
		return publicNetbootURL
	}
}
