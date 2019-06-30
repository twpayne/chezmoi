// +build !windows

package cmd

import (
	"os"
	"syscall"
)

func getOwner(info os.FileInfo) int {
	executableStat := info.Sys().(*syscall.Stat_t)
	return int(executableStat.Uid)
}
