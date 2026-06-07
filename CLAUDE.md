# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

BootWrangler is a GUI-first Linux provisioning and boot-media studio: create/validate/render unattended-install profiles, manage an OS image catalogue, flash USB/disk media, compose multi-boot media, and test installs in a QEMU lab. It ships as a **Wails desktop app** and a **CLI**, both built on the same Go core.

## Commands

Use the Makefile targets:

- `make build` — build the CLI (`dist/bootwrangler`)
- `make build-desktop` — build the Wails desktop app (`./scripts/build-desktop.sh`; needs the `wails` CLI on PATH)
- `make test` — `go test ./...` + frontend `vitest`
- `make lint` — `gofmt -l` check + `go vet ./...`
- `make fmt` — `gofmt -w .`

Frontend (in `frontend/`): `npm run dev` (Vite, :5173), `npm run lint` (`tsc --noEmit`), `npm test -- --run`, `npm run build`.
Run one Go test: `go test ./internal/usb/ -run TestParseLsblkOutput -v`.
Run one frontend test: `npx vitest run src/navigation.test.ts`.
Docs site (in `docs-site/`, Docusaurus): `make docs-dev` (:3000), `make docs-build`.

Restricted sandboxes block the default Go build cache. If `go` commands fail with cache/`no such file` errors, prefix with `GOCACHE=/private/tmp/bootwrangler-gocache GOMODCACHE=/private/tmp/bootwrangler-gomodcache` (run `go mod download` once after `go clean -modcache`).

## Architecture (the parts that span files)

**Core never depends on the GUI.** All real logic lives in `internal/*` packages. Two thin entry points call the same core:
- `main.go` (repo root) — Wails desktop app; binds `internal/app.Service`.
- `cmd/bootwrangler/main.go` — CLI; delegates to `internal/cli`.
Both blank-import `internal/renderers/all` so renderers register themselves. Never duplicate provisioning logic between GUI and CLI — add it to a core package and call it from both.

**Renderer registry (plugin pattern).** `internal/render` defines the `Renderer` interface + a global registry. Each distro renderer (`internal/renderers/{alpine,ubuntu,debian,rocky,fedora,arch,opensuse}`) self-registers in its `init()`. `internal/renderers/all` blank-imports them all. CLI/GUI look renderers up by OS family via the registry — they never call a distro package directly. Adding a distro = new package under `internal/renderers/` + add its blank import to `all`.

**GUI↔Go bridge.** Wails exposes `internal/app.Service` methods to the frontend as `window.go.app.Service.*`. `frontend/src/api/backend.ts` is the typed wrapper and the single boundary; it falls back to browser-preview behaviour when not running inside Wails (so `npm run dev` renders without a backend). Adding a backend capability = method on `Service` → wrapper fn in `backend.ts` → call from a page.

**Central data model.** `internal/profile` (the `Profile` struct + validation) is the hub: profiles are validated, rendered into installer assets + a `internal/manifest.Manifest`, versioned in the library, and tested in the lab. `internal/library` is the local `~/.bootwrangler/` workspace with git-backed version history.

**Platform-specific code uses build tags, not runtime branches.** `internal/usb/detect_{linux,darwin,windows}.go` (lsblk / diskutil / PowerShell) and `flash_{unix,windows}.go`; `internal/lab/sysattr_{unix,windows}.go`. After any change here, cross-compile-check all three: `GOOS=darwin/windows/linux go build ./...`.

**Disk safety is a hard invariant.** `internal/usb/safety.go` classifies devices (`Safe`/`SafeReason`); flashing refuses unsafe/system/mounted targets and is gated behind explicit confirmation. Keep destructive operations guarded and dry-runnable.

**QEMU lab resolution.** `internal/lab/locate.go` finds QEMU in priority order: bundled-with-app → system PATH → auto-installed (`~/.bootwrangler/qemu`). `internal/lab/install.go` does one-click install via brew/winget/apt-dnf-pacman. `scripts/bundle-qemu.sh` stages QEMU next to the desktop build.

**Frontend.** React 19 + Vite + TypeScript. One stylesheet, `frontend/src/styles/app.css` — the "Workbench" design system (light/dark, steel-blue accent, monospace-forward, flat — deliberately not translucent-chip SaaS). `App.tsx` owns navigation: 16 pages grouped into ~6 task sections with contextual sub-nav. Pages live in `frontend/src/pages/`.

**Docs split.** Public docs in `docs/` are surfaced by the Docusaurus app in `docs-site/`. The marketing website is a **separate private repo**; its only public contract is `docs/website-integration.md` — do not add marketing source here.

## Hard constraints

- **No AI attribution anywhere.** CI (`.github/workflows/ci.yml`) greps `*.go *.ts *.tsx *.md *.yaml` and **fails the build** on terms like `chatgpt`, `openai`, `co-authored-by`, `ai-generated`, etc. This applies to source, comments, docs, and commit messages. A secret-scan job similarly blocks private keys.
- CI also enforces `gofmt`, `go vet`, `go test`, and the full frontend lint/test/build — run `make lint && make test` before proposing a commit.
- Commit messages: concise conventional-commit style (`feat:`, `fix:`, `chore:`, `docs:`); commit per logical change.
- Generated/transient paths are gitignored: `frontend/wailsjs/`, `frontend/dist/` (keep `frontend/dist/gitkeep`), `docs-site/build`, `docs-site/.docusaurus/`. `build-desktop` regenerates Wails bindings; don't commit them.
