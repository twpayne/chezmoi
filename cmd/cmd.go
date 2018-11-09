package cmd

import (
	"os"

	"github.com/absfs/afero"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
)

type Config struct {
	SourceDir string
	TargetDir string
	Umask     int
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
	targetState := chezmoi.NewRootState(c.TargetDir, os.FileMode(c.Umask), c.SourceDir, c.Data)
	if err := targetState.Populate(fs); err != nil {
		return nil, err
	}
	return targetState, nil
}

func makeRunE(runCommand func(afero.Fs, *cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return runCommand(afero.NewOsFs(), cmd, args)
	}
}
