# Architecture

Rivu is a native Go/MyGo application. The HTML under `docs/design` is the accepted visual reference, not the application runtime.

## Ownership

`application.go` owns a single session, preferences store, updater, tray and optional windows. Closing a window cancels its draft/background work without closing the session. Native menu actions and the command palette share stable command IDs in `menu.go`.

`mainWindow`, `settingsWindow`, and `addWindow` own separate view state. Only native lifetime, status, and cancellable asynchronous work live in `windowBase`. Main list regions live in `view_*.go`; settings rendering and persistence are separate. All transfer views render immutable observations. `theme.go` translates the design reference's colors into MyGo theme values. SVG UI symbols originate in the approved local prototype; the product mark has its own editable vector source.

## Transfer state

`internal/session` serializes engine operations and state mutations behind a worker-side mutex. It publishes deep-copied snapshots atomically and notifies observers through coalescing capacity-one channels. Rendering does not acquire the session mutation lock. Measurements refresh every 500 ms; intent changes publish immediately after persistence.

Records contain durable run intent, pause-batch membership, file policy, metadata identity, destination, seed ratio, uploaded totals and completion epochs. Live engine phases never overwrite a manual pause. Engine handles are synchronous and discarded on stop, so a detached handle cannot publish a delayed callback into a replacement task.

The scheduler admits at most the configured number of downloading/verifying tasks and two unresolved metadata tasks. Seeding does not consume download slots. Payload tasks without useful byte progress and unresolved metadata tasks yield after 90 seconds and retry after one minute. Active verification and seeding are exempt. Admission is planned before handles change, so reprioritization also respects the metadata limit. Pausing detaches the torrent, closing peer activity. Resume reattaches and hashes existing pieces before trusting them.

`internal/torrent` adapts anacrolix/torrent v1.61.1-0.20260927071845-d913b30f520e, a fixed development commit. An unresolved magnet uses a separate client with DHT/PEX disabled; arbitrary metadata source URLs, web seeds and WebTorrent are disabled. Resolved private torrents retain the upstream library's private discovery policy.

Explicit magnet peers are retained when switching to the payload client and on
reattachment. Reprioritizing the queue assigns distinct ranks, so the latest
request takes precedence. File priority changes preserve completion epochs when
the selected file set remains unchanged.

## File boundary

Metainfo is limited to 16 MiB, member paths are validated, and conflicting/case-folded paths are rejected. `os.Root` anchors payload opens to the chosen directory. File ownership journals record the device/inode of newly created payloads; an unrelated replacement is not overwritten. Existing payloads are rejected by default. Explicit adoption records their device/inode separately from creation ownership, allows verification and repair, and never grants Trash authority. Relocation references an existing directory without moving original files, resets observations, and preserves paused intent. Every attachment rehashes the new location. New files are created under temporary names, journaled, then published with an exclusive rename; a failed journal write cannot strand an unclaimed payload at its final path. A crash before publication can leave an empty temporary file, retained conservatively.

The storage adapter handles piece/file overlap, including bytes in unselected files. Verified progress counts only selected bytes. Import space checks count the union of selected pieces intersected with actual file data, including shared-piece overlap but excluding v2 padding. Sources retain piece lengths and file offsets; older records without that geometry use a conservative full-size estimate. Every reattachment rehashes payload; no saved percentage is treated as integrity evidence.

Storage caches file geometry and searches for the files overlapping each I/O
range. Selected empty files are created through the ownership journal on
admission, without requesting neighboring pieces. Metadata validation rejects
malformed v2 roots and inconsistent hybrid piece counts before engine admission.

Default removal keeps payload. Trash first persists a removal record, detaches the engine, checks shared paths and file identity, then moves individual files through the native Trash API. Partial results are checkpointed. A failed move stays visible and is retried; permanent deletion is never used.

## Persistence and shutdown

Settings schema 2 and session schema 2 (migrated from schema 1) are separate JSON files. Metadata is written before its session reference. Same-directory temporary files are synced before atomic rename; the parent directory is then synced on macOS/Linux. An uncertain post-rename commit blocks writes until explicit reload. Files are private to the user. Upload totals checkpoint every 15 seconds; crash-time statistics can lag by that interval. Orphan metadata is retained conservatively.

Corrupt/future-schema files are preserved and block writes. The File menu exposes explicit reload after repair. A failed pause leaves the handle stopped with an error. A maintenance barrier blocks new admission during shutdown or update installation. Resume cannot release a persistence recovery hold; an explicit successful reload is required. Quit is intercepted synchronously, then saves/closes on a worker before requesting the final native quit.

Atomic replacement is not a transaction across settings, session, metadata and the filesystem. See the acceptance gaps in `implementation-status.md` before release.

## Feature modules

Session lifecycle remains in `manager.go`. `commands.go` validates complete batches before side effects; `import.go`, `persistence.go`, `recovery.go`, `removal.go`, `observation.go`, and `snapshot.go` isolate their respective responsibilities under the same mutex. `CommandKind`, `Phase`, and `PauseReason` have named constants. `internal/problem` carries language-independent failure codes and context; UI recovery text is translated when rendered.

`internal/torrent/preview.go` uses an isolated client and temporary storage root with payload admission permanently disabled. Previews time out after 45 seconds and close/cancel with their add window. Confirming the draft persists authenticated metadata and the selected file priorities.

Tags are durable task attributes, editable in the inspector and filterable from the toolbar. File trees cache path structure, search and expand folders, and apply directory priorities as one validated command. Only 60 file rows are rendered per page. Main transfers use MyGo's virtual `List` with stable task IDs.

`internal/config/schedule.go` computes daily local-time bandwidth limits, including overnight intervals. Manual slow mode remains an independent override. Seeding duration is cumulative active seeding time, checkpointed with session statistics; ratio or duration can stop a task. Explicit resume overrides a reached seeding goal, while scheduled limits never alter run intent.

`internal/watchfolder` reads at most 16 MiB from regular `.torrent` files, requires identical content in two 10-second scans, and passes those exact bytes to the parser. A separate atomic ledger prevents reimporting unchanged files after restart or task removal. A bad torrent does not block other files. Watching is opt-in, uses the default destination and configured tag, leaves source files intact, and never adopts pre-existing payloads automatically.

HTTP/UDP Tracker results and retry times are read from the pinned engine's `WriteStatus` table every five seconds. The version-specific parser is isolated in `internal/torrent/diagnostics.go` and verified against a local HTTP tracker. Raw diagnostic output, URL credentials, paths and query strings are never exposed in diagnostic rows. Original tracker URLs remain in source metadata so copied magnet links can retain private tracker credentials. Failure messages are reduced to DNS, timeout, refusal, authorization, or announce categories. An unknown output format leaves diagnostics unavailable instead of inventing state; dependency upgrades must run the integration contract test.
