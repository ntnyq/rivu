# Developing Rivu

[Documentation home](../index.md) · [User guide](../guide/index.md) · [Release guide](releasing.md)

## Requirements and dependencies

Use Go 1.27.2 or later on macOS, Windows, or Linux. The project pins MyGo and its CLI tool in `go.mod`; no Node.js package manager is needed for the native app. The documentation site has its own Node.js toolchain. Release packages target macOS (Apple Silicon and Intel), Windows x64, and Linux x64. Linux native windows need GTK 3 and tray integration needs AppIndicator. On Windows, use `go run .` and `go tool mygo build` directly if Make is not installed. See [desktop acceptance](desktop.md) for platform verification and current limitations.

MyGo v0.5.5 and anacrolix/torrent v1.61.1-0.20260927071845-d913b30f520e are pinned in `go.mod`. Torrent and dht/v2 use fixed development commits (`d913b30f520e` and `a9c967f9d68b`), recorded as Go pseudo-versions rather than release tags. Together they support `missinggo/v2` v2.11.0's roaring/v2 bitmap and remove the old DHT dependency on `stm`'s iterator API.

Run the following commands from the repository root.

## Run locally

```sh
go mod download
make run
```

Use `make dev` for MyGo development mode. To build a local app bundle without creating a disk image:

```sh
go tool mygo build -skip-dmg
```

On Apple Silicon, the resulting app is `build/darwin-arm64/Rivu.app`. Local builds are ad-hoc signed, not Developer ID signed or notarized. `make build` also produces the platform installer. See [releasing](releasing.md) for universal builds and distribution.

## Launch options

```sh
go run . /path/to/file.torrent 'magnet:?xt=urn:btih:…&tr=…'
go run . --data-dir /tmp/rivu-session
go run . --config /tmp/rivu-settings.json
go run . --version
```

`--data-dir` changes session and default preference storage; `--config` changes preferences only. Neither changes the single-instance application identity or MyGo window-state location. Close an existing Rivu instance before using an isolated test session. The [desktop verification guide](desktop.md) describes instance-socket isolation for native smoke checks.

## Validation

```sh
make check
make race
make integration
go tool mygo vet
make snapshots
```

`make check` runs unit tests, Go vet, and MyGo's translation checks. Integration tests use an isolated local swarm without public torrents. Snapshot tests render the application UI with synthetic tasks and cover both appearances and all supported languages.

Native changes also require the checks in [desktop behavior and acceptance](desktop.md). Rendered snapshots and cross-compilation do not validate platform permissions, menu bar positioning, accessibility, or signed updates.

To capture a packaged native empty window and exit, first close existing Rivu instances and ensure `snapshots/` exists:

```sh
mkdir -p snapshots
env RIVU_CAPTURE="$PWD/snapshots/native-empty.png" \
  build/darwin-arm64/Rivu.app/Contents/MacOS/Rivu --data-dir /tmp/rivu-smoke
```

Use a fresh data directory for an empty session. A `--data-dir` override does not isolate MyGo's window state.

## Documentation site

Use Node.js 22.12 or later and the pnpm version declared in the root `package.json`. This toolchain builds the VitePress site independently of the native app.

```sh
pnpm install
pnpm docs:dev
pnpm docs:build
pnpm docs:preview
```

The development server supports live editing. The production build checks internal dead links and writes `docs/.vitepress/dist`; preview serves that build locally.

English content lives in `docs/` and matching Simplified Chinese pages in `docs/zh/`. Keep file names and directory structure aligned so language switching reaches the corresponding page. Both homepages use `layout: home`, `hero`, and `features`. Navigation, sidebars, local search and translated theme labels live in `docs/.vitepress/config.ts`.

The production site is **https://rivu.ntnyq.dev**, served from `/`. Deploy `docs/.vitepress/dist` to the static host for that domain. The site uses default `.html` routes, so no extensionless URL rewrites are required. Canonical URLs and the sitemap use the production domain.

For Vercel, set the project's **Root Directory** to `docs` and enable **Include source files outside of the Root Directory in the Build Step** so the root pnpm workspace, lockfile, and package manager declaration are available. Select Node.js 24.x and set `ENABLE_EXPERIMENTAL_COREPACK=1` in the project's environment variables to use the pnpm version pinned in the root `package.json`. The checked-in `docs/vercel.json` selects VitePress, installs with `pnpm install --frozen-lockfile`, runs `pnpm run build` from `docs`, and publishes `.vitepress/dist` while preserving `.html` URLs. Connect `rivu.ntnyq.dev` in the project's domain settings. See [Vercel configuration](https://vercel.com/docs/project-configuration/vercel-json) for the supported settings.

The documentation is organized by audience:

