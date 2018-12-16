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

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
	vfs "github.com/twpayne/go-vfs"
)

// A Config represents a configuration.
type Config struct {
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
	dump             dumpCommandConfig
	edit             editCommandConfig
	importC          importCommandConfig
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

func (c *Config) applyArgs(fs vfs.FS, args []string, mutator chezmoi.Mutator) error {
	ts, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	if len(args) == 0 {
		return ts.Apply(fs, mutator)
	}
	entries, err := c.getEntries(ts, args)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := entry.Apply(fs, ts.TargetDir, ts.Umask, mutator); err != nil {
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

func (c *Config) getDefaultMutator(fs vfs.FS) chezmoi.Mutator {
	var mutator chezmoi.Mutator
	if c.DryRun {
		mutator = chezmoi.NullMutator
	} else {
		mutator = chezmoi.NewFSMutator(fs, c.TargetDir)
	}
	if c.Verbose {
		mutator = chezmoi.NewLoggingMutator(os.Stdout, mutator)
	}
	return mutator
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

	// user.LookupGroupId reads /etc/group, which is not populated on some
	// systems, causing lookup to fail. Instead of returning the error, simply
	// ignore it and only set group if lookup succeeds.
	if group, err := user.LookupGroupId(currentUser.Gid); err == nil {
		data["group"] = group.Name
	}

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

func (c *Config) getEntries(ts *chezmoi.TargetState, args []string) ([]chezmoi.Entry, error) {
	entries := []chezmoi.Entry{}
	for _, arg := range args {
		targetPath, err := filepath.Abs(arg)
		if err != nil {
			return nil, err
		}
		entry, err := ts.Get(targetPath)
		if err != nil {
			return nil, err
		}
		if entry == nil {
			return nil, fmt.Errorf("%s: not under management", arg)
		}
		entries = append(entries, entry)
	}
	return entries, nil
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
	ts := chezmoi.NewTargetState(c.TargetDir, os.FileMode(c.Umask), c.SourceDir, data, c.funcs)
	readOnlyFS := vfs.NewReadOnlyFS(fs)
	if err := ts.Populate(readOnlyFS); err != nil {
		return nil, err
	}
	return ts, nil
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
