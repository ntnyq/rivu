# Architecture

Rivu is a native Go/MyGo application. The HTML under `docs/design` is the accepted visual reference, not the application runtime.

## Ownership

`application.go` owns a single session, preferences store, updater, tray and optional windows. Closing a window cancels its draft/background work without closing the session. Native menu actions and the command palette share stable command IDs in `menu.go`.

`window.go` owns view-local search, sort, selection, inspector and settings drafts. `view.go`, `view_inspector.go`, `settings*.go`, `add_window.go` and `tray.go` render immutable observations. `theme.go` translates the design reference's colors into MyGo theme values. SVG UI symbols originate in the approved local prototype; the product mark has its own editable vector source.

## Transfer state

`internal/session` serializes engine operations and state mutations behind a worker-side mutex. It publishes deep-copied snapshots atomically and notifies observers through coalescing capacity-one channels. Rendering does not acquire the session mutation lock. Measurements refresh every 500 ms; intent changes publish immediately after persistence.

Records contain durable run intent, pause-batch membership, file policy, metadata identity, destination, seed ratio, uploaded totals and completion epochs. Live engine phases never overwrite a manual pause. Engine handles are synchronous and discarded on stop, so a detached handle cannot publish a delayed callback into a replacement task.

The scheduler admits at most the configured number of downloading/verifying tasks and two unresolved metadata tasks. Seeding does not consume download slots. Payload tasks without useful byte progress and unresolved metadata tasks yield after 90 seconds and retry after one minute. Active verification and seeding are exempt. Admission is planned before handles change, so reprioritization also respects the metadata limit. Pausing detaches the torrent, closing peer activity. Resume reattaches and hashes existing pieces before trusting them.

`internal/torrent` adapts anacrolix/torrent v1.61.0. An unresolved magnet uses a separate client with DHT/PEX disabled; arbitrary metadata source URLs, web seeds and WebTorrent are disabled. Resolved private torrents retain the upstream library's private discovery policy.

Explicit magnet peers are retained when switching to the payload client and on
reattachment. Reprioritizing the queue assigns distinct ranks, so the latest
request takes precedence. File priority changes preserve completion epochs when
the selected file set remains unchanged.

## File boundary

Metainfo is limited to 16 MiB, member paths are validated, and conflicting/case-folded paths are rejected. `os.Root` anchors payload opens to the chosen directory. File ownership journals record the device/inode of newly created payloads; an unrelated replacement is not overwritten. The first version deliberately rejects existing target payloads instead of silently adopting them. New files are created under temporary names, journaled, then published with an exclusive rename; a failed journal write cannot strand an unclaimed payload at its final path. A crash before publication can leave an empty temporary file, retained conservatively.

The storage adapter handles piece/file overlap, including bytes in unselected files. Verified progress counts only selected bytes. Import space checks count the union of selected pieces intersected with actual file data, including shared-piece overlap but excluding v2 padding. Sources retain piece lengths and file offsets; older records without that geometry use a conservative full-size estimate. Every reattachment rehashes payload; no saved percentage is treated as integrity evidence.

Storage caches file geometry and searches for the files overlapping each I/O
range. Selected empty files are created through the ownership journal on
admission, without requesting neighboring pieces. Metadata validation rejects
malformed v2 roots and inconsistent hybrid piece counts before engine admission.

Default removal keeps payload. Trash first persists a removal record, detaches the engine, checks shared paths and file identity, then moves individual files through the native Trash API. Partial results are checkpointed. A failed move stays visible and is retried; permanent deletion is never used.

## Persistence and shutdown

Settings schema 2 and session schema 1 are separate JSON files. Metadata is written before its session reference. Same-directory temporary files are synced before atomic rename; the parent directory is then synced on macOS/Linux. An uncertain post-rename commit blocks writes until explicit reload. Files are private to the user. Upload totals checkpoint every 15 seconds; crash-time statistics can lag by that interval. Orphan metadata is retained conservatively.

Corrupt/future-schema files are preserved and block writes. The File menu exposes explicit reload after repair. A failed pause leaves the handle stopped with an error. A maintenance barrier blocks new admission during shutdown or update installation. Resume cannot release a persistence recovery hold; an explicit successful reload is required. Quit is intercepted synchronously, then saves/closes on a worker before requesting the final native quit.

Atomic replacement is not a transaction across settings, session, metadata and the filesystem. See the acceptance gaps in `implementation-status.md` before release.
