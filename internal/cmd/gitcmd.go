package cmd

import (
	"github.com/spf13/cobra"
)

type gitCmdConfig struct {
	Command                   string `json:"command"                   mapstructure:"command"                   yaml:"command"`
	AutoAdd                   bool   `json:"autoadd"                   mapstructure:"autoadd"                   yaml:"autoadd"`
	AutoCommit                bool   `json:"autocommit"                mapstructure:"autocommit"                yaml:"autocommit"`
	AutoPush                  bool   `json:"autopush"                  mapstructure:"autopush"                  yaml:"autopush"`
	CommitMessageTemplate     string `json:"commitMessageTemplate"     mapstructure:"commitMessageTemplate"     yaml:"commitMessageTemplate"`
	CommitMessageTemplateFile string `json:"commitMessageTemplateFile" mapstructure:"commitMessageTemplateFile" yaml:"commitMessageTemplateFile"`
	LFS                       bool   `json:"lfs"                       mapstructure:"lfs"                       yaml:"lfs"`
}

func (c *Config) newGitCmd() *cobra.Command {
	gitCmd := &cobra.Command{
		GroupID: groupIDAdvanced,
		Use:     "git [arg]...",
		Short:   "Run git in the source directory",
		Long:    mustLongHelp("git"),
		Example: example("git"),
		RunE:    c.runGitCmd,
		Annotations: newAnnotations(
			createSourceDirectoryIfNeeded,
			persistentStateModeNone,
			requiresWorkingTree,
			runsCommands,
		),
	}

	return gitCmd
}

func (c *Config) runGitCmd(cmd *cobra.Command, args []string) error {
	return c.run(c.WorkingTreeAbsPath, c.Git.Command, args)
}
