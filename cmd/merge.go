package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/internal/chezmoi"
)

var mergeCmd = &cobra.Command{
	Use:     "merge targets...",
	Args:    cobra.MinimumNArgs(1),
	Short:   "Perform a three-way merge between the destination state, the source state, and the target state",
	Long:    mustGetLongHelp("merge"),
	Example: getExample("merge"),
	PreRunE: config.ensureNoError,
	RunE:    config.runMergeCmd,
}

type mergeConfig struct {
	Command string
	Args    []string
}

func init() {
	rootCmd.AddCommand(mergeCmd)
}

func (c *Config) runMergeCmd(cmd *cobra.Command, args []string) error {
	ts, err := c.getTargetState(nil)
	if err != nil {
		return err
	}

	entries, err := c.getEntries(ts, args)
	if err != nil {
		return err
	}

	// Create a temporary directory to store the target state and ensure that it
	// is removed afterwards. We cannot use fs as it lacks TempDir
	// functionality.
	tempDir, err := ioutil.TempDir("", "chezmoi")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	for i, entry := range entries {
		if err := c.runMergeCommand(args[i], entry, tempDir); err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) runMergeCommand(arg string, entry chezmoi.Entry, tempDir string) error {
	file, ok := entry.(*chezmoi.File)
	if !ok {
		return fmt.Errorf("%s: not a file", arg)
	}

	// By default, perform a two-way merge between the destination state and the
	// source state.
	args := append(
		append([]string{}, c.Merge.Args...),
		filepath.Join(c.DestDir, file.TargetName()),
		filepath.Join(c.SourceDir, file.SourceName()),
	)

	// Try to evaluate the target state. If this succeeds, perform a three-way
	// merge between the destination state, the source state, and the target
	// state. Target state evaluation might fail if the source state contains
	// template errors or cannot be decrypted.
	if contents, err := file.Contents(); err != nil {
		c.warn(fmt.Sprintf("%s: cannot evaluate target state: %v", arg, err))
	} else {
		targetStatePath := filepath.Join(tempDir, filepath.Base(file.TargetName()))
		if err := ioutil.WriteFile(targetStatePath, contents, 0600); err != nil {
			return err
		}
		args = append(args, targetStatePath)
	}

	if err := c.run("", c.Merge.Command, args...); err != nil {
		return fmt.Errorf("%s: %w", arg, err)
	}

	return nil
}
