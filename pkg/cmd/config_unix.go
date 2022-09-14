//go:build !windows
// +build !windows

package cmd

func (c *Config) windowsVersion() (map[string]any, error) {
	return nil, nil
}
