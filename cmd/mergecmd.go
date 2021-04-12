package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type mergeCmdConfig struct {
	Command string   `mapstructure:"command"`
	Args    []string `mapstructure:"args"`
}

func (c *Config) newMergeCmd() *cobra.Command {
	mergeCmd := &cobra.Command{
		Use:     "merge target...",
		Args:    cobra.MinimumNArgs(1),
		Short:   "Perform a three-way merge between the destination state, the source state, and the target state",
		Long:    mustLongHelp("merge"),
		Example: example("merge"),
		RunE:    c.makeRunEWithSourceState(c.runMergeCmd),
		Annotations: map[string]string{
			modifiesSourceDirectory: "true",
			requiresSourceDirectory: "true",
		},
	}

	return mergeCmd
}

func (c *Config) runMergeCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	targetRelPaths, err := c.targetRelPaths(sourceState, args, targetRelPathsOptions{
		mustBeInSourceState: false,
		recursive:           true,
	})
	if err != nil {
		return err
	}

	// Create a temporary directory to store the target state and ensure that it
	// is removed afterwards. We cannot use fs as it lacks TempDir
	// functionality.
	tempDir, err := os.MkdirTemp("", "chezmoi-merge")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)
	tempDirAbsPath := chezmoi.AbsPath(tempDir)

	for _, targetRelPath := range targetRelPaths {
		sourceStateEntry := sourceState.MustEntry(targetRelPath)
		// FIXME sourceStateEntry.TargetStateEntry eagerly evaluates the return
		// targetStateEntry's contents, which means that we cannot fallback to a
		// two-way merge if the source state's contents cannot be decrypted or
		// are an invalid template
		targetStateEntry, err := sourceStateEntry.TargetStateEntry(c.destSystem, c.destDirAbsPath.Join(targetRelPath))
		if err != nil {
			return fmt.Errorf("%s: %w", targetRelPath, err)
		}
		targetStateFile, ok := targetStateEntry.(*chezmoi.TargetStateFile)
		if !ok {
			// LATER consider handling symlinks?
			return fmt.Errorf("%s: not a file", targetRelPath)
		}
		contents, err := targetStateFile.Contents()
		if err != nil {
			return err
		}
		targetStatePath := tempDirAbsPath.Join(chezmoi.RelPath(targetRelPath.Base()))
		if err := c.baseSystem.WriteFile(targetStatePath, contents, 0o600); err != nil {
			return err
		}
		args := append(
			append([]string{}, c.Merge.Args...),
			string(c.destDirAbsPath.Join(targetRelPath)),
			string(c.sourceDirAbsPath.Join(sourceStateEntry.SourceRelPath().RelPath())),
			string(targetStatePath),
		)
		if err := c.persistentState.Close(); err != nil {
			return err
		}
		if err := c.run(c.destDirAbsPath, c.Merge.Command, args); err != nil {
			return fmt.Errorf("%s: %w", targetRelPath, err)
		}
	}

	return nil
}
