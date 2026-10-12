# Rivu

A quiet, native BitTorrent downloader for macOS, built with Go and MyGo. No WebView or JavaScript runtime is used in the application.

[Website](https://rivu.ntnyq.dev) · [Repository](https://github.com/ntnyq/rivu) · [Design reference](docs/design/README.md)

Rivu imports `.torrent` files and magnet links, manages a download queue, continues transfers after its windows close, and exposes the same session through the main window and menu bar. The interface follows the approved three-column design with light and dark appearances.

## Development

Requires Go 1.27.2+; MyGo v0.4.3 and anacrolix/torrent v1.61.0 are pinned in `go.mod`. macOS is the first supported runtime target.

```sh
go mod download
make run
make dev
make build
```

```sh
go run . /path/to/file.torrent 'magnet:?xt=urn:btih:…&tr=…'
go run . --data-dir /tmp/rivu-session
go run . --config /tmp/rivu-settings.json
go run . --version
```

`--data-dir` changes session and default preference storage; `--config` changes preferences only. Neither changes the single-instance application identity or MyGo window-state location. Close an existing Rivu instance before using an isolated test session.

## Implemented behavior

- Import review with per-item errors, destination validation, file selection, free-space checks and an immediate-start option.
- Real transfers, selected-file progress, high/normal/skipped file priorities, queue limits, pause/resume, recheck and ratio-based seeding.
- Durable intent separate from measurements; manual pauses and the latest Pause All batch survive restart. Resume Batch leaves manually paused tasks alone.
- App-owned transfers, native menus, command palette, drag-and-drop, `.torrent`/`magnet` associations, Dock reopening and a menu bar panel.
- Atomic session/preferences persistence, corrupt-file preservation, immutable metadata, rooted payload I/O and file ownership journals.
- Removal retains files by default. Optional Trash only acts on recorded owned files, records partial progress and never falls back to permanent deletion.
- General, download, connection, notification and update settings. Signed MyGo update coordination includes a session maintenance barrier and explicit restart.

Unresolved magnets use only explicit trackers or peers until metadata establishes the privacy flag. A magnet without either is rejected with an explanation. Existing destination files are not adopted automatically; choose an empty destination. Restored task data is rehashed. Skipped files can contain bytes from shared pieces.

## Validation

```sh
make check
make race
make integration             # Local swarm; no public torrent dependency
go tool mygo vet
make snapshots               # Native UI renderer; synthetic tasks only in tests
make icon                    # Regenerate PNG, ICNS, ICO and iconset from SVG
```

Capture a packaged native empty window and exit:

```sh
RIVU_CAPTURE="$PWD/snapshots/native-empty.png" \
  build/darwin-arm64/Rivu.app/Contents/MacOS/Rivu --data-dir /tmp/rivu-smoke
```

The SVG source is [resources/icon.svg](resources/icon.svg). Generated assets include `resources/rivu.icns`, `resources/rivu.ico`, the macOS iconset, PNG sizes 16–1024 and template menu bar images.

## Storage and recovery

Default data is under `os.UserConfigDir()/Rivu`, normally `~/Library/Application Support/Rivu` on macOS. It contains `settings.json`, `session.json`, immutable torrent metadata and ownership journals. Upload totals are checkpointed every 15 seconds; a forced termination can lose the latest statistics interval. Transfer completion is verified from data instead of trusting saved percentages.

Corrupt or future-schema files block mutation and remain untouched. Repair a backed-up copy, then use **File → 修复后重新读取会话与设置**. Settings version 1 can be migrated when explicitly loaded; the application does not alter another product's configuration directory.

## Release status

This repository contains a development implementation and local app packaging. The configured domain is product identity; it does not imply that hosting or DNS has been deployed. No release signing identity, notarization credentials, update feed or verification key is included. Update checks report unavailable until a trusted release configuration is supplied.

Windows/Linux builds are compilation checks, not supported downloadable releases. macOS native permission flows, multi-monitor tray behavior, forced-termination recovery, accessibility and signed-update replacement still require the acceptance matrix in [docs/desktop.md](docs/desktop.md). Do not equate successful snapshots with these platform checks.

Read [Architecture](docs/architecture.md), [Implementation status](docs/implementation-status.md), and [Release guidance](docs/releasing.md).
