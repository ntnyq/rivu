# Troubleshooting

If Rivu opens but a transfer fails, start with the task's error message and the **Connections** tab. Fix the underlying condition before retrying.

## macOS blocks Rivu from opening

Current macOS downloads do not have Apple Developer ID signing or Apple notarization. If macOS cannot verify the developer or check the app for malicious software, follow [Open Rivu when macOS blocks it](getting-started.md#open-rivu-when-macos-blocks-it) after confirming the download's source. The guide covers **Open Anyway** on macOS 13 or later and the corresponding macOS 12 settings.

If the verification alert offers only **Move to Trash** and **Done**, click **Done** to keep Rivu. Then go to **System Settings → Privacy & Security**, scroll to **Security**, and click **Open Anyway** beside the message about Rivu. Authenticate if prompted, then click **Open** in the confirmation dialog.

## A magnet link is rejected or cannot resolve

Check that the link contains an explicit tracker or peer address. Rivu uses only those addresses until metadata confirms whether the torrent is private. A bare info hash without a tracker or peer cannot be resolved in that preview.

If the preview times out after 45 seconds, check your network connection and the link's source, then retry. You can cancel a preview or close the sheet without writing download payload to your destination.

## A task is waiting for peers

Waiting does not necessarily indicate an error. Open **Connections** and check peer count and each tracker's announce result. Correct network or tracker problems before retrying; an unavailable swarm cannot provide missing pieces.

Diagnostics refresh every five seconds. Tracker rows hide credentials, but review any diagnostics before sharing them.

On Windows, check whether your proxy client is using a system proxy or TUN. HTTP/HTTPS trackers and TCP peers can use the system proxy, while UDP trackers and DHT require an available OS route or a TUN that carries UDP. See [system proxy and TUN](settings.md#system-proxy-and-tun). A task with no progress for 90 seconds yields its download slot and waits one minute before retrying, so **Waiting for peers** can temporarily change to **Queued**.

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

The default application-data directory depends on your system: `~/Library/Application Support/Rivu` on macOS, `%AppData%\Rivu` on Windows, and `$XDG_CONFIG_HOME/Rivu` (or `~/.config/Rivu`) on Linux. See [data locations](settings.md#where-your-data-lives).

It contains `settings.json`, `session.json`, immutable torrent metadata, and file-ownership journals. The watch-folder ledger is stored beside the configured settings file. Download payload lives separately in each task's destination.

Quit Rivu before making a backup and copy the application-data directory along with any settings file stored elsewhere. Keep downloaded files if you want to continue those transfers later.

If a corrupt file or unsupported newer schema blocks loading, back up the affected files before repairing them. After repair, choose **File → Reload Session and Settings After Repair**. Avoid replacing a session with an empty file just to dismiss an error.

Restored transfers recheck actual data. Upload totals and cumulative seeding time are saved every 15 seconds, so a forced termination can lose the latest statistics interval. Session schema 1 migrates to schema 2; older builds that only support schema 1 cannot open the migrated session.

## Update checks are unavailable

Use a release build installed in a writable location. Development mode, older builds without update configuration, and read-only installations cannot update themselves. On macOS, copy Rivu out of the DMG into Applications. On Linux, update a DEB installation through the package manager. Install an update-enabled release manually once if your older build has no update configuration. An unavailable result does not mean the build is current; see [update settings](settings.md#updates).

## Windows update check reports EOF or a timeout

`EOF` or `unexpected EOF` means the connection or response ended early. It does not mean your version is current or that an update signature failed. Windows release builds use `update-windows-amd64.json`.

Rivu retries interrupted or timed-out checks up to two times. On Windows, it uses the proxy environment variables when configured; otherwise it reads your current Windows system proxy, including automatic configuration scripts (PAC). Proxy exceptions are respected. You can enable a working system proxy and check again without enabling TUN mode. A browser reaching GitHub does not by itself prove that Rivu's proxy path is working.

If checks still fail, verify that the configured proxy is running and permits both `github.com` and the redirected release-download hosts. If you launch Rivu from a terminal, review `HTTP_PROXY`, `HTTPS_PROXY` and `NO_PROXY`, because they take precedence over system proxy selection. Check again even if you already installed the newest version: Rivu still retrieves the manifest before comparing versions. Report the complete final error if the problem persists, removing proxy credentials before sharing it.

## Still need help?

Search [GitHub Issues](https://github.com/ntnyq/rivu/issues), then report your operating system and version, Rivu version, steps to reproduce, and the exact error. Remove private paths and tracker credentials from screenshots or logs.
