package lab

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Availability describes the QEMU setup state for the GUI.
type Availability struct {
	Installed bool   `json:"installed"`
	Source    string `json:"source"`   // bundled|system|installed|none
	Version   string `json:"version"`  // e.g. "QEMU emulator version 8.2.0"
	Path      string `json:"path"`     // resolved qemu-system path
	Method    string `json:"method"`   // how Install() will obtain QEMU
	CanAuto   bool   `json:"can_auto"` // whether automatic install is possible now
	Hint      string `json:"hint"`     // manual instructions fallback
}

// CheckAvailability resolves QEMU and reports how it can be installed if absent.
func CheckAvailability() Availability {
	t := Locate()
	a := Availability{
		Installed: t.Available(),
		Source:    string(t.Source),
		Version:   t.Version,
		Path:      t.SystemBinary,
	}
	a.Method, a.CanAuto, a.Hint = installMethod()
	return a
}

func hasCmd(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// installMethod inspects the platform for an automatic install path.
func installMethod() (method string, canAuto bool, hint string) {
	switch runtime.GOOS {
	case "darwin":
		if hasCmd("brew") {
			return "brew install qemu", true,
				"Run in Terminal: brew install qemu"
		}
		return "homebrew-missing", false,
			"Install Homebrew from https://brew.sh, then run: brew install qemu"
	case "windows":
		if hasCmd("winget") {
			return "winget install QEMU", true,
				"Run in PowerShell: winget install SoftwareFreedomConservancy.QEMU"
		}
		return "winget-missing", false,
			"Download the QEMU installer from https://qemu.weilnetz.de/w64/"
	case "linux":
		switch {
		case hasCmd("apt-get"):
			return "apt-get install qemu-system-x86", hasCmd("pkexec"),
				"sudo apt-get install qemu-system-x86 qemu-utils"
		case hasCmd("dnf"):
			return "dnf install qemu-system-x86", hasCmd("pkexec"),
				"sudo dnf install qemu-system-x86 qemu-img"
		case hasCmd("pacman"):
			return "pacman -S qemu-base", hasCmd("pkexec"),
				"sudo pacman -S qemu-base"
		}
		return "manual", false,
			"Install qemu-system-x86 with your distribution's package manager"
	}
	return "manual", false, "Install QEMU from https://www.qemu.org/download/"
}

// Install attempts to install QEMU automatically using the platform package
// manager. progress receives human-readable status lines. It returns an error
// carrying manual instructions when automatic install is not possible.
func Install(progress func(string)) error {
	if progress == nil {
		progress = func(string) {}
	}
	if Locate().Available() {
		progress("QEMU is already available.")
		return nil
	}

	switch runtime.GOOS {
	case "darwin":
		if hasCmd("brew") {
			return runInstall(progress, "brew", "install", "qemu")
		}
		return fmt.Errorf("Homebrew is required for automatic install. Install it from https://brew.sh then run: brew install qemu")

	case "windows":
		if hasCmd("winget") {
			return runInstall(progress, "winget", "install", "--silent",
				"--accept-package-agreements", "--accept-source-agreements",
				"--id", "SoftwareFreedomConservancy.QEMU")
		}
		return fmt.Errorf("winget is required for automatic install. Download QEMU from https://qemu.weilnetz.de/w64/")

	case "linux":
		switch {
		case hasCmd("pkexec") && hasCmd("apt-get"):
			return runInstall(progress, "pkexec", "apt-get", "install", "-y",
				"qemu-system-x86", "qemu-utils")
		case hasCmd("pkexec") && hasCmd("dnf"):
			return runInstall(progress, "pkexec", "dnf", "install", "-y",
				"qemu-system-x86", "qemu-img")
		case hasCmd("pkexec") && hasCmd("pacman"):
			return runInstall(progress, "pkexec", "pacman", "-S", "--noconfirm",
				"qemu-base")
		}
		return fmt.Errorf("install QEMU with your package manager, e.g. sudo apt-get install qemu-system-x86 qemu-utils")
	}
	return fmt.Errorf("automatic install is not supported on this platform; install QEMU from https://www.qemu.org/download/")
}

func runInstall(progress func(string), name string, args ...string) error {
	progress(fmt.Sprintf("$ %s %s", name, strings.Join(args, " ")))
	out, err := exec.Command(name, args...).CombinedOutput()
	if trimmed := strings.TrimSpace(string(out)); trimmed != "" {
		progress(trimmed)
	}
	if err != nil {
		return fmt.Errorf("%s failed: %w", name, err)
	}
	if !Locate().Available() {
		return fmt.Errorf("install finished but QEMU was not found — you may need to restart BootWrangler")
	}
	progress("QEMU installed successfully.")
	return nil
}
