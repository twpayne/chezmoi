package cmd

import (
	"archive/tar"
	"os"

	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs"
)

var archiveCmd = &cobra.Command{
	Use:     "archive",
	Args:    cobra.NoArgs,
	Short:   "Write a tar archive of the target state to stdout",
	PreRunE: config.ensureNoError,
	RunE:    makeRunE(config.runArchiveCmd),
}

func init() {
	rootCmd.AddCommand(archiveCmd)
}

func (c *Config) runArchiveCmd(fs vfs.FS, args []string) error {
	ts, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	w := tar.NewWriter(c.Stdout())
	if err := ts.Archive(w, os.FileMode(c.Umask)); err != nil {
		return err
	}
	return w.Close()
}
