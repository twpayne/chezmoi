package cmd

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/absfs/afero"
	"github.com/spf13/cobra"
)

var editCommand = &cobra.Command{
	Use:   "edit",
	Args:  cobra.MinimumNArgs(1),
	Short: "Edit a file",
	RunE:  makeRunE(config.runEditCommandE),
}

func init() {
	rootCommand.AddCommand(editCommand)
}

func (c *Config) runEditCommandE(fs afero.Fs, command *cobra.Command, args []string) error {
	targetState, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	sourceFileNames, err := c.getSourceNames(targetState, args)
	if err != nil {
		return err
	}
	editor := os.Getenv("VISUAL")
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		editor = "vi"
	}
	editorPath, err := exec.LookPath(editor)
	if err != nil {
		return err
	}
	sourceFilePaths := []string{}
	for _, sourceFileName := range sourceFileNames {
		sourceFilePaths = append(sourceFilePaths, filepath.Join(c.SourceDir, sourceFileName))
	}
	argv := append([]string{editor}, sourceFilePaths...)
	if c.Verbose {
		log.Printf("exec %s", strings.Join(argv, " "))
	}
	if c.DryRun {
		return nil
	}
	return syscall.Exec(editorPath, argv, os.Environ())
}
