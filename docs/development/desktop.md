# Desktop behavior and acceptance

## Implemented paths

- Native MyGo main/settings/add windows. macOS keeps its inset integrated titlebar; Windows/Linux use default native decorations with application controls below them.
- Left-click menu bar flyout and temporary right-click menu. Linux falls back to a tray menu.
- Dock activation reopens the main window on macOS. Windows/Linux retain the session after closing windows only while a tray has been created; otherwise closing the last window uses the normal quit confirmation and save coordinator.
- `.torrent` and `magnet` associations in `mygo.json`; early OS-open events are queued until initialization completes.
- Second-instance arguments reuse import review. `--config` does not isolate application identity.
- Login registration uses MyGo `SetOpenAtLogin`; current boolean status is read on opening settings. Quiet login preserves the application session without showing a window only when a Dock/tray recovery entry is available.
- Notifications use task/completion-epoch IDs. Clicking routes to task selection. Successful display requests are not proof of delivery.
- Suspend saves and quiesces; resume reconciles unchanged durable intent.
- Quit confirmation and asynchronous saving also guard update restart.

## Platform policies

- Windows/Linux expose Settings, About and Quit in the File menu; shortcut hints use Control. Saving the tray preference applies immediately even when the language is unchanged. macOS keeps its existing application menu and Command hints.
- New default destinations use Windows Known Folders or Linux XDG user directories. Saved destinations stay unchanged. Import review and watched-folder imports create only the current or legacy Rivu default under an existing downloads directory. They never recreate missing parents or redirected volumes.
- Windows validates device names, reserved characters and control characters before admission. File ownership uses the volume serial and full 128-bit file ID queried from an open handle; filesystems without that query are rejected. Publication uses directory-relative NT rename with replacement disabled. Unix ownership journal IDs retain their existing format.
- Payload and journal contents are flushed before publication. Windows has no equivalent to an unprivileged directory fsync, so metadata durability after power loss still depends on the filesystem.
- `TestPlatformWindowDecorations` guards the platform options and unchanged macOS geometry. `TestNativeTitleBarContentBounds` covers all five languages at normal/minimum widths, with expanded/collapsed sidebars. Storage tests exercise stable identity, adoption, collision refusal, renamed-parent anchoring and missing-parent protection. Trash checks reuse component validation and handle identity rather than comparing canonical path spelling. CI runs storage/session and local-swarm tests on Windows/Linux as well as the macOS suite.

Native acceptance must additionally cover Windows DPI changes between monitors, maximize/restore, Snap layouts, window icons, drag areas, modal focus, long/UNC paths and network volumes. On Linux, cover GNOME/KDE under Wayland/X11, left/right/no titlebar buttons, fractional scaling, missing AppIndicator support, tray restoration, desktop associations and installed-app integration. Tray creation does not prove that a Linux desktop is displaying the indicator; verify the desktop host/extension as well. Software snapshots and cross-compilation do not establish these behaviors.

## Local verification

On the development macOS arm64 host, the app bundle builds with ad-hoc signing, starts with an isolated data directory, renders its native empty window and exits through the save coordinator. MyGo software-rendered screenshots cover default and minimum sizes, light/dark, settings, import and menu bar surfaces. Fixture tasks are confined to tests. Interactive native checks also verified that Add Download is an actual macOS sheet, invalid magnets remain unsubmitted, settings opens, unconfigured updates report unavailable, and Cmd+Q exits cleanly.

## Release acceptance still required

| Area          | Required native check                                                                          |
| ------------- | ---------------------------------------------------------------------------------------------- |
| Menu bar      | Anchoring/dismissal/right click across multiple monitors and display changes                   |
| Lifecycle     | Closing the last window during real transfer, Dock reopen, sleep/wake on hardware              |
| Associations  | Finder cold launch, browser magnet launch, duplicate launch events                             |
| Login         | Signed installed bundle, enabled/denied/revoked registration, quiet login                      |
| Notifications | Authorized/denied permission, focus mode, click routing from background                        |
| Accessibility | VoiceOver, keyboard-only navigation, Reduce Motion/Transparency                                |
| Updates       | Signed feed, wrong architecture/signature, interrupted replacement, restart after save failure |
| Storage       | External volume replacement, full disk, controlled forced termination during checkpoint/Trash  |

