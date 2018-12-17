package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs"
)

var addCommand = &cobra.Command{
	Use:   "add",
	Args:  cobra.MinimumNArgs(1),
	Short: "Add an existing file or directory",
	RunE:  makeRunE(config.runAddCommand),
}

// An AddCommandConfig is a configuration for the add command.
type addCommandConfig struct {
	empty     bool
	recursive bool
	template  bool
}

func init() {
	rootCommand.AddCommand(addCommand)

	persistentFlags := addCommand.PersistentFlags()
	persistentFlags.BoolVarP(&config.add.empty, "empty", "e", false, "add empty files")
	persistentFlags.BoolVarP(&config.add.recursive, "recursive", "r", false, "recurse in to subdirectories")
	persistentFlags.BoolVarP(&config.add.template, "template", "T", false, "add files as templates")
}

func (c *Config) runAddCommand(fs vfs.FS, command *cobra.Command, args []string) error {
	ts, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	mutator := c.getDefaultMutator(fs)
	info, err := fs.Stat(c.SourceDir)
	switch {
	case err == nil && info.Mode().IsDir():
		if info.Mode().Perm() != 0700 {
			if err := mutator.Chmod(c.SourceDir, 0700); err != nil {
				return err
			}
		}
	case os.IsNotExist(err):
		if err := mutator.Mkdir(c.SourceDir, 0700); err != nil {
			return err
		}
	case err == nil:
		return fmt.Errorf("%s: is not a directory", c.SourceDir)
	default:
		return err
	}
	for _, arg := range args {
		path, err := filepath.Abs(arg)
		if err != nil {
			return err
		}
		if c.add.recursive {
			if err := vfs.Walk(fs, path, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				return ts.Add(fs, path, info, c.add.empty, c.add.template, mutator)
			}); err != nil {
				return err
			}
		} else {
			if err := ts.Add(fs, path, nil, c.add.empty, c.add.template, mutator); err != nil {
				return err
			}
		}
	}
	return nil
}
