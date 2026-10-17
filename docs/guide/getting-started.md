# Get started

This guide takes you from installing Rivu to your first download.

## Get Rivu

Rivu supports **macOS 12 or later, Windows 10 / 11, and Linux**. Download the package for your system from the project's [GitHub Releases](https://github.com/ntnyq/rivu/releases).

| System                    | Package                                                  | Installation                                                                                                     |
| ------------------------- | -------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------- |
| macOS                     | Universal `.dmg` for Apple Silicon and Intel             | Open the disk image and drag **Rivu.app** to **Applications**, then open it from there.                          |
| Windows                   | x64 `Rivu Setup <version> amd64.exe`                     | Run the installer, then launch Rivu from the Start menu.                                                         |
| Linux (Debian / Ubuntu)   | x64 `rivu_<version>_amd64.deb`                           | Run `sudo apt install ./rivu_<version>_amd64.deb`, using the downloaded file name.                               |
| Other Linux distributions | x64 `rivu-<version>-linux-amd64.tar.gz` and `install.sh` | Download both into one folder and run `sh install.sh`. It installs Rivu for your user under `~/.local/rivu.app`. |

Linux native windows require GTK 3; the system tray also needs `libayatana-appindicator3` (on Debian / Ubuntu, install `libayatana-appindicator3-1`). Rivu uses native UI and does not require WebView2 or WebKitGTK for its windows, although MyGo's current DEB declares a WebKitGTK dependency.

### macOS installation

Current macOS downloads are not signed with an Apple Developer ID or notarized by Apple, so macOS may block the first launch. Follow [Open Rivu when macOS blocks it](#open-rivu-when-macos-blocks-it) below. New release builds support signed updates from GitHub Releases through [Settings → Updates & About](settings.md#updates). Install one update-enabled release manually if your older build reports updates as unavailable.

### Open Rivu when macOS blocks it

If macOS cannot verify the developer or check Rivu for malicious software, follow these steps only after confirming that your copy came from the project's [GitHub Releases](https://github.com/ntnyq/rivu/releases) and has not been modified.

The alert may say **Apple cannot check Rivu for malicious software** and offer only **Move to Trash** and **Done**, with no **Open** button. This means Apple cannot complete the verification; it does not mean malware has been detected. Click **Done** to keep the app, then open it through Privacy & Security:

1. Try opening **Rivu.app** from **Applications**. If the alert offers **Move to Trash** and **Done**, click **Done**, not **Move to Trash**. If it offers **Cancel** instead of Done, click **Cancel**.
2. Open **Apple menu → System Settings → Privacy & Security**, then scroll to **Security**.
3. Find the message about Rivu being blocked and click **Open Anyway**. Authenticate if prompted.
4. Click **Open** in the confirmation dialog. macOS saves an exception for this copy of Rivu, so you can open it normally afterward.

On **macOS 12 Monterey**, use **System Preferences → Security & Privacy → General** instead; unlock the pane if needed, then click **Open Anyway**.

If **Open Anyway** is missing, try launching Rivu again and return to that pane. A managed Mac may require your administrator's help. If the alert says the app **will damage your computer** or **is damaged**, stop and [report the exact message](https://github.com/ntnyq/rivu/issues).

See [Apple's instructions for opening downloaded apps](https://support.apple.com/en-us/102445).

## Build from source

To build Rivu yourself, install **Go 1.27.2 or later** and run:

```sh
git clone https://github.com/ntnyq/rivu.git
cd rivu
go mod download
go run .
```

To create a package for your current system:

```sh
go tool mygo build
```

Packages are written under `build/<os>-<arch>/`. On Apple Silicon, the app is `build/darwin-arm64/Rivu.app`; add `-skip-dmg` for a bundle without a disk image. Local macOS builds are ad-hoc signed. For other build targets and packaging details, see [development](../development/index.md).

## Add your first download

Shortcuts below use macOS notation; on Windows and Linux, use **Ctrl** in place of **⌘**.

1. Choose **Add Download** or press **⌘N**.
2. Select a `.torrent` file or paste a magnet link. You can also drag a torrent file into the main window.
3. Choose a destination and select **Read and Review**.
4. Review the name, size, and files. Keep at least one file selected, then add the task.

The default destination is `~/Downloads/Rivu`. Choose a location without conflicting files for a new download. To reuse existing data, follow [adopting existing files](adding-downloads.md#reuse-existing-files).

A magnet link must contain an explicit tracker or peer address so Rivu can start resolving it. See [magnet previews](adding-downloads.md#preview-a-magnet-link) if your link is rejected or times out.

## Follow its progress

Select the task to see its progress, files, and connections. Use sidebar filters to focus on active, finished, or paused transfers. The menu bar panel gives you a compact view while the main window is closed.

Closing the window keeps downloads running. **⌘Q** quits Rivu and stops transfers after saving the session.

## Make it yours

Open **Settings** with **⌘,** to choose your language, default destination, and bandwidth limits. Next, explore [queue and seeding controls](managing-transfers.md) or [automatic imports](automation.md).
