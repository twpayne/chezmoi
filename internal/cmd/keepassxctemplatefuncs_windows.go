package cmd

import "errors"

func (c *Config) keepassxcOutputOpen(command string, args ...string) ([]byte, error) {
	return nil, errors.New("not available on Windows")
}
