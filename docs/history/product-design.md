---
search: false
---

# Rivu Product and Interface Design

> Historical design reference (October 12, 2026). For the shipped code and current dependency versions, see [architecture](../development/architecture.md), [source structure](../development/implementation-structure.md), and [implementation status](../development/implementation-status.md).

Date: October 12, 2026. Status: proposed product specification and interactive design prototype. The download engine and operating system integrations described here are not implemented by this deliverable.

## Product direction

Rivu is a lightweight, macOS-first BitTorrent client. Use the menu bar for quick checks and controls, and the main window for managing tasks and inspecting details.

Keep the existing Go 1.27.2 + MyGo v0.4.3 native application stack. The review prototype in [interactive prototype](../design/README.md) is a standalone design aid, not the application's runtime architecture. It introduces no web frontend dependency into the Go application.

The initial interface uses Chinese copy, with room for English localization. Keep transfer logic and persistence platform-independent; prioritize macOS integration before other desktop platforms.

## Primary user journeys

1. Open a torrent file, paste a magnet link, or drop a torrent into Rivu. Choose a destination and files, then start downloading.
2. Read progress, speed, remaining time, and actionable errors in the task list.
3. Close the main window while transfers continue. Click the menu bar icon to inspect or pause them.
4. Receive a completion notification and reveal downloaded files in Finder. Continue seeding according to the selected policy.
5. Opt into launch at login. Start quietly on subsequent logins and resume tasks that were previously intended to run.
6. Check for updates manually or automatically. Preserve transfer state before a user-approved restart.

## Scope and priorities

| Area        | P0: initial release                                                                         | Later iterations                        |
| ----------- | ------------------------------------------------------------------------------------------- | --------------------------------------- |
| Import      | Torrent files, magnets, drag and drop, batch import, duplicate infohash detection           | Watch folders and RSS                   |
| Transfers   | Pause/resume, queue, restart recovery, verification, file selection, destinations           | Scheduling and sequential download      |
| Management  | Search, filters, sorting, multiple selection, batch actions, separate remove/delete choices | Editable labels, rules, bulk relocation |
| Details     | Overview, file progress and priority, trackers, connection information                      | Piece diagnostics and history charts    |
| Bandwidth   | Global upload/download limits, temporary speed limit mode, concurrency limit                | Per-task limits and schedules           |
| Seeding     | Uploaded bytes, ratio, ratio-based stop policy, manual resume                               | Time-based and label-based policies     |
| macOS       | Menu bar panel, Dock entry, native dialogs, notifications, shortcuts                        | Optional menu-bar-only mode             |
| Preferences | Appearance, directories, seeding, startup, notifications, connections, updates              | Advanced proxies and diagnostics        |
| Updates     | Manual/automatic checks, opt-in automatic installation, release notes, failure handling     | Preview release channel                 |

Exclude content search, media catalogs, built-in playback, accounts, cloud synchronization, and remote control from the first release. Use open-source distributions and openly licensed movies in examples. Do not claim anonymous networking or comprehensive proxy routing without verifying all traffic paths.

## Information architecture

```text
Rivu
├── Main window
│   ├── Sidebar: All / Downloading / Seeding / Completed / Paused / Needs Attention
│   ├── Toolbar: Add / Pause or Resume Selected / More / Search / Inspector
│   ├── Task list: Name / Progress / Size / State / Rates / ETA
│   ├── Inspector: Overview / Files / Connections
│   └── Status bar: Total rates / Speed limit mode / Network state
├── Menu bar panel
│   ├── Total download and upload rates
│   ├── Up to four active or paused tasks; attention indicator
│   └── Add / Pause All or Resume Batch / Speed Limit / Main Window / Settings
├── Add Download sheet
│   └── Source → Metadata → Destination → File Selection → Start or Queue
└── Settings window
    ├── General
    ├── Downloads and Seeding
    ├── Connections
    ├── Notifications
    └── Updates and About
```

## Screen specifications

### Main window