MyGo v0.4.8 exposes boolean login observation and notification support, not every OS authorization state. Rivu reports observed values and original errors without inferring richer permission states. The detailed native status adapters proposed in the plan remain a release gap.

Release packages target macOS, Windows, and Linux. Validate installation, startup, file associations, tray behavior, transfers, and clean exit on each OS before publishing the draft; cross-compilation alone does not cover these checks. The recorded native checks below were performed on macOS. Windows now has free-space, handle-based file identity and exclusive publication adapters. Native transfer and filesystem acceptance is still required; producing an installer does not verify transfer behavior.

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

On October 17, 2026, localization migrated to MyGo v0.5.5's ARB catalogs,
generated typed accessors, and application locale callbacks. Unit tests, Go vet,
race tests, the isolated swarm suite, MyGo vet, catalog checks, and localized
snapshots passed. The macOS arm64 app built successfully with all five language
resources derived from the catalogs. Seven isolated native startup/capture/exit
checks covered all five main-window languages, Japanese download settings, and
the Traditional Chinese tray panel (`snapshots/native-mygo-i18n-*.png`). A separate
isolated native session verified saving English → Japanese → Follow System,
including immediate menu, window-title, main-window, and save-result refresh.
`MYGO_LOCALE` prevented these checks from persisting native language preferences.

The October 12, 2026 macOS arm64 i18n build passed native startup/capture/exit checks
for all five explicit languages and the system default. These checks used an
isolated data directory and `XDG_RUNTIME_DIR` for the MyGo instance socket, leaving
the installed app running. Interactive save-and-switch verification remains a
native acceptance check; the local UI automation service timed out during that
check. Existing-view language changes are covered by the headless UI tests.

To capture the actual AppKit popover with an isolated session, quit other Rivu
instances, then run (the capture process quits after writing the PNG):

```sh
env RIVU_CAPTURE="$PWD/snapshots/native-tray.png" RIVU_CAPTURE_SURFACE=tray \
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

The semantic i18n key migration on October 13, 2026 passed unit tests, Go vet,
race tests, isolated swarm integration tests, MyGo vet, and localized UI snapshots.
All five catalogs retain their translations under 359 semantic keys; 21 unused
entries were removed. The macOS arm64 app and DMG built successfully on retry
after the first packaging attempt lost its temporary staging directory.
Native captures covered the main window in all five languages, English download
settings, and Simplified Chinese automation settings, using isolated data and
instance sockets. All seven processes exited cleanly without missing-key logs.
Captures are `snapshots/native-semantic-<language>-<surface>.png`; these smoke
checks do not replace the interactive and hardware release checks above.

The October 13 sidebar and empty-state adjustment passed unit tests, Go vet,
race tests, isolated swarm integration tests, MyGo vet, and localized snapshots.
The macOS arm64 app and DMG built successfully. An isolated native session
verified entering and leaving full screen: the brand moves into the top row
when window controls hide, and the compact Settings footer stays at the bottom.
The empty library shows one instruction and no details panel; the details toggle
is disabled without a selection. `TestDetailsRequireSelection` covers opening
details on request and removing the selected task at regular and narrow widths.
The native Chinese window capture is `snapshots/native-layout-main.png`.

On October 18, 2026, the cross-platform window/storage pass preserved the macOS
window geometry and added native Windows/Linux storage CI. A separate isolated
macOS native check saved Tray enabled → disabled while the language stayed
English; both changes immediately updated the native tray. The temporary smoke
harness was removed after verification. A Windows-only removal test covers
canonical filename casing, while shared tests reject symlinked parent paths and
prevent automatic recreation of missing downloads parents. Windows/Linux native
UI and filesystem acceptance remains required.