```text
docs/
├── index.md          # Product homepage
├── guide/            # Getting started, features, settings, and help
├── development/      # Build, architecture, verification, and releases
├── history/          # Archived design and implementation plans
├── design/           # Review-only visual prototype
├── images/           # Product screenshots
├── public/           # Static assets
├── zh/               # Matching Simplified Chinese pages
└── .vitepress/       # Site configuration and theme
```

User-guide sidebars and developer sidebars are separate. Historical plans and the visual prototype are excluded from local search with `search: false`. When moving a page, update links, navigation, and its translated counterpart. Building the site does not deploy it or enable app updates.

The site icon, `docs/public/icon.png`, is copied from `resources/icons/rivu-256.png`; refresh it when the app icon changes.

## Documentation screenshots

The product screenshots committed under `docs/images/` come from the application UI renderer, not the HTML design prototype. Their transfer names, rates, and progress are test fixtures. The README identifies them as demo transfers.

Refresh them after relevant UI changes:

```sh
make snapshots
cp snapshots/light-1280.png docs/images/main-light.png
cp snapshots/dark-1280.png docs/images/main-dark.png
cp snapshots/menu-bar-light.png docs/images/menu-bar.png
```

Inspect each image before committing. `snapshots/` remains ignored; only the curated images in `docs/images/` are tracked. Native captures are a separate verification step.

## Icons

The source of truth is [resources/icon.svg](https://github.com/ntnyq/rivu/blob/main/resources/icon.svg). Regenerate assets with:

```sh
make icon
```

Generated assets include PNG sizes 16–1024, `resources/rivu.icns`, `resources/rivu.ico`, the macOS iconset, and template menu bar images. Edit the SVG rather than generated bitmaps.

## Localization

The `language` preference accepts `system` (default), `zh-Hans`, `zh-Hant`, `en`, `ja`, or `ko`. Existing settings without this field continue to follow the system language.

MyGo embeds the five ARB JSON catalogs in `locales/`. English is the source and fallback language. Use flat lower camel case keys with a shared semantic prefix, such as `settingsSave` and `transferEtaSeconds`. Write complete ICU MessageFormat messages with named placeholders; declare their Go types in the source catalog's `@message` metadata. Keep distinct keys for distinct meanings. The reserved `mygo.*` keys supply framework menu, control, and dialog translations, including Japanese and Korean strings from the starter.

Update all five catalogs together, then run:

```sh
make generate
make i18n
```

Commit `locales/catalog_gen.go` with catalog changes; never edit it by hand. Call the generated functions directly (`locales.SettingsSave()`, `locales.TransferEtaSeconds(seconds)`) or use `locales.In(tag)` for an explicit language. MyGo's CLI also writes optional TypeScript output under the ignored `.mygo/` directory, keeping native source generation out of the documentation workspace.

The saved Rivu preference remains authoritative and applies through `App.SetLocale` after a successful settings save; `system` maps to MyGo's empty override. MyGo resolves the user's ordered preferred languages and updates native text, controls, and role menus together. `App.OnLocaleChanged` rebuilds application menus and tray content and refreshes window titles, including when system preferences change. `MYGO_LOCALE` supports temporary language previews without persisting them.

Translate phase labels when rendering. Structured errors carry a typed `func(locales.Messages) string` renderer and an explicit cause; the presentation boundary selects the current locale, while `Error()` uses English for stable diagnostics and `Unwrap()` preserves the cause. Async work returns display text, which must not be translated again. Catalog tests check complete translations and unused messages, and `mygo i18n check` validates ICU messages and placeholder compatibility. Compilation checks accessor names and argument types. Localized UI tests cover all five languages at normal and minimum window sizes and set the tester's locale for framework controls. Tests replace only the in-memory locale provider and run serially, without writing native language preferences. Operating-system dialogs and third-party diagnostic messages may use their own language.

## Collection benchmarks

```sh
go test ./internal/app ./internal/session -run '^$' -bench 'Benchmark(Main1000Tasks|FileTree10000Files|CachedFileTree10000Files|Snapshot1000Tasks10000Files)$' -benchmem
```

See [the refactor report](refactor-validation.md) for measured results, persistence migration and feature semantics. The local Tracker integration test checks the pinned engine's diagnostics output contract and must pass after engine upgrades.

Native capture also accepts `RIVU_CAPTURE_SURFACE=settings-downloads` or `settings-automation` to exercise the new settings surfaces.

## Further reading

- [Architecture](architecture.md): session ownership, engine integration, and persistence.
- [Implementation status](implementation-status.md): implemented behavior and remaining release gates.
- [Desktop verification](desktop.md): native behavior and acceptance checks.
- [Release guide](releasing.md): packaging, signing, updates, and licenses.
- [Design reference](../design/README.md): the review-only HTML prototype.
