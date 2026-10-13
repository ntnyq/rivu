# Refactor and feature validation

Date: October 13, 2026. Platform: macOS arm64 (Apple M1 Pro). This report covers the structure and P0–P2 feature recommendations from the project assessment.

## Delivered scope

| Recommendation | Implementation and behavior |
| --- | --- |
| Session responsibilities | Lifecycle, import, commands, validation, persistence, recovery, removal, observation and snapshots have separate files in one package. One Manager still owns serialized mutation and engine calls. |
| Window ownership | Main, settings and add windows have independent state; `windowBase` shares only native lifetime, status and cancellable work. |
| UI organization | Main regions and import rendering are separate; settings rendering, inputs and saving are separate. Root native entry points remain unchanged. |
| Typed commands, state and errors | Named command/phase/pause types, batch validation before side effects, structured error codes with deferred UI translation. |
| Documentation | Current index, architecture, implementation structure and status; original structure retained under `history`. |
| Magnet preview | Isolated metadata client, no payload admission or destination access, 45-second timeout, cancel/retry, authenticated file/name/size preview before submission. |
| Actionable failures | Current-language error details, task retry, directory relocation, and explicit session reload after storage repair. Waiting for peers has a separate connection hint. |
| Preference inputs | Decimal KB/s and MB/s; stop-on-completion, target ratio or unlimited seeding; optional cumulative seeding hours. |
| Existing data and relocation | Opt-in adoption records file identities without Trash authority. Reattachment hashes existing data before trusting it; damaged data is repaired on execution. Relocation does not move original files and does not resume a manually paused task. |
| File selection | Expand/collapse folders, search, paged rendering and atomic directory/file priority changes. Commands cannot skip every file. |
| Tracker diagnostics | Per-Tracker announce result, categorized error reason and engine-reported next announce time; refreshed every five seconds. HTTP/UDP state is parsed by a pinned-engine adapter, with local HTTP integration coverage. |
| Scheduled limits and seeding duration | Daily local-time interval including overnight; manual slow mode overrides it. Cumulative active seeding time checkpoints with session statistics; automatic policies preserve manual pauses. |
| Tags and watch folder | Durable task tags, inspector editing and toolbar filtering. Opt-in non-recursive `.torrent` folder import after two stable 10-second scans, with automatic tag and durable duplicate ledger. Source files remain intact. |
| Large collections | Stable-ID virtual task list, cached file-tree structure, 60 rendered file rows per page and repeatable benchmarks. |

## Data and recovery contracts

Session schema 1 is read and migrated to schema 2. Existing pause intent, task IDs, metadata references and ownership journals remain valid. Future/corrupt schemas block writes rather than resetting data. Added settings fields use defaults and retain settings schema 2.

Adoption grants permission to verify and repair the selected existing files; it does not grant permission to Trash them. The device/inode recorded at confirmation must still match before a write. New files retain the existing creation journal and exclusive-publication path. Relocation is a reference change: original files stay where they are, new-location observations are reset, and the next execution verifies the new location. If users want to move files, they move them externally and then select the new parent directory.

Watch-folder input is local and explicit. The watcher passes the exact stable bytes it read to the parser, bounds reads at 16 MiB, ignores symlinks/non-regular files, and does not automatically adopt existing payload. Invalid files are reported without blocking valid neighbors. The atomic watch ledger preserves decisions across restarts; a failed session save is retried and a corrupt ledger is never overwritten.

Structured error messages retain their translation key and arguments, not a language-specific formatted string. Session snapshots clone mutable slices, maps and error arguments. Changing tags does not clear an unrelated recovery hold.

## Measurements

Representative `go test -bench ... -benchmem -benchtime=500ms` results; these are headless Go/UI measurements, not hardware interaction guarantees:

| Workload | Before targeted optimization | After |
| --- | --- | --- |
| 1,000-task main view frame | 22.51 ms; 4.11 MB; 14,065 allocations | 0.885 ms; 3.77 MB; 2,177 allocations |
| 10,000-file tree structure | 3.357 ms; 2.59 MB; 10,881 allocations | Cached: 0.0214 ms; zero allocations |
| Snapshot of 1,000 tasks / 10,000 files | — | 0.369 ms; 1.52 MB; 2,001 allocations |

The snapshot cost did not justify splitting summary/detail contracts in this iteration. UI list construction did show a measurable cost and now uses MyGo's virtual List. The file-tree cache is invalidated by path, search, or expansion changes; progress and priority updates do not rebuild its structure.

Benchmarks: `BenchmarkMain1000Tasks`, `BenchmarkFileTree10000Files`, `BenchmarkCachedFileTree10000Files`, `BenchmarkSnapshot1000Tasks10000Files`.

## Validation

- Unit and native UI tests; Go vet and MyGo vet.
- Race detector across all packages.
- Isolated local swarm integration: v1/v2/hybrid integrity, magnet metadata, cancellation, adopted corrupt-file repair, selection across shared pieces, pause/restore, and HTTP Tracker diagnostics/redaction.
- Persistence/adoption/relocation/Trash tests, atomic batch validation, schema migration, scheduled interval boundaries, seeding duration, watch-folder stability/retry/restart, and error language refresh.
- UI screenshots in light/dark and existing five languages, plus download/automation settings captures. Virtual-list testing scrolls to the 1,000th task.
- Ad-hoc signed macOS arm64 app build; independent-data native captures of main, download settings and automation settings, each exiting through the save coordinator with exit code 0.
- Windows amd64 and Linux amd64 cross-compilation.

## Limits and remaining native acceptance

These changes do not replace the [desktop release checks](desktop.md): actual sleep/wake during transfer, external-volume replacement, disk-full/forced-termination injection, signed update replacement, VoiceOver, login authorization and multi-monitor tray behavior still require their documented hardware/system acceptance.

Unknown magnets still require an explicit Tracker or peer while their privacy status is unknown; DHT/PEX remain disabled until metadata is authenticated. Preview does not bypass that policy. Tracker diagnostics depend on the pinned engine's textual status contract; dependency upgrades must run the local tracker integration test. A format mismatch produces unavailable diagnostic rows instead of fabricated results.

Seeding/upload counters may lose up to one 15-second checkpoint interval after a crash. Watch folders are non-recursive and support one configured folder/tag. Windows payload ownership support remains unavailable even though the application cross-compiles. No signing/notarization, publishing or deployment credentials were introduced.
