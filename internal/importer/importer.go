// Package importer converts existing installer configs into BootWrangler profiles.
package importer

import (
	"fmt"
	"strings"

	"github.com/alexmchughdev/bootwrangler/internal/profile"
)

// ImportResult holds a best-effort converted profile and any fields that
// could not be mapped.
type ImportResult struct {
	Profile           profile.Profile
	UnsupportedFields []string
	Warnings          []string
}

// Format is the detected installer config format.
type Format string

const (
	FormatAlpineAnswerfile  Format = "alpine-answerfile"
	FormatUbuntuAutoinstall Format = "ubuntu-autoinstall"
	FormatDebianPreseed     Format = "debian-preseed"
	FormatKickstart         Format = "kickstart"
	FormatAutoYaST          Format = "autoyast"
	FormatUnknown           Format = "unknown"
)

// DetectFormat guesses the installer format from raw content.
func DetectFormat(content string) Format {
	content = strings.TrimSpace(content)
	switch {
	case strings.HasPrefix(content, "<?xml") && strings.Contains(content, "<profile"):
		return FormatAutoYaST
	case strings.HasPrefix(content, "#cloud-config") || strings.Contains(content, "autoinstall:"):
		return FormatUbuntuAutoinstall
	case strings.Contains(content, "d-i ") || strings.Contains(content, "debconf"):
		return FormatDebianPreseed
	case strings.HasPrefix(content, "#kickstart") || strings.Contains(content, "%packages") || strings.Contains(content, "rootpw"):
		return FormatKickstart
	case strings.Contains(content, "KEYMAPOPTS") || strings.Contains(content, "HOSTNAMEOPTS"):
		return FormatAlpineAnswerfile
	default:
		return FormatUnknown
	}
}

// Import converts installer config bytes into a BootWrangler profile.
// The result is always best-effort; check UnsupportedFields and Warnings.
func Import(content string) (ImportResult, error) {
	format := DetectFormat(content)
	switch format {
	case FormatAlpineAnswerfile:
		return importAlpine(content)
	case FormatUbuntuAutoinstall:
		return importUbuntu(content)
	case FormatDebianPreseed:
		return importDebianPreseed(content)
	case FormatKickstart:
		return importKickstart(content)
	case FormatAutoYaST:
		return importAutoYaST(content)
	default:
		return ImportResult{}, fmt.Errorf("importer: unrecognised installer config format")
	}
}

func baseProfile(family, version string) profile.Profile {
	return profile.Profile{
		OS: profile.OS{
			Family:       family,
			Version:      version,
			Architecture: "x86_64",
		},
		System:  profile.System{Timezone: "UTC", Keyboard: "us", Locale: "en_US.UTF-8"},
		Network: profile.Network{Mode: "dhcp"},
		Disk:    profile.Disk{Mode: "wipe", Target: "auto", InstallMode: "server", Filesystem: "ext4"},
		SSH:     profile.SSH{Enabled: true, PermitRootLogin: false, PasswordAuthentication: false},
	}
}

func extractLine(content, prefix string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix))
		}
	}
	return ""
}

func importAlpine(content string) (ImportResult, error) {
	p := baseProfile("alpine", "3.19")
	result := ImportResult{Profile: p}

	if h := extractLine(content, "HOSTNAMEOPTS="); h != "" {
		p.System.Hostname = strings.Trim(h, "\"'-n ")
	}
	if tz := extractLine(content, "TIMEZONEOPTS="); tz != "" {
		p.System.Timezone = strings.Trim(tz, "\"'-z ")
	}
	if kb := extractLine(content, "KEYMAPOPTS="); kb != "" {
		p.System.Keyboard = strings.Fields(strings.Trim(kb, "\""))[0]
	}
	result.Profile = p
	result.Warnings = append(result.Warnings, "alpine answerfile import is best-effort; review all fields")
	result.UnsupportedFields = append(result.UnsupportedFields, "DNSOPTS", "NTPOPTS", "APKREPOSOPTS", "SSHDOPTS", "ROOTSSHKEY", "IPADDRESS")
	return result, nil
}

func importUbuntu(content string) (ImportResult, error) {
	p := baseProfile("ubuntu", "24.04")
	result := ImportResult{Profile: p}
	result.Warnings = append(result.Warnings, "ubuntu autoinstall import is best-effort; review all fields")
	result.UnsupportedFields = append(result.UnsupportedFields, "storage layout details", "network bonds/bridges", "curtin commands")

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "hostname:") {
			p.System.Hostname = strings.TrimSpace(strings.TrimPrefix(line, "hostname:"))
		}
		if strings.HasPrefix(line, "timezone:") {
			p.System.Timezone = strings.TrimSpace(strings.TrimPrefix(line, "timezone:"))
		}
	}
	result.Profile = p
	return result, nil
}

func importDebianPreseed(content string) (ImportResult, error) {
	p := baseProfile("debian", "12")
	result := ImportResult{Profile: p}
	result.Warnings = append(result.Warnings, "debian preseed import is best-effort; review all fields")
	result.UnsupportedFields = append(result.UnsupportedFields, "partman recipes", "late_command detail", "apt mirror config")

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "d-i netcfg/get_hostname") {
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				p.System.Hostname = parts[len(parts)-1]
			}
		}
		if strings.HasPrefix(line, "d-i time/zone") {
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				p.System.Timezone = parts[len(parts)-1]
			}
		}
	}
	result.Profile = p
	return result, nil
}

func importKickstart(content string) (ImportResult, error) {
	// Used for both Rocky and Fedora.
	p := baseProfile("rocky", "9")
	result := ImportResult{Profile: p}
	result.Warnings = append(result.Warnings, "kickstart import is best-effort; review all fields")
	result.UnsupportedFields = append(result.UnsupportedFields, "partitioning details", "pre/post scripts", "repo config")

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "network --hostname=") {
			p.System.Hostname = strings.TrimPrefix(line, "network --hostname=")
		}
		if strings.HasPrefix(line, "timezone ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				p.System.Timezone = parts[1]
			}
		}
		if strings.HasPrefix(line, "keyboard ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				p.System.Keyboard = strings.TrimPrefix(parts[1], "--vckeymap=")
			}
		}
	}
	result.Profile = p
	return result, nil
}

func importAutoYaST(content string) (ImportResult, error) {
	p := baseProfile("opensuse", "15.5")
	result := ImportResult{Profile: p}
	result.Warnings = append(result.Warnings, "autoyast import is best-effort; review all fields")
	result.UnsupportedFields = append(result.UnsupportedFields, "partitioning", "scripts", "add-on products")

	if idx := strings.Index(content, "<timezone>"); idx >= 0 {
		end := strings.Index(content[idx:], "</timezone>")
		if end >= 0 {
			p.System.Timezone = content[idx+len("<timezone>") : idx+end]
		}
	}
	if idx := strings.Index(content, "<hostname>"); idx >= 0 {
		end := strings.Index(content[idx:], "</hostname>")
		if end >= 0 {
			p.System.Hostname = content[idx+len("<hostname>") : idx+end]
		}
	}
	result.Profile = p
	return result, nil
}
