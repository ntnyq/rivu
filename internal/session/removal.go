package session

import (
	"github.com/ntnyq/rivu/internal/problem"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ntnyq/rivu/internal/files"
)

func (m *Manager) finishRemoval(i int) error {
	r := m.state.Tasks[i]
	if err := m.stop(r.ID); err != nil {
		return err
	}
	if r.Trash {
		if m.trash == nil {
			return problem.Message(problem.Conflict, "系统废纸篓不可用")
		}
		if err := files.ValidateDestination(r.Destination); err != nil {
			return err
		}
		for _, f := range r.Source.Files {
			if r.Adopted[f.Path] != "" || slices.Contains(r.Trashed, f.Path) {
				continue
			}
			full := filepath.Join(r.Destination, filepath.FromSlash(f.Path))
			shared := false
			for _, other := range m.state.Tasks {
				if other.ID == r.ID {
					continue
				}
				for _, of := range other.Source.Files {
					if strings.EqualFold(full, filepath.Join(other.Destination, filepath.FromSlash(of.Path))) {
						shared = true
					}
				}
			}
			if shared {
				return problem.Message(problem.Conflict, "文件由其他任务共享，无法移到废纸篓")
			}
			real, err := filepath.EvalSymlinks(full)
			if err == nil {
				if real != full {
					return problem.Message(problem.Conflict, "文件已被符号链接替换")
				}
				st, er := os.Lstat(full)
				if er != nil {
					return er
				}
				if !st.Mode().IsRegular() {
					return problem.Message(problem.Conflict, "只允许移除任务的普通文件")
				}
				owners, er := files.ReadOwnership(r.Metadata + ".files.json")
				if er != nil {
					return er
				}
				identity, er := files.FileIdentity(st)
				if er != nil {
					return er
				}
				if owners[f.Path] != identity {
					return problem.Message(problem.Conflict, "文件归属已改变，已保留: %s", f.Path)
				}
				if err = m.trash(full); err != nil {
					return err
				}
			} else if !os.IsNotExist(err) {
				return err
			}
			r.Trashed = append(r.Trashed, f.Path)
			next := cloneState(m.state)
			next.Tasks[i] = r
			if err := m.save(next); err != nil {
				return err
			}
		}
	}
	next := cloneState(m.state)
	next.Tasks = slices.Delete(next.Tasks, i, i+1)
	if err := m.save(next); err != nil {
		return err
	}
	delete(m.samples, r.ID)
	delete(m.holds, r.ID)
	return nil
}
