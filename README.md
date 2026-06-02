# BootForge

BootForge is a GUI-first Linux provisioning and boot media studio. It is
intended to help users create, validate, test, version, share, and deploy Linux
installation profiles and boot media from a desktop application.

The project is under active development. The current foundation provides the
`bootforge` CLI entry point and version reporting. Provisioning, rendering,
media composition, device operations, and the desktop interface will be added
incrementally with explicit safety checks around destructive operations.

## Build

```sh
go build ./cmd/bootforge
```

## Test

```sh
go test ./...
```

## CLI

```sh
go run ./cmd/bootforge --help
go run ./cmd/bootforge version
```
