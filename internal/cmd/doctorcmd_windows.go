package cmd

import (
	"fmt"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/twpayne/chezmoi/internal/chezmoilog"
)

type (
	systeminfoCheck struct{}
	umaskCheck      struct{ omittedCheck }
	unameCheck      struct{ omittedCheck }
)

func (systeminfoCheck) Name() string {
	return "systeminfo"
}

func (systeminfoCheck) Run(config *Config) (checkResult, string) {
	cmd := exec.Command("systeminfo")
	data, err := chezmoilog.LogCmdOutput(slog.Default(), cmd)
	if err != nil {
		return checkResultFailed, err.Error()
	}

	var osName, osVersion string
	for line := range strings.Lines(string(data)) {
		switch key, value, found := strings.Cut(line, ":"); {
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
