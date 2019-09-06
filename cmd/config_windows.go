// +build windows

package cmd

import vfs "github.com/twpayne/go-vfs"

// exec, on windows, calls run since legit exec doesn't really exist.
func (c *Config) exec(fs vfs.FS, argv []string) error {
	return c.run(fs, "", argv[0], argv[1:]...)
}
