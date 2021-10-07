package cmd

import (
	"errors"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type updateCmdConfig struct {
	apply     bool
	exclude   *chezmoi.EntryTypeSet
	include   *chezmoi.EntryTypeSet
	init      bool
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
			requiresWorkingTree:          "true",
			runsCommands:                 "true",
		},
	}

	flags := updateCmd.Flags()
	flags.BoolVarP(&c.update.apply, "apply", "a", c.update.apply, "Apply after pulling")
	flags.VarP(c.update.exclude, "exclude", "x", "Exclude entry types")
	flags.VarP(c.update.include, "include", "i", "Include entry types")
	flags.BoolVar(&c.update.init, "init", c.update.init, "Recreate config file from template")
	flags.BoolVarP(&c.update.recursive, "recursive", "r", c.update.recursive, "Recurse into subdirectories")

	return updateCmd
}

func (c *Config) runUpdateCmd(cmd *cobra.Command, args []string) error {
	if c.UseBuiltinGit.Value(c.useBuiltinGitAutoFunc) {
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
	} else {
		args := []string{
			"pull",
			"--rebase",
			"--recurse-submodules",
		}
		if err := c.run(c.WorkingTreeAbsPath, c.Git.Command, args); err != nil {
			return err
		}
	}

	if c.update.apply {
		if err := c.applyArgs(cmd.Context(), c.destSystem, c.DestDirAbsPath, args, applyArgsOptions{
			include:      c.update.include.Sub(c.update.exclude),
			init:         c.update.init,
			recursive:    c.update.recursive,
			umask:        c.Umask,
			preApplyFunc: c.defaultPreApplyFunc,
		}); err != nil {
			return err
		}
	}

	return nil
}
