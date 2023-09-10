package cmd

import (
	"fmt"
	"io/fs"
	"strings"

	"golang.org/x/sys/windows/registry"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

const defaultEditor = "notepad.exe"

var defaultInterpreters = map[string]*chezmoi.Interpreter{
	"bat": {},
	"cmd": {},
	"com": {},
	"exe": {},
	"nu": {
		Command: "nu",
	},
	"pl": {
		Command: "perl",
	},
	"ps1": {
		Command: "powershell",
		Args:    []string{"-NoLogo"},
	},
	"py": {
		Command: "python3",
	},
	"rb": {
		Command: "ruby",
	},
}

func fileInfoUID(fs.FileInfo) int {
	return 0
}

func windowsVersion() (map[string]any, error) {
	registryKey, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Windows NT\CurrentVersion`,
		registry.QUERY_VALUE,
	)
	if err != nil {
		return nil, fmt.Errorf("registry.OpenKey: %w", err)
	}
	windowsVersion := make(map[string]any)
	for _, name := range []string{
		"CurrentBuild",
		"CurrentVersion",
		"DisplayVersion",
		"EditionID",
		"ProductName",
	} {
		if value, _, err := registryKey.GetStringValue(name); err == nil {
			key := strings.ToLower(name[:1]) + name[1:]
			windowsVersion[key] = value
		}
	}
	for _, name := range []string{
		"CurrentMajorVersionNumber",
		"CurrentMinorVersionNumber",
	} {
		if value, _, err := registryKey.GetIntegerValue(name); err == nil {
			key := strings.ToLower(name[:1]) + name[1:]
			windowsVersion[key] = value
		}
	}
	return windowsVersion, nil
}
