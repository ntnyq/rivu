# Implementation status

Date: October 13, 2026. Scope: macOS-first development build of Rivu 0.0.1.

## Implemented

The application now has a real application-owned torrent session, native main/settings/import/menu bar views, Chinese UI, persisted settings and tasks, queue admission, file priorities, pause/batch restore, recheck, ratio seeding, completion notification bookkeeping, conservative private-magnet discovery, rooted storage and journaled removal. Product identity, repository/domain links and editable/generated icons are configured.

The visual implementation follows `docs/design/styles.css` and the approved prototype rather than the old starter screens: 191 pt sidebar, solid blue navigation, 69 pt toolbar, 102 pt task rows, 271 pt inspector, segmented tabs, file symbols and small secondary labels. Settings use a 168 pt category sidebar and grouped rows/switches. Both appearances and the 880 × 560 minimum window are rendered in tests.

The engine is pinned to anacrolix/torrent v1.61.1-0.20260927071845-d913b30f520e. Controlled local swarm tests verify v1, v2-only and hybrid byte integrity, magnet metadata without payload writes, pause detachment, resume rehash and selected-file transfers across shared piece boundaries. The adapter retains the short-hash admission path introduced for v1.61.0, retaining/comparing full identity in Rivu before payload admission. No dependency cache source was modified.

The October 13 refactor adds isolated magnetic-link preview, structured recovery hints, bandwidth units and explicit seeding modes, existing-data adoption, directory relocation, file trees and directory priorities, per-tracker diagnostics, scheduled speed limits, cumulative seeding duration, task tags, and opt-in watch-folder importing. See [current source structure](implementation-structure.md) and [feature semantics and validation](refactor-validation.md).

## Verification performed

- Go unit/UI tests, race detector, Go vet and MyGo vet.
- Native-rendered screenshots for main default/minimum sizes, settings, import and menu bar, light/dark.
- Local swarm integration using generated legal test data, with no public tracker requirement.
- macOS arm64 app build, ad-hoc signature and native startup/capture/clean exit with an isolated session directory. Native interaction verified settings, unavailable-update feedback, modal sheet presentation, invalid magnet rejection and Cmd+Q.
- SVG-derived icon set: PNG 16–1024, macOS ICNS/iconset, Windows ICO and template tray images.

## Remaining release gates

This is not a claim that every milestone's hardware acceptance gate has passed. The original plan remains the acceptance reference.

- GitHub Releases update feeds and the embedded verification key are configured. Release builds require the matching Actions signing secret and verify signed archives before tagging. A local macOS universal build passed a disposable-bundle test: a bad signature retained the old app, and a valid update replaced it and launched the new executable. Public-release delivery, the complete settings/session/restart flow, and native Windows/Linux updates still require acceptance testing. Production Developer ID signing and notarization remain unavailable.
- Detailed macOS login/notification authorization adapters, multi-monitor flyout behavior, VoiceOver and permission-denial/revocation flows need packaged native acceptance.
- Forced-termination injection across every persistence/Trash boundary remains to be expanded. Large-list profiling and regression benchmarks are recorded in [the refactor report](refactor-validation.md).
- Upload totals can lag up to one checkpoint interval after a crash. Session error holds are runtime observations; durable user pauses survive restart, while environmental failures are detected again on admission.
- Atomic replacement syncs the file and parent directory on macOS/Linux. A post-rename sync failure blocks further mutation pending reload; hard power-loss recovery still needs filesystem-specific acceptance.
- Windows amd64 and Linux amd64 cross-compilation passed. These outputs are compilation checks only. Windows payload ownership/free-space adapters are unavailable and refuse download admission.
- DNS/website deployment, repository push and GitHub publishing have not been performed by this implementation task. No project license has been selected for Rivu's own code yet.

See `desktop.md` and `releasing.md` before distributing to users.
