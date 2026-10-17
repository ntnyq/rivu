# Settings & shortcuts

Press **⌘,** to open Settings. Save your changes to apply them.

## Appearance and language

In **General**, choose light, dark, or system appearance. Rivu supports **English, Simplified Chinese, Traditional Chinese, Japanese, and Korean**.

Language follows your system preferences by default, with English as the fallback. Choosing another language and saving updates the interface and native menus immediately. System dialogs and third-party diagnostic details may use their own language.

## Downloads and seeding

Choose a default destination, concurrent-download count, normal and temporary bandwidth limits, and default seeding policy in **Downloads & Seeding**.

The initial destination is `~/Downloads/Rivu`. You can choose a different destination for each new task. Setting a new default does not relocate existing downloads; use [Change Location](troubleshooting.md#the-download-directory-has-changed) for those.

See [managing transfers](managing-transfers.md) for limits and seeding behavior, and [automation](automation.md) for schedules and watched folders.

## Updates

The updater interface exists, but release signing and a trusted update feed are still being prepared. An **unavailable** result means Rivu cannot check a trusted feed; it does not establish that your build is up to date.

## Keyboard shortcuts

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

By default, preferences and session data are stored in `~/Library/Application Support/Rivu`. Downloaded files live in the destination selected for each task.

For backups and recovery, see [troubleshooting](troubleshooting.md#back-up-or-repair-a-session). Developers can override storage locations with [launch options](../development/index.md#launch-options).
