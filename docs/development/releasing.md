# Building and releasing Rivu

## Product identity

- Application: Rivu
- Bundle ID: `com.ntnyq.rivu`
- Repository: `https://github.com/ntnyq/rivu`
- Website: `https://rivu.ntnyq.dev`
- Platforms: macOS, Windows, Linux
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

Build packages for each release target with:

```sh
go tool mygo build -platform darwin/universal
go tool mygo build -platform windows/amd64
go tool mygo build -platform linux/amd64
```

macOS signing and DMG creation require a Mac. Build the Windows installer on Windows; MyGo locates or downloads NSIS automatically. Linux packaging produces a DEB, a `.tar.gz` archive and `install.sh`. Windows and Linux releases currently target x64; a source build can use another architecture supported by MyGo. On Windows without Make, run the corresponding `go` commands directly.

The icon's source of truth is `resources/icon.svg`. `make icon` renders all PNG sizes, `resources/rivu.icns`, `resources/rivu.iconset`, `resources/rivu.ico` and monochrome template tray images. Do not edit generated bitmap assets instead of the vector.

## Release workflow

Manually dispatch the Release workflow from the default branch with a stable version in `MAJOR.MINOR.PATCH` format, without `v`. It requires an existing default branch with commits and repository write permission.

1. **Prepare:** validate the version and branch, update `internal/app/run.go`, `mygo.json` and `CHANGELOG.md`, then run race tests, translation checks, local-swarm integration, Go vet and MyGo vet. Share these prepared files and `RELEASE_NOTES.md` with the remaining jobs.
2. **Build:** package the same prepared version on macOS, Windows and Linux runners in parallel, signing the update archives with `MYGO_UPDATER_PRIVATE_KEY`. Verify and collect each target's installer, archive and update manifest. A missing asset or an invalid signature, size, version or archive URL fails the build.
3. **Release:** after every build succeeds, verify all combined update assets again, generate `SHA256SUMS`, create the release commit and tag with an atomic push, and upload a **draft** GitHub Release. Its notes come from the same text as the new changelog entry.

| Runner               | Target             | Release assets                                                                |
| -------------------- | ------------------ | ----------------------------------------------------------------------------- |
| macOS                | `darwin/universal` | Universal `.dmg` for Apple Silicon and Intel                                  |
| Windows              | `windows/amd64`    | `Rivu Setup <version> amd64.exe`                                              |
| Linux (Ubuntu 24.04) | `linux/amd64`      | `rivu_<version>_amd64.deb`, `rivu-<version>-linux-amd64.tar.gz`, `install.sh` |

Each target also includes `rivu-<version>-<target>.tar.gz` and `update-<target>.json`; the Linux archive is both an installation and an update asset. The combined workflow artifact and draft also include `RELEASE_NOTES.md` and `SHA256SUMS`. Review the [native acceptance matrix and known platform limitations](desktop.md) before publishing a draft. Publish it as a stable release and mark it as latest to make it available to update checks.

### User-facing changelog

Both `CHANGELOG.md` and GitHub Release Notes include only `feat`, `fix` and `refactor` commits since the previous stable tag (or all history for the first release). Scoped and breaking-change forms such as `fix(storage): ...` and `feat!: ...` are included. Merge commits and other types such as `chore`, `docs`, `test`, `ci`, `build`, `style` and `perf` are omitted. Keep user-visible changes under one of the three included types.

If the range contains commits but none match, the release is still allowed and its notes say “No user-facing changes in this release.” A Full Changelog comparison link preserves access to every commit when a previous stable tag exists. Older changelog sections are preserved when adding a release.

Do not run the workflow merely to obtain a local development bundle: it mutates the repository and creates a tag. There are no signing secrets in this source tree.

## Signed updates

