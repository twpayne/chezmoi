package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
	"github.com/twpayne/go-vfs"
)

// An AddCommandConfig is a configuration for the add command.
type AddCommandConfig struct {
	Empty     bool
	Recursive bool
	Template  bool
}

// A Config represents a configuration.
type Config struct {
	SourceDir        string
	TargetDir        string
	Umask            int
	DryRun           bool
	Verbose          bool
	SourceVCSCommand string
	Data             map[string]interface{}
	Add              AddCommandConfig
}

func (c *Config) exec(argv []string) error {
	path, err := exec.LookPath(argv[0])
	if err != nil {
		return err
	}
	if c.Verbose {
		fmt.Printf("exec %s\n", strings.Join(argv, " "))
	}
	if c.DryRun {
		return nil
	}
	return syscall.Exec(path, argv, os.Environ())
}

func (c *Config) getDefaultActuator(fs vfs.FS) chezmoi.Actuator {
	var actuator chezmoi.Actuator
	if c.DryRun {
		actuator = chezmoi.NewNullActuator()
	} else {
		actuator = chezmoi.NewFSActuator(fs, c.TargetDir)
	}
	if c.Verbose {
		actuator = chezmoi.NewLoggingActuator(os.Stdout, actuator)
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

func (c *Config) getSourceNames(targetState *chezmoi.TargetState, targets []string) ([]string, error) {
	sourceNames := []string{}
	allEntries := targetState.AllEntries()
	for _, target := range targets {
		absTarget, err := filepath.Abs(target)
		if err != nil {
			return nil, err
		}
		targetName, err := filepath.Rel(c.TargetDir, absTarget)
		if err != nil {
			return nil, err
		}
		if filepath.HasPrefix(targetName, "..") {
			return nil, fmt.Errorf("%s: not in target directory", target)
		}
		entry, ok := allEntries[targetName]
		if !ok {
			return nil, fmt.Errorf("%s: not found", targetName)
		}
		sourceNames = append(sourceNames, entry.SourceName())
	}
	return sourceNames, nil
}

func (c *Config) getTargetState(fs vfs.FS) (*chezmoi.TargetState, error) {
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
	targetState := chezmoi.NewTargetState(c.TargetDir, os.FileMode(c.Umask), c.SourceDir, data)
	if err := targetState.Populate(fs); err != nil {
		return nil, err
	}
	return targetState, nil
}

func makeRunE(runCommand func(vfs.FS, *cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return runCommand(vfs.OSFS, cmd, args)
	}
}

func printErrorAndExit(err error) {
	fmt.Println(err)
	os.Exit(1)
}
