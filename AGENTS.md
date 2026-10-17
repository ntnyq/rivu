# Repository Guidelines

## Project Structure & Module Organization

Rivu is a native Go/MyGo BitTorrent client for macOS, Windows, and Linux. The root `main.go` delegates to `internal/app.Run`; `internal/app` owns startup, CLI, windows, views, menus, and tray integration as one package. Keep business logic in `internal/`: `session` owns transfer state, `torrent` adapts the engine, `config` stores preferences, and `files`, `storage`, `desktop`, and `windowtask` provide supporting services. `windowtask` manages cancellable window jobs, not transfers. Tests live beside implementations as `*_test.go`. Assets and their framework-independent embedding API belong in `resources/`; SVG parsing and UI controls stay in `internal/app/`; `tools/genicon/` generates icons and `tools/releaseassets/` verifies signed release assets. `docs/guide/` contains product documentation, `docs/development/` covers engineering and releases, and `docs/history/` preserves earlier plans; `docs/design/` is the visual reference. Generated packages and captures go in `build/` and `snapshots/`.

`locales/` contains MyGo ARB catalogs and generated typed Go accessors. Update all five languages together, use flat lower camel case keys and ICU named placeholders, then run `make generate` and `make i18n`. Commit `locales/catalog_gen.go`; do not edit generated accessors by hand. `make check` includes translation validation. Structured errors defer rendering through `locales.Messages` and preserve English diagnostics and wrapped causes. Tests that replace the global locale provider must remain serial and must not persist native locale preferences.

## Build, Test, and Development Commands

Use Go 1.27.2+ and the MyGo tool pinned in `go.mod`. Run these commands from the repository root:

- `go mod download`: fetch dependencies.
- `make run`: launch locally.
- `make dev`: start MyGo development mode.
- `make build`: package the application.
- `make fmt`: format Go sources.
- `make check`: run unit tests and Go vet.
- `make race`: test with the race detector.
- `make integration`: run the isolated local-swarm suite.
- `go tool mygo vet`: validate native UI usage.
- `make snapshots`: render UI captures.
- `make icon`: regenerate icon assets.

## Coding Style & Naming Conventions

Follow `.editorconfig`: UTF-8, LF, final newlines, tabs for Go/Makefiles, and two spaces elsewhere. Format with `gofmt`. Use lowercase package names, PascalCase exported identifiers, and camelCase unexported identifiers. Keep rendering based on immutable session snapshots; serialize transfer mutations through `internal/session`.

## Testing Guidelines

Use Go’s `testing` package, `TestXxx` names, and `t.TempDir()` fixtures. UI tests use MyGo’s `ui.NewTester`; swarm tests require the `integration` build tag and avoid public torrents. No numeric coverage threshold is configured. Test persistence failures, recovery, and changed transfer behavior. Before submission, run checks, race tests, integration tests, and MyGo vet. For native changes, follow `docs/development/desktop.md`; snapshots and cross-compilation cannot validate platform behavior.

## Commit & Pull Request Guidelines

History currently contains `chore: init commit`. Follow that type-prefixed style, such as `fix: preserve manual pauses`. PRs should describe behavior changes, link relevant issues, list validation and tested platforms, and include screenshots for visual changes.

## Documentation Toolchain

The native application builds with Go alone. The bilingual VitePress documentation under `docs/` uses the root pnpm workspace. Use Node.js 22.12+ and the pnpm version declared in `package.json`; run `pnpm install`, `pnpm docs:dev`, or `pnpm docs:build` from the repository root. Keep English and Chinese pages aligned.

## Structure Changes

Keep application code and its tests in `internal/app/` without splitting windows or views into separate packages prematurely. Keep `main.go` as the thin process entry point. After moving entry points or version metadata, update `Makefile`, `.github/workflows/release.yml`, and the bilingual development, structure, architecture, and release guides. `make snapshots` targets `internal/app`; the release workflow updates the version in `internal/app/run.go`.
