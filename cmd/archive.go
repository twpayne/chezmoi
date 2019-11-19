package cmd

import (
	"archive/tar"
	"os"

	"github.com/spf13/cobra"
)

var archiveCmd = &cobra.Command{
	Use:     "archive",
	Args:    cobra.NoArgs,
	Short:   "Write a tar archive of the target state to stdout",
	Long:    mustGetLongHelp("archive"),
	Example: getExample("archive"),
	PreRunE: config.ensureNoError,
	RunE:    config.runArchiveCmd,
}

func init() {
	rootCmd.AddCommand(archiveCmd)
}

func (c *Config) runArchiveCmd(cmd *cobra.Command, args []string) error {
	ts, err := c.getTargetState(nil)
	if err != nil {
		return err
	}
	w := tar.NewWriter(c.Stdout())
	if err := ts.Archive(w, os.FileMode(c.Umask)); err != nil {
		return err
	}
	return w.Close()
}