`mygo.json` configures `updates.github=ntnyq/rivu`, `tagPrefix=v` and the public verification key. The pinned MyGo build embeds the target's feed at `https://github.com/ntnyq/rivu/releases/latest/download/update-<target>.json`. The macOS universal build uses `darwin-universal` on both Apple Silicon and Intel. Full signed archives are used (`deltas=0`), so no previous release downloads are needed while building.

The repository Actions secret **`MYGO_UPDATER_PRIVATE_KEY`** must contain the matching private key. The workflow checks that it exists before preparing a release and exposes it only to that check and the package-build step. `tools/releaseassets` verifies the Ed25519 signature of each archive's SHA-256 digest using the committed public key before collecting any files and again after combining artifacts, before creating a tag. Missing signed assets block release; `SHA256SUMS` alone is not an update signature.

For a new fork, generate a dedicated key pair with `go tool mygo keygen`, set `updates.github` and `updates.publicKey`, adjust the repository URL checks in `tools/releaseassets`, and store the private key as that fork's Actions secret. Back up the private key outside Git; do not regenerate or replace a published application's key casually, because existing installations only trust their embedded public key. Never commit `mygo-update.key` or place it in workflow artifacts.

For a signed local build, supply the same private key through `MYGO_UPDATER_PRIVATE_KEY`. Without it, MyGo can still build a local app but does not create signed update assets, so release collection fails. Apple Developer ID signing and notarization are independent of update signatures and remain unconfigured.

Draft releases are invisible to users' update checks. The first update-enabled build must be installed manually on systems running an older unconfigured build. Development mode and installations without write permission report updates as unavailable; Linux DEB installations should use the package manager instead. Do not treat `Updater.Enabled() == false` as an up-to-date result.

Automatic checks and automatic installation are separate opt-ins. Automatic installation requires an idle session and blocks new admissions. Manual install explains the interruption, saves/quiesces transfers, downloads/verifies/replaces through MyGo, and restores permitted execution. Restart goes through the normal quit/save coordinator.

Update checks keep a 45-second overall network deadline, with at most three attempts of 12 seconds each and 0.5/1-second backoff. Dial and TLS handshake timeouts are 5 seconds; the response-header timeout is 10 seconds. Only EOF, truncated responses, connection resets/aborts and timeouts are retried. Cancellation, invalid JSON, certificate errors and HTTP errors such as 404 stop the check; failures do not advance the last successful check time. Installation is not retried as a whole.

Update requests and archive downloads use a dedicated transport. Explicit `HTTP_PROXY`/`HTTPS_PROXY` settings take precedence; `NO_PROXY` also bypasses the Windows system proxy. Otherwise Windows reads the current user's proxy settings for every request, including redirected URLs. A manual proxy is preferred over WPAD discovery; an explicit PAC script takes precedence over the manual proxy, and PAC `DIRECT` is honored. PAC/WPAD lookup uses WinHTTP, with a 5-second caller deadline and at most one outstanding native lookup. Unavailable or timed-out WPAD discovery falls back to the direct route and pauses discovery for one minute, so the default auto-detect setting on home networks does not block updates. Manual proxy and explicit PAC changes still apply immediately. Explicit PAC failures are reported rather than silently bypassing the proxy. The MyGo v0.5.5 adapter routes only marked update contexts through this transport, preserving other application HTTP traffic and MyGo's HTTPS and signature checks. CI runs the updater's real-MyGo EOF contract and Windows proxy/PAC tests on the existing Windows AMD64 runner.

Before enabling this for users, test a controlled signed update, bad signature, wrong architecture, low disk space, interrupted replacement and save failure at restart. A runnable previous bundle must survive a failed update. See `desktop.md` for the rest of native acceptance.

## Licenses

A license for Rivu's own code has not yet been selected in this repository. The engine is pinned to anacrolix/torrent v1.61.1-0.20260927071845-d913b30f520e (MPL-2.0); MyGo and transitive dependencies have their own licenses. `THIRD_PARTY_NOTICES.md` records primary components. Prepare a complete dependency license inventory and fulfill corresponding source/notice obligations before distributing binaries.
