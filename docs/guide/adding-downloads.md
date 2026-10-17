# Add downloads

Rivu accepts `.torrent` files and magnet links, including v1, v2-only, and hybrid torrents.

## Open a torrent

Choose **Add Download** (**⌘N**), use **Open Torrent** (**⌘O**), or drag a torrent file into the main window. Rivu can also handle torrent files and magnet links opened with the app.

Choose a destination, select **Read and Review**, and check the contents before adding the task. Add tags in the review sheet if you want to organize the transfer immediately.

## Preview a magnet link

Paste the magnet link into **Add Download** and select **Read and Review**. Rivu retrieves metadata so you can inspect the torrent name, size, and file tree before starting.

- The preview does not write download payload to your destination.
- You can cancel or retry it; closing the sheet cancels it too.
- A preview times out after 45 seconds if metadata cannot be resolved.

Until metadata establishes whether a torrent is private, Rivu uses only the trackers or peers explicitly included in the magnet. A link without either is rejected. Rivu does not fetch arbitrary metadata-source URLs or web seeds.

If resolution fails, check the source link and your connection, then [review the troubleshooting steps](troubleshooting.md).

## Choose files and priorities

Expand folders in the file tree, search for a name, and choose **High**, **Normal**, or **Skip** for files or whole directories. At least one file must remain selected. Large file lists are paginated.

Skipping a file does not always mean zero bytes on disk: adjacent files can share a torrent piece, and some bytes may be needed to finish a selected file.

You can adjust file priorities later in the task inspector.

## Reuse existing files

For a new download, use a destination without conflicting payload files. To continue from data you already have:

1. Select the directory that contains the torrent's top-level name.
2. Enable **Adopt and verify existing files**.
3. Review and add the task.

When the task runs, Rivu checks the data and downloads missing or damaged pieces. Adopted files are retained when you remove the task and cannot be moved to Trash through Rivu.

To point an existing task to another directory, use [Change Location](troubleshooting.md#the-download-directory-has-changed). That command updates the task's location; it does not move files for you.
