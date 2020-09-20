package chezmoi

import (
	"github.com/bmatcuk/doublestar/v2"
	vfs "github.com/twpayne/go-vfs"
)

type doubleStarOS struct {
	vfs.FS
}

func (os doubleStarOS) Open(name string) (doublestar.File, error) { return os.FS.Open(name) }
