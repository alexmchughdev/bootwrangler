package lab

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ToolSource indicates where the QEMU binaries were resolved from.
type ToolSource string

const (
	SourceBundled   ToolSource = "bundled"   // shipped alongside the app
	SourceSystem    ToolSource = "system"    // found on PATH
	SourceInstalled ToolSource = "installed" // auto-installed into the app data dir
	SourceNone      ToolSource = "none"      // not available
)

// Tools holds resolved paths to the QEMU binaries.
type Tools struct {
	SystemBinary string // qemu-system-x86_64[.exe]
	ImgBinary    string // qemu-img[.exe]
	Source       ToolSource
	Version      string
}

// Available reports whether QEMU is usable.
func (t Tools) Available() bool {
	return t.Source != SourceNone && t.SystemBinary != "" && t.ImgBinary != ""
}

func systemBinName() string {
	if runtime.GOOS == "windows" {
		return "qemu-system-x86_64.exe"
	}
	return "qemu-system-x86_64"
}

func imgBinName() string {
	if runtime.GOOS == "windows" {
		return "qemu-img.exe"
	}
	return "qemu-img"
}

// bundledDirs returns candidate directories for QEMU shipped with the app,
// covering the macOS .app layout, a sibling qemu/ folder, and the exe dir.
func bundledDirs() []string {
	var dirs []string
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		dirs = append(dirs,
			filepath.Join(exeDir, "qemu", "bin"),
			filepath.Join(exeDir, "qemu"),
			// macOS: Contents/MacOS/<exe> -> Contents/Resources/qemu/bin
			filepath.Join(exeDir, "..", "Resources", "qemu", "bin"),
			filepath.Join(exeDir, "..", "Resources", "qemu"),
		)
	}
	if env := os.Getenv("BOOTWRANGLER_QEMU_DIR"); env != "" {
		dirs = append([]string{env}, dirs...)
	}
	return dirs
}

// InstallDir is where auto-install places a portable QEMU.
func InstallDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".bootwrangler", "qemu")
	}
	return filepath.Join(home, ".bootwrangler", "qemu")
}

func installBinDir() string {
	return filepath.Join(InstallDir(), "bin")
}

func isFile(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

// resolve returns the QEMU binary paths and their source without probing the
// version. It checks bundled locations, then PATH, then the app-data install.
func resolve() (sys, img string, src ToolSource) {
	sysName := systemBinName()
	imgName := imgBinName()

	for _, d := range bundledDirs() {
		s := filepath.Join(d, sysName)
		i := filepath.Join(d, imgName)
		if isFile(s) && isFile(i) {
			return s, i, SourceBundled
		}
	}

	if s, err := exec.LookPath(sysName); err == nil {
		if i, err := exec.LookPath(imgName); err == nil {
			return s, i, SourceSystem
		}
	}

	s := filepath.Join(installBinDir(), sysName)
	i := filepath.Join(installBinDir(), imgName)
	if isFile(s) && isFile(i) {
		return s, i, SourceInstalled
	}

	return "", "", SourceNone
}

// qemuImgPath returns the resolved qemu-img path, or the bare name fallback.
func qemuImgPath() string {
	if _, i, src := resolve(); src != SourceNone {
		return i
	}
	return imgBinName()
}

// Locate resolves QEMU and probes its version. Use for status display.
func Locate() Tools {
	sys, img, src := resolve()
	t := Tools{SystemBinary: sys, ImgBinary: img, Source: src}
	if src != SourceNone {
		if out, err := exec.Command(sys, "--version").Output(); err == nil {
			t.Version = strings.TrimSpace(strings.SplitN(string(out), "\n", 2)[0])
		}
	}
	return t
}
