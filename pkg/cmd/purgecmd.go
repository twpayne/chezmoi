package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

type purgeCmdConfig struct {
	binary bool
}

func (c *Config) newPurgeCmd() *cobra.Command {
	purgeCmd := &cobra.Command{
		Use:     "purge",
		Short:   "Purge chezmoi's configuration and data",
		Long:    mustLongHelp("purge"),
		Example: example("purge"),
		Args:    cobra.NoArgs,
		RunE:    c.runPurgeCmd,
		Annotations: newAnnotations(
			modifiesDestinationDirectory,
			modifiesSourceDirectory,
		),
	}

	flags := purgeCmd.Flags()
	flags.BoolVarP(&c.purge.binary, "binary", "P", c.purge.binary, "Purge chezmoi binary")

	return purgeCmd
}

func (c *Config) runPurgeCmd(cmd *cobra.Command, args []string) error {
	return c.doPurge(&purgeOptions{
		binary: c.purge.binary,
	})
}

// doPurge is the core purge functionality. It removes all files and directories
// associated with chezmoi.
func (c *Config) doPurge(purgeOptions *purgeOptions) error {
	if c.persistentState != nil {
		if err := c.persistentState.Close(); err != nil {
			return err
		}
	}

	persistentStateFileAbsPath, err := c.persistentStateFile()
	if err != nil {
		return err
	}
	absPaths := []chezmoi.AbsPath{
		c.CacheDirAbsPath,
		c.configFileAbsPath.Dir(),
		c.configFileAbsPath,
		persistentStateFileAbsPath,
		c.WorkingTreeAbsPath,
		c.SourceDirAbsPath,
	}
	if purgeOptions != nil && purgeOptions.binary {
		executable, err := os.Executable()
		// Special case: do not purge the binary if it is a test binary created
		// by go test as this would break later tests.
		if err == nil && !strings.Contains(executable, "test") {
			absPaths = append(absPaths, chezmoi.NewAbsPath(executable))
		}
	}

	// Remove all paths that exist.
	for _, absPath := range absPaths {
		switch _, err := c.destSystem.Stat(absPath); {
		case errors.Is(err, fs.ErrNotExist):
			continue
		case err != nil:
			return err
		}

		if !c.force {
			switch choice, err := c.promptChoice(fmt.Sprintf("Remove %s", absPath), choicesYesNoAllQuit); {
			case err != nil:
				return err
			case choice == "yes":
			case choice == "no":
				continue
			case choice == "all":
				c.force = true
			case choice == "quit":
				return nil
			}
		}

		switch err := c.destSystem.RemoveAll(absPath); {
		case errors.Is(err, fs.ErrPermission):
			continue
		case err != nil:
			return err
		}
	}

	return nil
}
