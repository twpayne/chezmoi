package cmd

import (
	"errors"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"

	"chezmoi.io/chezmoi/internal/chezmoi"
)

type updateCmdConfig struct {
	Command           string   `json:"command"           mapstructure:"command"           yaml:"command"`
	Args              []string `json:"args"              mapstructure:"args"              yaml:"args"`
	Apply             bool     `json:"apply"             mapstructure:"apply"             yaml:"apply"`
	RecurseSubmodules bool     `json:"recurseSubmodules" mapstructure:"recurseSubmodules" yaml:"recurseSubmodules"`
	filter            *chezmoi.EntryTypeFilter
	init              bool
	parentDirs        bool
	recursive         bool
}

func (c *Config) newUpdateCmd() *cobra.Command {
	updateCmd := &cobra.Command{
		GroupID:           groupIDDaily,
		Use:               "update",
		Short:             "Pull and apply any changes",
		Long:              mustLongHelp("update"),
		Example:           example("update"),
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE:              c.runUpdateCmd,
		Annotations: newAnnotations(
			modifiesDestinationDirectory,
			persistentStateModeReadWrite,
			requiresSourceDirectory,
			requiresWorkingTree,
			runsCommands,
		),
	}

	updateCmd.Flags().BoolVarP(&c.Update.Apply, "apply", "a", c.Update.Apply, "Apply after pulling")
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

	return nil
}
