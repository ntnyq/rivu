---
layout: home
title: Let your downloads flow quietly
titleTemplate: Rivu

hero:
  name: Rivu
  text: Let your downloads flow quietly.
  tagline: A native BitTorrent downloader for macOS. Choose your files, organize your queue, and keep transfers close at hand in the menu bar.
  image:
    src: /icon.png
    alt: Rivu application icon
  actions:
    - theme: brand
      text: Get started
      link: /usage
    - theme: alt
      text: Build and run
      link: /development
    - theme: alt
      text: View on GitHub
      link: https://github.com/ntnyq/rivu

features:
  - title: Native on your Mac
    details: Go and MyGo power native windows, light and dark appearances, keyboard shortcuts, and a menu bar panel. No WebView or JavaScript runtime in the app.
    link: /desktop
  - title: Download what you need
    details: Open torrents or magnet links, review metadata, select files, and set priorities. Supports v1, v2-only, and hybrid torrents.
    link: /usage#add-a-download
  - title: A queue that remembers
    details: Control bandwidth and concurrency, pause a batch, and keep manual pauses across restarts. Restored transfers verify data before trusting progress.
    link: /usage#manage-transfers
  - title: Seed on your terms
    details: Stop on completion, set a share-ratio target, or seed without a ratio limit. Add a cumulative active-seeding-time limit for each task.
    link: /usage#manage-transfers
  - title: Make it automatic
    details: Schedule daily bandwidth limits, organize transfers with tags, and opt into importing torrent files from a watched folder.
    link: /usage#tags-schedules-and-watch-folders
  - title: Keep control of your files
    details: Adopt and verify existing data, change a task's location, and keep downloaded files when removing tasks by default.
    link: /usage#storage-and-recovery
---

## Your downloads, at a glance

![Rivu main window in light appearance](images/main-light.png)

<sub>Application UI rendered with demo transfers in Simplified Chinese.</sub>

## Try the development build

Rivu targets **macOS 12 or later** and is in active development. Follow the [development guide](development.md) to build and run it locally. Production signing, notarization, and a trusted update feed are still being prepared; Windows and Linux builds are compilation checks.

## Explore the documentation

- [User guide](usage.md): importing, queue controls, automation, and recovery.
- [Architecture](architecture.md) and [source structure](implementation-structure.md): state ownership, modules, and persistence.
- [Feature validation](refactor-validation.md) and [implementation status](implementation-status.md): verified behavior and remaining release work.
- [Desktop acceptance](desktop.md) and [release guide](releasing.md): native verification, packaging, and distribution.
- [Product design](product-design.md), [implementation plan](implementation-plan.md), [proposed source tree](history/implementation-structure-2026-10-12.md), and [visual prototype](design/README.md): historical references. Current code and implementation documents take precedence.
