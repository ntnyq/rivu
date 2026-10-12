# Implementation status

Date: October 12, 2026. Scope: macOS-first development build of Rivu 0.1.0.

## Implemented

The application now has a real application-owned torrent session, native main/settings/import/menu bar views, Chinese UI, persisted settings and tasks, queue admission, file priorities, pause/batch restore, recheck, ratio seeding, completion notification bookkeeping, conservative private-magnet discovery, rooted storage and journaled removal. Product identity, repository/domain links and editable/generated icons are configured.

The visual implementation follows `docs/design/styles.css` and the approved prototype rather than the old starter screens: 191 pt sidebar, solid blue navigation, 69 pt toolbar, 102 pt task rows, 271 pt inspector, segmented tabs, file symbols and small secondary labels. Settings use a 168 pt category sidebar and grouped rows/switches. Both appearances and the 880 × 560 minimum window are rendered in tests.

The engine is pinned to anacrolix/torrent v1.61.0. Controlled local swarm tests verify v1, v2-only and hybrid byte integrity, magnet metadata without payload writes, pause detachment, resume rehash and selected-file transfers across shared piece boundaries. The adapter handles a pinned upstream v2-only admission issue by entering through its documented short-hash path and retaining/comparing full identity in Rivu before payload admission. No dependency cache source was modified.

## Verification performed

- Go unit/UI tests, race detector, Go vet and MyGo vet.
- Native-rendered screenshots for main default/minimum sizes, settings, import and menu bar, light/dark.
- Local swarm integration using generated legal test data, with no public tracker requirement.
- macOS arm64 app build, ad-hoc signature and native startup/capture/clean exit with an isolated session directory. Native interaction verified settings, unavailable-update feedback, modal sheet presentation, invalid magnet rejection and Cmd+Q.
- SVG-derived icon set: PNG 16–1024, macOS ICNS/iconset, Windows ICO and template tray images.

## Remaining release gates

This is not a claim that every milestone's hardware acceptance gate has passed. The original plan remains the acceptance reference.

- Production Developer ID signing, notarization, update hosting and verification keys are not available. The updater coordinator exists, but a full signed replacement/restart has not been exercised.
- Detailed macOS login/notification authorization adapters, multi-monitor flyout behavior, VoiceOver and permission-denial/revocation flows need packaged native acceptance.
- The first version rejects pre-existing conflicting payloads; a deliberate compatible-data adoption workflow is not implemented. Unknown magnets can be submitted as pending default-all tasks, but there is no live metadata-only preview inside the import draft.
- Tracker hostnames and peer count are observable; per-tracker announce results are not exposed by this adapter.
- Forced-termination injection across every persistence/Trash boundary, 1,000-task/10,000-file profiling remain to be expanded.
- Upload totals can lag up to one checkpoint interval after a crash. Session error holds are runtime observations; durable user pauses survive restart, while environmental failures are detected again on admission.
- Atomic replacement syncs the file and parent directory on macOS/Linux. A post-rename sync failure blocks further mutation pending reload; hard power-loss recovery still needs filesystem-specific acceptance.
- Windows amd64 and Linux amd64 cross-compilation passed. These outputs are compilation checks only. Windows payload ownership/free-space adapters are unavailable and refuse download admission.
- DNS/website deployment, repository push and GitHub publishing have not been performed by this implementation task. No project license has been selected for Rivu's own code yet.

See `desktop.md` and `releasing.md` before distributing to users.
