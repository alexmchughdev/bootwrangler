// Package opensuse renders openSUSE AutoYaST installer assets.
package opensuse

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexmchughdev/bootwrangler/internal/manifest"
	"github.com/alexmchughdev/bootwrangler/internal/profile"
	"github.com/alexmchughdev/bootwrangler/internal/render"
)

// Renderer generates openSUSE AutoYaST assets.
type Renderer struct{}

// Name returns the unique renderer identifier.
func (r Renderer) Name() string { return "opensuse" }

// Family returns "opensuse".
func (r Renderer) Family() string { return "opensuse" }

// SupportedVersions returns nil, accepting any openSUSE version.
func (r Renderer) SupportedVersions() []string { return nil }

// Validate performs openSUSE-specific profile validation.
func (r Renderer) Validate(p profile.Profile) error {
	if p.OS.Family != "opensuse" {
		return fmt.Errorf("opensuse renderer: os family must be \"opensuse\", got %q", p.OS.Family)
	}
	if len(p.Users) == 0 {
		return fmt.Errorf("opensuse renderer: at least one user must be defined")
	}
	return nil
}

// Render generates openSUSE AutoYaST assets into opts.OutDir.
func (r Renderer) Render(p profile.Profile, opts render.Options) (manifest.Manifest, error) {
	m := manifest.Manifest{
		ProfileName: p.Name,
		OSFamily:    "opensuse",
		OSVersion:   p.OS.Version,
		Renderer:    r.Name(),
	}

	autoinstContent, err := renderAutoYaST(p)
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("opensuse: render autoinst.xml: %w", err)
	}

	ipxeContent, err := renderIPXE(opts.ServerBaseURL)
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("opensuse: render iPXE: %w", err)
	}

	m.AddFile("autoinst.xml", "AutoYaST XML profile for unattended installation")
	m.AddFile("custom.ipxe", "iPXE boot entry for network booting")

	if opts.DryRun {
		m.AddWarning("dry-run: no files written")
		return m, nil
	}

	if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
		return manifest.Manifest{}, fmt.Errorf("opensuse: create output directory: %w", err)
	}

	if err := os.WriteFile(filepath.Join(opts.OutDir, "autoinst.xml"), []byte(autoinstContent), 0o644); err != nil {
		return manifest.Manifest{}, fmt.Errorf("opensuse: write autoinst.xml: %w", err)
	}

	if err := os.WriteFile(filepath.Join(opts.OutDir, "custom.ipxe"), []byte(ipxeContent), 0o644); err != nil {
		return manifest.Manifest{}, fmt.Errorf("opensuse: write custom.ipxe: %w", err)
	}

	if err := m.ComputeChecksums(opts.OutDir); err != nil {
		return manifest.Manifest{}, fmt.Errorf("opensuse: compute checksums: %w", err)
	}

	return m, nil
}

// ---------------------------------------------------------------------------
// AutoYaST XML structs
// ---------------------------------------------------------------------------

// xmlBool wraps a boolean for AutoYaST config:type="boolean" output.
type xmlBool struct {
	XMLName xml.Name
	Type    string `xml:"config:type,attr"`
	Text    string `xml:",chardata"`
}

func newXMLBool(localName string, v bool) xmlBool {
	text := "false"
	if v {
		text = "true"
	}
	return xmlBool{
		XMLName: xml.Name{Local: localName},
		Type:    "boolean",
		Text:    text,
	}
}

// xmlSymbol wraps a value for AutoYaST config:type="symbol" output.
type xmlSymbol struct {
	XMLName xml.Name
	Type    string `xml:"config:type,attr"`
	Text    string `xml:",chardata"`
}

func newXMLSymbol(localName, value string) xmlSymbol {
	return xmlSymbol{
		XMLName: xml.Name{Local: localName},
		Type:    "symbol",
		Text:    value,
	}
}

// ---------------------------------------------------------------------------
// Top-level AutoYaST profile structs
// ---------------------------------------------------------------------------

type autoYaSTProfile struct {
	XMLName      xml.Name          `xml:"profile"`
	Xmlns        string            `xml:"xmlns,attr"`
	XmlnsConf    string            `xml:"xmlns:config,attr"`
	Language     ayLanguage        `xml:"language"`
	Keyboard     ayKeyboard        `xml:"keyboard"`
	Timezone     ayTimezone        `xml:"timezone"`
	Networking   ayNetworking      `xml:"networking"`
	Partitioning ayPartitioning    `xml:"partitioning"`
	Users        ayUsers           `xml:"users"`
	Software     aySoftware        `xml:"software"`
	Services     ayServicesManager `xml:"services-manager"`
	Scripts      ayScripts         `xml:"scripts"`
}

