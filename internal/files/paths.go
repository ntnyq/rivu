package files

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ntnyq/rivu/internal/problem"
	"github.com/ntnyq/rivu/locales"
)

func ValidateMembers(names []string) error {
	seen := map[string]bool{}
	for _, name := range names {
		if name == "" || !fs.ValidPath(name) || strings.ContainsAny(name, "\\:\x00") {
			return fmt.Errorf("unsafe torrent path %q", name)
		}
		for _, part := range strings.Split(name, "/") {
			if strings.TrimSpace(part) == "" || strings.HasSuffix(part, ".") || strings.HasSuffix(part, " ") || runtime.GOOS == "windows" && !validWindowsComponent(part) {
				return fmt.Errorf("invalid filename %q", part)
			}
		}
		key := strings.ToLower(name)
		if seen[key] {
			return fmt.Errorf("colliding torrent path %q", name)
		}
		seen[key] = true
	}
	for name := range seen {
		for parent := path.Dir(name); parent != "."; parent = path.Dir(parent) {
			if seen[parent] {
				return fmt.Errorf("file/directory collision %q", parent)
			}
		}
	}
	return nil
}

// CreateDownloadDirectory creates the application folder without recreating a
// missing downloads directory or an unmounted redirected volume.
func CreateDownloadDirectory(dir string) error {
	if err := os.Mkdir(dir, 0o755); err != nil && !os.IsExist(err) {
		return err
	}
	return ValidateDestination(dir)
}

// ValidateDestination never creates missing external volumes. The caller must
// explicitly create a default directory or pick an existing one before import.
func ValidateDestination(dir string) error {
	if !filepath.IsAbs(dir) {
		return problem.Message(problem.Destination, locales.Messages.FilesAbsolutePathRequired, nil)
	}
	stat, err := os.Stat(dir)
	if err != nil {
		return problem.New(problem.Destination, dir, err)
	}
	if !stat.IsDir() {
		return problem.Message(problem.Destination, locales.Messages.FilesNotFolder, nil)
	}
	entry, err := os.Lstat(dir)
	if err != nil {
		return err
	}
	if entry.Mode()&os.ModeSymlink != 0 {
		return problem.Message(problem.Destination, locales.Messages.FilesSymlinkFolder, nil)
	}
	return nil
}

func CheckConflicts(dir string, names []string) error {
	_, err := InspectExisting(dir, names, false)
	return err
}

// InspectExisting rejects links in every path component and records stable file identities.
func InspectExisting(dir string, names []string, allowExisting bool) (map[string]string, error) {
	if err := ValidateDestination(dir); err != nil {
		return nil, err
	}
	if err := ValidateMembers(names); err != nil {
		return nil, err
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	identities := map[string]string{}
	for _, name := range names {
		// Check every component: a pre-existing directory is fine, a symlink isn't.
		parts := strings.Split(name, "/")
		for i := range parts {
			rel := strings.Join(parts[:i+1], "/")
			st, e := root.Lstat(rel)
			if os.IsNotExist(e) {
				continue
			}
			if e != nil {
				return nil, e
			}
			if st.Mode()&os.ModeSymlink != 0 || (i == len(parts)-1 && (!allowExisting || !st.Mode().IsRegular())) || (i < len(parts)-1 && !st.IsDir()) {
				return nil, problem.Message(problem.Destination, func(m locales.Messages) string { return m.FilesUnsafeDestination(rel) }, nil)
			}
			if i == len(parts)-1 {
				file, err := root.Open(rel)
				if err != nil {
					return nil, err
				}
				opened, err := file.Stat()
				if err != nil {
					file.Close()
					return nil, err
				}
				if !os.SameFile(st, opened) {
					file.Close()
					return nil, fmt.Errorf("file changed while inspecting %q", rel)
				}
				id, err := FileIdentity(file)
				file.Close()
				if err != nil {
					return nil, err
				}
				identities[name] = id
			}
		}
	}
	return identities, nil
}

// Windows device names remain reserved even with extensions or superscript digits.
func validWindowsComponent(part string) bool {
	for _, ch := range part {
		if ch < 32 || strings.ContainsRune(`<>:"/\\|?*`, ch) {
			return false
		}
	}
	base, _, _ := strings.Cut(part, ".")
	base = strings.ToUpper(strings.TrimRight(base, " "))
	switch base {
	case "CON", "PRN", "AUX", "NUL", "CONIN$", "CONOUT$":
		return false
	}
	if strings.HasPrefix(base, "COM") || strings.HasPrefix(base, "LPT") {
		suffix := base[3:]
		if len(suffix) == 1 && suffix[0] >= '1' && suffix[0] <= '9' || suffix == "¹" || suffix == "²" || suffix == "³" {
			return false
		}
	}
	return part != "" && !strings.HasSuffix(part, ".") && !strings.HasSuffix(part, " ")
}
