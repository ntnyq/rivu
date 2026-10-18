---
search: false
---

# Rivu Implementation File Structure

Date: October 12, 2026. Status: proposed implementation structure.

Rivu will keep the existing Go 1.27.2 and MyGo v0.4.3 native stack. The root package composes application lifecycle, windows, views, and commands; `internal` owns transfer behavior, persistence, and native integration. The target below translates [the product design](product-design.md) into implementation files. It is a plan, not a description of files already implemented.

Read [the implementation plan](implementation-plan.md) for contracts, state transitions, sequencing, and acceptance criteria. Production views are Go code.

## Target source tree

Create these files as their corresponding feature lands. Tests sit beside implementations; the tree lists representative tests rather than requiring a test for every file.

```text
rivu/
├── AGENTS.md
├── go.mod                            # Keep pinned Go/MyGo; add validated torrent version
├── go.sum
├── mygo.json                         # Rivu identity, associations, packaging, update config
├── Makefile                          # Existing targets plus controlled integration tests
├── main.go                           # CLI bootstrap, single instance, MyGo event registration
├── application.go                    # App-owned session, tray, updater, window references
├── application_test.go               # Lifecycle coordination with controlled dependencies
├── lifecycle.go                      # Startup, login launch, quit/restart coordination
├── launch.go                         # Funnel CLI, file, URL, and second-instance imports
├── cli.go                            # Multiple torrent paths, magnets, launch flags
├── cli_test.go
├── commands.go                       # commands(), stable command IDs, enablement, dispatch
├── menu.go                           # Native menu and Dock menu from shared commands
├── window.go                         # Main window, local selection, subscription, window.run
├── window_actions.go                 # Capture UI inputs and dispatch session commands
├── view.go                           # Main window composition and responsive inspector
├── view_sidebar.go                   # Filters and overlapping counts
├── view_toolbar.go                   # Add, selection actions, search, inspector toggle
├── view_task_list.go                  # Stable task rows and multiple selection
├── view_inspector.go                  # Overview, file tree, trackers and peers
├── view_statusbar.go                  # Rates, limit mode, network condition
├── view_palette.go                   # Searchable shared command registry
├── view_test.go                      # Native presentation behavior and snapshots
├── add_window.go                     # Parent-modal add flow and cancellable draft work
├── add_view.go                       # Sources, metadata, destination, files, batch errors
├── remove_window.go                  # Keep-data default, explicit Trash option
├── settings_window.go                # Single settings window and local preference draft
├── settings_view.go                  # Settings navigation and common layout
├── settings_general.go               # Theme, tray, login launch and actual OS result
├── settings_downloads.go             # Directory, concurrency, bandwidth and seeding
├── settings_connections.go           # Port, DHT, PEX and diagnostics
├── settings_notifications.go         # Preferences and observed authorization status
├── settings_updates.go               # Update states, release notes, install/restart actions
├── tray.go                           # App-owned tray, left/right click and panel lifecycle
├── tray_view.go                      # At most four task rows from the shared snapshot
├── presentation.go                   # Shared filter/count/formatting and task display rules
├── presentation_test.go              # Progress, ETA, counts and stable sorting
├── theme.go                          # Semantic colors, spacing and typography tokens
├── icons.go                          # App symbols and monochrome template tray icon
├── capture.go                        # Native startup capture, updated for Rivu
├── internal/
│   ├── session/
│   │   ├── model.go                  # Task IDs, identity aliases, intent, state, file policy
│   │   ├── manager.go                # Serialized ownership of the transfer session
│   │   ├── command.go                # Typed commands and per-task operation results
│   │   ├── ports.go                  # Engine and import-parser contracts consumed here
│   │   ├── snapshot.go               # Immutable snapshots and coalesced subscriptions
│   │   ├── import.go                 # Drafts, deduplication, review and durable submission
│   │   ├── scheduler.go              # Download slots, verification and metadata limits
│   │   ├── pause.go                  # Individual and batch pause ownership
│   │   ├── bandwidth.go              # Normal/temporary limits and effective rate policy
│   │   ├── seeding.go                # Completion, ratio policy and explicit manual override
│   │   ├── recovery.go               # Restore intended state, verify data, handle missing disks
│   │   ├── persistence.go            # Versioned session JSON, ordered writes and tombstones
│   │   ├── remove.go                 # Stop engine, validate ownership, coordinate Trash
│   │   ├── manager_test.go
│   │   ├── import_test.go
│   │   ├── scheduler_test.go
│   │   ├── pause_test.go
│   │   ├── recovery_test.go
│   │   └── persistence_test.go
│   ├── torrent/
│   │   ├── engine.go                 # anacrolix adapter; library objects stay here
│   │   ├── source.go                 # Parse torrents/magnets, normalize supported identities
│   │   ├── metadata.go               # Metadata-only draft handles and their cleanup
│   │   ├── storage.go                # Per-task destinations and restricted file access
│   │   ├── observation.go            # Translate engine measurements into session types
│   │   ├── engine_test.go
│   │   └── engine_integration_test.go # Controlled local swarm, gated integration suite
│   ├── config/
│   │   ├── model.go                  # Typed preference groups and defaults
│   │   ├── store.go                  # Existing validated copy/write/publish behavior
│   │   ├── migrate.go                # Explicit starter-to-Rivu schema migration
│   │   └── store_test.go
│   ├── desktop/
│   │   ├── desktop.go                # Existing dialog/clipboard/browser/Finder entry points
│   │   ├── dialogs.go                # Multiple torrents and destination selection
│   │   ├── login.go                  # Registration, boolean OS observation and errors
│   │   ├── login_status_darwin.go    # Packaged macOS status detail when required
│   │   ├── login_status_other.go     # Explicit unsupported/unknown detail
│   │   ├── notifications.go          # Authorization, stable event IDs, action routing
│   │   ├── power.go                  # Sleep/wake subscriptions and environment observations
│   │   ├── trash.go                  # MyGo TrashItem wrapper; no permanent-delete fallback
│   │   └── desktop_test.go
│   ├── files/
│   │   ├── destination.go            # Destination validation and conflict detection
│   │   ├── paths.go                  # Root confinement and symlink handling
│   │   ├── ownership.go              # Task file ownership manifests and safe removal plans
│   │   ├── space_darwin.go           # Free space query for the first target platform
│   │   ├── space_other.go            # Supported alternatives or explicit unavailable result
│   │   └── paths_test.go
│   ├── updater/
│   │   ├── service.go                # App-owned update state and serialized operations
│   │   ├── policy.go                 # Consent, daily schedule, backoff and install eligibility
│   │   ├── mygo.go                   # Adapter over Enabled, Check and Install
│   │   └── policy_test.go
│   ├── storage/
│   │   ├── atomic.go                 # Reuse atomic same-volume file replacement
│   │   └── atomic_test.go
│   └── tasks/
│       ├── task.go                   # Existing one-shot UI job cancellation/generation guard
│       └── task_test.go
├── resources/
│   └── icon.png                      # Rivu application icon
├── tools/genicon/main.go
├── testdata/
│   ├── torrents/                     # Tiny, locally generated legal fixtures
│   └── sessions/                     # Old schemas, corrupt records and recovery fixtures
├── docs/
│   ├── architecture.md               # Current implementation; update as code lands
│   ├── desktop.md                    # Current native support; update as code lands
│   ├── releasing.md                  # Bundle signing, update signing, notarization, smoke tests
│   ├── product-design.md
│   ├── implementation-structure.md   # This proposed tree
│   └── implementation-plan.md        # Contracts, behavior, milestones and verification
├── build/                            # Ignored packaging output
└── snapshots/                        # Ignored native UI snapshots
```

