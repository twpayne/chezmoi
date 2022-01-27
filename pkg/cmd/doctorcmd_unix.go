//go:build !windows
// +build !windows

package cmd

import (
	"fmt"

	"golang.org/x/sys/unix"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

func (umaskCheck) Name() string {
	return "umask"
}

func (umaskCheck) Run(system chezmoi.System, homeDirAbsPath chezmoi.AbsPath) (checkResult, string) {
	umask := unix.Umask(0)
	unix.Umask(umask)
	result := checkResultOK
	if umask != 0o002 && umask != 0o022 {
		result = checkResultWarning
	}
	return result, fmt.Sprintf("%03o", umask)
}
