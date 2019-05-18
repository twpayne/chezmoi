// +build !snap

package cmd

// snapFix, when not built for snap, does nothing.
func (c *Config) snapFix() error {
	return nil
}
