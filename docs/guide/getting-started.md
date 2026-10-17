# Get started

This guide takes you from a local build to your first download.

## Get Rivu

Rivu currently requires a source build on **macOS 12 or later**, with **Go 1.27.2 or later** installed. Production signing, notarization, and automatic-update distribution are still being prepared.

```sh
git clone https://github.com/ntnyq/rivu.git
cd rivu
go mod download
make run
```

To create a local app bundle instead:

```sh
go tool mygo build -skip-dmg
```

On Apple Silicon, the bundle is written to `build/darwin-arm64/Rivu.app`. Local builds are ad-hoc signed. For other build targets and packaging details, see [development](../development/index.md).

## Add your first download

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
