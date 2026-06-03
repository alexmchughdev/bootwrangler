# Quickstart

## Requirements

- Go 1.22+
- Node 20+
- [Wails v2](https://wails.io) (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)
- `qemu-system-x86_64` (optional, for the Lab feature)

## Run in development mode

```sh
wails dev
```

This starts the Go backend and the Vite dev server with hot-reload.

## Build the desktop app

```sh
make build-desktop
```

The compiled binary lands in `build/bin/`.

## Build the CLI only

```sh
make build
# binary at dist/bootwrangler
```

## Run tests

```sh
make test
```

## First steps

1. Open the desktop app or run `./dist/bootwrangler`.
2. Go to **Profiles** → fill in OS family, version, hostname, SSH keys.
3. Click **Validate** then **Render** to generate unattended installer assets.
4. Insert a USB drive → go to **USB Devices** to confirm it is detected as safe.
5. Go to **Flash Image** → select an OS image and the USB device → confirm to flash.
