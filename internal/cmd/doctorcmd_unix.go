//go:build unix

package cmd

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"

	"golang.org/x/sys/unix"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
	"github.com/twpayne/chezmoi/v2/internal/chezmoilog"
)

type (
	systeminfoCheck struct{ skippedCheck }
	umaskCheck      struct{}
	unameCheck      struct{}
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

func (unameCheck) Name() string {
	return "uname"
}

func (unameCheck) Run(system chezmoi.System, homeDirAbsPath chezmoi.AbsPath) (checkResult, string) {
	if runtime.GOOS == "windows" {
		return checkResultSkipped, ""
	}
	cmd := exec.Command("uname", "-a")
	data, err := chezmoilog.LogCmdOutput(cmd)
	if err != nil {
		return checkResultFailed, err.Error()
	}
	return checkResultOK, string(bytes.TrimSpace(data))
}
