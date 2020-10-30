// +build !windows

package chezmoi

import (
	"os"
	"syscall"
)

var umask os.FileMode

func init() {
	umask = os.FileMode(syscall.Umask(0))
	syscall.Umask(int(umask))
}

// GetUmask returns the umask.
func GetUmask() os.FileMode {
	return umask
}

// SetUmask sets the umask.
func SetUmask(newUmask os.FileMode) {
	umask = newUmask
	syscall.Umask(int(umask))
}
