#!/usr/bin/env bash
#
# bundle-qemu.sh — stage a copy of QEMU next to the desktop build so the Lab
# works with no separate install. BootWrangler resolves QEMU in this order:
#
#   1. bundled   — <app>/qemu/bin (this script), or macOS Contents/Resources/qemu/bin
#   2. system    — qemu-system-x86_64 on PATH
#   3. installed — ~/.bootwrangler/qemu/bin (runtime auto-install fallback)
#
# This staging is best-effort: full macOS notarization of the QEMU binaries and
# their dylibs must happen in your signing pipeline. If bundling is skipped or
# incomplete, the app falls back to the one-click auto-install in the Lab.
#
# Usage: scripts/bundle-qemu.sh [dest-dir]
#   dest-dir defaults to build/qemu

set -euo pipefail

DEST="${1:-build/qemu}"
BIN_DEST="$DEST/bin"
OS="$(uname -s)"
ARCH="$(uname -m)"

mkdir -p "$BIN_DEST"

copy_with_deps_macos() {
  # Copy a Mach-O binary plus the Homebrew dylibs it links, then rewrite the
  # load paths to @executable_path/../lib so the bundle is self-contained.
  local bin="$1"
  local libdir="$DEST/lib"
  mkdir -p "$libdir"
  cp "$bin" "$BIN_DEST/"
  local name
  name="$(basename "$bin")"
  # Pull in non-system dylibs (Homebrew lives under /opt/homebrew or /usr/local).
  otool -L "$BIN_DEST/$name" | awk '/\/(opt\/homebrew|usr\/local)\// {print $1}' | while read -r dylib; do
    [ -f "$dylib" ] || continue
    cp -n "$dylib" "$libdir/" || true
    install_name_tool -change "$dylib" "@executable_path/../lib/$(basename "$dylib")" "$BIN_DEST/$name" || true
  done
}

case "$OS" in
Darwin)
  if ! command -v qemu-system-x86_64 >/dev/null 2>&1; then
    echo "qemu not found on PATH. Install with: brew install qemu" >&2
    echo "Skipping bundle — the app will auto-install QEMU at runtime instead." >&2
    exit 0
  fi
  echo "==> Bundling QEMU for macOS/$ARCH into $DEST"
  copy_with_deps_macos "$(command -v qemu-system-x86_64)"
  copy_with_deps_macos "$(command -v qemu-img)"
  # Copy the BIOS/firmware blobs QEMU needs at runtime.
  SHARE_SRC="$(dirname "$(command -v qemu-system-x86_64)")/../share/qemu"
  if [ -d "$SHARE_SRC" ]; then
    mkdir -p "$DEST/share/qemu"
    cp -R "$SHARE_SRC/." "$DEST/share/qemu/"
  fi
  echo "==> Done. NOTE: sign & notarize $BIN_DEST/* in your release pipeline."
  ;;
Linux)
  if ! command -v qemu-system-x86_64 >/dev/null 2>&1; then
    echo "qemu not found on PATH. Install with your package manager." >&2
    echo "Skipping bundle — the app will auto-install QEMU at runtime instead." >&2
    exit 0
  fi
  echo "==> Bundling QEMU for Linux/$ARCH into $DEST"
  cp "$(command -v qemu-system-x86_64)" "$BIN_DEST/"
  cp "$(command -v qemu-img)" "$BIN_DEST/"
  echo "==> Done (binaries only; relies on system shared libs)."
  ;;
*)
  echo "Windows bundling: download a portable build from https://qemu.weilnetz.de/w64/" >&2
  echo "and extract qemu-system-x86_64.exe + qemu-img.exe into $BIN_DEST" >&2
  exit 0
  ;;
esac
