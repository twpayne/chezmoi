package chezmoi

import (
	"os"
	"os/exec"
)

// A Mutator makes changes.
type Mutator interface {
	Chmod(name string, mode os.FileMode) error
	IdempotentCmdOutput(cmd *exec.Cmd) ([]byte, error)
	Mkdir(name string, perm os.FileMode) error
	RemoveAll(name string) error
	Rename(oldpath, newpath string) error
	RunCmd(cmd *exec.Cmd) error
	Stat(name string) (os.FileInfo, error)
	WriteFile(filename string, data []byte, perm os.FileMode, currData []byte) error
	WriteSymlink(oldname, newname string) error
}
