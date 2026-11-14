//go:build unix

package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/sys/unix"

	"chezmoi.io/chezmoi/v2/internal/chezmoilog"
)

type (
	binShCheck      struct{}
	hardlinkCheck   struct{}
	systeminfoCheck struct{ omittedCheck }
	umaskCheck      struct{}
	unameCheck      struct{}
)

func (c binShCheck) Name() string {
	return "bin-sh"
}

// Run checks /bin/sh.
func (c binShCheck) Run(config *Config) (checkResult, string) {
	const binSh = "/bin/sh"

	var messages []string

	// Firstly, check whether /bin/sh is a symlink to something else.
	switch target, err := config.baseSystem.Readlink(binSh); {
	case errors.Is(err, unix.EINVAL):
		// Not a symlink; do nothing.
	case errors.Is(err, fs.ErrNotExist):
		return checkResultError, "not found"
	case err != nil:
		return checkResultError, err.Error()
	default:
		// A symlink to something else, report it.
		messages = append(messages, "symlink to "+target)
	}

	// Secondly, check /bin/sh's environment variables. Report anything that
	// contains "VERSION" or "version" in the key.
	output := &bytes.Buffer{}
	cmd := exec.Command(binSh, "-c", "set")
	cmd.Stdout = output
	switch err := config.baseSystem.RunCmd(cmd); {
	case err != nil:
		return checkResultError, err.Error()
	default:
		for line := range bytes.Lines(output.Bytes()) {
			switch key, _, ok := bytes.Cut(line, []byte{'='}); {
			case !ok:
				// Do nothing.
			case bytes.Contains(key, []byte("CHEZMOI")):
				// Ignore chezmoi variables.
			case bytes.Contains(bytes.ToUpper(key), []byte("VERSION")):
				messages = append(messages, string(bytes.TrimSpace(line)))
			}
		}
	}

	return checkResultOK, strings.Join(messages, ", ")
}

func (hardlinkCheck) Name() string {
	return "hardlink"
}

// Run checks if hardlinks work between tempDir and sourceDir.
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

	if err := config.baseSystem.WriteFile(sourceAbsPath, nil, 0o666); err != nil {
		return checkResultFailed, err.Error()
	}

	if err := os.MkdirAll(hardlinkAbsPath.Dir().String(), 0o666); err != nil {
		return checkResultFailed, err.Error()
	}

	if err := config.baseSystem.Link(config.SourceDirAbsPath.JoinString(testFileName), hardlinkAbsPath); err != nil {
		errCleanUp := config.baseSystem.Remove(sourceAbsPath)
		return checkResultError, fmt.Sprintf(
			"failed creating hardlink from %s to %s: %s",
			config.SourceDirAbsPath,
			config.TempDir,
			errors.Join(err, errCleanUp),
		)
	}

	if err := config.baseSystem.Remove(sourceAbsPath); err != nil {
		return checkResultFailed, err.Error()
	}

	return checkResultOK, fmt.Sprintf("created hardlink from %s to %s", config.SourceDirAbsPath, config.TempDir)
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
