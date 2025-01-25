package cmd

import (
	"errors"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type updateCmdConfig struct {
	Command           string   `json:"command"           mapstructure:"command"           yaml:"command"`
	Args              []string `json:"args"              mapstructure:"args"              yaml:"args"`
	Apply             bool     `json:"apply"             mapstructure:"apply"             yaml:"apply"`
	RecurseSubmodules bool     `json:"recurseSubmodules" mapstructure:"recurseSubmodules" yaml:"recurseSubmodules"`
	every             time.Duration
	filter            *chezmoi.EntryTypeFilter
	init              bool
	parentDirs        bool
	recursive         bool
}

var (
	updateStateBucket = []byte("updateState")
	lastUpdateKey     = []byte("lastUpdate")
)

func (c *Config) newUpdateCmd() *cobra.Command {
	updateCmd := &cobra.Command{
		Use:     "update",
		Short:   "Pull and apply any changes",
		Long:    mustLongHelp("update"),
		Example: example("update"),
		Args:    cobra.NoArgs,
		RunE:    c.runUpdateCmd,
		Annotations: newAnnotations(
			modifiesDestinationDirectory,
			persistentStateModeReadWrite,
			requiresSourceDirectory,
			requiresWorkingTree,
			runsCommands,
		),
	}

	updateCmd.Flags().BoolVarP(&c.Update.Apply, "apply", "a", c.Update.Apply, "Apply after pulling")
	updateCmd.Flags().DurationVar(&c.Update.every, "every", c.Update.every, "Rate limit updates")
	updateCmd.Flags().VarP(c.Update.filter.Exclude, "exclude", "x", "Exclude entry types")
	updateCmd.Flags().VarP(c.Update.filter.Include, "include", "i", "Include entry types")
	updateCmd.Flags().BoolVar(&c.Update.init, "init", c.Update.init, "Recreate config file from template")
	updateCmd.Flags().BoolVarP(&c.Update.parentDirs, "parent-dirs", "P", c.Update.parentDirs, "Update all parent directories")
	updateCmd.Flags().
		BoolVar(&c.Update.RecurseSubmodules, "recurse-submodules", c.Update.RecurseSubmodules, "Recursively update submodules")
	updateCmd.Flags().BoolVarP(&c.Update.recursive, "recursive", "r", c.Update.recursive, "Recurse into subdirectories")

	return updateCmd
}

func (c *Config) runUpdateCmd(cmd *cobra.Command, args []string) error {
	// Determine whether to update.
	shouldUpdate, markUpdateDone := c.shouldUpdate()
	if !shouldUpdate {
		return nil
	}

	switch {
	case c.Update.Command != "":
		if err := c.run(c.WorkingTreeAbsPath, c.Update.Command, c.Update.Args); err != nil {
			return err
		}
	case c.UseBuiltinGit.Value(c.useBuiltinGitAutoFunc):
		rawWorkingTreeAbsPath, err := c.baseSystem.RawPath(c.WorkingTreeAbsPath)
		if err != nil {
			return err
		}
		repo, err := git.PlainOpen(rawWorkingTreeAbsPath.String())
		if err != nil {
			return err
		}
		wt, err := repo.Worktree()
		if err != nil {
			return err
		}
		if err := wt.Pull(&git.PullOptions{
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		}); err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
			return err
		}
	default:
		gitArgs := []string{
			"pull",
			"--autostash",
			"--rebase",
		}
		if c.Update.RecurseSubmodules {
			gitArgs = append(gitArgs,
				"--recurse-submodules",
			)
		}
		if err := c.run(c.WorkingTreeAbsPath, c.Git.Command, gitArgs); err != nil {
			return err
		}
	}

	if c.Update.Apply {
		if err := c.applyArgs(cmd.Context(), c.destSystem, c.DestDirAbsPath, args, applyArgsOptions{
			cmd:          cmd,
			filter:       c.Update.filter,
			init:         c.Update.init,
			parentDirs:   c.Update.parentDirs,
			recursive:    c.Update.recursive,
			umask:        c.Umask,
			preApplyFunc: c.defaultPreApplyFunc,
		}); err != nil {
			return err
		}
	}

	if markUpdateDone != nil {
		markUpdateDone()
	}

	return nil
}

// shouldUpdate returns whether to update and a function to be run
// after the update is successful.
func (c *Config) shouldUpdate() (bool, func()) {
	now := time.Now()
	markUpdateDone := func() {
		_ = c.persistentState.Set(updateStateBucket, lastUpdateKey, []byte(now.Format(time.RFC3339)))
	}

	// If the user has not rate-limited updates, then always run the update.
	if c.Update.every == 0 {
		return true, markUpdateDone
	}

	// Otherwise, determine when the last update was run. In case of any error,
	// ignore the error and run the update.
	lastUpdateValue, err := c.persistentState.Get(updateStateBucket, lastUpdateKey)
	if err != nil {
		return true, markUpdateDone
	}
	lastUpdate, err := time.Parse(time.RFC3339, string(lastUpdateValue))
	if err != nil {
		return true, markUpdateDone
	}

	// If the last update was run sufficiently recently then don't update.
	if lastUpdate.Add(c.Update.every).After(now) {
		return false, nil
	}

	// Otherwise, update.
	return true, markUpdateDone
}
