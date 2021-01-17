package chezmoi

import (
	"os"

	"github.com/bmatcuk/doublestar/v3"
	vfs "github.com/twpayne/go-vfs"
)

// A doubleStarOS embeds a vfs.FS into a value that implements doublestar.OS.
type doubleStarOS struct {
	vfs.FS
}

func (os doubleStarOS) Lstat(name string) (os.FileInfo, error) {
	return os.FS.Lstat(name)
}

func (os doubleStarOS) Open(name string) (doublestar.File, error) {
	return os.FS.Open(name)
}
