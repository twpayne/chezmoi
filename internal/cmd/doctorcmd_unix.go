//go:build unix

package cmd

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"golang.org/x/sys/unix"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
	"github.com/twpayne/chezmoi/v2/internal/chezmoilog"
)

type (
	systeminfoCheck struct{ omittedCheck }
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
	cmd := exec.Command("uname", "-a")
	cmd.Stderr = os.Stderr
	data, err := chezmoilog.LogCmdOutput(slog.Default(), cmd)
	if err != nil {
		return checkResultFailed, err.Error()
	}
	return checkResultOK, string(bytes.TrimSpace(data))
}
