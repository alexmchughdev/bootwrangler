// Package windows renders Windows unattended-setup assets (autounattend.xml)
// including optional Active Directory domain enrollment.
package windows

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/alexmchughdev/bootwrangler/internal/manifest"
	"github.com/alexmchughdev/bootwrangler/internal/profile"
	"github.com/alexmchughdev/bootwrangler/internal/render"
)

// Renderer generates Windows unattended-setup assets.
type Renderer struct{}

// Name returns the unique renderer identifier.
func (r Renderer) Name() string { return "windows" }

// Family returns "windows".
func (r Renderer) Family() string { return "windows" }

// SupportedVersions returns nil, accepting any Windows version.
func (r Renderer) SupportedVersions() []string { return nil }

// Validate performs Windows-specific profile validation.
func (r Renderer) Validate(p profile.Profile) error {
	if p.OS.Family != "windows" {
		return fmt.Errorf("windows renderer: os family must be \"windows\", got %q", p.OS.Family)
	}
	if len(p.Users) == 0 {
		return fmt.Errorf("windows renderer: at least one user is required (it becomes the local administrator)")
	}
	return nil
}

// Render generates Windows installer assets into opts.OutDir.
func (r Renderer) Render(p profile.Profile, opts render.Options) (manifest.Manifest, error) {
	m := manifest.Manifest{
		ProfileName: p.Name,
		OSFamily:    "windows",
		OSVersion:   p.OS.Version,
		Renderer:    r.Name(),
	}

	data := buildData(p, &m)

	unattend, err := renderTemplate("autounattend.xml", autounattendTmpl, data)
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("windows: render autounattend.xml: %w", err)
	}
	firstLogon, err := renderTemplate("FirstLogon.ps1", firstLogonTmpl, data)
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("windows: render FirstLogon.ps1: %w", err)
	}

	m.AddFile("autounattend.xml", "Windows Setup answer file (place at the root of the install USB)")
	m.AddFile("FirstLogon.ps1", "First-logon script: domain join, packages, post-install (place in sources\\$OEM$\\$$\\Setup\\Scripts\\)")

	m.AddWarning("place autounattend.xml at the root of the bootable Windows install media; copy FirstLogon.ps1 into sources\\$OEM$\\$$\\Setup\\Scripts\\ so Setup deploys it to C:\\Windows\\Setup\\Scripts")
	m.AddWarning("Windows ISOs are not redistributable — supply your own volume/retail media")
	m.AddWarning("validate the answer file in the Lab before deploying to hardware")

	if opts.DryRun {
		m.AddWarning("dry-run: no files written")
		return m, nil
	}

	if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
		return manifest.Manifest{}, fmt.Errorf("windows: create output directory: %w", err)
	}
	if err := os.WriteFile(filepath.Join(opts.OutDir, "autounattend.xml"), []byte(unattend), 0o644); err != nil {
		return manifest.Manifest{}, fmt.Errorf("windows: write autounattend.xml: %w", err)
	}
	if err := os.WriteFile(filepath.Join(opts.OutDir, "FirstLogon.ps1"), []byte(firstLogon), 0o644); err != nil {
		return manifest.Manifest{}, fmt.Errorf("windows: write FirstLogon.ps1: %w", err)
	}
	if err := m.ComputeChecksums(opts.OutDir); err != nil {
		return manifest.Manifest{}, fmt.Errorf("windows: compute checksums: %w", err)
	}
	return m, nil
}

// templateData holds everything the two templates need.
type templateData struct {
	Arch          string // amd64 | arm64
	Locale        string // e.g. en-US
	ComputerName  string
	Edition       string
	ProductKey    string
	Organization  string
	Owner         string
	AdminUser     string
	AdminPassword string
	SkipOOBE      bool
	EnableSSH     bool
	Packages      []string
	Scripts       []profile.Script

	// domain join
	JoinMethod    string // "" | offline-djoin | unattend | first-logon
	Domain        string
	OU            string
	JoinUser      string
	ProvisionBlob string
}

func buildData(p profile.Profile, m *manifest.Manifest) templateData {
	arch := "amd64"
	switch p.OS.Architecture {
	case "aarch64", "arm64":
		arch = "arm64"
	}

	locale := p.System.Locale
	if locale == "" {
		locale = "en-US"
	}

	// Windows computer names are limited to 15 characters.
	computerName := p.System.Hostname
	if computerName == "" {
		computerName = "WIN-PC"
	}
	if len(computerName) > 15 {
		m.AddWarning(fmt.Sprintf("hostname %q exceeds the 15-character NetBIOS limit; truncated to %q", computerName, computerName[:15]))
		computerName = computerName[:15]
	}

	win := p.OSSpecific.Windows
	adminPassword := ""
	edition := "Windows 11 Pro"
	org, owner := "", ""
	productKey := ""
	skipOOBE := true
	if win != nil {
		if win.Edition != "" {
			edition = win.Edition
		}
		productKey = win.ProductKey
		org = win.Organization
		owner = win.Owner
		adminPassword = win.AdminPassword
		if win.SkipOOBE {
			skipOOBE = true
		}
	}
	if adminPassword == "" {
		adminPassword = "BootWrangler!Change1"
		m.AddWarning("local administrator password is a generated placeholder — set os_specific.windows.admin_password or rotate it immediately after deployment")
	} else {
		m.AddWarning("admin_password is written in plaintext into autounattend.xml — treat the rendered output as a secret and rotate after deployment")
	}

	data := templateData{
		Arch:          arch,
		Locale:        locale,
		ComputerName:  computerName,
		Edition:       edition,
		ProductKey:    productKey,
		Organization:  org,
		Owner:         owner,
		AdminUser:     p.Users[0].Name,
		AdminPassword: adminPassword,
		SkipOOBE:      skipOOBE,
		EnableSSH:     p.SSH.Enabled,
		Packages:      p.Packages.Names,
		Scripts:       p.PostInstall.Scripts,
	}

	if win != nil && win.DomainJoin != nil {
		dj := win.DomainJoin
		data.JoinMethod = dj.Method
		data.Domain = dj.Domain
		data.OU = dj.OU
		data.JoinUser = dj.JoinUser
		data.ProvisionBlob = dj.ProvisionBlob
		switch dj.Method {
		case "offline-djoin":
			m.AddWarning("offline domain join: the provisioning blob is applied during setup; no credentials are stored")
		case "unattend", "first-logon":
			m.AddWarning(fmt.Sprintf("%s domain join requires a password for %q at run time; it is never stored in the profile or answer file", dj.Method, dj.JoinUser))
		}
	}

	return data
}

func renderTemplate(name, tmpl string, data templateData) (string, error) {
	t, err := template.New(name).Funcs(template.FuncMap{"xml": xmlEscape}).Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// xmlEscape escapes a value for safe inclusion in XML text/attributes.
func xmlEscape(s string) string {
	var buf bytes.Buffer
	_ = xml.EscapeText(&buf, []byte(s))
	return buf.String()
}

func init() {
	render.Register(Renderer{})
}
