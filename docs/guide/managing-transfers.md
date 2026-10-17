# Manage transfers

## Find a task

Use the sidebar to filter downloading, seeding, completed, paused, or attention-needed tasks. Search and sort the queue, or use the tag filter to narrow it further.

Select a task to inspect its progress, file tree, and connections. The Connections tab shows peer count, tracker announce results, failures, and the next announce time when the engine provides one. Diagnostics refresh every five seconds.

## Pause and resume

Pause or resume one task from its controls. To stop a group temporarily, choose **Pause All**, then **Resume Paused Batch** when you are ready.

Resuming a batch preserves tasks you had already paused individually. Manual pauses and the most recent paused batch survive app restarts.

## Set bandwidth and concurrency

Open **Settings → Downloads & Seeding** to set upload and download limits and the maximum number of concurrent downloads. Normal bandwidth limits accept KB/s or MB/s; **0 means unlimited**.

Use the speed-limit toggle to apply temporary limits. For limits that follow your daily routine, see [scheduled limits](automation.md#schedule-speed-limits).

## Choose a seeding policy

Seeding shares downloaded pieces with other peers. Choose a default policy in **Settings → Downloads & Seeding**, or override it for a task in the inspector.

| Policy              | Behavior                                                       |
| ------------------- | -------------------------------------------------------------- |
| Stop on completion  | Stop when the selected download completes.                     |
| Target share ratio  | Seed until the configured upload-to-download ratio is reached. |
| Unlimited ratio     | Continue seeding without a share-ratio limit.                  |
| Optional time limit | Stop after a cumulative amount of active seeding time.         |

The time limit can be used alongside a ratio policy. It counts active seeding time rather than elapsed time since import.

## Keep downloading in the background

Closing the main window leaves Rivu and its transfers running. Use the menu bar panel to view speeds and active tasks, pause transfers, toggle limits, add a download, or reopen the window.

Quitting Rivu saves the session and stops transfers. On restoration, downloaded data is verified instead of trusting a saved percentage.

## Remove a task

Removing a task **keeps downloaded files by default**. Moving files to Trash is a separate option, available only for files recorded as owned by that transfer. Adopted files remain on disk.

If moving files to Trash partly fails, Rivu records the completed steps and reports the problem. It does not fall back to permanent deletion. See [troubleshooting](troubleshooting.md) before retrying a failed removal.
