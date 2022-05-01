//go:build !windows
// +build !windows

package cmd

import (
	"errors"
	"os"

	"go.uber.org/multierr"
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

	var tty *os.File
	if tty, err = os.OpenFile("/dev/tty", os.O_RDWR, 0); err != nil {
		return
	}
	defer multierr.AppendInvoke(&err, multierr.Close(tty))
	if _, err = tty.Write([]byte(prompt)); err != nil {
		return
	}
	var passwordBytes []byte
	if passwordBytes, err = term.ReadPassword(int(tty.Fd())); err != nil && !errors.Is(err, term.ErrPasteIndicator) {
		return
	}
	if _, err = tty.Write([]byte{'\n'}); err != nil {
		return
	}
	password = string(passwordBytes)
	return
}
