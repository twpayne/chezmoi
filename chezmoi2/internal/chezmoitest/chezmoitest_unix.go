// +build !windows

package chezmoitest

import (
	"os"
	"syscall"
)

// Umask is the umask used in tests.
const Umask = os.FileMode(0o022)

func init() {
	syscall.Umask(int(Umask))
}