type ayLanguage struct {
	Language  string     `xml:"language"`
	Languages ayLangList `xml:"languages"`
}

type ayLangList struct {
	Type     string   `xml:"config:type,attr"`
	Language []string `xml:"language"`
}

type ayKeyboard struct {
	Keymap string `xml:"keymap"`
}

type ayTimezone struct {
	HWClock  string `xml:"hwclock"`
	Timezone string `xml:"timezone"`
}

type ayNetworking struct {
	SetupBeforeProposal xmlBool      `xml:"setup_before_proposal"`
	Managed             xmlBool      `xml:"managed"`
	Interfaces          ayInterfaces `xml:"interfaces"`
}

type ayInterfaces struct {
	Type      string        `xml:"config:type,attr"`
	Interface []ayInterface `xml:"interface"`
}

type ayInterface struct {
	Bootproto string `xml:"bootproto"`
	Device    string `xml:"device"`
	Startmode string `xml:"startmode"`
}

type ayPartitioning struct {
	Type  string    `xml:"config:type,attr"`
	Drive []ayDrive `xml:"drive"`
}

type ayDrive struct {
	Device     string       `xml:"device"`
	Use        string       `xml:"use"`
	Partitions ayPartitions `xml:"partitions"`
}

type ayPartitions struct {
	Type      string        `xml:"config:type,attr"`
	Partition []ayPartition `xml:"partition"`
}

type ayPartition struct {
	Create     xmlBool   `xml:"create"`
	Filesystem xmlSymbol `xml:"filesystem"`
	Mount      string    `xml:"mount"`
	Size       string    `xml:"size"`
}

type ayUsers struct {
	Type string   `xml:"config:type,attr"`
	User []ayUser `xml:"user"`
}

type ayUser struct {
	Username  string  `xml:"username"`
	Password  string  `xml:"user_password"`
	Encrypted xmlBool `xml:"encrypted"`
	Forename  string  `xml:"forename"`
	Surname   string  `xml:"surname"`
}

type aySoftware struct {
	Packages ayPackageList `xml:"packages"`
	Patterns ayPatternList `xml:"patterns"`
}

type ayPackageList struct {
	Type    string   `xml:"config:type,attr"`
	Package []string `xml:"package"`
}

type ayPatternList struct {
	Type    string   `xml:"config:type,attr"`
	Pattern []string `xml:"pattern"`
}

type ayServicesManager struct {
	Services ayServicesList `xml:"services"`
}

type ayServicesList struct {
	Enable ayEnableList `xml:"enable"`
}

type ayEnableList struct {
	Type    string   `xml:"config:type,attr"`
	Service []string `xml:"service"`
}

type ayScripts struct {
	PostScripts ayPostScripts `xml:"post-scripts"`
}

type ayPostScripts struct {
	Type   string     `xml:"config:type,attr"`
	Script []ayScript `xml:"script"`
}

type ayScript struct {
	Source      string  `xml:"source"`
	Filename    string  `xml:"filename"`
	Interpreter string  `xml:"interpreter"`
	Feedback    xmlBool `xml:"feedback"`
}

// ---------------------------------------------------------------------------
// Rendering helpers
// ---------------------------------------------------------------------------

