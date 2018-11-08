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
	RunE:  makeRunE(runEditCommandE),
}

func init() {
	rootCommand.AddCommand(editCommand)
}

func runEditCommandE(fs afero.Fs, command *cobra.Command, args []string) error {
	targetState, err := config.getTargetState(fs)
	if err != nil {
		return err
	}
	sourceFileNames, err := config.getSourceNames(targetState, args)
	if err != nil {
		return err
	}
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	editorPath, err := exec.LookPath(editor)
	if err != nil {
		return err
	}
	sourceFilePaths := []string{}
	for _, sourceFileName := range sourceFileNames {
		sourceFilePaths = append(sourceFilePaths, filepath.Join(config.SourceDir, sourceFileName))
	}
	argv := append([]string{editor}, sourceFilePaths...)
	if config.Verbose {
		log.Printf("exec %s", strings.Join(argv, " "))
	}
	if config.DryRun {
		return nil
	}
	return syscall.Exec(editorPath, argv, os.Environ())
}
