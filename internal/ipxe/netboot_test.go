package ipxe

import (
	"strings"
	"testing"
)

func TestNetbootChainURL_Public(t *testing.T) {
	opts := NetbootOptions{Mode: NetbootPublic}
	url := NetbootChainURL(opts)
	if url != publicNetbootURL {
		t.Errorf("NetbootChainURL = %q, want %q", url, publicNetbootURL)
	}
}

func TestNetbootChainURL_CustomPublic(t *testing.T) {
	opts := NetbootOptions{Mode: NetbootPublic, NetbootURL: "http://myserver/netboot.xyz"}
	url := NetbootChainURL(opts)
	if url != "http://myserver/netboot.xyz" {
		t.Errorf("NetbootChainURL = %q", url)
	}
}

func TestNetbootChainURL_SelfHosted(t *testing.T) {
	opts := NetbootOptions{Mode: NetbootSelfHosted, BaseURL: "http://192.168.1.1:8088"}
	url := NetbootChainURL(opts)
	if !strings.HasPrefix(url, "http://192.168.1.1:8088") {
		t.Errorf("NetbootChainURL = %q", url)
	}
	if !strings.HasSuffix(url, "custom.ipxe") {
		t.Errorf("NetbootChainURL missing custom.ipxe: %q", url)
	}
}

func TestNetbootChainURL_SelfHostedNoBase(t *testing.T) {
	opts := NetbootOptions{Mode: NetbootSelfHosted, BaseURL: ""}
	url := NetbootChainURL(opts)
	if url != publicNetbootURL {
		t.Errorf("NetbootChainURL with no base = %q, want %q", url, publicNetbootURL)
	}
}

func TestNetbootAssetPath_WithBase(t *testing.T) {
	opts := NetbootOptions{BaseURL: "http://192.168.1.1:8088"}
	path := NetbootAssetPath(opts, "user-data")
	if path != "http://192.168.1.1:8088/user-data" {
		t.Errorf("NetbootAssetPath = %q", path)
	}
}

func TestNetbootAssetPath_NoBase(t *testing.T) {
	opts := NetbootOptions{}
	path := NetbootAssetPath(opts, "user-data")
	if path != "user-data" {
		t.Errorf("NetbootAssetPath = %q", path)
	}
}
