// +build !windows

package cmd

import (
    "os"
)

func getOwner(info os.FileInfo) int {
	executableStat := info.Sys().(*syscall.Stat_t)
	return executableStat.Uid
}