Target a default native size of 1180 × 760 pt, with a minimum of 880 × 560 pt. Use a 190–220 pt sidebar and an approximately 292 pt optional inspector. Hide the inspector first when space becomes constrained; provide a details sheet when needed. The browser prototype also adapts to narrow review screens, which does not imply a mobile application.

Preserve native traffic-light controls and integrate the toolbar with the title bar. Use sidebar material for navigation and an opaque content surface for sustained reading. Reserve transparency for navigation and temporary panels.

Represent tasks as compact list rows, approximately 86–96 pt tall. Show the name first, then transferred/selected bytes and state, followed by a slim progress bar, rates, and ETA. Truncate long names only in the list; expose the complete name in the inspector and tooltip. Avoid large metric cards competing with the task list.

Single-click selects a task and updates the inspector. Double-click opens details rather than executing downloaded files. Support Cmd/Shift multiple selection and Space to pause or resume selected tasks in the native implementation.

Completed includes both actively seeding and stopped complete tasks, so counts can overlap. Downloading excludes queued tasks and metadata retrieval; All includes every state. Keep ordering stable across state changes. Default to newest additions, with name, progress, and state sorting available.

The context menu provides pause/resume, reveal in Finder, verify data, copy magnet link, and remove. Advanced tracker and peer information belongs in the inspector.

### Inspector

| Tab         | Contents                                                                                                           |
| ----------- | ------------------------------------------------------------------------------------------------------------------ |
| Overview    | Name, state, selected-file progress, rates, ETA, size, uploaded bytes, ratio, added time, location, seeding target |
| Files       | Tree, selection, per-file progress, high/normal/do-not-download priority, selected size                            |
| Connections | Tracker results, next announce, peer counts and sources                                                            |

Exclude unselected files from the completion denominator. Before metadata arrives, show an explicit retrieval state instead of invented sizes, files, or ETA. A failed tracker is a tracker-level result; only promote it to a task error when it prevents progress and needs user intervention.

### Menu bar panel

Use an approximately 18 pt monochrome template icon: a downward arrow entering a shallow tray is the initial concept. Let macOS tint it for light, dark, and selected states. Keep the icon still; communicate attention with a small indicator and text, not persistent animation.

Left-click toggles a panel anchored to the actual tray icon. Clicking elsewhere, clicking the icon again, or Escape dismisses it. Position within the current display's visible area; do not assume fixed coordinates or one monitor.

Use a 360 pt panel with content-dependent height. Show aggregate rates, up to four tasks with individual pause/resume controls, and an attention summary. The main-window action leads to the complete list. Clicking a task opens or focuses the main window and selects that task.

Both surfaces subscribe to the same application state. They must not maintain independent transfer state.

Pause All records the set of tasks it paused. Resume This Batch resumes only that set; it must not wake tasks that were already paused individually. Removing or manually resuming a task removes it from the batch.

Right-click opens a native menu with Open Rivu, Add Download, Pause All, Speed Limit Mode, Settings, and Quit. With no tasks, show an empty message and Add Download. With errors, show a count linking to Needs Attention.

### Add Download sheet

Cmd+O selects torrents. Cmd+N opens the sheet. Paste from the clipboard only in response to an explicit action; do not monitor the clipboard in the background.

Parse local torrents immediately. Magnets enter metadata retrieval first. Allow a magnet task to be queued before file metadata is available, but clearly show that the file list and size are pending.

When metadata is available, show the name, selected/total file count, selected size, file tree, destination, free space, and Start Immediately option. Default to `~/Downloads/Rivu`; confirm this location on the first transfer and allow per-task overrides.

Detect duplicates by torrent identity, not name. Focus the existing task for a duplicate infohash. Different torrents with the same name require destination conflict handling rather than automatic rejection or overwriting.

Enable submission only when source, selection, destination, and space checks permit the chosen action. If a destination is temporarily unavailable, allow creating a paused task only with an explicit explanation. At least one file must be selected when metadata is known.

Batch imports use one review step with individual errors; one invalid torrent must not fail the whole batch. On success, dismiss the sheet and select the new task without another success dialog.

### Settings and defaults

