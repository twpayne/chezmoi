package cmd

import (
	"fmt"

	"golang.org/x/sys/windows"
	"golang.org/x/term"
)

// readPassword reads a password.
func (c *Config) readPassword(prompt string) (string, error) {
	if c.noTTY {
		return c.readLine(prompt)
	}

	name, err := windows.UTF16PtrFromString("CONIN$")
	if err != nil {
		return "", err
	}
	handle, err := windows.CreateFile(name, windows.GENERIC_READ|windows.GENERIC_WRITE, windows.FILE_SHARE_READ, nil, windows.OPEN_EXISTING, 0, 0)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = windows.CloseHandle(handle)
	}()
	//nolint:forbidigo
	fmt.Print(prompt)
	password, err := term.ReadPassword(int(handle))
	if err != nil {
		return "", err
	}
	//nolint:forbidigo
	fmt.Println("")
	return string(password), nil
}
