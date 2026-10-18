package files

import "os"

// Windows has no unprivileged directory-fsync equivalent. Payload and journal
// contents are flushed before publication; metadata durability still depends on
// the filesystem. Do not call FlushFileBuffers on a read-only directory handle.
func SyncDirectory(*os.File) error { return nil }
