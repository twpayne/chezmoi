package cmd

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/windows/registry"

	"chezmoi.io/chezmoi/internal/chezmoi"
)

const defaultEditor = "notepad.exe"

// getPS1Interpreter returns the appropriate Interpreter for PowerShell
// scripts (.ps1) on Windows systems. It uses the provided findExecutable
// function to check for the presence of 'pwsh'. If 'pwsh' is not found, it
// returns an empty Interpreter, indicating no suitable interpreter is
// available.
func getPS1Interpreter(findExecutable func([]string, []string) (string, error)) chezmoi.Interpreter {
	paths := strings.Split(os.Getenv("PATH"), string(os.PathListSeparator))
	var interpreter chezmoi.Interpreter

	if pwshPath, _ := findExecutable([]string{"pwsh.exe", "pwsh"}, paths); pwshPath != "" {
		interpreter = chezmoi.Interpreter{
			Command: "pwsh",
			Args:    []string{"-NoLogo", "-File"},
		}
	} else if powershellPath, _ := findExecutable([]string{"powershell.exe", "powershell"}, paths); powershellPath != "" {
		interpreter = chezmoi.Interpreter{
			Command: "powershell",
			Args:    []string{"-NoLogo", "-File"},
		}
	}

	return interpreter
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
