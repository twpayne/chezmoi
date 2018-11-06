package cmd

import (
	"log"
	"path/filepath"

	"github.com/absfs/afero"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
)

var (
	sourceDir string
	targetDir string
)

var rootCommand = &cobra.Command{
	Use:   "chezmoi",
	Short: "chezmoi manages your home directory",
}

func init() {
	homeDir, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}
	sourceDir = filepath.Join(homeDir, ".chezmoi")
	targetDir = homeDir

	persistentFlags := rootCommand.PersistentFlags()
	persistentFlags.StringVar(&sourceDir, "source", sourceDir, "source directory")
	persistentFlags.StringVar(&targetDir, "target", targetDir, "target directory")
}

func Execute() {
	if err := rootCommand.Execute(); err != nil {
		log.Fatal(err)
	}
}

func getTargetState(fs afero.Fs) (*chezmoi.RootState, error) {
	targetState := chezmoi.NewRootState()
	if err := targetState.Populate(fs, sourceDir, nil); err != nil {
		return nil, err
	}
	return targetState, nil
}
