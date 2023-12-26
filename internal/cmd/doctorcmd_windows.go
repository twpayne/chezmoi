package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
	"github.com/twpayne/chezmoi/v2/internal/chezmoilog"
)

type (
	systeminfoCheck struct{}
	umaskCheck      struct{ skippedCheck }
	unameCheck      struct{ skippedCheck }
)

func (systeminfoCheck) Name() string {
	return "systeminfo"
}

func (systeminfoCheck) Run(system chezmoi.System, homeDirAbsPath chezmoi.AbsPath) (checkResult, string) {
	cmd := exec.Command("systeminfo")
	data, err := chezmoilog.LogCmdOutput(cmd)
	if err != nil {
		return checkResultFailed, err.Error()
	}

	var osName, osVersion string
	s := bufio.NewScanner(bytes.NewReader(data))
	for s.Scan() {
		switch key, value, found := strings.Cut(s.Text(), ":"); {
		case !found:
			// Do nothing.
		case key == "OS Name":
			osName = strings.TrimSpace(value)
		case key == "OS Version":
			osVersion = strings.TrimSpace(value)
		}
	}
	return checkResultOK, fmt.Sprintf("%s (%s)", osName, osVersion)
}
