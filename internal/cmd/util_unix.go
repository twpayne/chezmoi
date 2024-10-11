//go:build unix

package cmd

import (
	"io/fs"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

const defaultEditor = "vi"

var defaultInterpreters = make(map[string]chezmoi.Interpreter)

func fileInfoUID(info fs.FileInfo) int {
	return int(info.Sys().(*syscall.Stat_t).Uid) //nolint:forcetypeassert
}

func darwinVersion() (map[string]any, error) {
	if runtime.GOOS != "darwin" {
		return nil, nil
	}
	darwinVersion := make(map[string]any)
	for _, name := range []string{
		"buildVersion",
		"productName",
		"productVersion",
		"productVersionExtra",
	} {
		output, err := exec.Command("sw_vers", "--"+name).Output()
		if err != nil {
			return nil, err
		}
		darwinVersion[name] = strings.TrimSpace(string(output))
	}

	productVersion, ok := darwinVersion["productVersion"].(string)
	if ok {
		versionParts := strings.Split(productVersion, ".")
		var err error
		major, minor, patch := 0, 0, 0
		if len(versionParts) > 0 {
			major, err = strconv.Atoi(versionParts[0])
			if err != nil {
				return nil, err
			}
		}
		if len(versionParts) > 1 {
			minor, err = strconv.Atoi(versionParts[1])
			if err != nil {
				return nil, err
			}
		}
		if len(versionParts) > 2 {
			patch, err = strconv.Atoi(versionParts[2])
			if err != nil {
				return nil, err
			}
		}
		darwinVersion["productMajorVersion"] = major
		darwinVersion["productMinorVersion"] = minor
		darwinVersion["productPatchVersion"] = patch
	}
	return darwinVersion, nil
}

func windowsVersion() (map[string]any, error) {
	return nil, nil
}
