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

MyGo v0.4.3 exposes boolean login observation and notification support, not every OS authorization state. Rivu reports observed values and original errors without inferring richer permission states. The detailed native status adapters proposed in the plan remain a release gap.

Windows and Linux compilation do not establish desktop support. Safe file ownership currently has a macOS/Linux adapter; unsupported platforms refuse payload writes. Do not publish Windows packages as functional download clients until that adapter and native testing exist.
