package cmd

import (
	"log"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

var (
	sourceDir string
	targetDir string
)

var rootCommand = &cobra.Command{
	Use:   "chezmoi",
	Short: "chezmoi manages your home directory",
	Run:   makeRun(runRootCommand),
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

func runRootCommand(cmd *cobra.Command, args []string) error {
	return nil
}
