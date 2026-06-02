// Package profile defines and validates BootWrangler provisioning profiles.
package profile

// Profile describes the desired installed operating-system configuration.
type Profile struct {
	Name        string      `yaml:"name"`
	OS          OS          `yaml:"os"`
	System      System      `yaml:"system"`
	Network     Network     `yaml:"network"`
	Disk        Disk        `yaml:"disk"`
	Users       []User      `yaml:"users"`
	SSH         SSH         `yaml:"ssh"`
	Packages    Packages    `yaml:"packages,omitempty"`
	Services    Services    `yaml:"services,omitempty"`
	Validation  Validation  `yaml:"validation,omitempty"`
	PostInstall PostInstall `yaml:"post_install,omitempty"`
	OSSpecific  OSSpecific  `yaml:"os_specific,omitempty"`
}

// OS identifies the target operating system.
type OS struct {
	Family       string `yaml:"family"`
	Version      string `yaml:"version"`
	Architecture string `yaml:"architecture"`
}

// System describes common installed-system settings.
type System struct {
	Hostname string `yaml:"hostname"`
	Timezone string `yaml:"timezone"`
	Keyboard string `yaml:"keyboard"`
	Locale   string `yaml:"locale"`
}

// Network describes the initial network configuration.
type Network struct {
	Mode      string   `yaml:"mode"`
	Interface string   `yaml:"interface"`
	Address   string   `yaml:"address,omitempty"`
	Gateway   string   `yaml:"gateway,omitempty"`
	DNS       []string `yaml:"dns,omitempty"`
}

// Disk describes installer disk behavior. Runtime disk operations apply
// additional device-safety checks before destructive actions.
type Disk struct {
	Mode               string `yaml:"mode"`
	Target             string `yaml:"target"`
	InstallMode        string `yaml:"install_mode"`
	Filesystem         string `yaml:"filesystem"`
	ConfirmDestructive bool   `yaml:"confirm_destructive"`
}

// User describes one installed-system account.
type User struct {
	Name    string   `yaml:"name"`
	Shell   string   `yaml:"shell,omitempty"`
	Groups  []string `yaml:"groups,omitempty"`
	Sudo    bool     `yaml:"sudo,omitempty"`
	SSHKeys []string `yaml:"ssh_keys,omitempty"`
}

// SSH describes installed-system SSH daemon settings.
type SSH struct {
	Enabled                bool `yaml:"enabled"`
	PermitRootLogin        bool `yaml:"permit_root_login"`
	PasswordAuthentication bool `yaml:"password_authentication"`
}

// Packages lists package presets and explicit package names.
type Packages struct {
	Presets []string `yaml:"presets,omitempty"`
	Names   []string `yaml:"names,omitempty"`
}

// Services lists services to enable after installation.
type Services struct {
	Enable []string `yaml:"enable,omitempty"`
}

// Validation describes checks to perform against an installed host.
type Validation struct {
	SSH      SSHValidation       `yaml:"ssh,omitempty"`
	Commands []ValidationCommand `yaml:"commands,omitempty"`
}

// SSHValidation describes an SSH readiness check.
type SSHValidation struct {
	Enabled bool   `yaml:"enabled"`
	User    string `yaml:"user,omitempty"`
}

// ValidationCommand describes one post-install command check.
type ValidationCommand struct {
	Name           string `yaml:"name"`
	Command        string `yaml:"command"`
	ExpectExitCode int    `yaml:"expect_exit_code"`
}

// PostInstall lists scripts to run after installation.
type PostInstall struct {
	Scripts []Script `yaml:"scripts,omitempty"`
}

// Script is a named post-install script.
type Script struct {
	Name    string `yaml:"name"`
	Content string `yaml:"content"`
}

// OSSpecific holds explicitly separated distribution-specific overrides.
type OSSpecific struct {
	Alpine   *AlpineOverrides   `yaml:"alpine,omitempty"`
	Ubuntu   *UbuntuOverrides   `yaml:"ubuntu,omitempty"`
	RHELLike *RHELLikeOverrides `yaml:"rhel_like,omitempty"`
	Arch     *ArchOverrides     `yaml:"arch,omitempty"`
}

// AlpineOverrides describes Alpine-specific installer settings.
type AlpineOverrides struct {
	InstallMode  string   `yaml:"install_mode,omitempty"`
	Repositories []string `yaml:"repositories,omitempty"`
}

// UbuntuOverrides describes Ubuntu-specific autoinstall settings.
type UbuntuOverrides struct {
	AutoinstallVersion int    `yaml:"autoinstall_version,omitempty"`
	APTMirror          string `yaml:"apt_mirror,omitempty"`
}

// RHELLikeOverrides describes shared Rocky and Fedora settings.
type RHELLikeOverrides struct {
	SELinux  string `yaml:"selinux,omitempty"`
	Firewall string `yaml:"firewall,omitempty"`
}

// ArchOverrides describes Arch-specific installer settings.
type ArchOverrides struct {
	AURHelper string `yaml:"aur_helper,omitempty"`
}
