package cmd

import (
	"os"
	"os/user"
	"runtime"

	"github.com/absfs/afero"
	"github.com/mitchellh/go-homedir"
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
	Add       struct {
		Recursive bool
		Template  bool
	}
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

func getDefaultData() (map[string]interface{}, error) {
	currentUser, err := user.Current()
	if err != nil {
		return nil, err
	}
	group, err := user.LookupGroupId(currentUser.Gid)
	if err != nil {
		return nil, err
	}
	homedir, err := homedir.Dir()
	if err != nil {
		return nil, err
	}
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"arch":     runtime.GOARCH,
		"group":    group.Name,
		"homedir":  homedir,
		"hostname": hostname,
		"os":       runtime.GOOS,
		"username": currentUser.Username,
	}, nil
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
	defaultData, err := getDefaultData()
	if err != nil {
		return nil, err
	}
	data := map[string]interface{}{
		"chezmoi": defaultData,
	}
	for key, value := range c.Data {
		data[key] = value
	}
	targetState := chezmoi.NewRootState(c.TargetDir, os.FileMode(c.Umask), c.SourceDir, data)
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
