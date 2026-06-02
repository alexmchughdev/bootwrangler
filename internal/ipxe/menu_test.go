package ipxe_test

import (
	"strings"
	"testing"

	"github.com/alexmchughdev/bootwrangler/internal/ipxe"
)

func twoEntries() []ipxe.Entry {
	return []ipxe.Entry{
		{Label: "Alpine Auto Install", KernelLine: "kernel http://server/vmlinuz alpine_dev=eth0"},
		{Label: "Ubuntu Server 24.04", KernelLine: "kernel http://server/vmlinuz-ubuntu\ninitrd http://server/initrd.img"},
	}
}

// Test 1: RenderMenu with two entries — both labels and menu title appear.
func TestRenderMenu_TwoEntries(t *testing.T) {
	opts := ipxe.MenuOptions{
		Title:   "BootWrangler",
		Entries: twoEntries(),
	}
	out, err := ipxe.RenderMenu(opts)
	if err != nil {
		t.Fatalf("RenderMenu: unexpected error: %v", err)
	}
	for _, want := range []string{
		"BootWrangler Boot Menu",
		"Alpine Auto Install",
		"Ubuntu Server 24.04",
		":profile-0",
		":profile-1",
		"#!ipxe",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n\nFull output:\n%s", want, out)
		}
	}
}

// Test 2: RenderMenu with AddRescue: true — :shell block appears.
func TestRenderMenu_AddRescue(t *testing.T) {
	opts := ipxe.MenuOptions{
		Entries:   twoEntries(),
		AddRescue: true,
	}
	out, err := ipxe.RenderMenu(opts)
	if err != nil {
		t.Fatalf("RenderMenu: unexpected error: %v", err)
	}
	if !strings.Contains(out, ":shell") {
		t.Errorf("expected :shell block in output\n\nFull output:\n%s", out)
	}
	if !strings.Contains(out, "item shell") {
		t.Errorf("expected 'item shell' in menu section\n\nFull output:\n%s", out)
	}
}

// Test 3: RenderMenu with NetbootURL set — chain URL appears.
func TestRenderMenu_NetbootURL(t *testing.T) {
	const netbootURL = "https://boot.netboot.xyz"
	opts := ipxe.MenuOptions{
		Entries:    twoEntries(),
		NetbootURL: netbootURL,
	}
	out, err := ipxe.RenderMenu(opts)
	if err != nil {
		t.Fatalf("RenderMenu: unexpected error: %v", err)
	}
	if !strings.Contains(out, "chain "+netbootURL) {
		t.Errorf("expected 'chain %s' in output\n\nFull output:\n%s", netbootURL, out)
	}
	if !strings.Contains(out, ":netboot") {
		t.Errorf("expected :netboot block in output\n\nFull output:\n%s", out)
	}
}

// Test 4: RenderMenu with NetbootURL empty — no chain line.
func TestRenderMenu_NoNetbootURL(t *testing.T) {
	opts := ipxe.MenuOptions{
		Entries:    twoEntries(),
		NetbootURL: "",
	}
	out, err := ipxe.RenderMenu(opts)
	if err != nil {
		t.Fatalf("RenderMenu: unexpected error: %v", err)
	}
	if strings.Contains(out, "chain ") {
		t.Errorf("expected no 'chain' line when NetbootURL is empty\n\nFull output:\n%s", out)
	}
	if strings.Contains(out, ":netboot") {
		t.Errorf("expected no :netboot block when NetbootURL is empty\n\nFull output:\n%s", out)
	}
}

// Test 5: RenderSingleEntry — kernel line appears.
func TestRenderSingleEntry(t *testing.T) {
	entry := ipxe.Entry{
		Label:      "Alpine Auto Install",
		KernelLine: "kernel http://server/vmlinuz alpine_dev=eth0",
	}
	out, err := ipxe.RenderSingleEntry(entry, "http://server")
	if err != nil {
		t.Fatalf("RenderSingleEntry: unexpected error: %v", err)
	}
	if !strings.Contains(out, "kernel http://server/vmlinuz alpine_dev=eth0") {
		t.Errorf("expected kernel line in output\n\nFull output:\n%s", out)
	}
	if !strings.Contains(out, "#!ipxe") {
		t.Errorf("expected iPXE shebang in output\n\nFull output:\n%s", out)
	}
	if !strings.Contains(out, "boot") {
		t.Errorf("expected 'boot' directive in output\n\nFull output:\n%s", out)
	}
}

// Test 6: Output is deterministic — two calls with same args produce identical output.
func TestRenderMenu_Deterministic(t *testing.T) {
	opts := ipxe.MenuOptions{
		Title:      "BootWrangler",
		Entries:    twoEntries(),
		AddRescue:  true,
		NetbootURL: "https://boot.netboot.xyz",
	}
	out1, err := ipxe.RenderMenu(opts)
	if err != nil {
		t.Fatalf("first RenderMenu: %v", err)
	}
	out2, err := ipxe.RenderMenu(opts)
	if err != nil {
		t.Fatalf("second RenderMenu: %v", err)
	}
	if out1 != out2 {
		t.Errorf("output is non-deterministic:\nfirst:\n%s\n\nsecond:\n%s", out1, out2)
	}
}
