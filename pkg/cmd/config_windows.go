package cmd

import (
	"fmt"

	"go.uber.org/multierr"
	"golang.org/x/sys/windows"
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
		return
	}
	var handle windows.Handle
	if handle, err = windows.CreateFile(name, windows.GENERIC_READ|windows.GENERIC_WRITE, windows.FILE_SHARE_READ, nil, windows.OPEN_EXISTING, 0, 0); err != nil {
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
