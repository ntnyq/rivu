# Repository Guidelines

## Project Structure & Module Organization

Rivu is a native Go/MyGo BitTorrent client targeting macOS. Root Go files handle startup, CLI, windows, views, menus, and tray integration. Keep business logic in `internal/`: `session` owns transfer state, `torrent` adapts the engine, `config` stores preferences, and `files`, `storage`, `desktop`, and `tasks` provide supporting services. Tests live beside implementations as `*_test.go`. Assets belong in `resources/`; `tools/genicon/` generates icons. `docs/` describes architecture and releases; `docs/design/` is the visual reference. Generated packages and captures go in `build/` and `snapshots/`.

`locales/` contains MyGo ARB catalogs and generated typed Go accessors. Update all five languages together, use flat lower camel case keys and ICU named placeholders, then run `make generate` and `make i18n`. Commit `locales/catalog_gen.go`; do not edit generated accessors by hand. `make check` includes translation validation. Structured errors defer rendering through `locales.Messages` and preserve English diagnostics and wrapped causes. Tests that replace the global locale provider must remain serial and must not persist native locale preferences.

## Build, Test, and Development Commands

Use Go 1.27.2+ and the MyGo tool pinned in `go.mod`. Run shell commands through RTK:

- `rtk go mod download`: fetch dependencies.
- `rtk proxy make run`: launch locally.
- `rtk proxy make dev`: start MyGo development mode.
- `rtk proxy make build`: package the application.
- `rtk proxy make fmt`: format Go sources.
- `rtk proxy make check`: run unit tests and Go vet.
- `rtk proxy make race`: test with the race detector.
- `rtk proxy make integration`: run the isolated local-swarm suite.
- `rtk proxy go tool mygo vet`: validate native UI usage.
- `rtk proxy make snapshots`: render UI captures.
- `rtk proxy make icon`: regenerate icon assets.

## Coding Style & Naming Conventions

Follow `.editorconfig`: UTF-8, LF, final newlines, tabs for Go/Makefiles, and two spaces elsewhere. Format with `gofmt`. Use lowercase package names, PascalCase exported identifiers, and camelCase unexported identifiers. Keep rendering based on immutable session snapshots; serialize transfer mutations through `internal/session`.

## Testing Guidelines

Use Go’s `testing` package, `TestXxx` names, and `t.TempDir()` fixtures. UI tests use MyGo’s `ui.NewTester`; swarm tests require the `integration` build tag and avoid public torrents. No numeric coverage threshold is configured. Test persistence failures, recovery, and changed transfer behavior. Before submission, run checks, race tests, integration tests, and MyGo vet. For native changes, follow `docs/desktop.md`; snapshots and cross-compilation cannot validate platform behavior.

## Commit & Pull Request Guidelines

History currently contains `chore: init commit`. Follow that type-prefixed style, such as `fix: preserve manual pauses`. PRs should describe behavior changes, link relevant issues, list validation and tested platforms, and include screenshots for visual changes.

## Agent-Specific Instructions

Follow `@/Users/ntnyq/.codex/RTK.md`: prefix shell commands with `rtk`, except `pnpm typecheck`. This repository needs no pnpm. If used, prepend `/Users/ntnyq/Library/pnpm/bin` to the existing `PATH`, preserving the active Node.js version. Use the user-managed pnpm; never modify bundled pnpm or its runtime cache.
