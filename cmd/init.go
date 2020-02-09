package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/twpayne/chezmoi/internal/chezmoi"
	vfs "github.com/twpayne/go-vfs"
)

var initCmd = &cobra.Command{
	Args:    cobra.MaximumNArgs(1),
	Use:     "init [repo]",
	Short:   "Setup the source directory and update the destination directory to match the target state",
	Long:    mustGetLongHelp("init"),
	Example: getExample("init"),
	PreRunE: config.ensureNoError,
	RunE:    config.runInitCmd,
}

type initCmdConfig struct {
	apply bool
}

func init() {
	rootCmd.AddCommand(initCmd)

	persistentFlags := initCmd.PersistentFlags()
	persistentFlags.BoolVar(&config.init.apply, "apply", false, "update destination directory")
}

func (c *Config) runInitCmd(cmd *cobra.Command, args []string) error {
	vcs, err := c.getVCS()
	if err != nil {
		return err
	}

	if err := c.ensureSourceDirectory(); err != nil {
		return err
	}

	rawSourceDir, err := c.fs.RawPath(c.SourceDir)
	if err != nil {
		return err
	}

	switch len(args) {
	case 0: // init
		var initArgs []string
		if c.SourceVCS.Init != nil {
			switch v := c.SourceVCS.Init.(type) {
			case string:
				initArgs = strings.Split(v, " ")
			case []string:
				initArgs = v
			default:
				return fmt.Errorf("sourceVCS.init: cannot parse value")
			}
		} else {
			initArgs = vcs.InitArgs()
		}
		if err := c.run(c.SourceDir, c.SourceVCS.Command, initArgs...); err != nil {
			return err
		}
	case 1: // clone
		cloneArgs := vcs.CloneArgs(args[0], rawSourceDir)
		if cloneArgs == nil {
			return fmt.Errorf("%s: cloning not supported", c.SourceVCS.Command)
		}
		if err := c.run("", c.SourceVCS.Command, cloneArgs...); err != nil {
			return err
		}
		// FIXME this should be part of VCS
		if filepath.Base(c.SourceVCS.Command) == "git" {
			if _, err := c.fs.Stat(filepath.Join(c.SourceDir, ".gitmodules")); err == nil {
				for _, args := range [][]string{
					{"submodule", "init"},
					{"submodule", "update"},
				} {
					if err := c.run(c.SourceDir, c.SourceVCS.Command, args...); err != nil {
						return err
					}
				}
			}
		}
	}

	if err := c.createConfigFile(); err != nil {
		return err
	}

	if c.init.apply {
		persistentState, err := c.getPersistentState(nil)
		if err != nil {
			return err
		}
		if err := c.applyArgs(nil, persistentState); err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) createConfigFile() error {
	filename, ext, data, err := c.findConfigTemplate()
	if err != nil {
		return err
	}

	if filename == "" {
		// no config template file exists
		return nil
	}

	t, err := template.New(filename).Funcs(template.FuncMap{
		"promptString": c.promptString,
	}).Parse(data)
	if err != nil {
		return err
	}

	defaultData, err := c.getDefaultData()
	if err != nil {
		return err
	}

	contents := &bytes.Buffer{}
	if err = t.Execute(contents, map[string]interface{}{
		"chezmoi": defaultData,
	}); err != nil {
		return err
	}

	configDir := filepath.Join(c.bds.ConfigHome, "chezmoi")
	if err := vfs.MkdirAll(c.mutator, configDir, 0777&^os.FileMode(c.Umask)); err != nil {
		return err
	}

	configPath := filepath.Join(configDir, filename)
	if err := c.mutator.WriteFile(configPath, contents.Bytes(), 0600&^os.FileMode(c.Umask), nil); err != nil {
		return err
	}

	viper.SetConfigType(ext)
	if err := viper.ReadConfig(contents); err != nil {
		return err
	}
	return viper.Unmarshal(c)
}

func (c *Config) findConfigTemplate() (string, string, string, error) {
	for _, ext := range viper.SupportedExts {
		contents, err := c.fs.ReadFile(filepath.Join(c.SourceDir, ".chezmoi."+ext+chezmoi.TemplateSuffix))
		switch {
		case os.IsNotExist(err):
			continue
		case err != nil:
			return "", "", "", err
		}
		return "chezmoi." + ext, ext, string(contents), nil
	}
	return "", "", "", nil
}

func (c *Config) promptString(field string) string {
	fmt.Fprintf(c.Stdout(), "%s? ", field)
	value, err := bufio.NewReader(c.Stdin()).ReadString('\n')
	panicOnError(err)
	return strings.TrimSuffix(value, "\n")
}
