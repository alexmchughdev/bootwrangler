#!/bin/sh

set -eu

case "$(uname -s)" in
Darwin)
	export CGO_LDFLAGS="${CGO_LDFLAGS:+${CGO_LDFLAGS} }-framework UniformTypeIdentifiers -mmacosx-version-min=10.13"
	;;
Linux)
	;;
*)
	echo "desktop builds are supported on macOS and Linux; Windows support is roadmapped" >&2
	exit 1
	;;
esac

exec wails build "$@"
