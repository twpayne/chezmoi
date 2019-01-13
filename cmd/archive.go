package cmd

import (
	"archive/tar"
	"os"

	"github.com/spf13/cobra"
	"github.com/twpayne/go-vfs"
)

var archiveCommand = &cobra.Command{
	Use:   "archive",
	Args:  cobra.NoArgs,
	Short: "Write a tar archive of the target state to stdout",
	RunE:  makeRunE(config.runArchiveCommand),
}

func init() {
	rootCommand.AddCommand(archiveCommand)
}

func (c *Config) runArchiveCommand(fs vfs.FS, args []string) error {
	ts, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	w := tar.NewWriter(os.Stdout)
	if err := ts.Archive(w, os.FileMode(c.Umask)); err != nil {
		return err
	}
	return w.Close()
}
