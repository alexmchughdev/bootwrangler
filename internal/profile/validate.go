package profile

import (
	"fmt"
	"net"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/alexmchughdev/bootwrangler/internal/secrets"
)

var (
	hostnameLabelPattern   = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)
	identifierPattern      = regexp.MustCompile(`^[a-z_][a-z0-9_-]*$`)
	windowsHostnamePattern = regexp.MustCompile(`^[A-Za-z0-9-]{1,15}$`)
	windowsNamePattern     = regexp.MustCompile(`^[A-Za-z0-9 ._-]+$`)
	packagePattern         = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9+_.:@/-]*$`)
	servicePattern         = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9@_.:-]*$`)
	versionPattern         = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)
)

var (
	supportedFamilies      = setOf("alpine", "ubuntu", "debian", "rocky", "fedora", "arch", "opensuse", "windows")
	supportedArchitectures = setOf(
		"x86_64",
		"aarch64",
		"arm64",
		"armv7",
		"riscv64",
	)
	supportedDiskModes        = setOf("wipe", "preserve", "manual")
	supportedFilesystems      = setOf("ext4", "xfs", "btrfs", "f2fs", "ntfs")
	supportedDomainJoinMethod = setOf("offline-djoin", "unattend", "first-logon")
	supportedNetworkModes     = setOf("dhcp", "static")
	supportedPackagePresets   = setOf("minimal", "remote-admin", "vm-guest", "container-host", "developer", "security-basic")
	supportedAlpineModes      = setOf("sys", "diskless", "data")
	supportedSELinuxModes     = setOf("enforcing", "permissive", "disabled")
	supportedFirewallModes    = setOf("enabled", "disabled")
	supportedArchAURHelpers   = setOf("none", "yay", "paru")
)

// ValidationError reports deterministic profile validation problems.
type ValidationError struct {
	Problems []string
}

// Error implements error.
func (e *ValidationError) Error() string {
	return "invalid profile: " + strings.Join(e.Problems, "; ")
}

// Validate validates common profile fields.
func Validate(value Profile) error {
	validator := newValidator()

	validator.required("name", value.Name)
	if value.Name != "" && !identifierPattern.MatchString(value.Name) {
		validator.add("name must contain only lowercase letters, numbers, underscores, or hyphens and start with a letter or underscore")
	}

	validateOS(validator, value.OS)
	validateSystem(validator, value.System, value.OS.Family)
	validateNetwork(validator, value.Network)
	validateDisk(validator, value.Disk)
	users := validateUsers(validator, value.Users, value.OS.Family)
	validateSSH(validator, value.SSH, users)
	validatePackages(validator, value.Packages)
	validateServices(validator, value.Services)
	validateChecks(validator, value.Validation, users, value.SSH)
	validatePostInstall(validator, value.PostInstall)
	validateOSSpecific(validator, value.OSSpecific)

	return validator.err()
}

type validator struct {
	problems []string
}

func newValidator() *validator {
	return &validator{}
}

func (v *validator) add(format string, args ...any) {
	v.problems = append(v.problems, fmt.Sprintf(format, args...))
}

func (v *validator) required(path, value string) {
	if strings.TrimSpace(value) == "" {
		v.add("%s is required", path)
	}
}

func (v *validator) err() error {
	if len(v.problems) == 0 {
		return nil
	}
	return &ValidationError{Problems: v.problems}
}

func validateOS(v *validator, value OS) {
	v.required("os.family", value.Family)
	if value.Family != "" && !supportedFamilies.has(value.Family) {
		v.add("os.family %q is unsupported", value.Family)
	}
	v.required("os.version", value.Version)
	if value.Version != "" && !versionPattern.MatchString(value.Version) {
		v.add("os.version %q is invalid", value.Version)
	}
	v.required("os.architecture", value.Architecture)
	if value.Architecture != "" && !supportedArchitectures.has(value.Architecture) {
		v.add("os.architecture %q is unsupported", value.Architecture)
	}
}

func validateSystem(v *validator, value System, family string) {
	v.required("system.hostname", value.Hostname)
	if value.Hostname != "" {
		if family == "windows" {
			if !windowsHostnamePattern.MatchString(value.Hostname) {
				v.add("system.hostname %q is invalid (Windows names are 1-15 letters, digits, or hyphens)", value.Hostname)
			}
		} else if !validHostname(value.Hostname) {
			v.add("system.hostname %q is invalid", value.Hostname)
		}
	}
	v.required("system.timezone", value.Timezone)
	if value.Timezone != "" {
		if _, err := time.LoadLocation(value.Timezone); err != nil {
			v.add("system.timezone %q is invalid", value.Timezone)
		}
	}
	v.required("system.keyboard", value.Keyboard)
	v.required("system.locale", value.Locale)
}

