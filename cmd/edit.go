package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
	vfs "github.com/twpayne/go-vfs"
)

var editCmd = &cobra.Command{
	Use:   "edit targets...",
	Args:  cobra.MinimumNArgs(1),
	Short: "Edit the source state of a target",
	RunE:  makeRunE(config.runEditCmd),
}

type editCmdConfig struct {
	apply  bool
	diff   bool
	prompt bool
}

func init() {
	rootCmd.AddCommand(editCmd)

	persistentFlags := editCmd.PersistentFlags()
	persistentFlags.BoolVarP(&config.edit.apply, "apply", "a", false, "apply edit after editing")
	persistentFlags.BoolVarP(&config.edit.diff, "diff", "d", false, "print diff after editing")
	persistentFlags.BoolVarP(&config.edit.prompt, "prompt", "p", false, "prompt before applying (implies --diff)")
}

func (c *Config) runEditCmd(fs vfs.FS, args []string) error {
	if c.edit.prompt {
		c.edit.diff = true
	}
	ts, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	entries, err := c.getEntries(fs, ts, args)
	if err != nil {
		return err
	}
	argv := []string{}
	for _, entry := range entries {
		argv = append(argv, filepath.Join(c.SourceDir, entry.SourceName()))
	}
	if !c.edit.diff && !c.edit.apply {
		return c.execEditor(argv...)
	}
	if err := c.runEditor(argv...); err != nil {
		return err
	}
	readOnlyFS := vfs.NewReadOnlyFS(fs)
	applyMutator := c.getDefaultMutator(fs)
	for i, entry := range entries {
		anyMutator := chezmoi.NewAnyMutator(chezmoi.NullMutator)
		var mutator chezmoi.Mutator = anyMutator
		if c.edit.diff {
			mutator = chezmoi.NewLoggingMutator(c.Stdout(), mutator)
		}
		if err := entry.Apply(readOnlyFS, ts.DestDir, ts.TargetIgnore.Match, ts.Umask, mutator); err != nil {
			return err
		}
		if c.edit.apply && anyMutator.Mutated() {
			if c.edit.prompt {
				choice, err := c.prompt(fmt.Sprintf("Apply %s", args[i]), "ynqa")
				if err != nil {
					return err
				}
				switch choice {
				case 'y':
				case 'n':
					continue
				case 'q':
					return nil
				case 'a':
					c.edit.prompt = false
				}
			}
			if err := entry.Apply(readOnlyFS, ts.DestDir, ts.TargetIgnore.Match, ts.Umask, applyMutator); err != nil {
				return err
			}
		}
	}
	return nil
}
