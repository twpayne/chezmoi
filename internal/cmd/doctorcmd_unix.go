//go:build !windows
// +build !windows

package cmd

import (
	"fmt"

	"golang.org/x/sys/unix"
)

func (c *umaskCheck) Run() (checkResult, string) {
	umask := unix.Umask(0)
	unix.Umask(umask)
	result := checkResultOK
	if umask != 0o002 && umask != 0o022 {
		result = checkResultWarning
	}
	return result, fmt.Sprintf("%03o", umask)
}
