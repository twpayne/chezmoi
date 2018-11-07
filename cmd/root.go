package cmd

import (
	"log"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

var config Config

var rootCommand = &cobra.Command{
	Use:   "chezmoi",
	Short: "chezmoi manages your home directory",
}

func init() {
	homeDir, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}
	config.SourceDir = filepath.Join(homeDir, ".chezmoi")
	config.TargetDir = homeDir

	persistentFlags := rootCommand.PersistentFlags()
	persistentFlags.BoolVarP(&config.DryRun, "dry-run", "n", config.DryRun, "dry run")
	persistentFlags.StringVarP(&config.SourceDir, "source", "s", config.SourceDir, "source directory")
	persistentFlags.StringVarP(&config.TargetDir, "target", "t", config.TargetDir, "target directory")
	persistentFlags.BoolVarP(&config.Verbose, "verbose", "v", config.Verbose, "verbose")
}

func Execute() {
	if err := rootCommand.Execute(); err != nil {
		log.Fatal(err)
	}
}
