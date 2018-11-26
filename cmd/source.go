package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/twpayne/go-vfs"
)

var sourceCommand = &cobra.Command{
	Use:   "source",
	Short: "Run the source version control system command in the source directory",
	RunE:  makeRunE(config.runSourceCommand),
}

func init() {
	rootCommand.AddCommand(sourceCommand)
}

func (c *Config) runSourceCommand(fs vfs.FS, cmd *cobra.Command, args []string) error {
	if c.Verbose {
		fmt.Printf("cd %s\n", c.SourceDir)
	}
	if !c.DryRun {
		if err := os.Chdir(c.SourceDir); err != nil {
			return err
		}
	}
	return c.exec(append([]string{c.SourceVCSCommand}, args...))
}
