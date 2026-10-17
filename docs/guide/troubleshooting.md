# Troubleshooting

Start with the task's error message and the **Connections** tab. Fix the underlying condition before retrying.

## A magnet link is rejected or cannot resolve

Check that the link contains an explicit tracker or peer address. Rivu uses only those addresses until metadata confirms whether the torrent is private. A bare info hash without a tracker or peer cannot be resolved in that preview.

If the preview times out after 45 seconds, check your network connection and the link's source, then retry. You can cancel a preview or close the sheet without writing download payload to your destination.

## A task is waiting for peers

Waiting does not necessarily indicate an error. Open **Connections** and check peer count and each tracker's announce result. Correct network or tracker problems before retrying; an unavailable swarm cannot provide missing pieces.

Diagnostics refresh every five seconds. Tracker rows hide credentials, but review any diagnostics before sharing them.

## A destination contains existing files

For a new download, choose a location without conflicting payload files. To reuse existing data, enable **Adopt and verify existing files** in the add-download sheet. Rivu verifies those files when the task runs and repairs missing pieces.

Adopted files are kept when the task is removed. Rivu does not automatically adopt data during watch-folder imports.

## The download directory has changed

If a disk was disconnected or the data moved:

1. Make the disk available, or move the data to an accessible location.
2. Select the task and choose **Change Location…** in the inspector.
3. Select the directory containing the torrent's top-level name and confirm.
4. Retry or resume the task when ready.

Change Location does not move files from the previous directory. Data at the new location is verified when the task runs and retained when the task is removed. A manually paused task stays paused.

## A task does not resume with the batch

**Resume Paused Batch** restores the latest batch pause. It preserves tasks you had already paused individually. Resume those tasks using their own controls. Scheduled speed limits do not resume paused tasks either.

## Moving files to Trash failed

Check that the destination disk is available and writable, then follow the reported error. Rivu records completed removal steps and does not replace a failed Trash operation with permanent deletion. Adopted files cannot be moved to Trash through the task.

## Back up or repair a session

The default application-data directory is:

```text
~/Library/Application Support/Rivu
```

It contains `settings.json`, `session.json`, immutable torrent metadata, and file-ownership journals. The watch-folder ledger is stored beside the configured settings file. Download payload lives separately in each task's destination.

Quit Rivu before making a backup and copy the application-data directory along with any settings file stored elsewhere. Keep downloaded files if you want to continue those transfers later.

If a corrupt file or unsupported newer schema blocks loading, back up the affected files before repairing them. After repair, choose **File → Reload Session and Settings After Repair**. Avoid replacing a session with an empty file just to dismiss an error.

Restored transfers recheck actual data. Upload totals and cumulative seeding time are saved every 15 seconds, so a forced termination can lose the latest statistics interval. Session schema 1 migrates to schema 2; older builds that only support schema 1 cannot open the migrated session.

## Update checks are unavailable

A trusted release configuration is not yet supplied in the source tree. An unavailable update check does not mean the installed build is current. See [current availability](faq.md#where-can-i-download-rivu).

## Still need help?

Search [GitHub Issues](https://github.com/ntnyq/rivu/issues), then report your macOS version, Rivu version, steps to reproduce, and the exact error. Remove private paths and tracker credentials from screenshots or logs.
