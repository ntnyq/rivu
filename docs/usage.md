# Using Rivu

[Back to Rivu](../README.md) · [Build and run](development.md)

Rivu is a macOS BitTorrent downloader in active development. It supports `.torrent` files and magnet links, including v1, v2-only, and hybrid torrents.

## Add a download

Open **Add Download**, drop a torrent file into the main window, or open a torrent file or magnet link with Rivu. Review the source and choose a destination. When metadata is available, select the files you want before starting. An unresolved magnet can be added as a pending task with all files selected; live metadata preview inside the import sheet is not yet available.

Use an empty destination for a new download. Rivu does not automatically adopt existing files that conflict with the torrent's payload.

Unresolved magnets use only their explicit trackers or peers until metadata establishes whether the torrent is private. A magnet without either is rejected with an explanation. Rivu does not fetch arbitrary metadata-source URLs or web seeds.

## Manage transfers

Use the sidebar to filter downloading, seeding, completed, paused, or attention-needed tasks. Search and sort the list, and select a transfer to inspect its progress, files, and connections.

File priorities can be high, normal, or skipped. Files may share torrent pieces, so a skipped file can still contain bytes needed to complete an adjacent selected file.

Set concurrent-download limits, bandwidth limits, and share-ratio targets in Settings. Pause or resume individual transfers, or use **Pause All** and **Resume Paused Batch**. Resuming a batch leaves tasks that were already manually paused alone. Manual pauses and the most recent paused batch survive restarts.

Closing the main window keeps the application and its transfers running. Use the menu bar panel to check speeds, pause transfers, toggle speed limits, or reopen the main window. Quitting Rivu saves the session and stops transfers.

Removing a task keeps its files by default. Moving files to Trash is a separate option and applies only to files recorded as owned by that transfer. If a Trash operation partly fails, Rivu records that progress and does not fall back to permanent deletion.

## Appearance and language

Rivu offers light and dark appearances and supports Simplified Chinese, Traditional Chinese, English, Japanese, and Korean.

The default follows the system's preferred language, falling back to English when necessary. Change **Settings → General → Language** and save to update the interface and native menus immediately. System dialogs and diagnostic details from other software may use their own language.

## Storage and recovery

On macOS, Rivu normally stores application data in `~/Library/Application Support/Rivu` (`os.UserConfigDir()/Rivu`). This directory contains `settings.json`, `session.json`, immutable torrent metadata, and file-ownership journals. Downloaded files live in the destination selected for each task.

Restored transfers are rehashed: completion is verified from the actual data instead of trusting saved percentages. Upload totals are checkpointed every 15 seconds, so forced termination can lose the latest statistics interval.

Corrupt files or files with a newer, unsupported schema block changes and remain untouched. Back up the affected files before repairing them, then choose **File → Reload Session and Settings After Repair**. Settings version 1 can be migrated when explicitly loaded; Rivu does not alter another product's configuration directory.

For alternate session or settings locations, see [launch options](development.md#launch-options).

## Current availability

The repository provides a development implementation and local macOS app packaging. Production signing, notarization, and a trusted update feed are not configured in the source tree. Update checks report unavailable until a trusted release configuration is supplied.

Windows and Linux builds are compilation checks, not supported downloadable releases. Native permissions, multi-monitor menu bar behavior, accessibility, forced-termination recovery, and signed-update replacement still have release acceptance work documented in [desktop verification](desktop.md) and [implementation status](implementation-status.md).

The configured product domain does not by itself indicate that website hosting or DNS has been deployed. See [release guidance](releasing.md) for distribution status and requirements.
