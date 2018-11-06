package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/absfs/afero"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var editCommand = &cobra.Command{
	Use:   "edit",
	Args:  cobra.MinimumNArgs(1),
	Short: "Edit a file",
	Run:   makeRun(runEditCommand),
}

func init() {
	rootCommand.AddCommand(editCommand)
}

func runEditCommand(command *cobra.Command, args []string) error {
	targetState, err := getTargetState(afero.NewOsFs())
	if err != nil {
		return err
	}
	sourceFileNames := []string{}
	for _, arg := range args {
		fileState := targetState.FindSourceFile(arg)
		if fileState == nil {
			return errors.Errorf("%s: not found", arg)
		}
		sourceFileNames = append(sourceFileNames, filepath.Join(sourceDir, fileState.SourceName))
	}
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	editorPath, err := exec.LookPath(editor)
	if err != nil {
		return err
	}
	argv := append([]string{editor}, sourceFileNames...)
	return syscall.Exec(editorPath, argv, os.Environ())
}
