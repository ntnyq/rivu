# Settings & shortcuts

Press **⌘,** to open Settings. Save your changes to apply them.

## Appearance and language

In **General**, choose light, dark, or system appearance. Rivu supports **English, Simplified Chinese, Traditional Chinese, Japanese, and Korean**.

Language follows your system preferences by default, with English as the fallback. Choosing another language and saving updates the interface and native menus immediately. System dialogs and third-party diagnostic details may use their own language.

## Downloads and seeding

Choose a default destination, concurrent-download count, normal and temporary bandwidth limits, and default seeding policy in **Downloads & Seeding**.

The initial destination is `~/Downloads/Rivu`. You can choose a different destination for each new task. Setting a new default does not relocate existing downloads; use [Change Location](troubleshooting.md#the-download-directory-has-changed) for those.

See [managing transfers](managing-transfers.md) for limits and seeding behavior, and [automation](automation.md) for schedules and watched folders.

## System proxy and TUN

On Windows, HTTP/HTTPS trackers and outgoing TCP peer connections follow the current system proxy, including PAC rules and bypass lists. Explicit `HTTP_PROXY`, `HTTPS_PROXY` and `NO_PROXY` environment settings take precedence. TCP peers use the HTTPS proxy route for their address and port; HTTP/HTTPS proxies must allow CONNECT to the peer's port. SOCKS5 TCP proxies are also supported. A failed proxy connection is not retried directly.

Proxy settings are read for new requests and peer connections. When switching to TUN, disable the system proxy if you want connections to use only the OS route. Existing connections can remain open; pause and resume a task to reconnect immediately. Rivu does not configure the TUN device or its routing rules.

HTTP proxies cannot carry UDP. DHT, UDP trackers, and incoming connections continue to use the OS network, which a TUN can capture if configured for that traffic. Outgoing uTP is skipped when a proxy route is selected for that peer, so it cannot bypass the TCP proxy. This is not a mode that forces every BitTorrent packet through a proxy. If UDP is blocked on the direct route, use a TUN that supports UDP; enabling DHT/PEX alone cannot make an unreachable swarm accessible.

## Updates

In **Updates & About**, choose **Check for Updates** to check the latest published stable GitHub Release. New release builds verify downloaded update archives with an embedded public key before replacing the app. Draft releases are not offered.

Automatic checking and automatic installation are both off by default. Enable automatic checking and save to check approximately once a day. Automatic installation is a separate opt-in and waits until the transfer session is idle, including no active seeding. After installation, choose **Save and Restart**; Rivu does not automatically restart during transfers.

Use an installed app in a writable location: on macOS, copy Rivu to Applications rather than running it from the DMG. Linux DEB installations should be updated with the package manager; the per-user `install.sh` installation supports in-app updates when writable. Development mode and older builds without update configuration report **unavailable**, which does not mean the build is current. Install an update-enabled release manually once to migrate an older build.

## Keyboard shortcuts

The shortcuts below use macOS notation. On Windows and Linux, use **Ctrl** in place of **⌘**.

| Shortcut | Action                   |
| -------- | ------------------------ |
| ⌘N       | Add a download           |
| ⌘O       | Choose a torrent file    |
| ⌘1       | Open the main window     |
| ⌘K       | Open the command palette |
| ⌘,       | Open Settings            |
| ⌘Q       | Save and quit Rivu       |

The command palette provides another way to find application commands. Pause-all, batch-resume, and speed-limit commands are also available in the **Transfers** menu.

## Where your data lives

Preferences and session data use the system configuration directory:

| System  | Default location                                      |
| ------- | ----------------------------------------------------- |
| macOS   | `~/Library/Application Support/Rivu`                  |
| Windows | `%AppData%\Rivu`                                      |
| Linux   | `$XDG_CONFIG_HOME/Rivu`, or `~/.config/Rivu` if unset |

Downloaded files live in the destination selected for each task.

For backups and recovery, see [troubleshooting](troubleshooting.md#back-up-or-repair-a-session). Developers can override storage locations with [launch options](../development/index.md#launch-options).
