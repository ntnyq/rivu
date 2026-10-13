<p align="center">
  <img src="resources/icons/rivu-128.png" alt="Rivu icon" width="80" height="80">
</p>

<h1 align="center">Rivu</h1>

<p align="center"><strong>Let your downloads flow quietly.</strong></p>

<p align="center">
  A native BitTorrent downloader for macOS.<br>
  Choose your files, keep your queue in order, and let Rivu take care of the transfers.
</p>

<p align="center">
  <a href="#get-started">Get started</a> ·
  <a href="docs/usage.md">User guide</a> ·
  <a href="https://github.com/ntnyq/rivu/issues">Feedback</a>
</p>

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/images/main-dark.png">
  <img src="docs/images/main-light.png" alt="Rivu showing a download queue with progress, speeds, and details for the selected transfer">
</picture>

## Your downloads, at a glance

See what's downloading, seeding, finished, or waiting for attention. Search and sort your transfers, then open the details panel for progress, files, and connection information.

- **Add it your way.** Open a `.torrent` file, paste a magnet link, or drag a torrent into the window. Supports v1, v2, and hybrid torrents.
- **Download what you need.** Choose files, set their priorities, and pick a destination before starting. Magnet links resolve metadata in a cancellable preview.
- **Set the pace.** Limit bandwidth and concurrent downloads, pause individual transfers, or pause the whole queue. Resuming a paused batch respects tasks you had already paused yourself.
- **Pick up where you left off.** Tasks and pause choices survive restarts. Rivu checks downloaded data when restoring a transfer.
- **Seed on your terms.** Stop on completion, choose a share-ratio target, or seed without a ratio limit. Add a cumulative seeding-time limit when needed.
- **Make it automatic.** Schedule daily bandwidth limits, organize tasks with tags, or watch a folder for new torrent files.
- **Reuse your data.** Explicitly adopt existing downloads or point a task to a new location; Rivu verifies the data and preserves existing-file deletion rights.
- **Keep your files.** Removing a task keeps its files by default. Moving downloaded files to Trash is a separate choice.

## Close the window. Keep downloading.

Transfers continue while Rivu stays open in the background. The menu bar panel gives you a quick look at speeds and active tasks, with pause, speed-limit, and add-download controls close at hand.

<p align="center">
  <img src="docs/images/menu-bar.png" alt="Rivu menu bar panel with transfer progress, pause controls, and a speed-limit toggle" width="360">
</p>

## At home on your Mac

Rivu uses a native interface with light and dark appearances, familiar menus, keyboard shortcuts, and a command palette. Built with Go and MyGo, it runs without a WebView or JavaScript runtime.

Available in **English, 简体中文, 繁體中文, 日本語, and 한국어**. Rivu follows your system language by default; you can change it in **Settings → General → Language**.

<details>
  <summary>View the dark appearance</summary>

  <img src="docs/images/main-dark.png" alt="Rivu download queue and transfer details in dark appearance">
</details>

<sub>Screenshots show the application's UI rendered with demo transfers in Simplified Chinese.</sub>

## Get started

Rivu is in active development for **macOS 12 or later**. To try the current app, follow the [build and run guide](docs/development.md). Signed distribution and automatic updates are still being prepared; Windows and Linux are not supported releases.

Once Rivu is running:

1. Select **Add Download**, open a `.torrent` file, or paste a magnet link.
2. Review the download, choose a destination, and select files when metadata is available.
3. Start the transfer and follow its progress in the main window or menu bar.

See the [user guide](docs/usage.md) for magnet-link requirements, destination handling, and recovery. Report problems or share ideas in [GitHub Issues](https://github.com/ntnyq/rivu/issues).

## Explore the project

[Documentation index](docs/README.md) · [Development](docs/development.md) · [Architecture](docs/architecture.md) · [Release status](docs/implementation-status.md) · [Building releases](docs/releasing.md) · [Third-party notices](THIRD_PARTY_NOTICES.md)
