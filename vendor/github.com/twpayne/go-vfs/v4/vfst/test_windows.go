package vfst

import (
	"io/fs"
	"testing"

	"github.com/twpayne/go-vfs/v4"
)

// PermEqual returns if perm1 and perm2 represent the same permissions. On
// Windows, it always returns true.
func PermEqual(perm1, perm2 fs.FileMode) bool {
	return true
}

// TestSysNlink returns a PathTest that verifies that the the path's
// Sys().(*syscall.Stat_t).Nlink is equal to wantNlink. If path's Sys() cannot
// be converted to a *syscall.Stat_t, it does nothing.
func TestSysNlink(wantNlink int) PathTest {
	return func(*testing.T, vfs.FS, string) {
	}
}
