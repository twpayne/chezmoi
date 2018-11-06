package cmd

import (
	"log"
	"path/filepath"

	"github.com/absfs/afero"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
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

func getSourceFileStates(targetState *chezmoi.RootState, targetFileNames []string) ([]*chezmoi.FileState, error) {
	fileStates := []*chezmoi.FileState{}
	for _, targetFileName := range targetFileNames {
		fileState := targetState.FindSourceFile(targetFileName)
		if fileState == nil {
			return nil, errors.Errorf("%s: file not found", targetFileName)
		}
		fileStates = append(fileStates, fileState)
	}
	return fileStates, nil
}

func getSourceFileNames(targetState *chezmoi.RootState, targetFileNames []string) ([]string, error) {
	fileStates, err := getSourceFileStates(targetState, targetFileNames)
	if err != nil {
		return nil, err
	}
	sourceFileNames := []string{}
	for _, fileState := range fileStates {
		sourceFileName := filepath.Join(sourceDir, fileState.SourceName)
		sourceFileNames = append(sourceFileNames, sourceFileName)
	}
	return sourceFileNames, nil
}

func getTargetState(fs afero.Fs) (*chezmoi.RootState, error) {
	targetState := chezmoi.NewRootState()
	if err := targetState.Populate(fs, sourceDir, nil); err != nil {
		return nil, err
	}
	return targetState, nil
}
