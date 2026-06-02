# BootWrangler

BootWrangler is a GUI-first Linux provisioning and boot media studio. It is
intended to help users create, validate, test, version, share, and deploy Linux
installation profiles and boot media from a desktop application.

The project is under active development. The current foundation provides the
`bootwrangler` CLI entry point, version reporting, and a Wails desktop shell with a
React frontend. Provisioning, rendering, media composition, and device
operations will be added incrementally with explicit safety checks around
destructive operations.

## Build

```sh
go build ./cmd/bootwrangler
```

## Test

```sh
go test ./...
```

## Desktop Development

The desktop application currently supports macOS and Linux. Windows desktop
packaging is roadmapped.

Install the Wails CLI and frontend dependencies:

```sh
go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0
cd frontend
npm install
cd ..
```

Start the desktop development environment:

```sh
wails dev
```

Build the desktop application:

```sh
./scripts/build-desktop.sh
```

## Neovim

The repository includes `.nvim.lua` with project commands:

```text
:BootWranglerTest
:BootWranglerFrontendCheck
:BootWranglerFrontendTest
:BootWranglerDesktopBuild
```

Enable project-local configuration with `set exrc` in your Neovim
configuration, then trust this repository's `.nvim.lua` when Neovim prompts.

BootWrangler's desktop profile editor will also support opening profiles and
scripts in Neovim as the first external editor integration.

## CLI

```sh
go run ./cmd/bootwrangler --help
go run ./cmd/bootwrangler version
go run ./cmd/bootwrangler profile edit ./examples/profiles/ubuntu-server.yaml --editor nvim
go run ./cmd/bootwrangler profile validate ./examples/profiles/ubuntu-server.yaml
```
