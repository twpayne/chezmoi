package cmd

import (
	"errors"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/chezmoi2/internal/chezmoi"
)

type updateCmdConfig struct {
	apply     bool
	include   *chezmoi.IncludeSet
	recursive bool
}

func (c *Config) newUpdateCmd() *cobra.Command {
	updateCmd := &cobra.Command{
		Use:     "update",
		Short:   "Pull and apply any changes",
		Long:    mustLongHelp("update"),
		Example: example("update"),
		Args:    cobra.NoArgs,
		RunE:    c.runUpdateCmd,
		Annotations: map[string]string{
			modifiesDestinationDirectory: "true",
			persistentStateMode:          persistentStateModeReadWrite,
			requiresSourceDirectory:      "true",
			runsCommands:                 "true",
		},
	}

	flags := updateCmd.Flags()
	flags.BoolVarP(&c.update.apply, "apply", "a", c.update.apply, "apply after pulling")
	flags.VarP(c.update.include, "include", "i", "include entry types")
	flags.BoolVarP(&c.update.recursive, "recursive", "r", c.update.recursive, "recursive")

	return updateCmd
}

func (c *Config) runUpdateCmd(cmd *cobra.Command, args []string) error {
	switch useBuiltinGit, err := c.useBuiltinGit(); {
	case err != nil:
		return err
	case useBuiltinGit:
		rawSourceAbsPath, err := c.baseSystem.RawPath(c.sourceDirAbsPath)
		if err != nil {
			return err
		}
		repo, err := git.PlainOpen(string(rawSourceAbsPath))
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
		args := []string{
			"pull",
			"--rebase",
			"--recurse-submodules",
		}
		if err := c.run(c.sourceDirAbsPath, c.Git.Command, args); err != nil {
			return err
		}
	}

	if !c.update.apply {
		return nil
	}

	return c.applyArgs(c.destSystem, c.destDirAbsPath, args, applyArgsOptions{
		include:      c.update.include,
		recursive:    c.update.recursive,
		umask:        c.Umask,
		preApplyFunc: c.defaultPreApplyFunc,
	})
}
