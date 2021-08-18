//go:build !windows
// +build !windows

package cmd

import (
	"errors"
	"os"

	"golang.org/x/term"
)

func (c *Config) readPassword(prompt string) (string, error) {
	if c.noTTY {
		return c.readLine(prompt)
	}

	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = tty.Close()
	}()
	if _, err := tty.Write([]byte(prompt)); err != nil {
		return "", err
	}
	password, err := term.ReadPassword(int(tty.Fd()))
	if err != nil && !errors.Is(err, term.ErrPasteIndicator) {
		return "", err
	}
	if _, err := tty.Write([]byte{'\n'}); err != nil {
		return "", err
	}
	return string(password), nil
}
