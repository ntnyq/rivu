# Developing Rivu

[Back to Rivu](../README.md) · [User guide](usage.md) · [Release guide](releasing.md)

## Requirements and dependencies

Use Go 1.27.2 or later on macOS. The project pins MyGo and its CLI tool in `go.mod`; no Node.js package manager is needed. macOS is the supported runtime target. Windows and Linux builds currently serve as compilation checks.

MyGo v0.4.8 and anacrolix/torrent v1.61.1-0.20260927071845-d913b30f520e are pinned in `go.mod`. Torrent and dht/v2 use fixed development commits (`d913b30f520e` and `a9c967f9d68b`), recorded as Go pseudo-versions rather than release tags. Together they support `missinggo/v2` v2.11.0's roaring/v2 bitmap and remove the old DHT dependency on `stm`'s iterator API.

Commands below use RTK, following the repository guidelines. Run them from the repository root.

## Run locally

```sh
rtk go mod download
rtk proxy make run
```

Use `rtk proxy make dev` for MyGo development mode. To build a local app bundle without creating a disk image:

```sh
rtk proxy go tool mygo build -skip-dmg
```

On Apple Silicon, the resulting app is `build/darwin-arm64/Rivu.app`. Local builds are ad-hoc signed, not Developer ID signed or notarized. `rtk proxy make build` also produces the platform installer. See [releasing](releasing.md) for universal builds and distribution.

## Launch options

```sh
rtk proxy go run . /path/to/file.torrent 'magnet:?xt=urn:btih:…&tr=…'
rtk proxy go run . --data-dir /tmp/rivu-session
rtk proxy go run . --config /tmp/rivu-settings.json
rtk proxy go run . --version
```

`--data-dir` changes session and default preference storage; `--config` changes preferences only. Neither changes the single-instance application identity or MyGo window-state location. Close an existing Rivu instance before using an isolated test session. The [desktop verification guide](desktop.md) describes instance-socket isolation for native smoke checks.

## Validation

```sh
rtk proxy make check
rtk proxy make race
rtk proxy make integration
rtk proxy go tool mygo vet
rtk proxy make snapshots
```

`make check` runs unit tests and Go vet. Integration tests use an isolated local swarm without public torrents. Snapshot tests render the application UI with synthetic tasks and cover both appearances and all supported languages.

Native changes also require the checks in [desktop behavior and acceptance](desktop.md). Rendered snapshots and cross-compilation do not validate platform permissions, menu bar positioning, accessibility, or signed updates.

To capture a packaged native empty window and exit, first close existing Rivu instances and ensure `snapshots/` exists:

```sh
rtk proxy mkdir -p snapshots
rtk proxy env RIVU_CAPTURE="$PWD/snapshots/native-empty.png" \
  build/darwin-arm64/Rivu.app/Contents/MacOS/Rivu --data-dir /tmp/rivu-smoke
```

Use a fresh data directory for an empty session. A `--data-dir` override does not isolate MyGo's window state.

## Documentation screenshots

The product screenshots committed under `docs/images/` come from the application UI renderer, not the HTML design prototype. Their transfer names, rates, and progress are test fixtures. The README identifies them as demo transfers.

Refresh them after relevant UI changes:

```sh
rtk proxy make snapshots
rtk proxy cp snapshots/light-1280.png docs/images/main-light.png
rtk proxy cp snapshots/dark-1280.png docs/images/main-dark.png
rtk proxy cp snapshots/menu-bar-light.png docs/images/menu-bar.png
```

Inspect each image before committing. `snapshots/` remains ignored; only the curated images in `docs/images/` are tracked. Native captures are a separate verification step.

## Icons

The source of truth is [resources/icon.svg](../resources/icon.svg). Regenerate assets with:

```sh
rtk proxy make icon
```

Generated assets include PNG sizes 16–1024, `resources/rivu.icns`, `resources/rivu.ico`, the macOS iconset, and template menu bar images. Edit the SVG rather than generated bitmaps.

## Localization

The `language` preference accepts `system` (default), `zh-Hans`, `zh-Hant`, `en`, `ja`, or `ko`. Existing settings without this field continue to follow the system language.

Translations are embedded from `internal/i18n/locales`. Catalog tests check message coverage and format placeholders; localized UI tests cover normal and minimum window sizes. Operating-system dialogs and third-party diagnostic messages may use their own language.

## Further reading

- [Architecture](architecture.md): session ownership, engine integration, and persistence.
- [Implementation status](implementation-status.md): implemented behavior and remaining release gates.
- [Desktop verification](desktop.md): native behavior and acceptance checks.
- [Release guide](releasing.md): packaging, signing, updates, and licenses.
- [Design reference](design/README.md): the review-only HTML prototype.
