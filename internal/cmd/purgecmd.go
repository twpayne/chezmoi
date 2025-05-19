package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/internal/chezmoi"
)

type purgeCmdConfig struct {
	binary bool
}

func (c *Config) newPurgeCmd() *cobra.Command {
	purgeCmd := &cobra.Command{
		Use:               "purge",
		Short:             "Purge chezmoi's configuration and data",
		Long:              mustLongHelp("purge"),
		Example:           example("purge"),
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE:              c.runPurgeCmd,
		Annotations: newAnnotations(
			modifiesDestinationDirectory,
			modifiesSourceDirectory,
			persistentStateModeNone,
		),
	}

	purgeCmd.Flags().BoolVarP(&c.purge.binary, "binary", "P", c.purge.binary, "Purge chezmoi binary")

	return purgeCmd
}

func (c *Config) runPurgeCmd(cmd *cobra.Command, args []string) error {
	return c.doPurge(&doPurgeOptions{
		binary:          c.purge.binary,
		cache:           true,
		config:          true,
		persistentState: true,
		sourceDir:       true,
		workingTree:     true,
	})
}

// doPurge is the core purge functionality. It removes all files and directories
// associated with chezmoi.
func (c *Config) doPurge(options *doPurgeOptions) error {
	// absPaths contains the list of paths to purge, in order. The order is
	// assembled so that parent directories are purged before their children.
	var absPaths []chezmoi.AbsPath

	if options.cache {
		absPaths = append(absPaths, c.CacheDirAbsPath)
	}

	if options.config {
		configFileAbsPath, err := c.getConfigFileAbsPath()
		if err != nil {
			return err
		}
		absPaths = append(absPaths, configFileAbsPath.Dir(), configFileAbsPath)
	}

	if options.persistentState {
		if c.persistentState != nil {
			if err := c.persistentState.Close(); err != nil {
				return err
			}
		}

		persistentStateFileAbsPath, err := c.persistentStateFile()
		if err != nil {
			return err
		}

		absPaths = append(absPaths, persistentStateFileAbsPath)
	}

	if options.workingTree {
		absPaths = append(absPaths, c.WorkingTreeAbsPath)
	}

	if options.sourceDir {
		absPaths = append(absPaths, c.SourceDirAbsPath)
	}

	if options.binary {
		switch executable, err := os.Executable(); {
		case err != nil:
			return err
		case runtime.GOOS == "windows":
			// On Windows the binary of a running process cannot be removed.
			// Warn the user, but otherwise continue.
			c.errorf("cannot purge binary (%s) on Windows\n", executable)
		case strings.Contains(executable, "test"):
			// Special case: do not purge the binary if it is a test binary created
			// by go test as this will break later or concurrent tests.
		default:
			// Otherwise, remove the binary normally.
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
