//go:build !windows
// +build !windows

package cmd

import (
	"io"
	"io/fs"
	"syscall"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

const defaultEditor = "vi"

var defaultInterpreters = make(map[string]*chezmoi.Interpreter)

// enableVirtualTerminalProcessing does nothing on non-Windows systems.
func enableVirtualTerminalProcessing(w io.Writer) (func() error, error) {
	return nil, nil
}

func fileInfoUID(info fs.FileInfo) int {
	return int(info.Sys().(*syscall.Stat_t).Uid) //nolint:forcetypeassert
}