| Category      | Proposed defaults and controls                                                                                                |
| ------------- | ----------------------------------------------------------------------------------------------------------------------------- |
| General       | Follow system appearance; menu bar icon on; login launch off; background-only login launch on once login launch is enabled    |
| Downloads     | `~/Downloads/Rivu`; three concurrent downloads; restore previously running tasks on; completion notifications on              |
| Seeding       | Continue after completion; stop at ratio 2.0; allow unlimited or stop immediately; expose policy during import and in details |
| Bandwidth     | Normal mode unlimited; temporary mode 2 MB/s down and 512 KB/s up; editable in the native version                             |
| Connections   | Listening port, DHT/PEX respecting private torrents, diagnostics, progressively disclosed advanced options                    |
| Notifications | Completion and actionable failures on; show OS authorization separately                                                       |
| Updates       | Ask before enabling automatic checks; daily checks after consent; automatic installation off; manual check when supported     |

The prototype opens in light appearance for predictable review, although the product default is Follow System. It shows a post-consent automatic-check setting for demonstration.

Show real login-item registration results, including failure, system disablement, and authorization requirements. A local checkbox is not proof that the OS will launch the app. Disabling the menu bar icon must leave a Dock entry available.

Closing the main window keeps transfers running. Cmd+Q exits. An exit confirmation can explain interruption of active transfers and offer to remember the choice.

### Update experience

```text
Unavailable → Checking → Up to Date / Update Available / Check Failed
Update Available → Downloading and Installing → Restart Required / Install Failed
```

An unconfigured feed, development build, or unwritable installation location is an unavailable state, not “up to date.” Show the reason.

After consent, check at most once every 24 hours in the background with jitter and failure backoff. Manual checks bypass the schedule. Track the last successful check separately from attempts.

MyGo's current `Update.Install` combines download, verification, and replacement. Label the action “Download and Install Update”; do not promise an independently implemented download-only stage.

Automatic installation requires separate opt-in and waits until no downloading, verification, or seeding task is active. Allow explicit installation now. Seeding counts as activity, so explain why an update may be waiting.

Before restarting, persist destinations, file selections, queue positions, user intent, and pause reasons; stop scheduling, flush writes, and close the engine. Abort restart if state cannot be saved. Restore only tasks that were meant to run. Do not unexpectedly restart during transfers.

Distinguish invalid signature, wrong architecture, insufficient space, interrupted download, and installation failure. A corrupt update must not replace a runnable application. Configure a trusted HTTPS feed, signing keys, embedded public key, and macOS signing/notarization before shipping.

## Task state model

Separate observed engine state from desired running state. Otherwise a reconnect or restart can incorrectly resume a manually paused task.

| State               | Display                                            | Main actions                      |
| ------------------- | -------------------------------------------------- | --------------------------------- |
| Retrieving metadata | Retrieving file list; size unknown                 | Pause, remove                     |
| Queued              | Waiting for a slot; queue position                 | Start now, pause                  |
| Verifying           | Verification percentage                            | Pause, remove                     |
| Downloading         | Percentage, rates, estimated remaining time        | Pause, file selection             |
| Waiting for peers   | Waiting for available peers; no ETA                | Pause, inspect connections, retry |
| Paused              | Explain manual, policy, or shutdown pause          | Resume, remove                    |
| Seeding             | Complete, upload rate, ratio                       | Stop seeding, reveal files        |
| Completed           | Complete; seeding goal reached or stopped          | Resume seeding, reveal files      |
| Needs attention     | Disk full, permission denied, or disk disconnected | Choose location, retry, remove    |

Offline networking is recoverable environmental state, not a reason for repeated alerts. Reconnection and waking restore only tasks intended to run. Keep manual pauses intact.

Use decimal MB/GB consistently and rates per second. Progress measures verified bytes in selected files. Smooth rates before computing ETA; show a dash until an estimate is meaningful.

Removing a task retains data by default. A separate, unchecked “Also move downloaded files to Trash” option controls deletion. Validate ownership when tasks share destinations. If safe trashing fails, report it; never silently fall back to permanent deletion.

## Visual system

Use familiar macOS patterns, restrained chrome, and a quiet hierarchy. Task information should dominate. Avoid large statistics and decorative charts.

