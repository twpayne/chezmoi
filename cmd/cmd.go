package cmd

import (
	"log"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"syscall"

	"github.com/absfs/afero"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
)

// A Config represents a configuration.
type Config struct {
	SourceDir        string
	TargetDir        string
	Umask            int
	DryRun           bool
	Verbose          bool
	SourceVCSCommand string
	Data             map[string]interface{}
	Add              struct {
		Recursive bool
		Template  bool
	}
}

func (c *Config) exec(argv []string) error {
	path, err := exec.LookPath(argv[0])
	if err != nil {
		return err
	}
	if c.Verbose {
		log.Printf("exec %s", strings.Join(argv, " "))
	}
	if c.DryRun {
		return nil
	}
	return syscall.Exec(path, argv, os.Environ())
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
	data := map[string]interface{}{
		"arch": runtime.GOARCH,
		"os":   runtime.GOOS,
	}

	currentUser, err := user.Current()
	if err != nil {
		return nil, err
	}
	data["username"] = currentUser.Username

	group, err := user.LookupGroupId(currentUser.Gid)
	if err != nil {
		return nil, err
	}
	data["group"] = group.Name

	homedir, err := homedir.Dir()
	if err != nil {
		return nil, err
	}
	data["homedir"] = homedir

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	data["hostname"] = hostname

	return data, nil
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
