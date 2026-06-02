package profile

import (
	"strings"
	"testing"
)

const exampleSSHKey = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJ0EkzLnXFTKfQXv2bseD0jLxoc+R6NxRHWqnY15T9i6 example-only@bootwrangler.invalid"

func TestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		mutate      func(*Profile)
		wantProblem string
	}{
		{
			name: "valid profile",
		},
		{
			name: "unsupported family",
			mutate: func(value *Profile) {
				value.OS.Family = "gentoo"
			},
			wantProblem: `os.family "gentoo" is unsupported`,
		},
		{
			name: "invalid hostname",
			mutate: func(value *Profile) {
				value.System.Hostname = "Invalid Host"
			},
			wantProblem: `system.hostname "Invalid Host" is invalid`,
		},
		{
			name: "wipe confirmation required",
			mutate: func(value *Profile) {
				value.Disk.ConfirmDestructive = false
			},
			wantProblem: "disk.confirm_destructive must be true when disk.mode is wipe",
		},
		{
			name: "static address required",
			mutate: func(value *Profile) {
				value.Network.Mode = "static"
			},
			wantProblem: "network.address is required for static networking",
		},
		{
			name: "invalid SSH key",
			mutate: func(value *Profile) {
				value.Users[0].SSHKeys[0] = "not-a-key"
			},
			wantProblem: "users[0].ssh_keys[0] is not a valid SSH public key",
		},
		{
			name: "unknown package preset",
			mutate: func(value *Profile) {
				value.Packages.Presets = []string{"unknown"}
			},
			wantProblem: `packages.presets[0] "unknown" is unsupported`,
		},
		{
			name: "script path rejected",
			mutate: func(value *Profile) {
				value.PostInstall.Scripts[0].Name = "../script.sh"
			},
			wantProblem: `post_install.scripts[0].name "../script.sh" must be a file name without a path`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			value := validProfile()
			if test.mutate != nil {
				test.mutate(&value)
			}

			err := Validate(value)
			if test.wantProblem == "" {
				if err != nil {
					t.Fatalf("Validate() error = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatal("Validate() error = nil, want error")
			}
			if !strings.Contains(err.Error(), test.wantProblem) {
				t.Fatalf("Validate() error = %q, want problem %q", err, test.wantProblem)
			}
		})
	}
}

func validProfile() Profile {
	return Profile{
		Name: "test-profile",
		OS: OS{
			Family:       "ubuntu",
			Version:      "24.04",
			Architecture: "x86_64",
		},
		System: System{
			Hostname: "test-profile",
			Timezone: "Europe/London",
			Keyboard: "gb",
			Locale:   "en_GB.UTF-8",
		},
		Network: Network{
			Mode:      "dhcp",
			Interface: "auto",
		},
		Disk: Disk{
			Mode:               "wipe",
			Target:             "auto",
			InstallMode:        "server",
			Filesystem:         "ext4",
			ConfirmDestructive: true,
		},
		Users: []User{
			{
				Name:    "deploy",
				Shell:   "/bin/bash",
				Groups:  []string{"sudo"},
				Sudo:    true,
				SSHKeys: []string{exampleSSHKey},
			},
		},
		SSH: SSH{
			Enabled: true,
		},
		Packages: Packages{
			Presets: []string{"minimal"},
			Names:   []string{"git"},
		},
		Services: Services{
			Enable: []string{"ssh"},
		},
		Validation: Validation{
			SSH: SSHValidation{
				Enabled: true,
				User:    "deploy",
			},
		},
		PostInstall: PostInstall{
			Scripts: []Script{
				{
					Name:    "example.sh",
					Content: "echo example\n",
				},
			},
		},
	}
}
