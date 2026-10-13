# Desktop behavior and acceptance

## Implemented paths

- Native MyGo main/settings/add windows, with macOS titlebar controls supplied by the OS.
- Left-click menu bar flyout and temporary right-click menu. Linux falls back to a tray menu.
- Dock activation reopens the main window. Closing windows retains the session.
- `.torrent` and `magnet` associations in `mygo.json`; early OS-open events are queued until initialization completes.
- Second-instance arguments reuse import review. `--config` does not isolate application identity.
- Login registration uses MyGo `SetOpenAtLogin`; current boolean status is read on opening settings. Quiet login preserves the application session without showing a window.
- Notifications use task/completion-epoch IDs. Clicking routes to task selection. Successful display requests are not proof of delivery.
- Suspend saves and quiesces; resume reconciles unchanged durable intent.
- Quit confirmation and asynchronous saving also guard update restart.

## Local verification

On the development macOS arm64 host, the app bundle builds with ad-hoc signing, starts with an isolated data directory, renders its native empty window and exits through the save coordinator. MyGo software-rendered screenshots cover default and minimum sizes, light/dark, settings, import and menu bar surfaces. Fixture tasks are confined to tests. Interactive native checks also verified that Add Download is an actual macOS sheet, invalid magnets remain unsubmitted, settings opens, unconfigured updates report unavailable, and Cmd+Q exits cleanly.

## Release acceptance still required

| Area | Required native check |
| --- | --- |
| Menu bar | Anchoring/dismissal/right click across multiple monitors and display changes |
| Lifecycle | Closing the last window during real transfer, Dock reopen, sleep/wake on hardware |
| Associations | Finder cold launch, browser magnet launch, duplicate launch events |
| Login | Signed installed bundle, enabled/denied/revoked registration, quiet login |
| Notifications | Authorized/denied permission, focus mode, click routing from background |
| Accessibility | VoiceOver, keyboard-only navigation, Reduce Motion/Transparency |
| Updates | Signed feed, wrong architecture/signature, interrupted replacement, restart after save failure |
| Storage | External volume replacement, full disk, controlled forced termination during checkpoint/Trash |

MyGo v0.4.8 exposes boolean login observation and notification support, not every OS authorization state. Rivu reports observed values and original errors without inferring richer permission states. The detailed native status adapters proposed in the plan remain a release gap.

Windows and Linux compilation do not establish desktop support. Safe file ownership currently has a macOS/Linux adapter; unsupported platforms refuse payload writes. Do not publish Windows packages as functional download clients until that adapter and native testing exist.

## Visual parity checks

The menu bar content uses the design palette even when AppKit reports vibrancy.
Its height follows up to four preview rows (including paused tasks), with a compact
304 pt empty state. Pause and speed-limit actions sit above the footer; Add Download
and Settings sit beside Open Main Window. Main task rows use 102 pt spacing and
23 pt top-aligned controls, with the drop hint below the list.

`make snapshots` includes the 1280 × 818 design comparison size, empty/populated
menu bar panels, and a native-material rendering path in both appearances.
`TestTrayLayoutAndMaterial` checks material opacity and footer bounds.

Language preferences default to the native system locale. General settings offers
Simplified Chinese, Traditional Chinese, English, Japanese and Korean, with
changes applied after saving. `TestLocalizedSurfaces` covers every language at
normal and minimum window sizes and includes tray action bounds; `make snapshots`
writes `i18n-<language>-<surface>.png` captures. Catalog tests validate translation
coverage and formatting placeholders, while configuration tests cover migration
and persistence of language preferences.

The October 12, 2026 macOS arm64 i18n build passed native startup/capture/exit checks
for all five explicit languages and the system default. These checks used an
isolated data directory and `XDG_RUNTIME_DIR` for the MyGo instance socket, leaving
the installed app running. Interactive save-and-switch verification remains a
native acceptance check; the local UI automation service timed out during that
check. Existing-view language changes are covered by the headless UI tests.

To capture the actual AppKit popover with an isolated session, quit other Rivu
instances, then run (the capture process quits after writing the PNG):

```sh
rtk proxy env RIVU_CAPTURE="$PWD/snapshots/native-tray.png" RIVU_CAPTURE_SURFACE=tray \
  build/darwin-arm64/Rivu.app/Contents/MacOS/Rivu --data-dir /tmp/rivu-tray-review
```

This captures native content; the OS-owned arrow, outer corner shape, shadow and
multi-monitor positioning still require interactive inspection.

On October 12, 2026, the macOS arm64 build was rechecked with native main-window
inspection and a successful AppKit popover content capture. Unit tests, race tests,
the isolated swarm integration suite, Go vet and MyGo vet passed. The installed
`/Applications/Rivu.app` is separate from the rebuilt app under `build/darwin-arm64/`.

The subsequent control alignment pass replaced text disclosure characters with
centered SVG icons, gave settings selectors a shared width, and added hover/press
feedback to settings navigation. Sorting now opens a checked native menu instead
of cycling on click. Native macOS arm64 interaction verified selecting English,
saving, and immediately refreshing the settings window, main window, and menus;
selecting Progress from the sort menu also updated the visible control. These
checks used an isolated session without transfers. `TestSortMenuSelection` and
`TestSettingsLanguageMenu` cover menu selection and keyboard opening.

On October 13, 2026, the dependency upgrade to MyGo v0.4.8, Torrent commit
`d913b30f520e`, and DHT commit `a9c967f9d68b` passed unit tests, race tests, the
isolated swarm integration suite, Go vet, MyGo vet, and UI snapshot tests.
Windows and Linux cross-compilation and the ad-hoc signed macOS arm64 app
build also passed. Native main-window and AppKit popover content captures used
temporary data directories and isolated instance sockets; both processes exited
cleanly. These smoke checks do not replace the release acceptance checks above.

## October 13 refactor verification

The refactored build's main window, download settings, and automation settings were captured through actual AppKit windows using separate temporary data and instance-socket directories. Each process exited through the session save coordinator with exit code 0. Captures are `snapshots/native-refactor-main.png`, `native-refactor-settings-downloads.png`, and `native-refactor-settings-automation.png`. These isolated smoke checks do not claim the real-transfer/hardware acceptance rows above. Feature semantics, local-swarm checks and collection benchmarks are in [the refactor report](refactor-validation.md).
