package files

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/ntnyq/rivu/internal/i18n"
)

func ValidateMembers(names []string) error {
	seen := map[string]bool{}
	for _, name := range names {
		if name == "" || !fs.ValidPath(name) || strings.ContainsAny(name, "\\:\x00") {
			return fmt.Errorf("unsafe torrent path %q", name)
		}
		for _, part := range strings.Split(name, "/") {
			if strings.TrimSpace(part) == "" || strings.HasSuffix(part, ".") || strings.HasSuffix(part, " ") {
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

// ValidateDestination never creates missing external volumes. The caller must
// explicitly create a default directory or pick an existing one before import.
func ValidateDestination(dir string) error {
	if !filepath.IsAbs(dir) {
		return fmt.Errorf("%s", i18n.T("请选择绝对路径的下载目录"))
	}
	stat, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf(i18n.T("下载目录不可用: %w"), err)
	}
	if !stat.IsDir() {
		return fmt.Errorf("%s", i18n.T("下载路径不是目录"))
	}
	entry, err := os.Lstat(dir)
	if err != nil {
		return err
	}
	if entry.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("%s", i18n.T("请选择真实目录，不能使用符号链接"))
	}
	return nil
}

func CheckConflicts(dir string, names []string) error {
	if err := ValidateDestination(dir); err != nil {
		return err
	}
	if err := ValidateMembers(names); err != nil {
		return err
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return err
	}
	defer root.Close()
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
				return e
			}
			if st.Mode()&os.ModeSymlink != 0 || i == len(parts)-1 || !st.IsDir() {
				return fmt.Errorf(i18n.T("目标已存在或不安全，请选择另一个目录: %s"), rel)
			}
		}
	}
	return nil
}
