package cmd

import (
	"errors"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type updateCmdConfig struct {
	filter            *chezmoi.EntryTypeFilter
	Command           string   `json:"command"           mapstructure:"command"           yaml:"command"`
	Args              []string `json:"args"              mapstructure:"args"              yaml:"args"`
	RecurseSubmodules bool     `json:"recurseSubmodules" mapstructure:"recurseSubmodules" yaml:"recurseSubmodules"`
	apply             bool
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
		Annotations: newAnnotations(
			modifiesDestinationDirectory,
			persistentStateModeReadWrite,
			requiresSourceDirectory,
			requiresWorkingTree,
			runsCommands,
		),
	}

	flags := updateCmd.Flags()
	flags.BoolVarP(&c.Update.apply, "apply", "a", c.Update.apply, "Apply after pulling")
	flags.VarP(c.Update.filter.Exclude, "exclude", "x", "Exclude entry types")
	flags.VarP(c.Update.filter.Include, "include", "i", "Include entry types")
	flags.BoolVar(&c.Update.init, "init", c.Update.init, "Recreate config file from template")
	flags.BoolVar(
		&c.Update.RecurseSubmodules,
		"recurse-submodules",
		c.Update.RecurseSubmodules,
		"Recursively update submodules",
	)
	flags.BoolVarP(&c.Update.recursive, "recursive", "r", c.Update.recursive, "Recurse into subdirectories")

	registerExcludeIncludeFlagCompletionFuncs(updateCmd)

	return updateCmd
}

func (c *Config) runUpdateCmd(cmd *cobra.Command, args []string) error {
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

	if c.Update.apply {
		if err := c.applyArgs(cmd.Context(), c.destSystem, c.DestDirAbsPath, args, applyArgsOptions{
			cmd:          cmd,
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