func validateNetwork(v *validator, value Network) {
	v.required("network.mode", value.Mode)
	if value.Mode != "" && !supportedNetworkModes.has(value.Mode) {
		v.add("network.mode %q is unsupported", value.Mode)
	}
	v.required("network.interface", value.Interface)

	if value.Mode == "static" {
		if value.Address == "" {
			v.add("network.address is required for static networking")
		} else if _, _, err := net.ParseCIDR(value.Address); err != nil {
			v.add("network.address %q must be CIDR notation", value.Address)
		}
		if value.Gateway == "" {
			v.add("network.gateway is required for static networking")
		} else if net.ParseIP(value.Gateway) == nil {
			v.add("network.gateway %q is invalid", value.Gateway)
		}
	}

	for index, address := range value.DNS {
		if net.ParseIP(address) == nil {
			v.add("network.dns[%d] %q is invalid", index, address)
		}
	}
}

func validateDisk(v *validator, value Disk) {
	v.required("disk.mode", value.Mode)
	if value.Mode != "" && !supportedDiskModes.has(value.Mode) {
		v.add("disk.mode %q is unsupported", value.Mode)
	}
	v.required("disk.target", value.Target)
	v.required("disk.install_mode", value.InstallMode)
	v.required("disk.filesystem", value.Filesystem)
	if value.Filesystem != "" && !supportedFilesystems.has(value.Filesystem) {
		v.add("disk.filesystem %q is unsupported", value.Filesystem)
	}
	if value.Mode == "wipe" && !value.ConfirmDestructive {
		v.add("disk.confirm_destructive must be true when disk.mode is wipe")
	}
}

type validatedUsers struct {
	names    stringSet
	keyCount int
}

func validateUsers(v *validator, values []User, family string) validatedUsers {
	namePattern := identifierPattern
	if family == "windows" {
		namePattern = windowsNamePattern
	}
	result := validatedUsers{names: make(stringSet)}
	if len(values) == 0 {
		v.add("users must contain at least one user")
		return result
	}

	for index, user := range values {
		path := fmt.Sprintf("users[%d]", index)
		v.required(path+".name", user.Name)
		if user.Name != "" && !namePattern.MatchString(user.Name) {
			v.add("%s.name %q is invalid", path, user.Name)
		}
		if result.names.has(user.Name) {
			v.add("%s.name %q is duplicated", path, user.Name)
		}
		result.names.add(user.Name)

		if user.Shell != "" && !strings.HasPrefix(user.Shell, "/") {
			v.add("%s.shell %q must be an absolute path", path, user.Shell)
		}
		validateNamedList(v, path+".groups", user.Groups, namePattern)
		for keyIndex, key := range user.SSHKeys {
			if err := secrets.ValidateSSHPublicKey(key); err != nil {
				v.add("%s.ssh_keys[%d] is not a valid SSH public key: %v", path, keyIndex, err)
				continue
			}
			result.keyCount++
		}
	}

	return result
}

func validateSSH(v *validator, value SSH, users validatedUsers) {
	if value.Enabled && users.keyCount == 0 {
		v.add("ssh.enabled requires at least one valid users[].ssh_keys entry")
	}
}

func validatePackages(v *validator, value Packages) {
	validateKnownList(v, "packages.presets", value.Presets, supportedPackagePresets)
	validateNamedList(v, "packages.names", value.Names, packagePattern)
}

func validateServices(v *validator, value Services) {
	validateNamedList(v, "services.enable", value.Enable, servicePattern)
}

func validateChecks(v *validator, value Validation, users validatedUsers, ssh SSH) {
	if value.SSH.Enabled {
		if !ssh.Enabled {
			v.add("validation.ssh.enabled requires ssh.enabled")
		}
		v.required("validation.ssh.user", value.SSH.User)
		if value.SSH.User != "" && !users.names.has(value.SSH.User) {
			v.add("validation.ssh.user %q does not match a configured user", value.SSH.User)
		}
	}

	for index, command := range value.Commands {
		path := fmt.Sprintf("validation.commands[%d]", index)
		v.required(path+".name", command.Name)
		v.required(path+".command", command.Command)
		if command.ExpectExitCode < 0 || command.ExpectExitCode > 255 {
			v.add("%s.expect_exit_code must be between 0 and 255", path)
		}
	}
}

