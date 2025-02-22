//go:build unix && test

package chezmoitest

import (
	"golang.org/x/sys/unix"
)

func init() {
	Umask = mustParseFileMode(umaskStr)
	unix.Umask(int(Umask))
}
