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
	"github.com/spf13/viper"
	"github.com/twpayne/chezmoi/lib/chezmoi"
	vfs "github.com/twpayne/go-vfs"
	xdg "github.com/twpayne/go-xdg"
)

// A Config represents a configuration.
type Config struct {
	SourceDir        string
	DestDir          string
	Umask            octalIntValue
	DryRun           bool
	Verbose          bool
	SourceVCSCommand string
	Bitwarden        bitwardenCommandConfig
	LastPass         LastPassCommandConfig
	Data             map[string]interface{}
	funcs            template.FuncMap
	add              addCommandConfig
	dump             dumpCommandConfig
	edit             editCommandConfig
	_import          importCommandConfig
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
		if err := entry.Apply(fs, ts.DestDir, ts.TargetIgnore.Match, ts.Umask, mutator); err != nil {
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

func (c *Config) execEditor(argv ...string) error {
	return c.exec(append([]string{c.getEditor()}, argv...))
}

func (c *Config) getDefaultMutator(fs vfs.FS) chezmoi.Mutator {
	var mutator chezmoi.Mutator
	if c.DryRun {
		mutator = chezmoi.NullMutator
	} else {
		mutator = chezmoi.NewFSMutator(fs, c.DestDir)
	}
	if c.Verbose {
		mutator = chezmoi.NewLoggingMutator(os.Stdout, mutator)
	}
	return mutator
}

func (c *Config) getEditor() string {
	if editor := os.Getenv("VISUAL"); editor != "" {
		return editor
	}
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	return "vi"
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
			return nil, fmt.Errorf("%s: not in source state", arg)
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
	ts := chezmoi.NewTargetState(c.DestDir, os.FileMode(c.Umask), c.SourceDir, data, c.funcs)
	readOnlyFS := vfs.NewReadOnlyFS(fs)
	if err := ts.Populate(readOnlyFS); err != nil {
		return nil, err
	}
	return ts, nil
}

func (c *Config) runEditor(argv ...string) error {
	editor := c.getEditor()
	if c.Verbose {
		fmt.Printf("%s %s\n", editor, strings.Join(argv, " "))
	}
	if c.DryRun {
		return nil
	}
	cmd := exec.Command(editor, argv...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func getDefaultConfigFile(x *xdg.XDG, homeDir string) string {
	// Search XDG config directories first.
	for _, configDir := range x.ConfigDirs {
		for _, extension := range viper.SupportedExts {
			configFilePath := filepath.Join(configDir, "chezmoi", "chezmoi."+extension)
			if _, err := os.Stat(configFilePath); err == nil {
				return configFilePath
			}
		}
	}
	// Fallback to XDG default.
	return filepath.Join(x.ConfigHome, "chezmoi", "chezmoi.yaml")
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

func getDefaultSourceDir(x *xdg.XDG, homeDir string) string {
	// Check for XDG data directories first.
	for _, dataDir := range x.DataDirs {
		sourceDir := filepath.Join(dataDir, "chezmoi")
		if _, err := os.Stat(sourceDir); err == nil {
			return sourceDir
		}
	}
	// Fallback to XDG default.
	return filepath.Join(x.DataHome, "chezmoi")
}

func makeRunE(runCommand func(vfs.FS, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return runCommand(vfs.OSFS, args)
	}
}

func printErrorAndExit(err error) {
	fmt.Printf("chezmoi: %v\n", err)
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
