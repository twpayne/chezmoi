package cmd

import (
	"log"
	"os"
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
	Data      map[string]interface{}
}

func (c *Config) getDefaultActuator(fs afero.Fs) chezmoi.Actuator {
	var actuator chezmoi.Actuator
	if c.DryRun {
		actuator = chezmoi.NewNullActuator()
	} else {
		actuator = chezmoi.NewFsActuator(fs, c.TargetDir)
	}
	if c.Verbose {
		actuator = chezmoi.NewLoggingActuator(actuator)
	}
	return actuator
}

func (c *Config) getSourceNames(targetState *chezmoi.RootState, targetNames []string) ([]string, error) {
	sourceNames := []string{}
	allStates := targetState.AllStates()
	for _, targetName := range targetNames {
		state, ok := allStates[targetName]
		if !ok {
			return nil, errors.Errorf("%s: not found", targetName)
		}
		sourceNames = append(sourceNames, state.SourceName())
	}
	return sourceNames, nil
}

func (c *Config) getTargetState(fs afero.Fs) (*chezmoi.RootState, error) {
	targetState := chezmoi.NewRootState()
	if err := targetState.Populate(fs, c.SourceDir, c.Data); err != nil {
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
