package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
	vfs "github.com/twpayne/go-vfs"
)

var addCommand = &cobra.Command{
	Use:   "add",
	Args:  cobra.MinimumNArgs(1),
	Short: "Add an existing file, directory, or symlink to the source state",
	RunE:  makeRunE(config.runAddCommand),
}

type addCommandConfig struct {
	recursive bool
	options   chezmoi.AddOptions
}

func init() {
	rootCommand.AddCommand(addCommand)

	persistentFlags := addCommand.PersistentFlags()
	persistentFlags.BoolVarP(&config.add.options.Empty, "empty", "e", false, "add empty files")
	persistentFlags.BoolVarP(&config.add.options.Exact, "exact", "x", false, "add directories exactly")
	persistentFlags.BoolVarP(&config.add.recursive, "recursive", "r", false, "recurse in to subdirectories")
	persistentFlags.BoolVarP(&config.add.options.Template, "template", "T", false, "add files as templates")
}

func (c *Config) runAddCommand(fs vfs.FS, args []string) error {
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
				return ts.Add(fs, c.add.options, path, info, mutator)
			}); err != nil {
				return err
			}
		} else {
			if err := ts.Add(fs, c.add.options, path, nil, mutator); err != nil {
				return err
			}
		}
	}
	return nil
}