func validatePostInstall(v *validator, value PostInstall) {
	names := make(stringSet)
	for index, script := range value.Scripts {
		path := fmt.Sprintf("post_install.scripts[%d]", index)
		v.required(path+".name", script.Name)
		if script.Name != "" && (filepath.Base(script.Name) != script.Name || script.Name == "." || script.Name == "..") {
			v.add("%s.name %q must be a file name without a path", path, script.Name)
		}
		if names.has(script.Name) {
			v.add("%s.name %q is duplicated", path, script.Name)
		}
		names.add(script.Name)
		v.required(path+".content", script.Content)
	}
}

func validateOSSpecific(v *validator, value OSSpecific) {
	if value.Alpine != nil {
		if value.Alpine.InstallMode != "" && !supportedAlpineModes.has(value.Alpine.InstallMode) {
			v.add("os_specific.alpine.install_mode %q is unsupported", value.Alpine.InstallMode)
		}
		validateNamedList(v, "os_specific.alpine.repositories", value.Alpine.Repositories, packagePattern)
	}
	if value.Ubuntu != nil && value.Ubuntu.AutoinstallVersion < 0 {
		v.add("os_specific.ubuntu.autoinstall_version must not be negative")
	}
	if value.RHELLike != nil {
		if value.RHELLike.SELinux != "" && !supportedSELinuxModes.has(value.RHELLike.SELinux) {
			v.add("os_specific.rhel_like.selinux %q is unsupported", value.RHELLike.SELinux)
		}
		if value.RHELLike.Firewall != "" && !supportedFirewallModes.has(value.RHELLike.Firewall) {
			v.add("os_specific.rhel_like.firewall %q is unsupported", value.RHELLike.Firewall)
		}
	}
	if value.Arch != nil && value.Arch.AURHelper != "" && !supportedArchAURHelpers.has(value.Arch.AURHelper) {
		v.add("os_specific.arch.aur_helper %q is unsupported", value.Arch.AURHelper)
	}
	if value.Windows != nil && value.Windows.DomainJoin != nil {
		dj := value.Windows.DomainJoin
		v.required("os_specific.windows.domain_join.domain", dj.Domain)
		v.required("os_specific.windows.domain_join.method", dj.Method)
		if dj.Method != "" && !supportedDomainJoinMethod.has(dj.Method) {
			v.add("os_specific.windows.domain_join.method %q is unsupported", dj.Method)
		}
		if dj.Method == "offline-djoin" && dj.ProvisionBlob == "" {
			v.add("os_specific.windows.domain_join.provision_blob is required for offline-djoin")
		}
		if (dj.Method == "unattend" || dj.Method == "first-logon") && dj.JoinUser == "" {
			v.add("os_specific.windows.domain_join.join_user is required for the %q method", dj.Method)
		}
	}
}

func validateKnownList(v *validator, path string, values []string, supported stringSet) {
	seen := make(stringSet)
	for index, value := range values {
		if value == "" {
			v.add("%s[%d] is empty", path, index)
		} else if !supported.has(value) {
			v.add("%s[%d] %q is unsupported", path, index, value)
		}
		if seen.has(value) {
			v.add("%s[%d] %q is duplicated", path, index, value)
		}
		seen.add(value)
	}
}

func validateNamedList(v *validator, path string, values []string, pattern *regexp.Regexp) {
	seen := make(stringSet)
	for index, value := range values {
		if !pattern.MatchString(value) {
			v.add("%s[%d] %q is invalid", path, index, value)
		}
		if seen.has(value) {
			v.add("%s[%d] %q is duplicated", path, index, value)
		}
		seen.add(value)
	}
}

func validHostname(value string) bool {
	if len(value) > 253 {
		return false
	}
	for _, label := range strings.Split(value, ".") {
		if !hostnameLabelPattern.MatchString(label) {
			return false
		}
	}
	return true
}

type stringSet map[string]struct{}

func setOf(values ...string) stringSet {
	result := make(stringSet, len(values))
	for _, value := range values {
		result.add(value)
	}
	return result
}

func (s stringSet) add(value string) {
	s[value] = struct{}{}
}

func (s stringSet) has(value string) bool {
	_, ok := s[value]
	return ok
}
