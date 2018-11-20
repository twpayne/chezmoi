package cmd

import (
	"log"
	"os"

	"github.com/absfs/afero"
	"github.com/spf13/cobra"
)

var sourceCommand = &cobra.Command{
	Use:   "source",
	Short: "Run the source version control system command in the source directory",
	RunE:  makeRunE(config.runSourceCommand),
}

func init() {
	rootCommand.AddCommand(sourceCommand)
}

func (c *Config) runSourceCommand(fs afero.Fs, cmd *cobra.Command, args []string) error {
	if c.Verbose {
		log.Printf("cd %s", c.SourceDir)
	}
	if !c.DryRun {
		if err := os.Chdir(c.SourceDir); err != nil {
			return err
		}
	}
	return c.exec(append([]string{c.SourceVCSCommand}, args...))
}
