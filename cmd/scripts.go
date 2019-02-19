package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
	vfs "github.com/twpayne/go-vfs"
)

var scriptsCmd = &cobra.Command{
	Use:   "scripts [targets...]",
	Short: "Run scripts that need to run",
	RunE:  makeRunE(config.runScriptsCmd),
}

type scriptsCmdConfig struct {
	force  bool
	prompt bool
}

func init() {
	rootCmd.AddCommand(scriptsCmd)

	persistentFlags := scriptsCmd.PersistentFlags()
	persistentFlags.BoolVarP(&config.scripts.force, "force", "f", false, "run all scripts")
	persistentFlags.BoolVarP(&config.scripts.prompt, "prompt", "p", false, "prompt before running each script")
}

func (c *Config) runScriptsCmd(fs vfs.FS, args []string) error {
	if config.DryRun {
		println("chezmoi: the scripts subcommand doesn't support dry-run yet")
		return nil
	}

	ts, err := c.getTargetState(fs)
	if err != nil {
		return err
	}

	if len(args) == 0 && !config.scripts.prompt {
		return ts.ApplyScripts(fs, config.scripts.force)
	}

	var quit int
	defer func() {
		if r := recover(); r != nil {
			if p, ok := r.(*int); ok && p == &quit {
				err = nil
			} else {
				panic(r)
			}
		}
	}()
	var scripts []*chezmoi.Script
	if len(args) == 0 {
		for _, s := range ts.Scripts {
			if config.scripts.prompt {
				choice, err := c.prompt(fmt.Sprintf("Run %s", s.Name), "ynqa")
				if err != nil {
					return err
				}
				switch choice {
				case 'a':
					c.scripts.prompt = false
					fallthrough
				case 'y':
					scripts = append(scripts, s)
				case 'n':
				case 'q':
					panic(&quit) // abort vfs.Walk by panicking
				}
			} else {
				scripts = append(scripts, s)
			}
		}
	} else {
		for _, arg := range args {
			s, ok := ts.Scripts[chezmoi.StripTemplateExtension(arg)]
			if ok {
				if !config.scripts.prompt {
					scripts = append(scripts, s)
				} else {
					choice, err := c.prompt(fmt.Sprintf("Run %s", s.Name), "ynqa")
					if err != nil {
						return err
					}
					switch choice {
					case 'y':
						scripts = append(scripts, s)
					case 'n':
					case 'q':
						panic(&quit) // abort vfs.Walk by panicking
					case 'a':
						c.scripts.prompt = false
					}
				}
			}
		}
	}
	for _, s := range scripts {
		if err := s.Apply(); err != nil {
			return err
		}
	}
	return nil
}
