// Package profile defines and validates BootWrangler provisioning profiles.
package profile

// Profile describes the desired installed operating-system configuration.
type Profile struct {
	Name        string      `json:"name" yaml:"name"`
	OS          OS          `json:"os" yaml:"os"`
	System      System      `json:"system" yaml:"system"`
	Network     Network     `json:"network" yaml:"network"`
	Disk        Disk        `json:"disk" yaml:"disk"`
	Users       []User      `json:"users" yaml:"users"`
	SSH         SSH         `json:"ssh" yaml:"ssh"`
	Packages    Packages    `json:"packages" yaml:"packages,omitempty"`
	Services    Services    `json:"services" yaml:"services,omitempty"`
	Validation  Validation  `json:"validation" yaml:"validation,omitempty"`
	PostInstall PostInstall `json:"post_install" yaml:"post_install,omitempty"`
	OSSpecific  OSSpecific  `json:"os_specific" yaml:"os_specific,omitempty"`
}

// OS identifies the target operating system.
type OS struct {
	Family       string `json:"family" yaml:"family"`
	Version      string `json:"version" yaml:"version"`
	Architecture string `json:"architecture" yaml:"architecture"`
}

// System describes common installed-system settings.
type System struct {
	Hostname string `json:"hostname" yaml:"hostname"`
	Timezone string `json:"timezone" yaml:"timezone"`
	Keyboard string `json:"keyboard" yaml:"keyboard"`
	Locale   string `json:"locale" yaml:"locale"`
}

// Network describes the initial network configuration.
type Network struct {
	Mode      string   `json:"mode" yaml:"mode"`
	Interface string   `json:"interface" yaml:"interface"`
	Address   string   `json:"address" yaml:"address,omitempty"`
	Gateway   string   `json:"gateway" yaml:"gateway,omitempty"`
	DNS       []string `json:"dns" yaml:"dns,omitempty"`
}

// Disk describes installer disk behavior. Runtime disk operations apply
// additional device-safety checks before destructive actions.
type Disk struct {
	Mode               string `json:"mode" yaml:"mode"`
	Target             string `json:"target" yaml:"target"`
	InstallMode        string `json:"install_mode" yaml:"install_mode"`
	Filesystem         string `json:"filesystem" yaml:"filesystem"`
	ConfirmDestructive bool   `json:"confirm_destructive" yaml:"confirm_destructive"`
}

// User describes one installed-system account.
type User struct {
	Name    string   `json:"name" yaml:"name"`
	Shell   string   `json:"shell" yaml:"shell,omitempty"`
	Groups  []string `json:"groups" yaml:"groups,omitempty"`
	Sudo    bool     `json:"sudo" yaml:"sudo,omitempty"`
	SSHKeys []string `json:"ssh_keys" yaml:"ssh_keys,omitempty"`
}

// SSH describes installed-system SSH daemon settings.
type SSH struct {
	Enabled                bool `json:"enabled" yaml:"enabled"`
	PermitRootLogin        bool `json:"permit_root_login" yaml:"permit_root_login"`
	PasswordAuthentication bool `json:"password_authentication" yaml:"password_authentication"`
}

// Packages lists package presets and explicit package names.
type Packages struct {
	Presets []string `json:"presets" yaml:"presets,omitempty"`
	Names   []string `json:"names" yaml:"names,omitempty"`
}

// Services lists services to enable after installation.
type Services struct {
	Enable []string `json:"enable" yaml:"enable,omitempty"`
}

// Validation describes checks to perform against an installed host.
type Validation struct {
	SSH      SSHValidation       `json:"ssh" yaml:"ssh,omitempty"`
	Commands []ValidationCommand `json:"commands" yaml:"commands,omitempty"`
}

// SSHValidation describes an SSH readiness check.
type SSHValidation struct {
	Enabled bool   `json:"enabled" yaml:"enabled"`
	User    string `json:"user" yaml:"user,omitempty"`
}

// ValidationCommand describes one post-install command check.
type ValidationCommand struct {
	Name           string `json:"name" yaml:"name"`
	Command        string `json:"command" yaml:"command"`
	ExpectExitCode int    `json:"expect_exit_code" yaml:"expect_exit_code"`
}

// PostInstall lists scripts to run after installation.
type PostInstall struct {
	Scripts []Script `json:"scripts" yaml:"scripts,omitempty"`
}

// Script is a named post-install script.
type Script struct {
	Name    string `json:"name" yaml:"name"`
	Content string `json:"content" yaml:"content"`
}

// OSSpecific holds explicitly separated distribution-specific overrides.
type OSSpecific struct {
	Alpine   *AlpineOverrides   `json:"alpine,omitempty" yaml:"alpine,omitempty"`
	Ubuntu   *UbuntuOverrides   `json:"ubuntu,omitempty" yaml:"ubuntu,omitempty"`
	RHELLike *RHELLikeOverrides `json:"rhel_like,omitempty" yaml:"rhel_like,omitempty"`
	Arch     *ArchOverrides     `json:"arch,omitempty" yaml:"arch,omitempty"`
}

// AlpineOverrides describes Alpine-specific installer settings.
type AlpineOverrides struct {
	InstallMode  string   `json:"install_mode" yaml:"install_mode,omitempty"`
	Repositories []string `json:"repositories" yaml:"repositories,omitempty"`
}

// UbuntuOverrides describes Ubuntu-specific autoinstall settings.
type UbuntuOverrides struct {
	AutoinstallVersion int    `json:"autoinstall_version" yaml:"autoinstall_version,omitempty"`
	APTMirror          string `json:"apt_mirror" yaml:"apt_mirror,omitempty"`
}

// RHELLikeOverrides describes shared Rocky and Fedora settings.
type RHELLikeOverrides struct {
	SELinux  string `json:"selinux" yaml:"selinux,omitempty"`
	Firewall string `json:"firewall" yaml:"firewall,omitempty"`
}

// ArchOverrides describes Arch-specific installer settings.
type ArchOverrides struct {
	AURHelper string `json:"aur_helper" yaml:"aur_helper,omitempty"`
}
