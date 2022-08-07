package cmd

import (
	"fmt"
	"strings"

	"go.uber.org/multierr"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
	"golang.org/x/term"
)

// readPassword reads a password.
func (c *Config) readPassword(prompt string) (password string, err error) {
	if c.noTTY {
		password, err = c.readLine(prompt)
		return
	}

	if c.PINEntry.Command != "" {
		return c.readPINEntry(prompt)
	}

	var name *uint16
	name, err = windows.UTF16PtrFromString("CONIN$")
	if err != nil {
		err = fmt.Errorf("windows.UTF16PtrFromString: %w", err)
		return
	}
	var handle windows.Handle
	if handle, err = windows.CreateFile(name, windows.GENERIC_READ|windows.GENERIC_WRITE, windows.FILE_SHARE_READ, nil, windows.OPEN_EXISTING, 0, 0); err != nil {
		err = fmt.Errorf("windows.CreateFile: %w", err)
		return
	}
	defer func() {
		err = multierr.Append(err, windows.CloseHandle(handle))
	}()
	//nolint:forbidigo
	fmt.Print(prompt)
	var passwordBytes []byte
	if passwordBytes, err = term.ReadPassword(int(handle)); err != nil {
		return
	}
	//nolint:forbidigo
	fmt.Println("")
	password = string(passwordBytes)
	return
}

func (c *Config) windowsVersion() (map[string]any, error) {
	registryKey, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
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
