package cmd

import (
	"archive/tar"
	"os"

	"github.com/absfs/afero"
	"github.com/spf13/cobra"
)

var archiveCommand = &cobra.Command{
	Use:   "archive",
	Args:  cobra.NoArgs,
	Short: "Write a tar archive of the target state to stdout",
	RunE:  makeRunE(runArchiveCommandE),
}

func init() {
	rootCommand.AddCommand(archiveCommand)
}

func runArchiveCommandE(fs afero.Fs, command *cobra.Command, args []string) error {
	targetState, err := config.getTargetState(fs)
	if err != nil {
		return err
	}
	w := tar.NewWriter(os.Stdout)
	if err := targetState.Archive(w); err != nil {
		return err
	}
	return w.Close()
}
