//go:build unix

package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"golang.org/x/sys/unix"

	"github.com/twpayne/chezmoi/internal/chezmoilog"
)

type (
	// Check if hardlinks work between tempDir and sourceDir.
	hardlinkCheck   struct{}
	systeminfoCheck struct{ omittedCheck }
	umaskCheck      struct{}
	unameCheck      struct{}
)

func (hardlinkCheck) Name() string {
	return "hardlink"
}

func (hardlinkCheck) Run(config *Config) (checkResult, string) {
	if !config.Edit.Hardlink {
		return checkResultInfo, "edit.hardlink disabled"
	}

	testFileName := ".chezmoi-doctor-hardlink-test"

	tempDirAbsPath, err := config.tempDir("chezmoi-doctor")
	if err != nil {
		return checkResultFailed, err.Error()
	}

	hardlinkAbsPath := tempDirAbsPath.JoinString(testFileName)
	sourceAbsPath := config.SourceDirAbsPath.JoinString(testFileName)

	if err := config.baseSystem.WriteFile(sourceAbsPath, []byte(""), 0o700); err != nil {
		return checkResultFailed, err.Error()
	}

	if err := os.MkdirAll(hardlinkAbsPath.Dir().String(), 0o700); err != nil {
		return checkResultFailed, err.Error()
	}

	if err := config.baseSystem.Link(config.SourceDirAbsPath.JoinString(testFileName), hardlinkAbsPath); err != nil {
		errCleanUp := config.baseSystem.Remove(sourceAbsPath)
		return checkResultError, fmt.Sprintf(
			"Failed creating hardlink %s -> %s: %s",
			config.SourceDirAbsPath,
			config.TempDir,
			errors.Join(err, errCleanUp),
		)
	}

	if err := config.baseSystem.Remove(sourceAbsPath); err != nil {
		return checkResultFailed, err.Error()
	}

	return checkResultOK, fmt.Sprintf("Can create hardlinks %s -> %s", config.SourceDirAbsPath, config.TempDir)
}

func (umaskCheck) Name() string {
	return "umask"
}

func (umaskCheck) Run(config *Config) (checkResult, string) {
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

func (unameCheck) Run(config *Config) (checkResult, string) {
	cmd := exec.Command("uname", "-a")
	cmd.Stderr = os.Stderr
	data, err := chezmoilog.LogCmdOutput(slog.Default(), cmd)
	if err != nil {
		return checkResultFailed, err.Error()
	}
	return checkResultOK, string(bytes.TrimSpace(data))
}
