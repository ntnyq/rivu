# Building and releasing Rivu

## Product identity

- Application: Rivu
- Bundle ID: `com.ntnyq.rivu`
- Repository: `https://github.com/ntnyq/rivu`
- Website: `https://rivu.ntnyq.dev`
- Initial version: `0.1.0`
- macOS minimum: 12.0

These values are configured locally. DNS, website hosting and a GitHub release are separate deployment steps.

## Local build

```sh
make check
make race
make integration
go tool mygo vet
make snapshots
make icon
go tool mygo build -skip-dmg
```

On Apple Silicon, the app is `build/darwin-arm64/Rivu.app`. The default local signing is ad hoc; it is not Developer ID signing or notarization. Use `go tool mygo build` to produce the platform installer, and `-platform darwin/universal` for the universal macOS package.

The icon's source of truth is `resources/icon.svg`. `make icon` renders all PNG sizes, `resources/rivu.icns`, `resources/rivu.iconset`, `resources/rivu.ico` and monochrome template tray images. Do not edit generated bitmap assets instead of the vector.

## Release workflow

The manually dispatched Release workflow validates a stable version, updates `main.go`, `mygo.json` and the changelog, tests and packages macOS, creates the release commit/tag, and uploads a **draft** GitHub release. It requires an existing default branch with commits and repository write permission. It does not publish Windows/Linux installers. Review the native acceptance matrix before publishing a draft.

Do not run the workflow merely to obtain a local development bundle: it mutates the repository and creates a tag. There are no signing secrets in this source tree.

## Signed updates

The update coordinator is implemented, but a production feed and keys are intentionally unconfigured. Supply a trusted HTTPS feed, the corresponding public verification key, signed artifacts and Developer ID/notarization configuration using the pinned MyGo release tooling. Keep private keys in release secrets. Do not treat `Updater.Enabled() == false` as an up-to-date result.

Automatic checks and automatic installation are separate opt-ins. Automatic installation requires an idle session and blocks new admissions. Manual install explains the interruption, saves/quiesces transfers, downloads/verifies/replaces through MyGo, and restores permitted execution. Restart goes through the normal quit/save coordinator.

Before enabling this for users, test a controlled signed update, bad signature, wrong architecture, low disk space, interrupted replacement and save failure at restart. A runnable previous bundle must survive a failed update. See `desktop.md` for the rest of native acceptance.

## Licenses

A license for Rivu's own code has not yet been selected in this repository. The engine is pinned to anacrolix/torrent v1.61.0 (MPL-2.0); MyGo and transitive dependencies have their own licenses. `THIRD_PARTY_NOTICES.md` records primary components. Prepare a complete dependency license inventory and fulfill corresponding source/notice obligations before distributing binaries.
