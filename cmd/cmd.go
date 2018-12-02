package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"text/template"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
	"github.com/twpayne/go-vfs"
)

// A Version represents a version.
type Version struct {
	Version string
	Commit  string
	Date    string
}

// A Config represents a configuration.
type Config struct {
	version          Version
	SourceDir        string
	TargetDir        string
	Umask            int
	DryRun           bool
	Verbose          bool
	SourceVCSCommand string
	LastPass         LastPassCommandConfig
	Data             map[string]interface{}
	funcs            template.FuncMap
	add              addCommandConfig
	keyring          keyringCommandConfig
}

func (c *Config) addFunc(key string, value interface{}) {
	if c.funcs == nil {
		c.funcs = make(template.FuncMap)
	}
	if _, ok := c.funcs[key]; ok {
		panic(fmt.Sprintf("Config.addFunc: %s already defined", key))
	}
	c.funcs[key] = value
}

func (c *Config) applyArgs(fs vfs.FS, args []string, actuator chezmoi.Actuator) error {
	targetState, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	if len(args) == 0 {
		return targetState.Apply(fs, actuator)
	}
	for _, arg := range args {
		targetPath, err := filepath.Abs(arg)
		if err != nil {
			return err
		}
		entry, err := targetState.Get(targetPath)
		if err != nil {
			return err
		}
		if entry == nil {
			return fmt.Errorf("%s: not under management", arg)
		}
		if err := targetState.ApplyOne(fs, targetPath, entry, actuator); err != nil {
			return err
		}
	}
	return nil
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
	targetState := chezmoi.NewTargetState(c.TargetDir, os.FileMode(c.Umask), c.SourceDir, data, c.funcs)
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

func prompt(s, choices string) (byte, error) {
	r := bufio.NewReader(os.Stdin)
	for {
		_, err := fmt.Printf("%s [%s]? ", s, strings.Join(strings.Split(choices, ""), ","))
		if err != nil {
			return 0, err
		}
		line, err := r.ReadString('\n')
		if err != nil {
			return 0, err
		}
		if len(line) == 2 && strings.IndexByte(choices, line[0]) != -1 {
			return line[0], nil
		}
	}
}
