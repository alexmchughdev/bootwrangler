package bootmenu

import (
	"strings"
	"testing"
)

func TestBuild_Empty(t *testing.T) {
	m := Build(MenuOptions{Title: "Test"}, nil)
	if m.Title != "Test" {
		t.Errorf("Title = %q", m.Title)
	}
	if len(m.Entries) != 0 {
		t.Errorf("Entries = %d", len(m.Entries))
	}
}

func TestBuild_WithNetbootAndShell(t *testing.T) {
	opts := MenuOptions{
		Title:      "BootWrangler",
		AddNetboot: true,
		AddShell:   true,
		AddReboot:  true,
	}
	entries := []Entry{
		{Label: "Ubuntu Server", Kind: EntryLocalProfile, URL: "http://server/ubuntu.ipxe"},
	}
	m := Build(opts, entries)

	if len(m.Entries) != 4 { // netboot + ubuntu + shell + reboot
		t.Errorf("expected 4 entries, got %d", len(m.Entries))
	}
	if m.Entries[0].Kind != EntryNetboot {
		t.Errorf("first entry should be netboot, got %v", m.Entries[0].Kind)
	}
	if m.Entries[1].Label != "Ubuntu Server" {
		t.Errorf("second entry label = %q", m.Entries[1].Label)
	}
	if m.Entries[2].Kind != EntryShell {
		t.Errorf("third entry should be shell, got %v", m.Entries[2].Kind)
	}
	if m.Entries[3].Kind != EntryReboot {
		t.Errorf("fourth entry should be reboot, got %v", m.Entries[3].Kind)
	}
}

func TestBuild_DefaultNetbootURL(t *testing.T) {
	m := Build(MenuOptions{AddNetboot: true}, nil)
	if m.Entries[0].URL != DefaultNetbootURL {
		t.Errorf("default netboot URL = %q, want %q", m.Entries[0].URL, DefaultNetbootURL)
	}
}

func TestBuild_CustomNetbootURL(t *testing.T) {
	m := Build(MenuOptions{AddNetboot: true, NetbootURL: "http://myserver/netboot.xyz"}, nil)
	if m.Entries[0].URL != "http://myserver/netboot.xyz" {
		t.Errorf("custom netboot URL = %q", m.Entries[0].URL)
	}
}

func TestRenderIPXE_Basic(t *testing.T) {
	m := Build(MenuOptions{
		Title:      "Test Menu",
		AddNetboot: true,
		AddShell:   true,
	}, []Entry{
		{Label: "Alpine Auto", Kind: EntryLocalProfile, URL: "http://server/alpine.ipxe"},
	})

	out, err := RenderIPXE(m)
	if err != nil {
		t.Fatalf("RenderIPXE: %v", err)
	}
	if !strings.Contains(out, "#!ipxe") {
		t.Error("missing #!ipxe header")
	}
	if !strings.Contains(out, "Test Menu") {
		t.Error("missing menu title")
	}
	if !strings.Contains(out, "netboot.xyz") {
		t.Error("missing netboot entry")
	}
	if !strings.Contains(out, "Alpine Auto") {
		t.Error("missing profile entry")
	}
	if !strings.Contains(out, "Shell") {
		t.Error("missing shell entry")
	}
}

func TestRenderIPXE_ProfileEntry(t *testing.T) {
	m := Build(MenuOptions{}, []Entry{
		{Label: "Ubuntu", Kind: EntryLocalProfile, Kernel: "/vmlinuz", Initrd: "/initrd.img", Cmdline: "quiet"},
	})
	out, err := RenderIPXE(m)
	if err != nil {
		t.Fatalf("RenderIPXE: %v", err)
	}
	if !strings.Contains(out, "kernel /vmlinuz") {
		t.Error("missing kernel line")
	}
	if !strings.Contains(out, "initrd /initrd.img") {
		t.Error("missing initrd line")
	}
}

func TestRenderIPXE_Deterministic(t *testing.T) {
	m := Build(MenuOptions{Title: "Menu", AddShell: true}, []Entry{
		{Label: "Entry A", Kind: EntryLocalImage, URL: "http://a"},
		{Label: "Entry B", Kind: EntryLocalImage, URL: "http://b"},
	})
	out1, _ := RenderIPXE(m)
	out2, _ := RenderIPXE(m)
	if out1 != out2 {
		t.Error("RenderIPXE output is not deterministic")
	}
}

func TestRenderGRUB_Basic(t *testing.T) {
	m := Build(MenuOptions{Title: "Test"}, []Entry{
		{Label: "Ubuntu", Kind: EntryLocalProfile, Kernel: "/vmlinuz", Initrd: "/initrd.img"},
	})
	out, err := RenderGRUB(m)
	if err != nil {
		t.Fatalf("RenderGRUB: %v", err)
	}
	if !strings.Contains(out, "menuentry") {
		t.Error("missing menuentry")
	}
	if !strings.Contains(out, "Ubuntu") {
		t.Error("missing entry label")
	}
}