## Ownership and import rules

| Module      | Owns                                                             | Must not own                                         |
| ----------- | ---------------------------------------------------------------- | ---------------------------------------------------- |
| Root `main` | MyGo lifecycle, windows, tray handles, commands and presentation | Torrent internals or a second copy of transfer state |
| `session`   | Intent, queue, stable task records, orchestration and snapshots  | MyGo objects or anacrolix objects                    |
| `torrent`   | Protocol execution, source parsing and engine storage adapter    | Window lifetime or product pause/queue policy        |
| `config`    | Preferences, validation and schema migration                     | Transfer counters and active engine state            |
| `desktop`   | Native calls, capability/status observations                     | Scheduling or independent transfer control           |
| `files`     | Destination safety, free space and ownership rules               | UI state or protocol execution                       |
| `updater`   | Update lifecycle and policy                                      | Direct task mutation or unconditional process exit   |
| `storage`   | Atomic file primitives                                           | Product schemas                                      |
| `tasks`     | Short UI operations and stale callback suppression               | Long-lived torrent sessions                          |

`session` defines the engine contracts it consumes. `torrent` imports those ordinary Go types and implements the contracts; `session` never imports `torrent`. `main` constructs the implementations and passes them into `session`. A deterministic in-memory engine in session tests is the second adapter at this seam. Avoid a generic event bus, a dependency injection framework, and packages that only forward calls.

