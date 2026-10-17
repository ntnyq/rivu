# Frequently asked questions

## Which platforms are supported?

Rivu targets **macOS 12 or later**. Windows and Linux builds are currently compilation checks, rather than supported releases.

## Where can I download Rivu?

The current way to try Rivu is a [local source build](getting-started.md). Signed distribution and automatic updates are still being prepared. The documentation site is [rivu.ntnyq.dev](https://rivu.ntnyq.dev).

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

Check [troubleshooting](troubleshooting.md) and existing [GitHub Issues](https://github.com/ntnyq/rivu/issues), then open an issue with your macOS version, Rivu version, expected behavior, and steps to reproduce. Remove tracker credentials and private paths from attached diagnostics.
