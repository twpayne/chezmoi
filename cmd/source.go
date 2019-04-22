package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs"
)

var sourceCmd = &cobra.Command{
	Use:     "source [args...]",
	Short:   "Run the source version control system command in the source directory",
	Long:    mustGetLongHelp("source"),
	Example: getExample("source"),
	PreRunE: config.ensureNoError,
	RunE:    makeRunE(config.runSourceCmd),
}

func init() {
	rootCmd.AddCommand(sourceCmd)
}

func (c *Config) runSourceCmd(fs vfs.FS, args []string) error {
	if c.Verbose {
		fmt.Printf("cd %s\n", c.SourceDir)
	}
	if !c.DryRun {
		if err := os.Chdir(c.SourceDir); err != nil {
			return err
		}
	}
	return c.exec(append([]string{c.SourceVCS.Command}, args...))
}
