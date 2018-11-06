package cmd

import (
	"archive/tar"
	"os"

	"github.com/absfs/afero"
	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
)

var archiveCommand = &cobra.Command{
	Use:   "archive",
	Args:  cobra.NoArgs,
	Short: "Write a tar archive of the target state to stdout",
	Run:   makeRun(runArchiveCommand),
}

func init() {
	rootCommand.AddCommand(archiveCommand)
}

func runArchiveCommand(command *cobra.Command, args []string) error {
	fs := afero.NewOsFs()
	targetState, err := chezmoi.ReadTargetDirState(fs, sourceDir, nil)
	if err != nil {
		return err
	}
	w := tar.NewWriter(os.Stdout)
	if err := targetState.Archive(w); err != nil {
		return err
	}
	return w.Close()
}
