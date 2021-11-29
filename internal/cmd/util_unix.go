//go:build !windows
// +build !windows

package cmd

import (
	"io"
	"io/fs"
	"syscall"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

const defaultEditor = "vi"

var defaultInterpreters = make(map[string]*chezmoi.Interpreter)

// enableVirtualTerminalProcessing does nothing.
func enableVirtualTerminalProcessing(w io.Writer) error {
	return nil
}

func fileInfoUID(info fs.FileInfo) int {
	return int(info.Sys().(*syscall.Stat_t).Uid)
}
