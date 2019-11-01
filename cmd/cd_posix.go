// +build !windows

package cmd

import (
	shell "github.com/twpayne/go-shell"
	vfs "github.com/twpayne/go-vfs"
)

func (c *Config) runCDCmd(fs vfs.FS, args []string) error {
	mutator := c.getDefaultMutator(fs)
	if err := c.ensureSourceDirectory(fs, mutator); err != nil {
		return err
	}

	shell, _ := shell.CurrentUserShell()
	return c.run(fs, c.SourceDir, shell)
}