| Token     | Light     | Dark      | Role                         |
| --------- | --------- | --------- | ---------------------------- |
| Content   | `#FFFFFF` | `#202226` | Main surface                 |
| Sidebar   | `#EFF0F3` | `#27292E` | Material fallback            |
| Text      | `#20232A` | `#F1F2F5` | Primary text                 |
| Secondary | `#626975` | `#A8AEB9` | Secondary labels             |
| Accent    | `#0868DB` | `#5AA2FF` | Selection, actions, progress |
| Separator | `#E3E5E9` | `#3B3E45` | Structural dividers          |

Use semantic green for complete/seeding, orange for attention, and red for errors, always accompanied by text or an icon.

Use the macOS system font, SF Pro with PingFang SC fallback where appropriate. Target 13 pt body text, 11–12 pt secondary text, and 17–20 pt titles. Rates and byte counts use tabular numerals. The browser review compresses some secondary metadata; validate native readability at real point sizes.

Use a 4 pt spacing grid with 8/12/16/24 pt gaps, approximately 6 pt control corners, and 8 pt selection corners. Let macOS draw the popover shell. Respect Reduce Transparency and Reduce Motion.

Provide visible focus, accessible names for icon actions, non-hover access to essential operations, and VoiceOver task summaries. Avoid announcing continuously changing rates every second.

```text
┌─────────────────────────────────────────────────────────────────────┐
│ ● ● ●  Rivu       All Tasks     +   Pause   More          Search     │
├───────────────┬────────────────────────────────┬────────────────────┤
│ All Tasks   6 │ Ubuntu Desktop                 │ Ubuntu Desktop     │
│ Downloading 2 │ 3.68 GB / 5.90 GB      62.4%   │ Overview Files Net │
│ Seeding     1 │ ━━━━━━━━━━━━━━━─────           │ State / Rates / ETA│
│ Completed   2 │ ↓ 8.4 MB/s       About 4 min   │                    │
│ Paused      1 ├────────────────────────────────┤ Save location      │
│ Attention   1 │ Blender Open Movies            │                    │
│               │ …                              │ Ratio / Seed goal  │
│ Settings      │                                │                    │
├───────────────┴────────────────────────────────┴────────────────────┤
│ ↓ 12.6 MB/s   ↑ 1.2 MB/s                       Speed limit  Online  │
└─────────────────────────────────────────────────────────────────────┘
```

## Implementation with the existing stack

This mapping is based on the repository and locally installed MyGo v0.4.3 source, including its menu bar example, login APIs, and updater implementation.

| Requirement                  | Current support                                           | Proposed work                                                                                                                     |
| ---------------------------- | --------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| Windows, themes, menus       | Starter includes these                                    | Replace demonstration pages; retain `commands()` and window restoration                                                           |
| Menu bar panel               | MyGo `NewTray` and `NewFlyout`                            | Use `FlyoutOptions{Tray: tray, Popover: true, Focusable: true}` following `examples/menubar-flyout`                               |
| Tray clicks                  | Left and right clicks                                     | Do not permanently attach a menu that consumes `OnClick`; attach the right-click menu temporarily                                 |
| Login launch                 | `SetOpenAtLogin`, `OpenAtLogin`, `WasOpenedAtLogin`       | Verify registration in a packaged app; branch to background launch                                                                |
| Updates                      | `Updater.Enabled/Check`, `Update.Install`, `App.Relaunch` | Implement policies and UI; configure feed and key; prefer this updater over adding Sparkle                                        |
| Transfers after window close | Current starter does not satisfy this                     | Move engine, repository, and tray ownership to application scope                                                                  |
| BitTorrent engine            | None yet                                                  | Evaluate and pin `anacrolix/torrent`; verify license, protocol support, and recovery before integration                           |
| File/URL associations        | Not configured                                            | Register `.torrent` and `magnet:`; reuse single-instance routing; validate magnet separately from HTTP-only `desktop.ValidateURL` |

Suggested responsibilities:

