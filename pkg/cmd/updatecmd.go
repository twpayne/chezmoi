package cmd

import (
	"errors"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

type updateCmdConfig struct {
	RecurseSubmodules bool
	apply             bool
	filter            *chezmoi.EntryTypeFilter
	init              bool
	recursive         bool
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
	flags.BoolVarP(&c.Update.apply, "apply", "a", c.Update.apply, "Apply after pulling")
	flags.VarP(c.Update.filter.Exclude, "exclude", "x", "Exclude entry types")
	flags.VarP(c.Update.filter.Include, "include", "i", "Include entry types")
	flags.BoolVar(&c.Update.init, "init", c.Update.init, "Recreate config file from template")
	flags.BoolVar(&c.Update.RecurseSubmodules, "recurse-submodules", c.Update.RecurseSubmodules, "Recursively update submodules") //nolint:lll
	flags.BoolVarP(&c.Update.recursive, "recursive", "r", c.Update.recursive, "Recurse into subdirectories")

	registerExcludeIncludeFlagCompletionFuncs(updateCmd)

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
			"--autostash",
			"--rebase",
		}
		if c.Update.RecurseSubmodules {
			args = append(args,
				"--recurse-submodules",
			)
		}
		if err := c.run(c.WorkingTreeAbsPath, c.Git.Command, args); err != nil {
			return err
		}
	}

	if c.Update.apply {
		if err := c.applyArgs(cmd.Context(), c.destSystem, c.DestDirAbsPath, args, applyArgsOptions{
			filter:       c.Update.filter,
			init:         c.Update.init,
			recursive:    c.Update.recursive,
			umask:        c.Umask,
			preApplyFunc: c.defaultPreApplyFunc,
		}); err != nil {
			return err
		}
	}

	return nil
}
