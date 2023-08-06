//go:build !windows
// +build !windows

package vfst

import (
	"io/fs"
	"syscall"
	"testing"

	vfs "github.com/twpayne/go-vfs/v4"
)

func init() {
	umask = fs.FileMode(syscall.Umask(0))
	syscall.Umask(int(umask))
}

// PermEqual returns if perm1 and perm2 represent the same permissions. On
// Windows, it always returns true.
func PermEqual(perm1, perm2 fs.FileMode) bool {
	return perm1&fs.ModePerm&^umask == perm2&fs.ModePerm&^umask
}

// TestSysNlink returns a PathTest that verifies that the path's
// Sys().(*syscall.Stat_t).Nlink is equal to wantNlink. If path's Sys() cannot
// be converted to a *syscall.Stat_t, it does nothing.
func TestSysNlink(wantNlink int) PathTest {
	return func(t *testing.T, fileSystem vfs.FS, path string) {
		t.Helper()
		info, err := fileSystem.Lstat(path)
		if err != nil {
			t.Errorf("fileSystem.Lstat(%q) == %+v, %v, want !<nil>, <nil>", path, info, err)
			return
		}
		if stat, ok := info.Sys().(*syscall.Stat_t); ok && int(stat.Nlink) != wantNlink {
			t.Errorf("fileSystem.Lstat(%q).Sys().(*syscall.Stat_t).Nlink == %d, want %d", path, stat.Nlink, wantNlink)
		}
	}
}
