package cmd

import (
	"log"
	"os"
	"path/filepath"
	"syscall"

	"github.com/absfs/afero"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
)

type Config struct {
	SourceDir string
	TargetDir string
	DryRun    bool
	Verbose   bool
}

func (c *Config) getDefaultActuator(fs afero.Fs) chezmoi.Actuator {
	var actuator chezmoi.Actuator
	if c.DryRun {
		actuator = chezmoi.NewNullActuator()
	} else {
		actuator = chezmoi.NewFsActuator(fs)
	}
	if c.Verbose {
		actuator = chezmoi.NewLoggingActuator(actuator)
	}
	return actuator
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

func (c *Config) getSourceFileNames(targetState *chezmoi.RootState, targetFileNames []string) ([]string, error) {
	fileStates, err := getSourceFileStates(targetState, targetFileNames)
	if err != nil {
		return nil, err
	}
	sourceFileNames := []string{}
	for _, fileState := range fileStates {
		sourceFileName := filepath.Join(c.SourceDir, fileState.SourceName)
		sourceFileNames = append(sourceFileNames, sourceFileName)
	}
	return sourceFileNames, nil
}

func (c *Config) getTargetState(fs afero.Fs) (*chezmoi.RootState, error) {
	targetState := chezmoi.NewRootState()
	if err := targetState.Populate(fs, c.SourceDir, nil); err != nil {
		return nil, err
	}
	return targetState, nil
}

func makeRun(runCommand func(afero.Fs, *cobra.Command, []string) error) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		if err := runCommand(afero.NewOsFs(), cmd, args); err != nil {
			log.Fatal(err)
		}
	}
}

func getUmask() os.FileMode {
	// FIXME should we call runtime.LockOSThread or similar?
	umask := syscall.Umask(0)
	syscall.Umask(umask)
	return os.FileMode(umask)
}