```text
application
├── session.Service        Transfer intent, queue, commands, snapshots
│   ├── torrent.Engine     Small adapter around the selected library
│   ├── repository         Versioned metadata and recovery records
│   └── bandwidth          Global limits and concurrency policy
├── desktop.Service        Tray, login item, notifications, dialogs
├── updater.Service        Check/install policy and update state
├── mainWindow             List and inspector
├── trayPanel              Compact snapshot view
└── settingsWindow         Preferences and actual system state
```

These are proposed boundaries, not a requirement to create every package immediately. Keep modules as small as real needs permit and avoid a generic event bus. The current `internal/windowtask` cancellation model suits window operations, not long-lived torrent sessions.

Run engine work in the background and expose immutable snapshots. Coalesce UI updates to approximately 2–4 Hz in the main window and 1–2 Hz in the panel. Closing a view cancels its subscription, not its transfer.

Apply UI changes through `Window.Update`; keep native blocking operations behind `window.run`. Reject stale callbacks for closed windows.

Persist stable ID/infohash, source metadata, destination, file selection/priority, desired state, pause reason, queue position, uploaded totals, and errors. Do not serialize peer objects. Reuse atomic storage with schema migrations and separate recovery data as needed. Persist user changes promptly and throttle statistics. Revalidate disk contents after restart; cached percentages are not proof of integrity.

Use `~/Library/Application Support/Rivu` for application data. Update app identity and module branding during implementation. Preserve corrupt settings and offer recovery instead of overwriting them.

Validate torrent paths against traversal, absolute-path writes, and writes outside the destination. Handle symlinks, shared paths, existing files, changing free space, and disconnected drives. Respect private torrent discovery rules and redact tracker credentials from logs.

## Delivery stages and acceptance

1. **Native presentation:** list, inspector, add sheet, settings, and panel with sample state. Validate light/dark appearance, keyboard use, small windows, long names, empty and error states.
2. **Transfer loop:** engine, persistence, file selection, pause/resume, restart recovery, and seeding. Use authorized data and a controlled local swarm for transfer/interruption tests.
3. **Desktop integration:** close-window behavior, Dock reopening, login launch, associations, notifications, sleep/wake, and panel anchoring across displays.
4. **Release readiness:** signing/notarization, updates, persistence before restart, corrupt-update rejection, regression and performance validation.

Required acceptance cases:

- Closing the main window leaves transfers running; reopening creates no duplicate sessions or tasks. Quit ends the session.
- Login launch defaults off. Enabling, revoking, or disabling it through the OS is reflected accurately. Background login does not steal focus.
- Manual pause, batch pause, disconnection, sleep, and restart preserve running intent.
- Disk full, disconnected drives, permission denial, duplicates, missing peers, and corrupt torrents have recovery paths.
- Partial file selection produces correct progress and clear storage requirements.
- Automatic updates never force a restart during transfers. Failed checks do not become “up to date.” Invalid signatures leave the current app runnable.
- During native implementation, run `make check`, `make race`, `go tool mygo vet`, and `make snapshots`. Separately test packaged macOS tray behavior, login items, associations, notifications, and updates. Cross-compilation does not validate these integrations.

See [the prototype guide](../design/README.md) for supported review interactions and limitations.

## References

- [Apple MenuBarExtra](https://developer.apple.com/documentation/swiftui/menubarextra): interaction reference for data-rich panels; this does not require switching to SwiftUI.
- [Apple Toolbars](https://developer.apple.com/design/human-interface-guidelines/toolbars): organize controls by function and frequency.
- [Apple SMAppService](https://developer.apple.com/documentation/servicemanagement/smappservice): system login-item behavior.
- [MyGo v0.4.3 menu bar example](https://github.com/egoist/mygo/blob/v0.4.3/examples/menubar-flyout/main.go): checked against installed source.
- [MyGo login API](https://github.com/egoist/mygo/blob/v0.4.3/login.go) and [updater API](https://github.com/egoist/mygo/blob/v0.4.3/updater.go): the pinned version is implementation authority.
- [anacrolix/torrent](https://github.com/anacrolix/torrent): candidate Go engine; validate the selected version's behavior and performance.
