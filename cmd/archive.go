package cmd

import (
	"archive/tar"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type archiveCmdConfig struct {
	output string
}

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

	persistentFlags := archiveCmd.PersistentFlags()
	persistentFlags.StringVarP(&config.archive.output, "output", "o", "", "output filename")
	panicOnError(archiveCmd.MarkPersistentFlagFilename("output"))
}

func (c *Config) runArchiveCmd(cmd *cobra.Command, args []string) error {
	ts, err := c.getTargetState(nil)
	if err != nil {
		return err
	}

	output := &strings.Builder{}
	w := tar.NewWriter(output)
	if err := ts.Archive(w, os.FileMode(c.Umask)); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	if c.archive.output == "" {
		_, err := c.Stdout.Write([]byte(output.String()))
		return err
	}
	return c.fs.WriteFile(c.archive.output, []byte(output.String()), 0o666)
}
