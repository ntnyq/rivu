# Automation

## Organize with tags

Enter comma-separated tags in a task's inspector, then use the toolbar's tag filter to find related transfers. You can also assign tags when reviewing a new download.

For example, use `linux, images` for operating-system downloads and `archives` for data collections.

## Schedule speed limits

1. Open **Settings → Automation**.
2. Enable the daily speed-limit interval.
3. Enter start and end times in `HH:MM`, using local time.
4. Set the temporary download and upload limits, then save.

An interval can cross midnight, such as **22:00–08:00**. During that interval, Rivu applies the temporary limits.

The manual speed-limit toggle uses the same temporary limits. Turning it on keeps them active outside the schedule; turning it off during a scheduled interval leaves the schedule in effect. To return to normal limits during that interval, disable the schedule.

Schedules never resume manually paused tasks.

## Import from a watched folder

1. Open **Settings → Automation** and enable **Watch folder**.
2. Enter the folder's absolute path.
3. Optionally add an automatic tag.
4. Save the settings and place `.torrent` files in that folder.

Rivu checks the folder every 10 seconds and waits until file content is identical in two scans before importing. New tasks use your default download directory.

The watched folder:

- Imports regular `.torrent` files in that folder, without scanning subfolders.
- Keeps the original torrent files.
- Skips unchanged files that were already imported, even after a task is removed or Rivu restarts.
- Leaves invalid files available for correction without blocking other imports.
- Does not automatically adopt conflicting download data.

If a torrent does not appear, check that watching is enabled, the path is accessible, the file is complete, and the destination has no conflicting data. To reuse data explicitly, [add the torrent manually](adding-downloads.md#reuse-existing-files).
