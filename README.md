<p align="center">
  <img src="resources/icons/rivu-128.png" alt="Rivu icon" width="96" height="96">
</p>

<h1 align="center">Rivu</h1>

<p align="center"><strong>Let your downloads flow quietly.</strong></p>
<p align="center">A native BitTorrent app for macOS. Choose your files, manage your queue, and keep downloading in the background.</p>

<p align="center">
  <a href="https://rivu.ntnyq.dev">Website</a> ·
  <a href="https://rivu.ntnyq.dev/guide/getting-started.html">Get started</a> ·
  <a href="https://rivu.ntnyq.dev/guide/">Documentation</a> ·
  <a href="https://github.com/ntnyq/rivu/issues">Feedback</a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/macOS-12%2B-111827?logo=apple" alt="macOS 12 or later">
  <img src="https://img.shields.io/badge/status-in%20development-2563eb" alt="In development">
</p>

<p align="center">English · <a href="README.zh-CN.md">简体中文</a></p>

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/images/main-dark.png">
  <img src="docs/images/main-light.png" alt="Rivu download queue with transfer progress, speeds, and a file inspector">
</picture>

## Why Rivu?

- **Made for macOS.** Native windows, light and dark appearances, keyboard shortcuts, and a command palette. Built with Go and MyGo.
- **Download only what you need.** Add torrent files or magnet links, preview their contents, and choose files and priorities. Supports v1, v2, and hybrid torrents.
- **Keep your queue in order.** Search, sort, and tag transfers. Set bandwidth and concurrent-download limits, pause a batch, and preserve individual pauses.
- **Pick up where you left off.** Restore tasks after restarting and verify existing data before continuing. Reuse downloads or point a task to a different location.
- **Choose when to seed.** Stop on completion, at a share-ratio target, or after a seeding-time limit; keep seeding without a ratio limit if you prefer.
- **Automate the routine.** Schedule daily speed limits and import torrents from a watched folder.

Removing a task keeps its files by default. Moving downloaded files to Trash is a separate choice.

## Close the window. Keep downloading.

Rivu stays available in the menu bar while transfers run in the background. Check progress, pause the queue, switch speed limits, or add a download without reopening the main window.

<p align="center">
  <img src="docs/images/menu-bar.png" alt="Rivu menu bar panel with active transfers and quick controls" width="360">
</p>

<details>
  <summary>View the dark appearance</summary>

  <img src="docs/images/main-dark.png" alt="Rivu main window in dark appearance">
</details>

<sub>Screenshots use the application UI renderer with demo transfers in Simplified Chinese.</sub>

## Get started

Rivu targets **macOS 12 or later** and is in active development. The current way to try it is to build from source. Signed distribution and automatic updates are still being prepared.

The current application version is **0.0.1**. The project pins **MyGo v0.5.5** and its CLI tool in `go.mod`; no separate MyGo installation is needed.

With Go 1.27.2 or later installed on your Mac:

```sh
git clone https://github.com/ntnyq/rivu.git
cd rivu
go mod download
make run
```

To package a local app bundle, run `go tool mygo build -skip-dmg`. See the [getting started guide](https://rivu.ntnyq.dev/guide/getting-started.html) for your first download and the [development guide](https://rivu.ntnyq.dev/development/) for build details.

## Speaks your language

**English · 简体中文 · 繁體中文 · 日本語 · 한국어**

Rivu follows your system language. You can choose another language in **Settings → General → Language**. The documentation is available in [English](https://rivu.ntnyq.dev) and [简体中文](https://rivu.ntnyq.dev/zh/).

## Documentation & feedback

- [User guide](https://rivu.ntnyq.dev/guide/) — adding downloads, managing transfers, automation, and settings.
- [Troubleshooting](https://rivu.ntnyq.dev/guide/troubleshooting.html) — magnet links, connection issues, and recovering a session.
- [Development](https://rivu.ntnyq.dev/development/) — building, testing, architecture, and release preparation.
- [GitHub Issues](https://github.com/ntnyq/rivu/issues) — report a bug or suggest a feature. Include your macOS version, Rivu version, and steps to reproduce.

To work on the documentation locally, run `pnpm install` and `pnpm docs:dev`. Use `pnpm docs:build` to validate the site. Keep English and Chinese pages aligned.

## Acknowledgments & licensing

Rivu is built with [MyGo](https://github.com/egoist/mygo) and [anacrolix/torrent](https://github.com/anacrolix/torrent). See [third-party notices](THIRD_PARTY_NOTICES.md) for dependency licenses. A license for Rivu's own code has not yet been selected.
