// +build !windows

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	vfs "github.com/twpayne/go-vfs"
)

func (c *Config) exec(fs vfs.FS, argv []string) error {
	path, err := exec.LookPath(argv[0])
	if err != nil {
		return err
	}
	if c.Verbose {
		fmt.Printf("exec %s\n", strings.Join(argv, " "))
	}
	if c.DryRun {
		return nil
	}

	return syscall.Exec(path, argv, os.Environ())
}
