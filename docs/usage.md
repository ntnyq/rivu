# Using Rivu

[Documentation home](index.md) · [Build and run](development.md)

Rivu is a macOS BitTorrent downloader in active development. It supports `.torrent` files and magnet links, including v1, v2-only, and hybrid torrents.

## Add a download

Open **Add Download**, drop a torrent file into the main window, or open a torrent file or magnet link with Rivu. Review the source and choose a destination. Select **Read and Review** to resolve metadata, then inspect the name, sizes and file tree before adding the task. Magnet previews never download payload to your destination, time out after 45 seconds, and can be canceled or retried. Closing the sheet cancels its preview.

Use an empty destination for a new download. To reuse existing data, explicitly select **Adopt and verify existing files**. Rivu checks their contents when the task runs and repairs missing or damaged pieces. Adopted files never gain Trash permission; task removal retains them. Without this option, conflicting payloads are rejected.

Unresolved magnets use only their explicit trackers or peers until metadata establishes whether the torrent is private. A magnet without either is rejected with an explanation. Rivu does not fetch arbitrary metadata-source URLs or web seeds.

## Manage transfers

Use the sidebar to filter downloading, seeding, completed, paused, or attention-needed tasks. Search and sort the list, and select a transfer to inspect its progress, files, and connections.

The file tree supports folder expansion, search, and directory-wide priority changes. Large file lists use pages. File priorities can be high, normal, or skipped, but at least one file must remain selected. Files may share torrent pieces, so a skipped file can still contain bytes needed to complete an adjacent selected file.

Set concurrent-download limits and bandwidth limits in **Settings → Downloads & Seeding**. Rate inputs use KB/s or MB/s; a normal limit of 0 means unlimited. Choose stop-on-completion, a target share ratio, or unlimited seeding, with an optional cumulative active-seeding-hour limit. The inspector can override a task's seeding policy. Pause or resume individual transfers, or use **Pause All** and **Resume Paused Batch**. Resuming a batch leaves tasks that were already manually paused alone. Manual pauses and the most recent paused batch survive restarts.

Closing the main window keeps the application and its transfers running. Use the menu bar panel to check speeds, pause transfers, toggle speed limits, or reopen the main window. Quitting Rivu saves the session and stops transfers.

Removing a task keeps its files by default. Moving files to Trash is a separate option and applies only to files recorded as owned by that transfer. If a Trash operation partly fails, Rivu records that progress and does not fall back to permanent deletion.

## Tags, schedules and watch folders

Edit comma-separated tags in the selected task's inspector, then use the toolbar's tag filter. New imports can receive tags in their review sheet.

In **Settings → Automation**, enable a daily local-time limit interval using `HH:MM`. Overnight intervals are supported. During the interval Rivu uses the temporary download/upload limits. Manual slow mode can keep those limits active outside the interval; disabling manual slow mode does not disable an active schedule. Disable the schedule in settings to return to normal limits during that interval. Scheduling never resumes manually paused tasks.

Enable **Watch folder**, enter an absolute folder path and optional automatic tag, then save. Rivu checks regular `.torrent` files in that folder every 10 seconds, waits for identical content in two scans, and automatically imports to the default download directory. It does not scan subfolders, delete source torrents, or adopt conflicting payload automatically. A durable ledger prevents unchanged files from reappearing after task removal or app restart. Invalid files remain available for correction without blocking other imports.

## Appearance and language

Rivu offers light and dark appearances and supports Simplified Chinese, Traditional Chinese, English, Japanese, and Korean.

The default follows the system's preferred language, falling back to English when necessary. Change **Settings → General → Language** and save to update the interface and native menus immediately. System dialogs and diagnostic details from other software may use their own language.

## Storage and recovery

On macOS, Rivu normally stores application data in `~/Library/Application Support/Rivu` (`os.UserConfigDir()/Rivu`). This directory contains `settings.json`, `session.json`, immutable torrent metadata, and file-ownership journals. The watch-folder ledger lives beside the configured settings file. Downloaded files live in the destination selected for each task.

Restored transfers are rehashed: completion is verified from the actual data instead of trusting saved percentages. Upload totals and cumulative seeding time are checkpointed every 15 seconds, so forced termination can lose the latest statistics interval.

Corrupt files or files with a newer, unsupported schema block changes and remain untouched. Back up the affected files before repairing them, then choose **File → Reload Session and Settings After Repair**. Settings version 1 can be migrated when explicitly loaded; Rivu does not alter another product's configuration directory.

A task error offers a retry after the underlying condition is repaired. If a disk or destination changed, use **Change Location…** in its inspector, pick the directory containing the torrent's top-level name, and confirm. This does not move files from the old directory. Existing data at the new location is verified on execution and retained on removal. A manually paused task stays paused.

The Connections tab shows peer count and per-Tracker announce state, categorized failures and the next engine-reported announce time. Credentials are hidden from diagnostic rows. Waiting for peers can be normal; check the network and tracker results before retrying. Diagnostic state refreshes every five seconds.

Session schema 1 migrates to schema 2 while preserving tasks and manual pause intent. Older application builds that only understand schema 1 cannot open the migrated session.

For alternate session or settings locations, see [launch options](development.md#launch-options).

## Current availability

The repository provides a development implementation and local macOS app packaging. Production signing, notarization, and a trusted update feed are not configured in the source tree. Update checks report unavailable until a trusted release configuration is supplied.

Windows and Linux builds are compilation checks, not supported downloadable releases. Native permissions, multi-monitor menu bar behavior, accessibility, forced-termination recovery, and signed-update replacement still have release acceptance work documented in [desktop verification](desktop.md) and [implementation status](implementation-status.md).

The configured product domain does not by itself indicate that website hosting or DNS has been deployed. See [release guidance](releasing.md) for distribution status and requirements.