`session` may import `config`, `files`, and `storage`; these packages must not import `session`. `updater` receives activity observations and restart coordination from `main`, so it need not import the concrete session manager. Shared view helpers remain in `main` until an actual second package needs them.

## Mapping screens to files

| Product surface                 | View and controller                        | Main behavior owner                                           |
| ------------------------------- | ------------------------------------------ | ------------------------------------------------------------- |
| Main window                     | `window.go`, `view*.go`, `presentation.go` | `session.Manager`                                             |
| Overview, files and connections | `view_inspector.go`                        | Session snapshots plus selected-task detail observation       |
| Menu bar panel                  | `tray.go`, `tray_view.go`                  | Same session and shared command registry                      |
| Add Download                    | `add_window.go`, `add_view.go`             | `session/import.go`, `torrent/source.go`, `files`             |
| Remove confirmation             | `remove_window.go`                         | `session/remove.go`, `files/ownership.go`, `desktop/trash.go` |
| Settings                        | `settings_*.go`                            | `config`, session policy, desktop status and updater          |
| Update and restart              | `settings_updates.go`, `lifecycle.go`      | `updater` and application shutdown coordinator                |

## Migration from the starter

| Existing file or behavior                 | Implementation change                                                                                                             |
| ----------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| `main.go`                                 | Keep startup and event registration; move application ownership into `application.go`; route file/URL opens to import review      |
| `window.go`                               | Move application, tray and global shortcut ownership out; replace document/checksum state with local task selection and snapshots |
| `view.go`                                 | Replace five demo pages with the task layout; split substantial screen sections into the listed root files                        |
| `menu.go`                                 | Replace index-based action lookup with stable IDs in `commands()`; menus, palette and tray use the same handlers                  |
| `internal/files/files.go`                 | Remove text-preview/checksum examples after their views and tests are replaced; add destination/ownership behavior                |
| `internal/tasks/task.go`                  | Retain for dialogs and draft work; do not expand it into a multi-transfer engine                                                  |
| `internal/config/store.go`                | Preserve corrupt-file protection and atomic publication; migrate theme and introduce typed Rivu preferences                       |
| `internal/storage/atomic.go`              | Retain and document durability limitations; add durability changes only where recovery requirements need them                     |
| `mygo.json`, imports, capture variables   | Adopt Rivu identity together; proposed module `github.com/ntnyq/rivu`, bundle ID `com.ntnyq.rivu`, environment prefix `RIVU_`     |
| `docs/architecture.md`, `docs/desktop.md` | Keep accurate for current code until each implementation change lands; do not describe this target as already working             |

## Runtime data tree

The default macOS data root is `~/Library/Application Support/Rivu`. Other targets derive the root from the OS configuration/data conventions. Downloads remain outside application metadata.

```text
Application Support/Rivu/
├── settings.json                     # Preference schema, consent, normal/temporary limits
├── session.json                      # Authoritative task records, batch pause sets, operation state
├── metadata/
│   └── <task-id>.torrent              # Verified immutable metainfo, including resolved magnets
└── engine/
    └── <task-id>/                     # Versioned disposable engine resume/piece cache

~/Downloads/Rivu/
└── <torrent-root>/                   # User payload; conflicts checked before attaching storage
```

Window position remains under MyGo's window-state mechanism. Transfer progress caches never substitute for data verification. Authentication-bearing magnets and tracker URLs belong only in permission-restricted records and must be redacted from logs. No payload belongs in `docs`, `resources`, `build`, or `snapshots`.
