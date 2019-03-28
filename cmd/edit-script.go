package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
	vfs "github.com/twpayne/go-vfs"
)

var editScriptsCmd = &cobra.Command{
	Use:   "edit-scripts scripts",
	Args:  cobra.MinimumNArgs(1),
	Short: "Edit scripts in the scripts folder",
	RunE:  makeRunE(config.runEditScriptCmd),
}

type editScriptsCmdConfig struct {
	run    bool
	prompt bool
}

func init() {
	rootCmd.AddCommand(editScriptsCmd)

	persistentFlags := editScriptsCmd.PersistentFlags()
	persistentFlags.BoolVarP(&config.editScripts.run, "run", "r", false, "run after editing")
	persistentFlags.BoolVarP(&config.editScripts.prompt, "prompt", "p", false, "prompt before running")
}

func (c *Config) runEditScriptCmd(fs vfs.FS, args []string) error {
	ts, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	var scripts []*chezmoi.Script
	for _, arg := range args {
		if s, ok := ts.Scripts[chezmoi.StripTemplateExtension(arg)]; ok {
			scripts = append(scripts, s)
		}
	}

	argv := []string{}
	for _, s := range scripts {
		argv = append(argv, s.SourcePath)
	}
	if !c.editScripts.run {
		return c.execEditor(argv...)
	}
	if err := c.runEditor(argv...); err != nil {
		return err
	}

	for _, s := range scripts {
		if c.editScripts.prompt {
			choice, err := prompt(fmt.Sprintf("Apply %s", s.Name), "ynqa")
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
				c.editScripts.prompt = false
			}
		}
		if err := s.Apply(c.DestDir, c.DryRun); err != nil {
			return err
		}
	}
	return nil
}
