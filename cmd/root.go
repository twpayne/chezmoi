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
	dryRun    = false
	verbose   = false
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
	persistentFlags.BoolVar(&dryRun, "dry-run", dryRun, "dry run")
	persistentFlags.StringVar(&sourceDir, "source", sourceDir, "source directory")
	persistentFlags.StringVar(&targetDir, "target", targetDir, "target directory")
	persistentFlags.BoolVar(&verbose, "verbose", verbose, "verbose")
}

func Execute() {
	if err := rootCommand.Execute(); err != nil {
		log.Fatal(err)
	}
}

func getDefaultActuator(fs afero.Fs) chezmoi.Actuator {
	var actuator chezmoi.Actuator
	if dryRun {
		actuator = chezmoi.NewNullActuator()
	} else {
		actuator = chezmoi.NewFsActuator(fs)
	}
	if verbose {
		actuator = chezmoi.NewLoggingActuator(actuator)
	}
	return actuator
}
