# Frequently asked questions

## Which platforms are supported?

Rivu supports **macOS 12 or later, Windows 10 / 11, and Linux**. Releases include a universal macOS DMG, a Windows x64 installer, and Linux x64 DEB and archive packages. Linux needs GTK 3, plus AppIndicator for the tray. See [installation](getting-started.md#get-rivu) for details.

## Where can I download Rivu?

Download Rivu from the project's [GitHub Releases](https://github.com/ntnyq/rivu/releases), or [build it from source](getting-started.md#build-from-source). Current macOS downloads do not have Apple Developer ID signing or Apple notarization; see [how to open Rivu if macOS blocks it](getting-started.md#open-rivu-when-macos-blocks-it). New release builds support signed updates from GitHub Releases; see [update settings](settings.md#updates). The documentation site is [rivu.ntnyq.dev](https://rivu.ntnyq.dev).

## What can Rivu download?

Rivu accepts `.torrent` files and magnet links for v1, v2-only, and hybrid BitTorrent torrents. Direct HTTP downloads, RSS subscriptions, remote control, and streaming playback are outside the current feature set.

## Does closing the window stop downloads?

No. Transfers continue while Rivu is running. Use the menu bar panel to check progress or reopen the window. Quitting the app saves the session and stops transfers.

## Does removing a task delete my files?

Files are kept by default. Moving files to Trash is a separate option, limited to files owned by that transfer. Existing files adopted by Rivu are always retained on removal.

## Why does Rivu check files after restarting?

Rivu verifies the actual data before continuing. Saved percentages alone cannot establish that files are complete or unchanged. A large download can spend time checking before its transfer resumes.

## Why did a skipped file appear on disk?

Torrent pieces can span more than one file. Rivu may need some bytes belonging to a skipped file to complete a neighboring selected file.

## Which languages are available?

The app supports English, Simplified Chinese, Traditional Chinese, Japanese, and Korean. Choose a language in **Settings → General → Language**. The documentation is available in English and Simplified Chinese.

## How do I report a problem?

Check [troubleshooting](troubleshooting.md) and existing [GitHub Issues](https://github.com/ntnyq/rivu/issues), then open an issue with your operating system and version, Rivu version, expected behavior, and steps to reproduce. Remove tracker credentials and private paths from attached diagnostics.