func renderAutoYaST(p profile.Profile) (string, error) {
	locale := p.System.Locale
	if len(locale) >= 5 {
		locale = locale[:5]
	}
	if locale == "" {
		locale = "en_US"
	}

	keyboard := p.System.Keyboard
	if keyboard == "" {
		keyboard = "english-us"
	}

	iface := p.Network.Interface
	if iface == "" || iface == "auto" {
		iface = "eth0"
	}

	bootproto := p.Network.Mode
	if bootproto == "" {
		bootproto = "dhcp"
	}

	diskTarget := p.Disk.Target
	if diskTarget == "" || diskTarget == "auto" {
		diskTarget = "/dev/sda"
	}

	filesystem := p.Disk.Filesystem
	if filesystem == "" {
		filesystem = "ext4"
	}

	var users []ayUser
	for _, u := range p.Users {
		users = append(users, ayUser{
			Username:  u.Name,
			Password:  "!",
			Encrypted: newXMLBool("encrypted", true),
			Forename:  "",
			Surname:   "",
		})
	}

	postScriptSource := buildPostScript(p)

	prof := autoYaSTProfile{
		Xmlns:     "http://www.suse.com/1.0/yast2ns",
		XmlnsConf: "http://www.suse.com/1.0/configns",
		Language: ayLanguage{
			Language: locale,
			Languages: ayLangList{
				Type:     "list",
				Language: []string{locale},
			},
		},
		Keyboard: ayKeyboard{
			Keymap: keyboard,
		},
		Timezone: ayTimezone{
			HWClock:  "UTC",
			Timezone: p.System.Timezone,
		},
		Networking: ayNetworking{
			SetupBeforeProposal: newXMLBool("setup_before_proposal", true),
			Managed:             newXMLBool("managed", false),
			Interfaces: ayInterfaces{
				Type: "list",
				Interface: []ayInterface{
					{
						Bootproto: bootproto,
						Device:    iface,
						Startmode: "auto",
					},
				},
			},
		},
		Partitioning: ayPartitioning{
			Type: "list",
			Drive: []ayDrive{
				{
					Device: diskTarget,
					Use:    "all",
					Partitions: ayPartitions{
						Type: "list",
						Partition: []ayPartition{
							{
								Create:     newXMLBool("create", true),
								Filesystem: newXMLSymbol("filesystem", filesystem),
								Mount:      "/",
								Size:       "max",
							},
						},
					},
				},
			},
		},
		Users: ayUsers{
			Type: "list",
			User: users,
		},
		Software: aySoftware{
			Packages: ayPackageList{
				Type:    "list",
				Package: p.Packages.Names,
			},
			Patterns: ayPatternList{
				Type:    "list",
				Pattern: []string{"enhanced_base"},
			},
		},
		Services: ayServicesManager{
			Services: ayServicesList{
				Enable: ayEnableList{
					Type:    "list",
					Service: p.Services.Enable,
				},
			},
		},
		Scripts: ayScripts{
			PostScripts: ayPostScripts{
				Type: "list",
				Script: []ayScript{
					{
						Source:      postScriptSource,
						Filename:    "post-install.sh",
						Interpreter: "shell",
						Feedback:    newXMLBool("feedback", false),
					},
				},
			},
		},
	}

	out, err := xml.MarshalIndent(prof, "", "  ")
	if err != nil {
		return "", err
	}

	header := `<?xml version="1.0"?>` + "\n" + `<!DOCTYPE profile>` + "\n"
	return header + string(out) + "\n", nil
}

// buildPostScript generates the post-install bash script content.
func buildPostScript(p profile.Profile) string {
	var sb strings.Builder

	sb.WriteString("#!/bin/bash\nset -eu\n")

	// SSH authorized_keys setup for each user.
	for _, u := range p.Users {
		if len(u.SSHKeys) == 0 {
			continue
		}
		home := "/home/" + u.Name
		sb.WriteString(fmt.Sprintf("\n# SSH keys for %s\n", u.Name))
		sb.WriteString(fmt.Sprintf("install -d -m 0700 %s/.ssh\n", home))
		sb.WriteString(fmt.Sprintf("cat > %s/.ssh/authorized_keys << 'SSHEOF'\n", home))
		for _, key := range u.SSHKeys {
			sb.WriteString(key + "\n")
		}
		sb.WriteString("SSHEOF\n")
		sb.WriteString(fmt.Sprintf("chmod 0600 %s/.ssh/authorized_keys\n", home))
		sb.WriteString(fmt.Sprintf("chown -R %s:%s %s/.ssh\n", u.Name, u.Name, home))
	}

	// sshd_config hardening.
	sb.WriteString("\n# Configure sshd\n")
	if !p.SSH.PermitRootLogin {
		sb.WriteString("sed -i 's|^#\\?PermitRootLogin.*|PermitRootLogin no|' /etc/ssh/sshd_config\n")
	}
	if !p.SSH.PasswordAuthentication {
		sb.WriteString("sed -i 's|^#\\?PasswordAuthentication.*|PasswordAuthentication no|' /etc/ssh/sshd_config\n")
	}

	// Additional profile scripts.
	for _, s := range p.PostInstall.Scripts {
		sb.WriteString(fmt.Sprintf("\n# %s\n", s.Name))
		sb.WriteString(s.Content)
		if !strings.HasSuffix(s.Content, "\n") {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// ---------------------------------------------------------------------------
// iPXE
// ---------------------------------------------------------------------------

const ipxeTmplWithURL = `#!ipxe
set base-url %s
kernel ${base-url}/boot/x86_64/loader/linux autoyast=${base-url}/autoinst.xml install=${base-url} quiet
initrd ${base-url}/boot/x86_64/loader/initrd
boot
`

const ipxeTmplNoURL = `#!ipxe
# configure ServerBaseURL in BootWrangler settings
kernel /boot/x86_64/loader/linux autoyast=/autoinst.xml quiet
initrd /boot/x86_64/loader/initrd
boot
`

func renderIPXE(serverBaseURL string) (string, error) {
	if serverBaseURL == "" {
		return ipxeTmplNoURL, nil
	}
	return fmt.Sprintf(ipxeTmplWithURL, serverBaseURL), nil
}

func init() {
	render.Register(Renderer{})
}
